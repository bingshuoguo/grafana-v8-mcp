# v84 Remaining 5 Tools — Technical Design

## Context

After P0/P1/P2 implementation, five Grafana 8.4.7-compatible tools remain unimplemented.
This document covers the technical design for implementing them in a single batch.

---

## Tools Summary

| Tool | File | Registration | API Endpoint |
|------|------|-------------|-------------|
| `delete_annotation` | `annotations_extra.go` | `registerV84CoreWriteTools` | `DELETE /api/annotations/{id}` |
| `get_dashboard_versions` | `dashboard_helpers.go` | `registerV84CoreReadOnlyExtensions` | `GET /api/dashboards/id/{id}/versions` |
| `query_datasource_expressions` | `query.go` | `registerV84CoreReadTools` | `POST /api/ds/query` |
| `get_firing_alerts` | `alerting_unified.go` | `registerV84P2OptionalTools` | `GET /api/alertmanager/grafana/api/v2/alerts` |
| `get_alert_rules_with_state` | `alerting_unified.go` | `registerV84P2OptionalTools` | `GET /api/prometheus/grafana/api/v1/rules` |

---

## 1. `delete_annotation`

### API

```
DELETE /api/annotations/{id}
```

Response (200 OK):
```json
{"message": "Annotation deleted"}
```

### Types

```go
type DeleteAnnotationRequest struct {
    ID int64 `json:"id" jsonschema:"required,description=Annotation ID to delete"`
}
type DeleteAnnotationResponse struct {
    Message string `json:"message,omitempty"`
    ID      int64  `json:"id"`
}
```

### Validation

- `ID <= 0` → error "id is required"

### Registration

Added to `registerV84CoreWriteTools` alongside existing write annotation tools.

### Risk

Low. Mirrors existing UpdateAnnotationTool pattern. Pure DELETE with ID-in-path.

---

## 2. `get_dashboard_versions`

### API

Grafana 8.4.7 only supports **numeric-ID-based** version history:
```
GET /api/dashboards/id/{dashboardId}/versions?limit=10&start=0
```

UID-based (`/api/dashboards/uid/{uid}/versions`) was added in Grafana 9.0.

### Implementation

1. Call `getDashboardByUID(ctx, uid)` → get `meta["id"]` (numeric ID)
2. Call `GET /api/dashboards/id/{id}/versions?limit=N&start=N`

### Types

```go
type GetDashboardVersionsRequest struct {
    UID   string `json:"uid" jsonschema:"required"`
    Limit *int   `json:"limit,omitempty"` // default 10
    Start *int   `json:"start,omitempty"` // pagination offset
}
type DashboardVersionInfo struct {
    ID            int    `json:"id"`
    DashboardID   int    `json:"dashboardId"`
    Version       int    `json:"version"`
    ParentVersion int    `json:"parentVersion"`
    RestoredFrom  int    `json:"restoredFrom"`
    Created       string `json:"created"`
    CreatedBy     string `json:"createdBy"`
    Message       string `json:"message,omitempty"`
}
```

### Risk

Low. Two sequential API calls (getDashboardByUID already tested).
Requires `net/url` and `strconv` imports added to `dashboard_helpers.go`.

---

## 3. `query_datasource_expressions`

### Motivation

`query_datasource` uses `/api/tsdb/query` (legacy format, numeric datasource ID).
`query_datasource_expressions` uses `/api/ds/query` (modern format, UID+type, data frames, supports expressions).

### API

```
POST /api/ds/query
```

Request:
```json
{
  "queries": [
    {
      "datasource": {"uid": "abc", "type": "prometheus"},
      "refId": "A",
      "expr": "up"
    }
  ],
  "from": "now-1h",
  "to": "now"
}
```

Response (data frames format):
```json
{
  "results": {
    "A": {
      "frames": [
        {"schema": {"fields": [...]}, "data": {"values": [...]}}
      ]
    }
  }
}
```

### Auto-injection

If `datasource` param is provided, resolve it to `{uid, type}` and inject into any query
that does not already have a `datasource` field.

### Types

```go
type QueryDatasourceExpressionsRequest struct {
    From       string           `json:"from" jsonschema:"required"`
    To         string           `json:"to" jsonschema:"required"`
    Queries    []map[string]any `json:"queries" jsonschema:"required"`
    Datasource *DatasourceRef   `json:"datasource,omitempty"`
}
type QueryDatasourceExpressionsResponse struct {
    Raw     json.RawMessage            `json:"raw,omitempty"`
    Results map[string]json.RawMessage `json:"results,omitempty"`
}
```

