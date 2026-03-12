# 技术实现方案：v84 Prometheus 工具集

## 一、整体架构

- **新增文件**：`tools/prometheus.go`（约 300 行）
- **修改文件**：`tools/registry.go`（3 行追加）

所有新工具复用现有基础设施：

- `doAPIRequest` → HTTP 调用（自动处理认证、TLS、超时）
- `resolveDatasourceRef` → id > uid > name 数据源解析
- `MustTool` + `DatasourceRef` → 与现有工具完全一致的注册模式

---

## 二、代理 URL 构造逻辑

`doAPIRequest` 构造 URL 的方式（`common.go:152`）：

```
{grafana_url}/api + path
```

因此传入 path 为：

- `/datasources/proxy/{id}/api/v1/query_range` → 完整请求到 Prometheus

这与 8.4.7 源码路由 `apiRoute.Any("/datasources/proxy/:id/*", ...)` 完全对应，无需任何 hack。

---

## 三、时间解析设计

Prometheus API 要求 Unix 秒（float64 字符串）。需要支持三种输入格式：

| 输入 | 解析结果 |
|------|----------|
| `"now"` | `time.Now()` |
| `"now-1h"` | `time.Now().Add(-1h)` |
| `"now-30m"` | `time.Now().Add(-30m)` |
| `"2026-02-26T13:21:47+08:00"` | `time.Parse(time.RFC3339, s)` |

不引入外部依赖，在 `prometheus.go` 内自实现 `parsePromTime(s string) (time.Time, error)`，约 30 行：

```go
func parsePromTime(s string) (time.Time, error) {
    s = strings.TrimSpace(s)
    if s == "" || s == "now" {
        return time.Now(), nil
    }
    // RFC3339
    if t, err := time.Parse(time.RFC3339, s); err == nil {
        return t, nil
    }
    // "now-Xunit" or "now+Xunit"
    if strings.HasPrefix(s, "now") {
        suffix := s[3:]   // e.g. "-1h", "-30m", "+5m"
        if suffix == "" {
            return time.Now(), nil
        }
        d, err := time.ParseDuration(suffix) // "-1h", "-30m" → Go 支持负数 duration
        if err != nil {
            return time.Time{}, fmt.Errorf("invalid time expression %q: %w", s, err)
        }
        return time.Now().Add(d), nil
    }
    return time.Time{}, fmt.Errorf("unsupported time format: %q", s)
}
```

---

## 四、工具详细设计

### 4.1 query_prometheus（P0）

**HTTP 调用：**

- Range 查询：
  ```http
  GET /api/datasources/proxy/{id}/api/v1/query_range
      ?query=<PromQL>&start=<unix_sec>&end=<unix_sec>&step=<seconds>
  ```
- Instant 查询：
  ```http
  GET /api/datasources/proxy/{id}/api/v1/query
      ?query=<PromQL>&time=<unix_sec>
  ```

**入参结构：**

```go
type QueryPrometheusRequest struct {
    Datasource *DatasourceRef `json:"datasource,omitempty" jsonschema:"description=Datasource reference (id/uid/name). Omit to use default."`
    Expr      string `json:"expr" jsonschema:"required,description=PromQL expression"`
    Start     string `json:"start" jsonschema:"required,description=Start time: 'now-1h', 'now', or RFC3339"`
    End       string `json:"end,omitempty" jsonschema:"description=End time (default: now). Required for range queries."`
    Step      string `json:"step,omitempty" jsonschema:"description=Step interval (default: 60s). E.g. '30s', '1m'."`
    QueryType string `json:"queryType,omitempty" jsonschema:"description=Query type: 'range' (default) or 'instant'"`
}
```

**返回结构（LLM 可直接读取）：**

原始 Prometheus 响应中 `values` 是 `[[unix_ts, "0.95"], ...]`，转换为带可读时间的格式，方便 LLM 直接分析：

```go
type PromSample struct {
    Time  string `json:"time"`  // "2026-02-26T13:21:00Z"
    Value string `json:"value"` // "0.9500"
}

type PromSeriesResult struct {
    Metric map[string]string `json:"metric"`       // label set
    Values []PromSample      `json:"values"`       // range query
    Value  *PromSample       `json:"value,omitempty"` // instant query
}

type QueryPrometheusResponse struct {
    ResultType string            `json:"resultType"` // "matrix" or "vector"
    Result     []PromSeriesResult `json:"result"`
    Hints      []string          `json:"hints,omitempty"` // 空结果提示
}
```

