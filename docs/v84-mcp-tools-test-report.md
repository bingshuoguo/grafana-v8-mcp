# v84 Grafana MCP 42 工具测试报告

测试环境：
- 测试日期：2026-03-03
- MCP 服务器：`grafana`（Codex 连接）
- Grafana 版本：`8.4.7`（`get_health` 返回）
- 测试范围：v84 读路径 42 tools（Core + Extensions + P1）

---

## 1. 双轮结果总览（修复前 vs 修复后）

| 指标 | 第一轮（修复前） | 第二轮（修复后复测） | 变化 |
| --- | ---: | ---: | ---: |
| ✅ 严格成功 | 28 | 30 | +2 |
| ❌ 失败 | 8 | 6 | -2 |
| ⚠️ 环境限制 | 6 | 6 | 0 |
| Reachability `(成功+环境限制)/42` | 80.95% | 85.71% | +4.76pp |
| Strict Success `成功/42` | 66.67% | 71.43% | +4.76pp |

结论：修复后第二轮通过率提升，主要收益来自 `query_datasource` 与 `query_datasource_expressions` 两项由失败转成功。

---

## 2. 第二轮 42 tools 逐项结果

### 2.1 Core / 组织 /通用查询

| 工具 | 第二轮结果 | 备注 |
| --- | --- | --- |
| `get_health` | ✅ | `database=ok`, `version=8.4.7` |
| `get_current_user` | ⚠️ | `[404] getSignedInUserNotFound`（当前 token 无有效 signed-in user 上下文） |
| `get_current_org` | ✅ | `id=36`, `name=Map` |
| `search_dashboards` | ✅ | 正常返回 dashboard 列表 |
| `get_dashboard_by_uid` | ✅ | 正常返回 dashboard 详情 |
| `list_folders` | ✅ | 正常 |
| `list_datasources` | ✅ | 正常（含分页） |
| `get_datasource` | ✅ | 正常 |
| `resolve_datasource_ref` | ✅ | 正常 |
| `query_datasource` | ✅ | 已修复，`/tsdb/query` 返回 2xx |
| `query_datasource_expressions` | ✅ | 已修复，`/ds/query` 兼容回退逻辑生效 |
| `query_prometheus` | ✅ | 在 `qrVnYb2nz(gcp_container)` 返回 2xx（可为空结果） |
| `list_prometheus_label_values` | ✅ | 2xx（空集合） |
| `list_prometheus_metric_names` | ✅ | 2xx（空集合） |
| `get_annotations` | ✅ | 正常返回注解 |
| `list_legacy_alerts` | ✅ | 正常 |
| `list_legacy_notification_channels` | ✅ | 正常 |
| `list_org_users` | ✅ | 正常（数据量大） |
| `list_teams` | ✅ | 正常 |

### 2.2 Extensions（兼容/辅助）

| 工具 | 第二轮结果 | 备注 |
| --- | --- | --- |
| `get_datasource_by_uid` | ✅ | 正常 |
| `get_datasource_by_name` | ✅ | 正常 |
| `list_users_by_org` | ✅ | 正常 |
| `search_folders` | ✅ | 正常 |
| `get_dashboard_panel_queries` | ✅ | 正常 |
| `get_dashboard_property` | ✅ | 正常 |
| `get_dashboard_summary` | ✅ | 正常 |
| `get_dashboard_versions` | ❌ | 仍报 `could not determine numeric dashboard ID for uid` |
| `get_annotation_tags` | ❌ | 仍超时（含重试与小 limit 变体） |
| `generate_deeplink` | ✅ | 正常 |
| `get_query_examples` | ✅ | `prometheus`/`clickhouse`/`loki` 均可返回 |

### 2.3 P1 数据源工具

| 工具 | 第二轮结果 | 备注 |
| --- | --- | --- |
| `list_prometheus_label_names` | ✅ | 2xx（空集合） |
| `list_prometheus_metric_metadata` | ✅ | 2xx（空集合） |
| `query_prometheus_histogram` | ✅ | 2xx（空结果 + hints） |
| `list_loki_label_names` | ⚠️ | `datasource not found`（当前组织无 Loki） |
| `list_loki_label_values` | ⚠️ | 同上 |
| `query_loki_logs` | ⚠️ | 同上 |
| `query_loki_stats` | ⚠️ | 同上 |
| `query_loki_patterns` | ⚠️ | 同上 |
| `query_clickhouse` | ❌ | `clickhouse query: Query data error` |
| `list_clickhouse_tables` | ❌ | 同上 |
| `describe_clickhouse_table` | ❌ | 同上 |
| `search_logs`（ClickHouse） | ❌ | `search clickhouse: clickhouse query: Query data error` |

---

## 3. 修复验证结果

### 3.1 已验证修复生效

| 工具 | 第一轮 | 第二轮 | 结论 |
| --- | --- | --- | --- |
| `query_datasource` | ❌ | ✅ | 修复生效 |
| `query_datasource_expressions` | ❌ | ✅ | 修复生效 |

### 3.2 未修复/非工具逻辑问题

| 工具 | 当前状态 | 归因 |
| --- | --- | --- |
| `get_dashboard_versions` | ❌ | 仍存在 UID -> numeric dashboard id 解析问题 |
| `get_annotation_tags` | ❌ | 服务端/链路超时（客户端已有延长超时+重试） |
| `query_clickhouse` / `list_clickhouse_tables` / `describe_clickhouse_table` / `search_logs` | ❌ | ClickHouse 后端查询链路错误（`Query data error`） |
| 5 个 Loki 工具 | ⚠️ | 当前组织无 Loki datasource |
| `get_current_user` | ⚠️ | 当前 token 无 signed-in user 上下文 |

---

## 4. 第二轮实测关键参数

- Dashboard UID：`aUyw7Xnnz`（另复测 `N9uZBy8Wz`）
- Prometheus（可用）：`uid=qrVnYb2nz`, `id=1191`, `name=gcp_container`
- Prometheus（后端 400）：`uid=RoRqN4iVz`, `id=3849`, `name=AI-Claim`
- ClickHouse：`uid=ktP56dnDk`, `name=ClickHouse`
- Loki：`list_datasources type=loki` 返回空

---

## 5. 本轮结论

修复后 42 tools 严格成功率从 `66.67%` 提升到 `71.43%`，Reachability 从 `80.95%` 提升到 `85.71%`。通过率提升由两项查询兼容修复直接贡献；剩余失败项主要是 `get_dashboard_versions` 逻辑残留与外部后端/环境限制（ClickHouse/Loki/用户上下文）。
