# Claude Code

This guide helps you set up the `mcp-grafana` server for the Claude Code CLI.

This fork is primarily intended for **Grafana v8** deployments and is validated against **Grafana 8.4.7**. Other v8 releases with the same API surface may work, but Grafana v9+ is not the primary target of this repository.

## Prerequisites

- Claude Code CLI installed (`npm install -g @anthropic-ai/claude-code`)
- Grafana v8, ideally 8.4.7
- `mcp-grafana` binary in your PATH

## Install the binary

Use one of these methods before configuring Claude Code:

```bash
go install github.com/bingshuoguo/grafana-v8-mcp/cmd/mcp-grafana@latest
```

Or download the archive for your platform from [GitHub Releases](https://github.com/bingshuoguo/grafana-v8-mcp/releases) and put `mcp-grafana` in your `PATH`.

If Claude Code cannot resolve `mcp-grafana`, replace it with the absolute binary path in the examples below.

## One-command setup

Claude Code uses local `stdio` servers, so the recommended setup is to launch the binary directly:

```bash
claude mcp add-json "grafana" '{"command":"mcp-grafana","args":[],"env":{"GRAFANA_URL":"http://localhost:3000","GRAFANA_SERVICE_ACCOUNT_TOKEN":"<your-token>"}}'
```

## Manual configuration

Claude Code stores MCP configuration alongside other settings. Use the CLI to manage servers.

```bash
# List configured servers
claude mcp list

# Add a server
claude mcp add grafana -- mcp-grafana

# Remove a server
claude mcp remove grafana
```

## Scope options

Claude Code supports three scopes for MCP servers:

| Scope             | Description                              |
| :---------------- | :--------------------------------------- |
| `local` (default) | Available only to you in current project |
| `project`         | Shared with team via `.mcp.json` file    |
| `user`            | Available to you across all projects     |

```bash
# Add for all your projects
claude mcp add grafana --scope user -- mcp-grafana

# Add for current project only (default)
claude mcp add grafana --scope local -- mcp-grafana
```

## Full configuration with environment variables

```bash
claude mcp add-json "grafana" '{
  "command": "mcp-grafana",
  "args": [],
  "env": {
    "GRAFANA_URL": "http://localhost:3000",
    "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<your-token>"
  }
}'
```

## Docker setup

```bash
claude mcp add-json "grafana" '{
  "command": "docker",
  "args": ["run", "--rm", "-i", "-e", "GRAFANA_URL", "-e", "GRAFANA_SERVICE_ACCOUNT_TOKEN", "bingshuoguo/grafana-v8-mcp:latest"],
  "env": {
    "GRAFANA_URL": "http://host.docker.internal:3000",
    "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<your-token>"
  }
}'
```

## Debug mode

```bash
claude mcp add-json "grafana" '{
  "command": "mcp-grafana",
  "args": ["--debug"],
  "env": {
    "GRAFANA_URL": "http://localhost:3000",
    "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<your-token>"
  }
}'
```

Then run Claude Code with debug output:

```bash
claude --debug
```

## Verify configuration

1.  Start a new Claude Code session:

    ```bash
    claude
    ```

2.  Ask: "List my Grafana dashboards"
3.  Claude should use the Grafana MCP tools automatically

## View current configuration

```bash
claude mcp list --json
```

## Troubleshooting

**Server not found:**

- Verify binary path: `which mcp-grafana`
- Use full path in configuration if needed

**Permission errors:**

- Check Grafana service account token
- Verify token has required RBAC permissions

## Read-only mode

```bash
claude mcp add-json "grafana" '{
  "command": "mcp-grafana",
  "args": ["--disable-write"],
  "env": {
    "GRAFANA_URL": "http://localhost:3000",
    "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<your-token>"
  }
}'
```

## Tool-level filtering

```bash
claude mcp add-json "grafana" '{
  "command": "mcp-grafana",
  "args": [
    "--enable-tools=get_health,create_folder,get_panel_image",
    "--disable-tools=create_folder"
  ],
  "env": {
    "GRAFANA_URL": "http://localhost:3000",
    "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<your-token>"
  }
}'
```

Use exact public tool names only. Both flags also support repeated usage such as `--enable-tools get_health --enable-tools search_dashboards`.

**Allowlist overrides disable-write:** When you set `--enable-tools`, only those tools are registered, so you can expose specific write tools even with `--disable-write`, e.g. `"--disable-write"`, `"--enable-tools=get_health,search_dashboards,upsert_dashboard"` in `args`.
