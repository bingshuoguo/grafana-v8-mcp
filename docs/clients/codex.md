# Codex CLI

This guide helps you set up the `mcp-grafana` server for the OpenAI Codex CLI.

This fork is primarily intended for **Grafana v8** deployments and is validated against **Grafana 8.4.7**. Other v8 releases with the same API surface may work, but Grafana v9+ is not the primary target of this repository.

## Prerequisites

- Codex CLI installed (`npm install -g @openai/codex`)
- Grafana v8, ideally 8.4.7
- `mcp-grafana` binary in your PATH

## Install the binary

```bash
go install github.com/bingshuoguo/grafana-v8-mcp/cmd/mcp-grafana@latest
```

Or download the archive for your platform from [GitHub Releases](https://github.com/bingshuoguo/grafana-v8-mcp/releases) and put `mcp-grafana` in your `PATH`.

If Codex cannot find `mcp-grafana`, switch `command` to the absolute binary path.

## Important: TOML format

Codex uses **TOML** configuration, not JSON. Configuration file: `~/.codex/config.toml`

## Configuration

### CLI setup (recommended)

Codex works best when it launches the local binary over `stdio`:

```bash
codex mcp add grafana -- mcp-grafana
```

Add environment variables:

```bash
codex mcp add grafana \
  --env GRAFANA_URL=http://localhost:3000 \
  --env GRAFANA_SERVICE_ACCOUNT_TOKEN=<your-token> \
  -- mcp-grafana
```

### Manual configuration

Create or edit `~/.codex/config.toml`:

```toml
[mcp_servers.grafana]
command = "mcp-grafana"
args = []
env = { GRAFANA_URL = "http://localhost:3000", GRAFANA_SERVICE_ACCOUNT_TOKEN = "<your-token>" }
```

**Note:** Use `mcp_servers` (underscore, not hyphen).

## Debug mode

```toml
[mcp_servers.grafana]
command = "mcp-grafana"
args = ["--debug"]
env = { GRAFANA_URL = "http://localhost:3000", GRAFANA_SERVICE_ACCOUNT_TOKEN = "<your-token>" }
```

## Docker setup

```toml
[mcp_servers.grafana]
command = "docker"
args = ["run", "--rm", "-i", "-e", "GRAFANA_URL", "-e", "GRAFANA_SERVICE_ACCOUNT_TOKEN", "bingshuoguo/grafana-v8-mcp:latest"]
env = { GRAFANA_URL = "http://host.docker.internal:3000", GRAFANA_SERVICE_ACCOUNT_TOKEN = "<your-token>" }
```

## Verify configuration

```bash
# List configured servers
codex mcp list

# Show specific server config
codex mcp get grafana

# Start Codex and test
codex
```

Then ask: "List my Grafana dashboards"

## Timeout settings

If Grafana operations take time, increase timeout:

```toml
[mcp_servers.grafana]
command = "mcp-grafana"
args = []
env = { GRAFANA_URL = "http://localhost:3000", GRAFANA_SERVICE_ACCOUNT_TOKEN = "<your-token>" }
startup_timeout_ms = 20000
tool_timeout_ms = 120000
```

## Troubleshooting

**Server not found in Codex:**

- Verify TOML syntax (no trailing commas, use `=` not `:`)
- Check key is `mcp_servers` not `mcp-servers`
- Restart Codex after configuration changes

**Config shared across CLI and IDE:**
Codex CLI and VS Code extension share `~/.codex/config.toml`. A syntax error breaks both.

**Common TOML mistakes:**

```toml
# Wrong - JSON-style
env = {"GRAFANA_URL": "http://localhost:3000"}

# Correct - TOML-style
env = { GRAFANA_URL = "http://localhost:3000" }
```

## Read-only mode

```toml
[mcp_servers.grafana]
command = "mcp-grafana"
args = ["--disable-write"]
env = { GRAFANA_URL = "http://localhost:3000", GRAFANA_SERVICE_ACCOUNT_TOKEN = "<your-token>" }
```

## Tool-level filtering

```toml
[mcp_servers.grafana]
command = "mcp-grafana"
args = [
  "--enable-tools=get_health,create_folder,get_panel_image",
  "--disable-tools=create_folder",
]
env = { GRAFANA_URL = "http://localhost:3000", GRAFANA_SERVICE_ACCOUNT_TOKEN = "<your-token>" }
```

Use exact public tool names only. Both flags also support repeated entries, for example multiple `--enable-tools` items in the `args` array.

**Allowlist overrides disable-write:** When you set `--enable-tools`, only those tools are registered, so you can expose specific write tools even with `--disable-write`:

```toml
args = ["--disable-write", "--enable-tools=get_health,search_dashboards,upsert_dashboard"]
```
