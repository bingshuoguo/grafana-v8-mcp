# v84 缺失工具补齐设计技术方案

更新时间：2026-02-27
关联文档：
- [v84-missing-tools-api-feasibility.md](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/docs/v84-missing-tools-api-feasibility.md)
- [grafana-8.4.7-mcp-tool-spec.md](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/docs/grafana-8.4.7-mcp-tool-spec.md)

## 1. 目标与范围

本方案目标是基于 v84 profile，补齐缺失工具中“可落地”的部分，并保持 8.4.7 兼容性与默认稳定性。

本期范围：
- P0：13 个低风险工具（兼容别名 + 直接可实现工具）
- P1：12 个 UID -> ID 改造工具（Prometheus/Loki/ClickHouse/Search Logs）
- P2：7 个条件工具（Unified Alerting + Rendering）

暂不纳入：
- P3 插件/云能力依赖工具（OnCall/Incident/Sift/Asserts/Pyroscope/RBAC）

## 2. 设计原则

1. `ID-first`：数据源访问统一 `id > uid > name`，禁止新增 UID-only 路径依赖。
2. `v84 默认保守`：默认只注册稳定可用工具；条件工具走 capability gating。
3. `最小侵入`：优先复用 `tools/v84/common.go`、`resolveDatasourceRef`、`doAPIRequest`。
4. `错误统一`：统一返回 `APIError`，保留 `HardError` 语义。
5. `可回归`：新增工具必须有 unit 测试覆盖主路径与错误路径。

## 3. 当前基线

`tools/v84/registry.go` 当前注册 23 个工具，`enableOptionalTools` 尚未生效。  
v84 已具备关键底座：
- OpenAPI client 封装：`getGrafanaClient`
- raw HTTP 访问：`doAPIRequest`
- datasource 解析：`resolveDatasourceRef`
- Prometheus 三件套：`query_prometheus`、`list_prometheus_label_values`、`list_prometheus_metric_names`

这意味着本次新增主要是“补工具 + 组装能力 + 兼容改造”，不是重构核心框架。

## 4. 总体架构设计

### 4.1 分层注册模型

将 `AddV84Tools` 拆分为 4 层注册函数：
- `registerV84CoreReadTools`
- `registerV84CoreWriteTools`
- `registerV84CompatAliasTools`
- `registerV84OptionalTools`

`enableWriteTools` 控制写工具，`enableOptionalTools` 控制条件工具与插件工具。

### 4.2 能力探测层（新增）

新增 `tools/v84/capability.go`，启动时探测并缓存能力：
- `HasUnifiedAlerting`
- `HasImageRenderer`
- `HasDatasourceProxyByID`
- `HasPlugin(name)`（为 P3 预留）

探测失败策略：
- 默认 fail-open for core（不影响基础工具启动）
- optional 工具按能力缺失直接不注册

### 4.3 Datasource Proxy 统一访问层（新增）

新增 `tools/v84/datasource_proxy.go`：
- `doDatasourceProxyRequest(ctx, dsRef, method, subpath, query, body)`
- 统一：解析 datasource -> 获取 id -> 请求 `/api/datasources/proxy/{id}/{subpath}`

P1 所有工具通过该层访问，避免重复实现 ID 解析和错误处理。

## 5. 工具实现设计

## 5.1 P0（13 个）设计

### A. 兼容别名工具（4 个）

新增 `tools/v84/compat_alias.go`：
- `update_dashboard` -> 直接调用 `upsertDashboard`
- `list_users_by_org` -> 直接调用 `listOrgUsers`
- `get_datasource_by_uid` -> 参数转 `GetDatasourceRequest{UID: ...}`
- `get_datasource_by_name` -> 参数转 `GetDatasourceRequest{Name: ...}`

约束：
- 仅作为兼容入口，不复制业务逻辑
- 响应结构遵循旧名称语义，但数据字段直接复用 v84 model

### B. 直接可实现工具（9 个）

新增文件建议：
- `tools/v84/search_folders.go`
- `tools/v84/dashboard_helpers.go`
- `tools/v84/annotations_extra.go`
- `tools/v84/navigation.go`
- `tools/v84/examples.go`

关键实现点：
- `search_folders`：`/api/search` with `type=dash-folder`
- `get_dashboard_panel_queries`：解析 dashboard JSON 中 panel.targets 与 datasource
- `get_dashboard_property`：支持 JSONPath（建议复用现有 `dashboard.go` 中思路）
- `get_dashboard_summary`：汇总 title/panel_count/panel_types/variables/tags/meta
- `create_graphite_annotation`：调用注解 API 的 graphite 兼容参数
- `update_annotation`：PUT `/api/annotations/{id}`
- `get_annotation_tags`：GET `/api/annotations/tags`（带 `tag` 与 `limit`）
- `generate_deeplink`：纯本地 URL 拼装
- `get_query_examples`：静态数据返回

## 5.2 P1（12 个）设计

新增文件建议：
- `tools/v84/prometheus_extra.go`
- `tools/v84/loki.go`
- `tools/v84/clickhouse.go`
- `tools/v84/search_logs.go`

统一策略：
- 入参 datasource 一律 `DatasourceRef`
- 走 `resolveDatasourceRef` 与 ID proxy
- 所有上游错误经 `wrapRawAPIError` 归一

Prometheus 增补（3）：
- `list_prometheus_label_names` -> `/api/v1/labels`
- `list_prometheus_metric_metadata` -> `/api/v1/metadata`
- `query_prometheus_histogram` -> 复用 `query_prometheus` 结果做 histogram 分桶摘要

