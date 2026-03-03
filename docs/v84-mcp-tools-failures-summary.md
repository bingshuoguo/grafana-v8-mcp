# v84 MCP Tools 失败项原因总结

基线：2026-03-03 第二轮 42 tools 实测结果（修复后复测）。

## 1. 失败工具清单（6/42）

| Tool | 失败现象 | 原因判断 | 可修复性 |
| --- | --- | --- | --- |
| `get_dashboard_versions` | `could not determine numeric dashboard ID for uid` | `/api/dashboards/uid/:uid` 返回结构在当前环境下未能稳定解析到 numeric `dashboard.id`，导致后续 `/api/dashboards/id/:id/versions` 无法发起 | 高（工具逻辑可继续兼容增强） |
| `get_annotation_tags` | `context deadline exceeded` / `Client.Timeout exceeded while awaiting headers` | `/api/annotations/tags` 在当前 Grafana 实例侧响应超时，已出现“未返回 header 即超时” | 中（需服务端/网关联动） |
| `query_clickhouse` | `clickhouse query: Query data error` | ClickHouse datasource 可发现但查询链路失败，属于后端连接/权限/插件配置问题 | 低-中（主要依赖环境配置） |
| `list_clickhouse_tables` | `clickhouse query: Query data error` | 同上，元数据查询也失败，说明不是单条 SQL 语句问题 | 低-中（主要依赖环境配置） |
| `describe_clickhouse_table` | `clickhouse query: Query data error` | 同上，表结构查询链路失败 | 低-中（主要依赖环境配置） |
| `search_logs`（ClickHouse 路径） | `search clickhouse: clickhouse query: Query data error` | 依赖 ClickHouse 底层查询能力，随上游失败而失败 | 低-中（先修复 ClickHouse 数据源） |

## 2. 根因归类

1. 工具侧兼容问题（1项）
- `get_dashboard_versions`

2. 服务端性能/稳定性问题（1项）
- `get_annotation_tags`

3. 外部数据源环境问题（4项）
- ClickHouse 相关：`query_clickhouse`、`list_clickhouse_tables`、`describe_clickhouse_table`、`search_logs`

## 3. 建议修复顺序

1. `get_dashboard_versions`（优先）
- 增强 UID -> numeric ID 解析容错：兼容更多响应字段/类型，并输出原始响应关键字段日志（脱敏后）用于定位。

2. `get_annotation_tags`
- 与 Grafana/网关侧联调：确认 `/api/annotations/tags` 在大数据量下的查询耗时、网关超时阈值、是否需要服务端索引优化。

3. ClickHouse 四项
- 先在 Grafana UI/后端验证 datasource 健康度（连接、账号权限、database、readonly 策略、插件版本兼容）。
- 只有在环境链路恢复后，再复测 4 个 tool。

## 4. 备注

- 本轮已修复并转成功的失败项：`query_datasource`、`query_datasource_expressions`。
- 当前文档仅统计“严格失败（❌）”，未包含“环境限制（⚠️）”项（如无 Loki datasource、无 signed-in user 上下文）。
