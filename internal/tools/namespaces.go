package tools

import (
	"context"
	"encoding/json"

	"github.com/k8s-mcp-server/internal/k8s"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RegisterNamespaces(s *server.MCPServer, c *k8s.Client) {
	tool := mcp.NewTool("list_namespaces",
		mcp.WithDescription("List all namespaces in the cluster."),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nss, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		out := make([]map[string]string, 0, len(nss.Items))
		for _, ns := range nss.Items {
			out = append(out, map[string]string{
				"name":   ns.Name,
				"status": string(ns.Status.Phase),
			})
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})
}
