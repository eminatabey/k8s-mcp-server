package main

import (
	"flag"
	"log"
	"os"

	"github.com/k8s-mcp-server/internal/k8s"
	"github.com/k8s-mcp-server/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	addr := flag.String("addr", envOr("MCP_ADDR", ":8080"), "SSE listen address")
	baseURL := flag.String("base-url", os.Getenv("MCP_BASE_URL"), "Public base URL for SSE clients (optional)")
	flag.Parse()

	client, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("k8s client: %v", err)
	}
	log.Printf("k8s client ready (mode=%s)", client.Mode)

	mcpSrv := server.NewMCPServer(
		"k8s-mcp-server",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	tools.RegisterPods(mcpSrv, client)
	tools.RegisterDeployments(mcpSrv, client)
	tools.RegisterNamespaces(mcpSrv, client)

	opts := []server.SSEOption{}
	if *baseURL != "" {
		opts = append(opts, server.WithBaseURL(*baseURL))
	}
	sse := server.NewSSEServer(mcpSrv, opts...)

	log.Printf("MCP SSE server listening on %s", *addr)
	if err := sse.Start(*addr); err != nil {
		log.Fatalf("sse server: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