Loki（5）：
- `list_loki_label_names` -> `/loki/api/v1/labels`
- `list_loki_label_values` -> `/loki/api/v1/label/{name}/values`
- `query_loki_logs` -> `/loki/api/v1/query_range` or `/query`
- `query_loki_stats` -> `/loki/api/v1/index/stats`
- `query_loki_patterns` -> `/loki/api/v1/patterns`

ClickHouse（3）：
- `query_clickhouse` -> `/api/ds/query`（注入 `datasourceId`）
- `list_clickhouse_tables`、`describe_clickhouse_table` 通过 SQL 模板生成并走 `query_clickhouse`

Search Logs（1）：
- 作为 routing 工具，根据 datasource type 分派至 Loki/ClickHouse 子实现

## 5.3 P2（7 个）设计

新增文件建议：
- `tools/v84/alerting_unified.go`
- `tools/v84/rendering.go`

Unified Alerting（6）：
- 仅当 `HasUnifiedAlerting=true` 时注册
- 实现 `list/get/create/update/delete alert rule` + `list_contact_points`
- 支持 `datasourceUid` 输入时先转 datasource ID，再走兼容路径或 Grafana-managed 路径

**重要：Grafana 8.4.7 不可用 Provisioning API，必须改用 Ruler API**

主版本 `tools/alerting.go` 调用的 `c.Provisioning.GetAlertRules()` 对应
`GET /api/v1/provisioning/alert-rules`，该接口是 **Grafana 9.1 引入的**，在 8.4.7 不存在（返回 404）。
`alerting_unified.go` 必须直接调用以下 8.4.7 原生接口，不得复用主版本 provisioning 路径：

| 操作 | API 路径 |
|------|---------|
| 列出所有规则组 | `GET /api/ruler/grafana/api/v1/rules` |
| 获取单个规则组 | `GET /api/ruler/grafana/api/v1/rules/{namespace}/{group}` |
| 创建或更新规则组 | `POST /api/ruler/grafana/api/v1/rules/{namespace}` |
| 删除规则组 | `DELETE /api/ruler/grafana/api/v1/rules/{namespace}/{group}` |
| 列出联系点 | `GET /api/alertmanager/grafana/api/v2/receivers` |

Ruler API 的响应结构为 `map[namespace][]RuleGroup`（按 namespace/group 嵌套），与
Provisioning API 的扁平列表完全不同，需单独实现解析逻辑。`get_alert_rule_by_uid` 需遍历
所有规则组后按 UID 过滤，不存在直接按 UID 查询的端点。

Loki patterns 额外说明：
- `/loki/api/v1/patterns` 是 **Loki 2.8+** 引入的功能，与 Grafana 版本无关
- 若 Loki 版本低于 2.8，该端点返回 404；capability 探测时应单独处理

Rendering（1）：
- `get_panel_image`
- 仅当 `HasImageRenderer=true` 注册
- 提供明确错误：renderer 未安装、超时、dashboard/panel 不存在

## 6. 目录与文件变更清单（建议）

新增：
- `tools/v84/capability.go`
- `tools/v84/datasource_proxy.go`
- `tools/v84/compat_alias.go`
- `tools/v84/search_folders.go`
- `tools/v84/dashboard_helpers.go`
- `tools/v84/annotations_extra.go`
- `tools/v84/navigation.go`
- `tools/v84/examples.go`
- `tools/v84/prometheus_extra.go`
- `tools/v84/loki.go`
- `tools/v84/clickhouse.go`
- `tools/v84/search_logs.go`
- `tools/v84/alerting_unified.go`
- `tools/v84/rendering.go`

修改：
- `tools/v84/registry.go`（注册分层 + optional 开关生效）
- `tools/v84/types.go`（补充新增工具输出结构）
- `tools/v84/v84_unit_test.go`（新增单测）

## 7. 错误模型与返回契约

沿用现有 `APIError` 契约：
- `statusCode`
- `message`
- `upstream`

规则：
- 配置缺失/鉴权失败：返回 `HardError`
- HTTP 非 2xx：`wrapRawAPIError`
- OpenAPI client 失败：`wrapOpenAPIError`
- 工具级参数错误：返回可读 `fmt.Errorf`，由工具层转换为可恢复错误

## 8. 测试方案

### 8.1 单测覆盖要求

每个新增工具至少覆盖：
- 正常路径 1 个
- 参数校验失败 1 个
- 上游 API 失败 1 个

关键工具额外覆盖：
- 兼容别名是否正确委派
- datasource 解析优先级 `id > uid > name`
- optional capability 为 false 时不注册

### 8.2 推荐命令

按仓库约定执行：

```bash
GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn spkit run go test ./tools/v84 -tags unit
GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn spkit run go test ./... -tags unit
```

## 9. 发布与回滚策略

分 5 个 PR 发布：
1. PR-1：P0（13）
2. PR-2：Prometheus 增补（3）
3. PR-3：Loki + Search Logs（6）
4. PR-4：ClickHouse（3）
5. PR-5：P2 optional（7）

回滚策略：
- 按 PR 维度可逆
- optional 工具由注册开关控制，可在不回滚代码情况下快速禁用

## 10. 验收标准

功能验收：
- P0、P1、P2 工具在 `list_tools` 可见性符合开关设计
- P0/P1 在 Grafana 8.4.7 环境可调用成功
- P2 在 capability 不满足时不注册，满足时可调用

质量验收：
- `go test`（unit）通过
- 不引入 `golangci-lint` 新告警
- 文档与工具注册保持一致

## 11. 风险与缓解

风险：
- 不同 8.4.7 部署对 unified alerting 支持差异大
- datasource proxy 响应格式在插件链路下不稳定
- 大结果集（日志/时序）导致响应体过大

缓解：
- capability gating + optional 注册
- 统一 proxy 访问层加响应体大小限制与错误归一
- 查询类工具引入 `limit/page/time range` 保护参数与默认值
