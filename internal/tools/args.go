package tools

import "github.com/mark3labs/mcp-go/mcp"

func stringArg(req mcp.CallToolRequest, key, def string) string {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return def
	}
	v, ok := args[key]
	if !ok || v == nil {
		return def
	}
	s, ok := v.(string)
	if !ok {
		return def
	}
	return s
}

func intArg(req mcp.CallToolRequest, key string, def int) int {
	args, ok := req.Params.Arguments.(map[string]any)
	if !ok {
		return def
	}
	v, ok := args[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return def
	}
}
