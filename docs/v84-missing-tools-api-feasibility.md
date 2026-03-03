# v84 缺失 60 个工具 API 可实现性研究

更新时间：2026-02-27

## 1. 结论摘要

- 对比结果：官方工具 71 个，v84 已实现 23 个，v84 缺失 60 个。
- 这 60 个工具中：
  - 4 个已有等价能力（仅命名变化，可做兼容别名）。
  - 9 个可直接基于 8.4.7 core API 或本地逻辑实现。
  - 12 个可通过 UID -> ID 兼容改造实现。
  - 6 个依赖 unified alerting 配置，条件可实现。
  - 21 个依赖插件/云能力，需环境支持。
  - 7 个 RBAC 工具在 8.4.7 OSS 上大概率不可用或不稳定。
  - 1 个渲染工具需 image-renderer 运行前提。

## 2. 研究口径

以“是否存在可调用 API + 是否满足 Grafana 8.4.7 默认部署前提”为判定标准，逐项查看缺失工具在源码中的实际调用路径。

主要证据来源：

- [tools/prometheus.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/prometheus.go)
- [tools/loki.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/loki.go)
- [tools/clickhouse.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/clickhouse.go)
- [tools/pyroscope.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/pyroscope.go)
- [tools/search_logs.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/search_logs.go)
- [tools/admin.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/admin.go)
- [tools/alerting.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/alerting.go)
- [tools/alerting_client.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/alerting_client.go)
- [tools/oncall.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/oncall.go)
- [tools/incident.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/incident.go)
- [tools/sift.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/sift.go)
- [tools/asserts.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/asserts.go)
- [tools/rendering.go](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/tools/rendering.go)
- [docs/grafana-8.4.7-mcp-tool-spec.md](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/docs/grafana-8.4.7-mcp-tool-spec.md)
- [docs/grafana-9plus-tools.md](/Users/bingshuo.guo/Work/AIOps/mcp-grafana-8.4.7/docs/grafana-9plus-tools.md)

## 3. 可实现性分层（60 个）

### A. 已有等价能力（命名兼容即可）4 个

- `update_dashboard` -> `upsert_dashboard`
- `list_users_by_org` -> `list_org_users`
- `get_datasource_by_uid` -> `get_datasource`
- `get_datasource_by_name` -> `get_datasource`

说明：这 4 个不需要新 API，建议做兼容 wrapper/tool alias。

### B. 8.4.7 可直接实现（低风险）9 个

- `search_folders`
- `create_graphite_annotation`
- `update_annotation`
- `get_annotation_tags`
- `get_dashboard_panel_queries`
- `get_dashboard_property`
- `get_dashboard_summary`
- `generate_deeplink`
- `get_query_examples`

说明：

- `search_folders` 可直接走 `/api/search?type=dash-folder`。
- dashboard 三个 helper 由 `get_dashboard_by_uid` 做二次处理即可。
- `generate_deeplink` 是本地 URL 组装。
- `get_query_examples` 为静态示例数据。

### C. 需 UID -> ID 兼容改造（中风险）12 个

- `list_prometheus_label_names`
- `list_prometheus_metric_metadata`
- `query_prometheus_histogram`
- `list_loki_label_names`
- `list_loki_label_values`
- `query_loki_logs`
- `query_loki_stats`
- `query_loki_patterns`
- `query_clickhouse`
- `list_clickhouse_tables`
- `describe_clickhouse_table`
- `search_logs`

说明：

- 当前实现多使用 `/api/datasources/uid/{uid}` 或 `/api/datasources/proxy/uid/{uid}`。
- v84 需改为 ID-first：`resolve_datasource_ref` -> `/api/datasources/proxy/{id}/...`。

### D. 条件可实现（统一告警能力）6 个

- `list_alert_rules`
- `get_alert_rule_by_uid`
- `create_alert_rule`
- `update_alert_rule`
- `delete_alert_rule`
- `list_contact_points`

说明：

- API 路径存在，但 8.4.7 部署可能仅 legacy alerting 可用。
- 建议作为 optional tools，运行时探测后注册。

### E. 插件/云能力依赖（需环境支持）21 个

OnCall（7）：

- `list_oncall_schedules`
- `get_oncall_shift`
- `get_current_oncall_users`
- `list_oncall_teams`
- `list_oncall_users`
- `list_alert_groups`
- `get_alert_group`

Incident（4）：

- `list_incidents`
- `create_incident`
- `add_activity_to_incident`
- `get_incident`

Sift（5）：

- `get_sift_investigation`
- `get_sift_analysis`
- `list_sift_investigations`
- `find_error_pattern_logs`
- `find_slow_requests`

Asserts（1）：

- `get_assertions`

Pyroscope（4）：

- `list_pyroscope_label_names`
- `list_pyroscope_label_values`
- `list_pyroscope_profile_types`
- `fetch_pyroscope_profile`

说明：这组依赖 `grafana-irm-app`、`grafana-ml-app`、`grafana-asserts-app`、Pyroscope 等插件或云产品能力，不属于 vanilla 8.4.7 默认能力。

### F. 8.4.7 OSS 大概率不可用/不稳定（RBAC）7 个

- `list_all_roles`
- `get_role_details`
- `get_role_assignments`
- `list_user_roles`
- `list_team_roles`
- `get_resource_permissions`
- `get_resource_description`

说明：该组依赖 `access_control` / RBAC 相关接口，和 8.4.7 默认 OSS 能力不完全一致。

### G. 有 API 但有运行前提 1 个

- `get_panel_image`

说明：需安装并可访问 Grafana image-renderer 服务，否则工具返回 renderer 不可用错误。

## 4. 可落地优先级实现清单

### P0（先做，价值高、风险低）13 个

- `update_dashboard`（兼容别名）
- `list_users_by_org`（兼容别名）
- `get_datasource_by_uid`（兼容别名）
- `get_datasource_by_name`（兼容别名）
- `search_folders`
- `get_dashboard_panel_queries`
- `get_dashboard_property`
- `get_dashboard_summary`
- `create_graphite_annotation`
- `update_annotation`
- `get_annotation_tags`
- `generate_deeplink`
- `get_query_examples`

### P1（第二批，需兼容改造）12 个

- `list_prometheus_label_names`
- `list_prometheus_metric_metadata`
- `query_prometheus_histogram`
- `list_loki_label_names`
- `list_loki_label_values`
- `query_loki_logs`
- `query_loki_stats`
- `query_loki_patterns`
- `query_clickhouse`
- `list_clickhouse_tables`
- `describe_clickhouse_table`
- `search_logs`

### P2（按部署能力开关）7 个

- `list_alert_rules`
- `get_alert_rule_by_uid`
- `create_alert_rule`
- `update_alert_rule`
- `delete_alert_rule`
- `list_contact_points`
- `get_panel_image`

### P3（暂缓，环境依赖重）28 个

- OnCall 7 + Incident 4 + Sift 5 + Asserts 1 + Pyroscope 4 + RBAC 7

## 5. 建议的 PR 切分

1. PR-1：P0 全量（13 个），优先做兼容别名 + 低风险新工具。
2. PR-2：Prometheus 3 个（P1 子集）。
3. PR-3：Loki 5 + Search Logs 1（P1 子集）。
4. PR-4：ClickHouse 3（P1 子集）。
5. PR-5：P2 条件工具 + feature flag + runtime capability 探测。

## 6. 备注

- 本文“可实现”指工程上存在可落地 API 路径，不等于默认部署一定可用。
- 对插件/云能力工具，建议统一放到 `optional-tools` 维度，并在启动时进行能力探测后再注册。
