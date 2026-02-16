# Grafana 9.0+ Tools 识别结果

## 判定口径
本文将“Grafana 9.0 版本后的 tools”定义为：在当前代码实现中，依赖 Grafana 9.0 引入的 datasource UID 相关接口（如 `/api/datasources/uid/{uid}`）才能正常工作的 tools。

判定依据：
- `README.md` 明确说明 `/datasources/uid/{uid}` 是 Grafana 9.0 引入，低版本会失败（`README.md:889`）。
- 代码中凡是调用 `getDatasourceByUID(...)`（`tools/datasources.go:125`）或直接拼接 UID 资源路径（如 `tools/prometheus.go:40`）的 tool，视为 9.0+ 依赖。

## 一、必然依赖 Grafana 9.0+ 的 tools（20 个）

### 1) Datasource
- `get_datasource_by_uid`

### 2) Prometheus（6）
- `list_prometheus_metric_metadata`
- `query_prometheus`
- `list_prometheus_metric_names`
- `list_prometheus_label_names`
- `list_prometheus_label_values`
- `query_prometheus_histogram`

证据：`promClientFromContext` 先调用 `getDatasourceByUID`，并使用 `/api/datasources/uid/%s/resources`（`tools/prometheus.go:32-40`）。

### 3) Loki（5）
- `list_loki_label_names`
- `list_loki_label_values`
- `query_loki_logs`
- `query_loki_stats`
- `query_loki_patterns`

证据：`newLokiClient` 调用 `getDatasourceByUID`，并走 `/api/datasources/proxy/uid/%s`（`tools/loki.go:62-70`）。

### 4) ClickHouse（3）
- `query_clickhouse`
- `list_clickhouse_tables`
- `describe_clickhouse_table`

证据：`query_clickhouse` 使用 `newClickHouseClient`，后者调用 `getDatasourceByUID`（`tools/clickhouse.go:85-89`, `tools/clickhouse.go:308-310`）。

### 5) Search Logs（1）
- `search_logs`

证据：处理函数直接调用 `getDatasourceByUID`（`tools/search_logs.go:314-317`）。

### 6) Pyroscope（4）
- `list_pyroscope_label_names`
- `list_pyroscope_label_values`
- `list_pyroscope_profile_types`
- `fetch_pyroscope_profile`

证据：`newPyroscopeClient` 调用 `getDatasourceByUID`，并走 `/api/datasources/proxy/uid/...`（`tools/pyroscope.go:304`, `tools/pyroscope.go:313`）。

## 二、条件性依赖 Grafana 9.0+ 的 tools（2 个）
- `list_alert_rules`
- `list_contact_points`

说明：
- 仅当传入 `datasourceUid` 参数时，才会走 datasource UID 逻辑（`tools/alerting.go:65-67`, `tools/alerting.go:436-437`），并调用 `getDatasourceByUID`（`tools/alerting.go:279`, `tools/alerting.go:489`）。
- 不传 `datasourceUid` 时，两者可走 Grafana-managed 路径，不必然触发 9.0+ 依赖。

## 三、补充（推断项）
`proxied` 动态工具（运行时发现并注册）使用 `/api/datasources/proxy/uid/...` 进行探测与转发（`proxied_tools.go:132`, `proxied_tools.go:159`）。  
这类工具名称运行时动态生成，无法在静态代码中给出固定列表；但从实现路径看，**高概率同样依赖 9.0+ 的 UID 能力**。