**执行流程：**

1. 校验 `expr`、`start` 必填
2. `resolveDatasourceRef` → 得到 `datasourceID` (int64)
3. `parsePromTime(start)` → startTs
4. `parsePromTime(end)` → endTs（默认 now）
5. 解析 step → 秒数（默认 60）
6. 构造 `url.Values{query, start, end, step}`
7. `doAPIRequest("GET", "/datasources/proxy/{id}/api/v1/query_range", query, nil)`
8. 解析 Prometheus JSON → 转换时间格式
9. 若 Result 为空 → 生成 hints（检查 PromQL 语法、时间范围建议）

**关键：step 的默认值逻辑**

根据时间跨度自动选择合理 step，避免数据点过多：

```go
func defaultStep(start, end time.Time) int {
    duration := end.Sub(start)
    switch {
    case duration <= 30*time.Minute: return 30   // 30s
    case duration <= 3*time.Hour:   return 60   // 1m
    case duration <= 24*time.Hour:  return 300  // 5m
    default:                        return 3600 // 1h
    }
}
```

---

### 4.2 list_prometheus_label_values（P2）

**HTTP 调用：**

```http
GET /api/datasources/proxy/{id}/api/v1/label/{labelName}/values
    [?start=<unix_sec>&end=<unix_sec>]
```

**入参结构：**

```go
type ListPrometheusLabelValuesRequest struct {
    Datasource *DatasourceRef `json:"datasource,omitempty" jsonschema:"description=Datasource reference (id/uid/name)"`
    LabelName  string         `json:"labelName" jsonschema:"required,description=Label name to query values for"`
    Start      string         `json:"start,omitempty" jsonschema:"description=Optional start time for filtering"`
    End        string         `json:"end,omitempty" jsonschema:"description=Optional end time for filtering"`
    Limit      int            `json:"limit,omitempty" jsonschema:"description=Max values to return (default 100)"`
}
```

**返回结构：**

```go
type ListPrometheusLabelValuesResponse struct {
    LabelName  string   `json:"labelName"`
    Values     []string `json:"values"`
    Total      int      `json:"total"`
    Truncated  bool     `json:"truncated"` // 是否因 limit 截断
}
```

**执行流程：**

1. `resolveDatasourceRef` → datasourceID
2. path = `/datasources/proxy/{id}/api/v1/label/{labelName}/values`
3. 可选加 start/end query params
4. `doAPIRequest("GET", path, query, nil)`
5. 解析 `{"status":"success","data":["v1","v2",...]}`
6. 应用 limit 截断

---

### 4.3 list_prometheus_metric_names（P3）

本质：`list_prometheus_label_values` 的特例，`labelName` 固定为 `__name__`，增加正则过滤。

**HTTP 调用：**

```http
GET /api/datasources/proxy/{id}/api/v1/label/__name__/values
```

**入参结构：**

```go
type ListPrometheusMetricNamesRequest struct {
    Datasource *DatasourceRef `json:"datasource,omitempty" jsonschema:"description=Datasource reference (id/uid/name)"`
    Regex      string        `json:"regex,omitempty" jsonschema:"description=Optional regex filter on metric names"`
    Limit      int           `json:"limit,omitempty" jsonschema:"description=Max results (default 50)"`
    Page       int           `json:"page,omitempty" jsonschema:"description=Page number for pagination (default 1)"`
}
```

**返回结构：**

```go
type ListPrometheusMetricNamesResponse struct {
    Metrics []string `json:"metrics"`
    Total   int      `json:"total"`   // 过滤后总数
    Page    int      `json:"page"`
    HasMore bool     `json:"hasMore"`
}
```

**执行流程：**

1. `resolveDatasourceRef` → datasourceID
2. `doAPIRequest` GET `/datasources/proxy/{id}/api/v1/label/__name__/values`
3. 若 `regex != ""` → Go 端过滤（Prometheus API 不支持服务端 regex 过滤 label values）
4. 分页：`(page-1)*limit` 到 `page*limit`

---

## 五、Prometheus 响应解析结构

供内部 `json.Unmarshal` 使用（`tools/prometheus.go` 内部类型，不导出）：

```go
type promResponse struct {
    Status    string   `json:"status"`
    Data      promData `json:"data"`
    Error     string   `json:"error,omitempty"`
    ErrorType string   `json:"errorType,omitempty"`
}

type promData struct {
    ResultType string       `json:"resultType"`
    Result     []promSeries `json:"result"`
}

type promSeries struct {
    Metric map[string]string  `json:"metric"`
    Values [][2]interface{}   `json:"values,omitempty"` // range: [[ts, "v"], ...]
    Value  [2]interface{}     `json:"value,omitempty"`   // instant: [ts, "v"]
}
```

