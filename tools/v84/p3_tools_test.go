//go:build unit

package v84

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── tool registrations ───────────────────────────────────────────────────────

func TestP3ToolsRegistered(t *testing.T) {
	assert.NotNil(t, DeleteAnnotationTool)
	assert.NotNil(t, GetDashboardVersionsTool)
	assert.NotNil(t, QueryDatasourceExpressionsTool)
	assert.NotNil(t, GetFiringAlertsTool)
	assert.NotNil(t, GetAlertRulesWithStateTool)
}

// ─── delete_annotation ────────────────────────────────────────────────────────

func TestDeleteAnnotationValidation(t *testing.T) {
	_, err := deleteAnnotation(nil, DeleteAnnotationRequest{ID: 0})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestDeleteAnnotationNegativeID(t *testing.T) {
	_, err := deleteAnnotation(nil, DeleteAnnotationRequest{ID: -5})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

// ─── get_dashboard_versions ───────────────────────────────────────────────────

func TestGetDashboardVersionsValidation(t *testing.T) {
	_, err := getDashboardVersions(nil, GetDashboardVersionsRequest{UID: ""})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "uid is required")
}

// ─── query_datasource_expressions ─────────────────────────────────────────────

func TestQueryDatasourceExpressionsValidation(t *testing.T) {
	t.Run("missing from returns error", func(t *testing.T) {
		_, err := queryDatasourceExpressions(nil, QueryDatasourceExpressionsRequest{
			To:      "now",
			Queries: []map[string]any{{"refId": "A"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "from and to")
	})

	t.Run("missing to returns error", func(t *testing.T) {
		_, err := queryDatasourceExpressions(nil, QueryDatasourceExpressionsRequest{
			From:    "now-1h",
			Queries: []map[string]any{{"refId": "A"}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "from and to")
	})

	t.Run("empty queries returns error", func(t *testing.T) {
		_, err := queryDatasourceExpressions(nil, QueryDatasourceExpressionsRequest{
			From:    "now-1h",
			To:      "now",
			Queries: []map[string]any{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "queries is required")
	})

	t.Run("nil queries returns error", func(t *testing.T) {
		_, err := queryDatasourceExpressions(nil, QueryDatasourceExpressionsRequest{
			From: "now-1h",
			To:   "now",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "queries is required")
	})
}

// ─── get_firing_alerts ────────────────────────────────────────────────────────

func TestParseFiringAlertsResponse(t *testing.T) {
	raw := `[
		{
			"labels": {"alertname": "HighCPU", "severity": "critical"},
			"annotations": {"summary": "CPU is high"},
			"startsAt": "2024-01-01T12:00:00Z",
			"endsAt": "0001-01-01T00:00:00Z",
			"fingerprint": "abc123",
			"status": {
				"state": "active",
				"silencedBy": [],
				"inhibitedBy": ["rule-1"]
			}
		},
		{
			"labels": {"alertname": "LowMemory"},
			"annotations": {},
			"startsAt": "2024-01-01T11:00:00Z",
			"fingerprint": "def456",
			"status": {"state": "suppressed", "silencedBy": ["silence-1"], "inhibitedBy": []}
		}
	]`

	var alerts []FiringAlert
	require.NoError(t, json.Unmarshal([]byte(raw), &alerts))
	require.Len(t, alerts, 2)

	assert.Equal(t, "HighCPU", alerts[0].Labels["alertname"])
	assert.Equal(t, "critical", alerts[0].Labels["severity"])
	assert.Equal(t, "abc123", alerts[0].Fingerprint)
	require.NotNil(t, alerts[0].Status)
	assert.Equal(t, "active", alerts[0].Status.State)
	assert.Empty(t, alerts[0].Status.SilencedBy)
	assert.Equal(t, []string{"rule-1"}, alerts[0].Status.InhibitedBy)

	assert.Equal(t, "LowMemory", alerts[1].Labels["alertname"])
	assert.Equal(t, "def456", alerts[1].Fingerprint)
	require.NotNil(t, alerts[1].Status)
	assert.Equal(t, "suppressed", alerts[1].Status.State)
	assert.Equal(t, []string{"silence-1"}, alerts[1].Status.SilencedBy)
}

func TestParseFiringAlertsEmptyResponse(t *testing.T) {
	var alerts []FiringAlert
	require.NoError(t, json.Unmarshal([]byte("[]"), &alerts))
	assert.Empty(t, alerts)
}

// ─── get_alert_rules_with_state ───────────────────────────────────────────────

func TestParsePrometheusRulesResponse(t *testing.T) {
	raw := `{
		"status": "success",
		"data": {
			"groups": [
				{
					"name": "group-a",
					"file": "my-folder",
					"interval": 60,
					"rules": [
						{
							"name": "HighCPU",
							"state": "firing",
							"health": "ok",
							"lastEvaluation": "2024-01-01T12:00:00Z",
							"evaluationTime": 0.123,
							"type": "alerting",
							"labels": {"severity": "critical"},
							"annotations": {"summary": "CPU too high"},
							"alerts": [
								{
									"labels": {"severity": "critical"},
									"state": "firing",
									"activeAt": "2024-01-01T11:00:00Z",
									"value": "1"
								}
							]
						},
						{
							"name": "LowDisk",
							"state": "inactive",
							"health": "ok",
							"type": "alerting",
							"labels": {},
							"annotations": {},
							"alerts": []
						}
					]
				},
				{
					"name": "group-b",
					"file": "other-folder",
					"interval": 120,
					"rules": [
						{
							"name": "NetworkDown",
							"state": "pending",
							"health": "ok",
							"type": "alerting",
							"labels": {},
							"annotations": {},
							"alerts": []
						}
					]
				}
			]
		}
	}`

	var resp prometheusRulesResponse
	require.NoError(t, json.Unmarshal([]byte(raw), &resp))
	assert.Equal(t, "success", resp.Status)
	require.Len(t, resp.Data.Groups, 2)

	g0 := resp.Data.Groups[0]
	assert.Equal(t, "group-a", g0.Name)
	assert.Equal(t, "my-folder", g0.File)
	require.Len(t, g0.Rules, 2)
	assert.Equal(t, "HighCPU", g0.Rules[0].Name)
	assert.Equal(t, "firing", g0.Rules[0].State)
	assert.InDelta(t, 0.123, g0.Rules[0].EvaluationTime, 1e-9)
	require.Len(t, g0.Rules[0].Alerts, 1)
	assert.Equal(t, "firing", g0.Rules[0].Alerts[0].State)
	assert.Equal(t, "1", g0.Rules[0].Alerts[0].Value)
	assert.Equal(t, "inactive", g0.Rules[1].State)

	g1 := resp.Data.Groups[1]
	assert.Equal(t, "other-folder", g1.File)
	assert.Equal(t, "pending", g1.Rules[0].State)
}

func TestGetAlertRulesWithStateNoConfig(t *testing.T) {
	// Confirm the function reaches the HTTP call and fails with a config error
	// (no URL configured) rather than panicking. Filtering logic is validated
	// via TestParsePrometheusRulesResponse above.
	_, err := getAlertRulesWithState(context.Background(), GetAlertRulesWithStateRequest{})
	require.Error(t, err)
}
