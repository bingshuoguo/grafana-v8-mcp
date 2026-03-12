//go:build unit

package tools

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── capability ───────────────────────────────────────────────────────────────

func TestCapabilityProbesExist(t *testing.T) {
	// Just ensure the functions are callable (signature check).
	// Real probing requires a live Grafana instance.
	assert.NotNil(t, ProbeUnifiedAlerting)
	assert.NotNil(t, ProbeImageRenderer)
}

// ─── alerting_unified ─────────────────────────────────────────────────────────

func TestAlertingToolsRegistered(t *testing.T) {
	assert.NotNil(t, ListAlertRulesTool)
	assert.NotNil(t, GetAlertRuleByUIDTool)
	assert.NotNil(t, CreateAlertRuleTool)
	assert.NotNil(t, UpdateAlertRuleTool)
	assert.NotNil(t, DeleteAlertRuleTool)
	assert.NotNil(t, ListContactPointsTool)
}

func TestFlattenRulerRules(t *testing.T) {
	resp := rulerRulesResponse{
		"my-folder": {
			{
				Name:     "group-a",
				Interval: "1m",
				Rules: []RulerRule{
					{UID: "uid-1", Title: "Alert One", For: "5m"},
					{UID: "uid-2", Title: "Alert Two"},
				},
			},
		},
		"other-folder": {
			{
				Name:  "group-b",
				Rules: []RulerRule{{UID: "uid-3", Title: "Alert Three"}},
			},
		},
	}

	rules := flattenRulerRules(resp)
	assert.Len(t, rules, 3)

	// Build a map by UID for order-independent assertion.
	byUID := make(map[string]AlertRuleSummary, len(rules))
	for _, r := range rules {
		byUID[r.UID] = r
	}

	require.Contains(t, byUID, "uid-1")
	assert.Equal(t, "Alert One", byUID["uid-1"].Title)
	assert.Equal(t, "my-folder", byUID["uid-1"].Namespace)
	assert.Equal(t, "group-a", byUID["uid-1"].Group)
	assert.Equal(t, "5m", byUID["uid-1"].For)

	require.Contains(t, byUID, "uid-3")
	assert.Equal(t, "other-folder", byUID["uid-3"].Namespace)
	assert.Equal(t, "group-b", byUID["uid-3"].Group)
}

func TestFlattenRulerRulesEmpty(t *testing.T) {
	rules := flattenRulerRules(rulerRulesResponse{})
	assert.Nil(t, rules) // nil is fine; listAlertRules coerces to []
}

func TestCreateAlertRuleRequestValidation(t *testing.T) {
	t.Run("missing namespace returns error", func(t *testing.T) {
		_, err := createAlertRule(nil, CreateAlertRuleRequest{
			Namespace: "",
			Group:     RulerRuleGroup{Name: "g"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "namespace")
	})

	t.Run("missing group name returns error", func(t *testing.T) {
		_, err := createAlertRule(nil, CreateAlertRuleRequest{
			Namespace: "ns",
			Group:     RulerRuleGroup{Name: ""},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "group.name")
	})
}

func TestGetAlertRuleByUIDValidation(t *testing.T) {
	_, err := getAlertRuleByUID(nil, GetAlertRuleByUIDRequest{UID: "  "})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "uid is required")
}

func TestDeleteAlertRuleValidation(t *testing.T) {
	t.Run("missing namespace", func(t *testing.T) {
		_, err := deleteAlertRule(nil, DeleteAlertRuleRequest{Namespace: "", Group: "g"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "namespace")
	})

	t.Run("missing group", func(t *testing.T) {
		_, err := deleteAlertRule(nil, DeleteAlertRuleRequest{Namespace: "ns", Group: ""})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "group")
	})
}

func TestParseContactPoints(t *testing.T) {
	raw := `[
		{
			"name": "email",
			"grafana_managed_receiver_configs": [
				{"name": "email-cfg", "type": "email", "settings": {"addresses": "a@b.com"}}
			]
		},
		{
			"name": "slack"
		}
	]`
	var receivers []alertmanagerReceiver
	require.NoError(t, json.Unmarshal([]byte(raw), &receivers))
	assert.Len(t, receivers, 2)
	assert.Equal(t, "email", receivers[0].Name)
	require.Len(t, receivers[0].Configs, 1)
	assert.Equal(t, "email", receivers[0].Configs[0].Type)
	assert.Equal(t, "slack", receivers[1].Name)
	assert.Empty(t, receivers[1].Configs)
}

// ─── rendering ────────────────────────────────────────────────────────────────

func TestRenderingToolRegistered(t *testing.T) {
	assert.NotNil(t, GetPanelImageTool)
}

func TestBuildV84RenderURL(t *testing.T) {
	panelID := 42
	width := 800
	height := 400
	timeout := 30

	url := buildV84RenderURL("http://grafana:3000", GetPanelImageRequest{
		DashboardUID: "abc123",
		PanelID:      &panelID,
		Width:        &width,
		Height:       &height,
		From:         "now-1h",
		To:           "now",
		Theme:        "light",
		Variables:    map[string]string{"var-env": "prod"},
		Timeout:      &timeout,
	})

	assert.Contains(t, url, "http://grafana:3000/render/d/abc123")
	assert.Contains(t, url, "viewPanel=42")
	assert.Contains(t, url, "width=800")
	assert.Contains(t, url, "height=400")
	assert.Contains(t, url, "from=now-1h")
	assert.Contains(t, url, "to=now")
	assert.Contains(t, url, "theme=light")
	assert.Contains(t, url, "var-env=prod")
	assert.Contains(t, url, "kiosk=true")
	// No /api/ prefix
	assert.NotContains(t, url, "/api/render")
}

func TestBuildV84RenderURLDefaults(t *testing.T) {
	url := buildV84RenderURL("http://grafana:3000/", GetPanelImageRequest{
		DashboardUID: "uid1",
	})

	assert.Contains(t, url, "http://grafana:3000/render/d/uid1")
	assert.Contains(t, url, "width=1000")
	assert.Contains(t, url, "height=500")
	// No panelId without PanelID
	assert.NotContains(t, url, "viewPanel")
	// Trailing slash stripped from base
	assert.NotContains(t, url, "//render")
}

func TestBuildV84RenderURLNoPanelID(t *testing.T) {
	url := buildV84RenderURL("http://g:3000", GetPanelImageRequest{DashboardUID: "d1"})
	assert.NotContains(t, url, "viewPanel")
	assert.Contains(t, url, "/render/d/d1")
}