### Risk

Low. Same `doAPIRequest` pattern as ClickHouse (`/ds/query`). Returns raw frames — caller handles format.

---

## 4. `get_firing_alerts`

### API

```
GET /api/alertmanager/grafana/api/v2/alerts
```

Query parameters:
- `filter`: repeated, label matchers (e.g. `alertname=HighCPU`)
- `silenced`: bool (default server-side: true)
- `inhibited`: bool (default server-side: true)
- `active`: bool (default server-side: true)

Response:
```json
[
  {
    "labels": {"alertname": "HighCPU", "severity": "critical"},
    "annotations": {"summary": "..."},
    "startsAt": "2024-01-01T12:00:00Z",
    "endsAt": "0001-01-01T00:00:00Z",
    "fingerprint": "abc123",
    "status": {"state": "active", "silencedBy": [], "inhibitedBy": []}
  }
]
```

### Types

```go
type FiringAlertStatus struct {
    State       string   `json:"state"`
    SilencedBy  []string `json:"silencedBy"`
    InhibitedBy []string `json:"inhibitedBy"`
}
type FiringAlert struct {
    Labels       map[string]string  `json:"labels"`
    Annotations  map[string]string  `json:"annotations"`
    StartsAt     string             `json:"startsAt"`
    EndsAt       string             `json:"endsAt,omitempty"`
    GeneratorURL string             `json:"generatorURL,omitempty"`
    Fingerprint  string             `json:"fingerprint,omitempty"`
    Status       *FiringAlertStatus `json:"status,omitempty"`
}
```

### Risk

Medium. Requires Unified Alerting (gated by `enableOptionalTools`). Returns empty slice when
no alerts are firing — never nil.

---

## 5. `get_alert_rules_with_state`

### API

```
GET /api/prometheus/grafana/api/v1/rules?type=alert
```

Prometheus-compat endpoint. Returns rules with runtime evaluation state.
Differs from Ruler API (`/ruler/...`) which returns rule definitions without state.

Response (abridged):
```json
{
  "status": "success",
  "data": {
    "groups": [
      {
        "name": "group-a",
        "file": "folder-name",
        "rules": [
          {
            "name": "HighCPU",
            "state": "firing",
            "health": "ok",
            "lastEvaluation": "...",
            "evaluationTime": 0.123,
            "type": "alerting",
            "labels": {},
            "annotations": {},
            "alerts": [{"labels": {}, "state": "firing", "activeAt": "..."}]
          }
        ]
      }
    ]
  }
}
```

### Filtering

Client-side filtering after API response:
- `state`: exact match (firing/pending/inactive)
- `ruleName`: partial substring match

Groups with 0 matching rules are omitted from the output.

### Types

```go
type AlertInstanceWithState struct {
    Labels      map[string]string `json:"labels"`
    Annotations map[string]string `json:"annotations,omitempty"`
    State       string            `json:"state"`
    ActiveAt    string            `json:"activeAt,omitempty"`
    Value       string            `json:"value,omitempty"`
}
type AlertRuleWithState struct {
    Name, State, Health string
    LastEvaluation      string
    EvaluationTime      float64
    Labels, Annotations map[string]string
    Alerts              []AlertInstanceWithState
}
type AlertRuleGroupWithState struct {
    Name, Namespace string
    Rules           []AlertRuleWithState
}
```

### Risk

Medium. Requires Unified Alerting. `strconv` import needed in `alerting_unified.go`
(for `get_firing_alerts` query building). Internal struct `prometheusRulesResponse` is unexported.

---

## Registry Changes

```
registerV84CoreWriteTools:
  + DeleteAnnotationTool

registerV84CoreReadOnlyExtensions:
  + GetDashboardVersionsTool (after GetDashboardSummaryTool)

registerV84CoreReadTools:
  + QueryDatasourceExpressionsTool (after QueryDatasourceTool)

registerV84P2OptionalTools:
  + GetFiringAlertsTool
  + GetAlertRulesWithStateTool
```

---

## Test File

`tools/v84/p3_tools_test.go` covers:
- All 5 tool vars not nil (`TestP3ToolsRegistered`)
- Input validation for delete_annotation, get_dashboard_versions, query_datasource_expressions
- JSON parsing for FiringAlert (Alertmanager v2 response)
- JSON parsing for prometheusRulesResponse
