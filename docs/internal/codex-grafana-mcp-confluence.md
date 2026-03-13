# 公司内部 AI 客户端接入 Grafana MCP 说明

## 1. 文档目的

本文档说明如何在公司内部使用 **Cursor**、**Claude Code**、**Codex CLI**、**Gemini CLI** 接入 Grafana MCP，以便通过自然语言查询 Shopee Monitoring Grafana 中的仪表板、指标、日志与告警信息。

当前推荐接入方式为：

- 本地启动 `mcp-grafana` 二进制
- 通过 `stdio` 方式由各客户端拉起 MCP 进程
- 默认开启只读模式（`--disable-write`），避免误操作

## 2. 适用范围

适用于以下场景：

- 在 Cursor / Claude Code / Codex / Gemini 中查询 Grafana 仪表板、Panel、Datasource、Folders
- 查询 Prometheus 指标
- 查询 Loki 日志
- 查询 ClickHouse 日志或表结构
- 生成 Grafana deeplink，辅助问题排查

不建议用于以下场景：

- 直接在生产环境执行写操作
- 将 Grafana token 明文写入代码仓库、脚本或共享文档

## 3. 接入效果

接入完成后，对应 AI 客户端可以直接调用 Grafana MCP 工具，例如：

- 查询某个 dashboard
- 查看某个指标最近 1 小时趋势
- 按关键字搜索 Loki 或 ClickHouse 日志
- 生成某个 dashboard 或 panel 的 Grafana 链接

当前 Grafana MCP 已支持的主要能力包括：

- 健康检查、当前用户、当前组织信息
- dashboard 搜索与读取
- datasource 查询与解析
- Prometheus / Loki / ClickHouse 查询
- 日志搜索
- annotations、legacy alerting、org users、teams 等辅助能力

## 4. 推荐接入方案

### 4.1 推荐模式

公司内部推荐使用 **本地二进制 + stdio** 方式接入，所有客户端均适用。

原因如下：

- 配置简单
- 与各客户端的 MCP 模式兼容
- 不需要额外暴露 HTTP 服务
- 结合 `--disable-write` 可以默认只读，更安全

### 4.2 当前内部使用的 Grafana 地址

当前公司内部 Grafana MCP 应指向的地址为：

```text
https://monitoring.infra.sz.shopee.io/grafana
```

若后续监控域名、租户或访问方式有调整，请以 AIOps 实际发放的信息为准。

## 5. 前置条件

接入前请确认：

- 已安装目标 AI 客户端（Cursor / Claude Code / Codex / Gemini CLI 之一）
- 已具备本仓库代码或可执行二进制 `mcp-grafana`
- 本机可以访问内部 Grafana 地址
- 已申请 Grafana 只读 Service Account Token

建议：

- 优先申请只读 token
- 不要使用个人账号密码直接接入
- 不要在共享配置、截图或聊天记录中传播 token

## 6. 安装 mcp-grafana 二进制

使用 go install 安装，所有客户端共用同一二进制：

```bash
go install github.com/bingshuoguo/grafana-v8-mcp/cmd/mcp-grafana@latest
```

配置中 `command` 填 `mcp-grafana` 即可（依赖安装后的 `PATH`）；若某客户端无法解析到该命令，再改为本机 `mcp-grafana` 的**绝对路径**。

## 7. 各客户端配置

下表为各客户端的配置文件位置与格式，公司内部统一使用**内部 Grafana 地址**与 **`--disable-write`**。

| 客户端        | 配置文件路径           | 配置格式 |
|---------------|------------------------|----------|
| Cursor        | `~/.cursor/mcp.json` 或 `.cursor/mcp.json` | JSON     |
| Claude Code   | 通过 `claude mcp add-json` 写入其配置      | JSON     |
| Codex CLI     | `~/.codex/config.toml`                     | TOML     |
| Gemini CLI    | `~/.gemini/settings.json`                | JSON     |

以下各小节给出**公司内部推荐配置示例**（Grafana 地址与只读已填好，仅需替换 token；`command` 使用 `mcp-grafana`，若客户端找不到再改绝对路径）。

### 7.1 Cursor

1. 打开 Cursor：Settings → Tools & Integrations → New MCP Server，或直接编辑配置文件。
2. 全局配置：`~/.cursor/mcp.json`；项目级配置：项目根目录 `.cursor/mcp.json`。