**时间戳转换：**

`[[1700000000, "0.95"]]` → `{Time:"2026-02-26T13:21:00Z", Value:"0.9500"}`

```go
func convertSample(raw [2]interface{}) (PromSample, error) {
    tsFloat, ok := raw[0].(float64)
    if !ok { return PromSample{}, fmt.Errorf("invalid timestamp") }
    val, _ := raw[1].(string)
    t := time.Unix(int64(tsFloat), 0).UTC()
    return PromSample{Time: t.Format(time.RFC3339), Value: val}, nil
}
```

---

## 六、registry.go 变更

在 `tools/registry.go` 中追加 3 行：

```go
func AddV84Tools(m *server.MCPServer, enableWriteTools, enableOptionalTools bool) {
    // ... 现有 15 行 ...

    // Prometheus 专用工具（只读）
    QueryPrometheusTool.Register(m)
    ListPrometheusLabelValuesTool.Register(m)
    ListPrometheusMetricNamesTool.Register(m)

    if enableWriteTools { ... }
}
```

---

## 七、工具描述字符串设计（影响 Agent 调用质量）

```go
var QueryPrometheusTool = mcpgrafana.MustTool(
    "query_prometheus",
    "WORKFLOW: list_prometheus_metric_names → list_prometheus_label_values → query_prometheus. "+
    "Execute a PromQL query against a Prometheus datasource via Grafana proxy. "+
    "Returns human-readable time series with ISO8601 timestamps. "+
    "For AIOps: set start/end around the alert time (e.g. start='now-30m', end='now').",
    queryPrometheus,
    mcp.WithTitleAnnotation("Query Prometheus metrics"),
    mcp.WithReadOnlyHintAnnotation(true),
    mcp.WithIdempotentHintAnnotation(true),
)

var ListPrometheusLabelValuesTool = mcpgrafana.MustTool(
    "list_prometheus_label_values",
    "Get all values for a specific label in a Prometheus datasource. "+
    "Use to resolve Grafana dashboard template variables (e.g. $instance, $job) "+
    "before constructing PromQL queries.",
    listPrometheusLabelValues,
    mcp.WithTitleAnnotation("List Prometheus label values"),
    mcp.WithReadOnlyHintAnnotation(true),
    mcp.WithIdempotentHintAnnotation(true),
)

var ListPrometheusMetricNamesTool = mcpgrafana.MustTool(
    "list_prometheus_metric_names",
    "DISCOVERY: Call this first to find available metrics before querying. "+
    "Lists metric names in a Prometheus datasource with optional regex filter and pagination.",
    listPrometheusMetricNames,
    mcp.WithTitleAnnotation("List Prometheus metric names"),
    mcp.WithReadOnlyHintAnnotation(true),
    mcp.WithIdempotentHintAnnotation(true),
)
```

---

## 八、单元测试方案

参照 `tools/v84_unit_test.go` 现有模式（build tag `//go:build unit`），新增：

| 测试用例 | 覆盖范围 |
|----------|----------|
| `TestParsePromTime` | now / now-1h / RFC3339 / 无效输入 |
| `TestDefaultStep` | 各时间跨度 |
| `TestConvertSample` | 时间戳转换正确性 |
| `TestQueryPrometheus_*` | mock HTTP server 返回 matrix / vector / 空 / error |
| `TestListPrometheusLabelValues_*` | mock HTTP server 覆盖正常 / 分页 / 空 |
| `TestListPrometheusMetricNames_*` | regex 过滤、分页 |

---

## 九、方案总结

| 维度 | 设计决策 |
|------|----------|
| 新增文件 | 1 个：`tools/prometheus.go` |
| 新增工具 | 3 个（P0/P2/P3） |
| 外部依赖 | 零新增，复用现有 `doAPIRequest` / `resolveDatasourceRef` |
| 时间格式 | 入参支持相对时间和 RFC3339；返回值统一转为 RFC3339（LLM 可读） |
| 8.4.7 兼容 | 全部走 `/datasources/proxy/:id/*` 路由，源码已确认 |
| get_dashboard_by_id | 无需新增（list_legacy_alerts 已返回 dashboardUid，告警 URL 也含 UID） |
