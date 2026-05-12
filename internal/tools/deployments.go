package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/k8s-mcp-server/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func RegisterDeployments(s *server.MCPServer, c *k8s.Client) {
	describe := mcp.NewTool("describe_deployment",
		mcp.WithDescription("Describe a deployment (replicas, conditions, image)."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Deployment namespace")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Deployment name")),
	)
	s.AddTool(describe, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns := stringArg(req, "namespace", "")
		name := stringArg(req, "name", "")
		if ns == "" || name == "" {
			return mcp.NewToolResultError("namespace and name are required"), nil
		}
		d, err := c.Clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		images := []string{}
		for _, cn := range d.Spec.Template.Spec.Containers {
			images = append(images, fmt.Sprintf("%s=%s", cn.Name, cn.Image))
		}
		conditions := []map[string]string{}
		for _, cd := range d.Status.Conditions {
			conditions = append(conditions, map[string]string{
				"type":    string(cd.Type),
				"status":  string(cd.Status),
				"reason":  cd.Reason,
				"message": cd.Message,
			})
		}
		out := map[string]any{
			"name":              d.Name,
			"namespace":         d.Namespace,
			"replicas":          d.Status.Replicas,
			"readyReplicas":     d.Status.ReadyReplicas,
			"updatedReplicas":   d.Status.UpdatedReplicas,
			"availableReplicas": d.Status.AvailableReplicas,
			"strategy":          string(d.Spec.Strategy.Type),
			"images":            images,
			"conditions":        conditions,
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	scale := mcp.NewTool("scale_deployment",
		mcp.WithDescription("Scale a deployment to N replicas."),
		mcp.WithString("namespace", mcp.Required()),
		mcp.WithString("name", mcp.Required()),
		mcp.WithNumber("replicas", mcp.Required(), mcp.Description("Target replica count (>=0)")),
	)
	s.AddTool(scale, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns := stringArg(req, "namespace", "")
		name := stringArg(req, "name", "")
		replicas := int32(intArg(req, "replicas", -1))
		if ns == "" || name == "" || replicas < 0 {
			return mcp.NewToolResultError("namespace, name and replicas (>=0) are required"), nil
		}
		sc, err := c.Clientset.AppsV1().Deployments(ns).GetScale(ctx, name, metav1.GetOptions{})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		sc.Spec.Replicas = replicas
		updated, err := c.Clientset.AppsV1().Deployments(ns).UpdateScale(ctx, name, sc, metav1.UpdateOptions{})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("scaled %s/%s to %d replicas", ns, name, updated.Spec.Replicas)), nil
	})

	restart := mcp.NewTool("rollout_restart",
		mcp.WithDescription("Trigger a rollout restart (kubectl rollout restart equivalent)."),
		mcp.WithString("namespace", mcp.Required()),
		mcp.WithString("name", mcp.Required()),
	)
	s.AddTool(restart, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns := stringArg(req, "namespace", "")
		name := stringArg(req, "name", "")
		if ns == "" || name == "" {
			return mcp.NewToolResultError("namespace and name are required"), nil
		}
		patch := fmt.Sprintf(
			`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`,
			time.Now().UTC().Format(time.RFC3339),
		)
		_, err := c.Clientset.AppsV1().Deployments(ns).Patch(
			ctx, name, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{},
		)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("rollout restart triggered for %s/%s", ns, name)), nil
	})
}