```json
{
  "mcpServers": {
    "grafana": {
      "command": "mcp-grafana",
      "args": ["--disable-write"],
      "env": {
        "GRAFANA_URL": "https://monitoring.infra.sz.shopee.io/grafana",
        "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<REDACTED_TOKEN>"
      }
    }
  }
}
```

### 7.2 Claude Code

在终端执行（将 token 替换为实际值）：

```bash
claude mcp add-json "grafana" '{
  "command": "mcp-grafana",
  "args": ["--disable-write"],
  "env": {
    "GRAFANA_URL": "https://monitoring.infra.sz.shopee.io/grafana",
    "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<REDACTED_TOKEN>"
  }
}'
```

或手动编辑 Claude Code 使用的配置文件（具体路径以 Claude Code 文档为准），内容结构与上述 JSON 一致。

### 7.3 Codex CLI

**注意：Codex 使用 TOML，不是 JSON。** 编辑 `~/.codex/config.toml`：

```toml
[mcp_servers.grafana]
command = "mcp-grafana"
args = ["-t", "stdio", "--disable-write"]

[mcp_servers.grafana.env]
GRAFANA_URL = "https://monitoring.infra.sz.shopee.io/grafana"
GRAFANA_SERVICE_ACCOUNT_TOKEN = "<REDACTED_TOKEN>"
```

说明：`-t stdio` 表示由 Codex 直接拉起本地 MCP 进程。若 Codex 找不到 `mcp-grafana`，将 `command` 改为本机绝对路径。

可选：若启动或查询较慢，可增加超时（仅 Codex 支持时）：

```toml
startup_timeout_sec = 60.0
tool_timeout_ms = 120000
```

### 7.4 Gemini CLI

编辑 `~/.gemini/settings.json`：

```json
{
  "mcpServers": {
    "grafana": {
      "command": "mcp-grafana",
      "args": ["--disable-write"],
      "env": {
        "GRAFANA_URL": "https://monitoring.infra.sz.shopee.io/grafana",
        "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<REDACTED_TOKEN>"
      }
    }
  }
}
```

说明：若 Gemini CLI 找不到 `mcp-grafana`，将 `command` 改为本机绝对路径。

## 8. 验证方式

### 8.1 按客户端检查 MCP 是否生效

- **Cursor**：Settings → Tools & Integrations，查看是否有名为 `grafana` 的 MCP server，状态为可用；在 Composer 中提问如「列出当前 Grafana 健康状态」。
- **Claude Code**：在终端执行 `claude mcp list`，应能看到 `grafana`；在对话中请求调用 Grafana 工具。
- **Codex**：在终端执行 `codex mcp list`，应能看到 `grafana`。
- **Gemini CLI**：启动 `gemini` 后执行 `/mcp` 查看是否有 Grafana 相关工具，并尝试提问。

### 8.2 通用验证提示词

可直接尝试以下提示词验证工具是否被调用：

- 「请列出当前 Grafana 的健康状态」
- 「请搜索标题里包含 error rate 的 dashboard」
- 「请查看最近 1 小时某个 Prometheus 指标趋势」

若工具调用成功，说明 Grafana MCP 已接入完成。

## 9. 推荐使用方式

### 9.1 建议优先使用只读能力

推荐主要使用以下只读能力：

- `get_health`
- `search_dashboards`
- `get_dashboard_by_uid`
- `list_datasources`
- `get_datasource`
- `query_prometheus`
- `query_loki_logs`
- `query_clickhouse`
- `search_logs`
- `generate_deeplink`

### 9.2 建议的提问方式

建议在提示词中明确：

- 时间范围
- 服务名 / dashboard 名 / datasource
- 查询目标是指标、日志还是 dashboard

示例：

- 「查询过去 1 小时 payment 服务的错误率趋势」
- 「搜索 Loki 中最近 30 分钟包含 timeout 的日志」
- 「帮我生成某个 dashboard 的 Grafana deeplink」

## 10. 安全要求

请遵循以下规范：

- 不要把 `GRAFANA_SERVICE_ACCOUNT_TOKEN` 提交到 Git
- 不要把 token 写入 README、Confluence、脚本或聊天群
- 推荐默认保留 `--disable-write`
- 如确需开启写工具，必须说明用途并限制使用范围
- 离职、转岗或权限变更时，及时回收或轮换 token

