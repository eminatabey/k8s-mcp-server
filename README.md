# k8s-mcp-server

MCP server (SSE transport) exposing read/operate tools against a Kubernetes cluster, written in Go with [`mcp-go`](https://github.com/mark3labs/mcp-go) and `client-go`.

## Tools

| Tool | Description |
|------|-------------|
| `list_pods` | List pods in a namespace (default `default`). |
| `get_logs` | Fetch container logs (`tail_lines`, `container` optional). |
| `describe_deployment` | Replicas, conditions, container images. |
| `scale_deployment` | Scale to N replicas (uses `/scale` subresource). |
| `rollout_restart` | `kubectl rollout restart` equivalent (annotation patch). |
| `list_namespaces` | List all namespaces. |

## Dual-mode auth

The client tries `rest.InClusterConfig()` first; if it fails, falls back to `KUBECONFIG` or `~/.kube/config`.

- **In-cluster**: deploy with `deploy/rbac.yaml` + `deploy/deployment.yaml` (ServiceAccount + minimal ClusterRole).

## Build

```bash
go build ./cmd/server
docker build -t k8s-mcp-server:dev .
```

## Run (local)

```bash
./server --addr :8080
# or
docker run --rm -p 8080:8080 -v $HOME/.kube:/home/nonroot/.kube:ro k8s-mcp-server:dev
```

SSE endpoint: `http://localhost:8080/sse`

## Connect to an AI agent

The server speaks MCP over SSE, so any MCP-capable client can use it. Point the client at the `/sse` endpoint of a running instance (local or in-cluster, e.g. via `kubectl port-forward`).

### Claude Code

```bash
claude mcp add --transport sse k8s http://localhost:8080/sse
```

Then list/verify:

```bash
claude mcp list
```

### Claude Desktop

Edit `claude_desktop_config.json` (macOS: `~/Library/Application Support/Claude/`, Windows: `%APPDATA%\Claude\`). Native SSE support is recent; for older builds bridge through [`mcp-remote`](https://www.npmjs.com/package/mcp-remote):

```json
{
  "mcpServers": {
    "k8s": {
      "command": "npx",
      "args": ["-y", "mcp-remote", "http://localhost:8080/sse"]
    }
  }
}
```

### Cursor / Windsurf / other SSE clients

Add an entry to the client's MCP config (typically `~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "k8s": {
      "url": "http://localhost:8080/sse"
    }
  }
}
```

### In-cluster access

When the server runs in the cluster, expose it to your local agent with:

```bash
kubectl port-forward svc/k8s-mcp-server 8080:8080
```

then use `http://localhost:8080/sse` as above. For remote use, front it with an Ingress (TLS recommended) and set `--base-url https://<host>` so SSE clients receive the correct public URL.

## Deploy (in-cluster)

```bash
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/deployment.yaml
```

## CI

`.github/workflows/docker.yml` builds multi-arch (amd64/arm64) and pushes to `ghcr.io/eminatabey/k8s-mcp-server` on push to `main` and on `v*` tags.

## RBAC scope (minimal)

| Group | Resource | Verbs |
|-------|----------|-------|
| `""` | `namespaces`, `pods` | get, list, watch |
| `""` | `pods/log` | get |
| `apps` | `deployments` | get, list, watch, patch |
| `apps` | `deployments/scale` | get, update, patch |
