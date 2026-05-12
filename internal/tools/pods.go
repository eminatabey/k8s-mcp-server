package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/k8s-mcp-server/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RegisterPods(s *server.MCPServer, c *k8s.Client) {
	listPods := mcp.NewTool("list_pods",
		mcp.WithDescription("List pods in a namespace (defaults to 'default')."),
		mcp.WithString("namespace", mcp.Description("Kubernetes namespace")),
	)
	s.AddTool(listPods, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns := stringArg(req, "namespace", "default")
		pods, err := c.Clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out := make([]map[string]any, 0, len(pods.Items))
		for _, p := range pods.Items {
			out = append(out, map[string]any{
				"name":      p.Name,
				"namespace": p.Namespace,
				"phase":     string(p.Status.Phase),
				"node":      p.Spec.NodeName,
				"ready":     readyContainers(p),
			})
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	getLogs := mcp.NewTool("get_logs",
		mcp.WithDescription("Fetch logs from a pod container."),
		mcp.WithString("namespace", mcp.Required(), mcp.Description("Pod namespace")),
		mcp.WithString("pod", mcp.Required(), mcp.Description("Pod name")),
		mcp.WithString("container", mcp.Description("Container name (optional)")),
		mcp.WithNumber("tail_lines", mcp.Description("Last N lines (default 200)")),
	)
	s.AddTool(getLogs, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ns := stringArg(req, "namespace", "")
		name := stringArg(req, "pod", "")
		if ns == "" || name == "" {
			return mcp.NewToolResultError("namespace and pod are required"), nil
		}
		container := stringArg(req, "container", "")
		tail := int64(intArg(req, "tail_lines", 200))

		opts := &corev1.PodLogOptions{TailLines: &tail}
		if container != "" {
			opts.Container = container
		}
		stream, err := c.Clientset.CoreV1().Pods(ns).GetLogs(name, opts).Stream(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer stream.Close()
		buf, err := io.ReadAll(stream)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("read logs: %v", err)), nil
		}
		return mcp.NewToolResultText(string(buf)), nil
	})
}

func readyContainers(p corev1.Pod) string {
	ready := 0
	for _, cs := range p.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, len(p.Status.ContainerStatuses))
}