## 11. 常见问题

### 11.1 客户端无法识别 `grafana` MCP

- **Cursor**：检查 `~/.cursor/mcp.json` 或 `.cursor/mcp.json` 的 JSON 语法（如是否有非法尾逗号），重启 Cursor；若用绝对路径，确认该路径存在。
- **Claude Code**：确认已执行 `claude mcp add-json` 或手动配置正确；修改后必要时重启 Claude Code。
- **Codex**：确认 `~/.codex/config.toml` 为合法 TOML；节名为 `[mcp_servers.grafana]`；`command` 路径存在；修改后重新启动 Codex。
- **Gemini CLI**：确认 `~/.gemini/settings.json` 语法正确，且 `mcpServers.grafana` 存在；修改后重新启动 Gemini。

### 11.2 启动时报找不到 `mcp-grafana`

- 是否已执行 `go install github.com/bingshuoguo/grafana-v8-mcp/cmd/mcp-grafana@latest`
- 确认 `go env GOPATH/bin` 或安装所在目录在系统 `PATH` 中
- 若客户端仍找不到，将配置中 `command` 改为 `mcp-grafana` 的绝对路径（如 `$(go env GOPATH)/bin/mcp-grafana`）

### 11.3 查询报鉴权失败

- token 是否有效、是否过期或已被回收
- token 是否具备目标 Grafana 的访问权限
- `GRAFANA_URL` 是否指向正确实例（公司内部为 `https://monitoring.infra.sz.shopee.io/grafana`）

### 11.4 查询不到数据

- 时间范围是否过窄
- datasource 是否正确
- 指标名 / 日志关键字 / dashboard UID 是否正确
- 当前账号或 token 是否有权限访问对应数据源

### 11.5 日志搜索无结果

对于 ClickHouse 日志，请优先确认：

- 实际表名是否为 `otel_logs`
- 列名是否与默认假设一致

建议优先调用：

- `list_clickhouse_tables`
- `describe_clickhouse_table`

## 12. 变更与维护建议

建议由 AIOps 或 MCP 维护人统一维护以下内容：

- Grafana MCP 二进制版本
- 默认 tool 开关策略
- 内部 Grafana 地址
- token 申请方式与最小权限规范

如需从只读升级为可写模式，请在团队内部评估以下风险：

- dashboard / folder / annotation 被误改
- 生产环境操作不可回溯
- LLM 误调用写工具带来的变更风险

## 13. 附录：各客户端最小可用配置

以下为仅替换 `<REDACTED_TOKEN>` 即可使用的最小配置；`command` 使用 `mcp-grafana`，若客户端找不到再改为本机绝对路径。

### Cursor（`~/.cursor/mcp.json`）

```json
{
  "mcpServers": {
    "grafana": {
      "command": "mcp-grafana",
      "args": ["--disable-write"],
      "env": {
        "GRAFANA_URL": "https://monitoring.infra.sz.shopee.io/grafana",
        "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<REDACTED_TOKEN>"
      }
    }
  }
}
```

### Claude Code（通过 add-json 或等效配置）

```json
{
  "command": "mcp-grafana",
  "args": ["--disable-write"],
  "env": {
    "GRAFANA_URL": "https://monitoring.infra.sz.shopee.io/grafana",
    "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<REDACTED_TOKEN>"
  }
}
```

### Codex（`~/.codex/config.toml`）

```toml
[mcp_servers.grafana]
command = "mcp-grafana"
args = ["-t", "stdio", "--disable-write"]

[mcp_servers.grafana.env]
GRAFANA_URL = "https://monitoring.infra.sz.shopee.io/grafana"
GRAFANA_SERVICE_ACCOUNT_TOKEN = "<REDACTED_TOKEN>"
```

### Gemini CLI（`~/.gemini/settings.json`）

```json
{
  "mcpServers": {
    "grafana": {
      "command": "mcp-grafana",
      "args": ["--disable-write"],
      "env": {
        "GRAFANA_URL": "https://monitoring.infra.sz.shopee.io/grafana",
        "GRAFANA_SERVICE_ACCOUNT_TOKEN": "<REDACTED_TOKEN>"
      }
    }
  }
}
```

以上配置可作为团队内部推广时的默认模板，按所用客户端择一使用即可。
