# Tests

This directory contains Python E2E tests for the current Grafana 8.4.7 (`v84`) MCP server profile.

The suite exercises the tools that are still exposed by the server today:

- health and org/admin reads
- dashboards and navigation
- Loki and ClickHouse queries
- rendering
- write-tool gating via `--disable-write`

Tests use [`uv`] for dependency management and `deepeval` for the LLM-driven scenarios.

## Prerequisites

- Docker installed and running
- local Grafana test services started
- API keys for the LLM providers used by the test suite

## Setup

1. Install dependencies:
   ```bash
   uv sync --all-groups
   ```

2. Create a `.env` file with LLM credentials:
   ```env
   OPENAI_API_KEY=sk-...
   ANTHROPIC_API_KEY=sk-ant-...
   ```

3. Start the local test stack:
   ```bash
   make run-test-services
   ```

4. Start the MCP server with optional v84 tools enabled:
   ```bash
   GRAFANA_USERNAME=admin \
   GRAFANA_PASSWORD=admin \
   go run ./cmd/mcp-grafana -t sse --enable-v84-optional-tools
   ```

5. Run the tests:
   ```bash
   uv run pytest
   ```

## Notes

- The old Tempo/proxied E2E tests were removed because proxied datasource tools are forcibly disabled in the current v84 runtime.
- For `stdio` transport, `tests/conftest.py` starts the server with `--enable-v84-optional-tools` so rendering and unified-alerting tests can see those tools.

[`uv`]: https://docs.astral.sh/uv/
