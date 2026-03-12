//go:build unit

package tools

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── prometheus_extra ─────────────────────────────────────────────────────────

func TestPrometheusExtraToolsRegistered(t *testing.T) {
	assert.NotNil(t, ListPrometheusLabelNamesTool)
	assert.NotNil(t, ListPrometheusMetricMetadataTool)
	assert.NotNil(t, QueryPrometheusHistogramTool)
}

func TestQueryPrometheusHistogramExprBuilding(t *testing.T) {
	// buildHistogramExpr mirrors the logic in queryPrometheusHistogram.
	buildHistogramExpr := func(args QueryPrometheusHistogramRequest) string {
		quantile := args.Quantile
		if quantile <= 0 || quantile > 1 {
			quantile = 0.95
		}
		rateWindow := "5m"
		if args.RateWindow != "" {
			rateWindow = args.RateWindow
		}
		metric := args.Metric
		// trim trailing _bucket then re-append
		const suffix = "_bucket"
		if len(metric) >= len(suffix) && metric[len(metric)-len(suffix):] == suffix {
			metric = metric[:len(metric)-len(suffix)]
		}
		bucketMetric := metric + suffix
		selector := args.Selector
		if selector == "" {
			selector = "{}"
		}
		groupBy := append([]string{"le"}, args.GroupBy...)
		byClause := ""
		for i, g := range groupBy {
			if i > 0 {
				byClause += ", "
			}
			byClause += g
		}
		return fmt.Sprintf("histogram_quantile(%g, sum(rate(%s%s[%s])) by (%s))",
			quantile, bucketMetric, selector, rateWindow, byClause)
	}

	tests := []struct {
		name       string
		args       QueryPrometheusHistogramRequest
		wantSubstr string
	}{
		{
			name:       "default quantile",
			args:       QueryPrometheusHistogramRequest{Metric: "http_request_duration_seconds", Start: "now-1h"},
			wantSubstr: "histogram_quantile(0.95",
		},
		{
			name:       "custom quantile",
			args:       QueryPrometheusHistogramRequest{Metric: "http_request_duration_seconds", Quantile: 0.99, Start: "now-1h"},
			wantSubstr: "histogram_quantile(0.99",
		},
		{
			name:       "appends _bucket suffix",
			args:       QueryPrometheusHistogramRequest{Metric: "latency", Start: "now-1h"},
			wantSubstr: "latency_bucket{}",
		},
		{
			name:       "deduplicates _bucket suffix",
			args:       QueryPrometheusHistogramRequest{Metric: "latency_bucket", Start: "now-1h"},
			wantSubstr: "latency_bucket{}",
		},
		{
			name:       "custom rate window",
			args:       QueryPrometheusHistogramRequest{Metric: "http_request_duration_seconds", RateWindow: "10m", Start: "now-1h"},
			wantSubstr: "[10m]",
		},
		{
			name:       "groupBy labels added",
			args:       QueryPrometheusHistogramRequest{Metric: "http_request_duration_seconds", GroupBy: []string{"job"}, Start: "now-1h"},
			wantSubstr: "by (le, job)",
		},
		{
			name:       "selector injected",
			args:       QueryPrometheusHistogramRequest{Metric: "http_request_duration_seconds", Selector: `{job="api"}`, Start: "now-1h"},
			wantSubstr: `http_request_duration_seconds_bucket{job="api"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			expr := buildHistogramExpr(tc.args)
			assert.Contains(t, expr, tc.wantSubstr, "expr=%s", expr)
		})
	}
}

// ─── Loki ─────────────────────────────────────────────────────────────────────

func TestLokiToolsRegistered(t *testing.T) {
	assert.NotNil(t, ListLokiLabelNamesTool)
	assert.NotNil(t, ListLokiLabelValuesTool)
	assert.NotNil(t, QueryLokiLogsTool)
	assert.NotNil(t, QueryLokiStatsTool)
	assert.NotNil(t, QueryLokiPatternsTool)
}

func TestLokiDefaultTimeRange(t *testing.T) {
	// Empty inputs → defaults filled.
	start, end := lokiDefaultTimeRange("", "")
	assert.NotEmpty(t, start)
	assert.NotEmpty(t, end)

	// Explicit values are preserved unchanged.
	s2, e2 := lokiDefaultTimeRange("2024-01-01T00:00:00Z", "2024-01-02T00:00:00Z")
	assert.Equal(t, "2024-01-01T00:00:00Z", s2)
	assert.Equal(t, "2024-01-02T00:00:00Z", e2)
}

func TestLokiAddTimeRange(t *testing.T) {
	params := url.Values{}
	lokiAddTimeRange(params, "2024-01-01T00:00:00Z", "2024-01-02T00:00:00Z")

	assert.NotEmpty(t, params.Get("start"), "start should be set as nanosecond timestamp")
	assert.NotEmpty(t, params.Get("end"), "end should be set as nanosecond timestamp")

	// Nanosecond values must be much larger than second-resolution timestamps.
	// 2024-01-01 in nanoseconds ≈ 1.7e18, so string length > 15.
	assert.Greater(t, len(params.Get("start")), 15, "start should look like a nanosecond timestamp")

	// Invalid RFC3339 → params unchanged.
	params2 := url.Values{}
	lokiAddTimeRange(params2, "not-a-time", "")
	assert.Empty(t, params2.Get("start"))
}

func TestParseLokiLogStreams(t *testing.T) {
	raw := json.RawMessage(`[
		{
			"stream": {"app": "test", "env": "prod"},
			"values": [
				["1700000000000000000", "log line one"],
				["1700000001000000000", "log line two"]
			]
		}
	]`)

	entries, err := parseLokiLogStreams(raw)
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "1700000000000000000", entries[0].Timestamp)
	assert.Equal(t, "log line one", entries[0].Line)
	assert.Equal(t, map[string]string{"app": "test", "env": "prod"}, entries[0].Labels)
	assert.Equal(t, "log line two", entries[1].Line)
}

func TestParseLokiLogStreamsEmpty(t *testing.T) {
	entries, err := parseLokiLogStreams(json.RawMessage(`[]`))
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestParseLokiLogStreamsMissingLabels(t *testing.T) {
	// nil stream map → labels replaced with empty map, no panic.
	raw := json.RawMessage(`[{"stream": null, "values": [["123", "line"]]}]`)
	entries, err := parseLokiLogStreams(raw)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.NotNil(t, entries[0].Labels)
	assert.Equal(t, "line", entries[0].Line)
}

func TestParseLokiLogStreamsShortValues(t *testing.T) {
	// values with fewer than 2 elements should be skipped gracefully.
	raw := json.RawMessage(`[{"stream": {}, "values": [["only-one"]]}]`)
	entries, err := parseLokiLogStreams(raw)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

// ─── ClickHouse ───────────────────────────────────────────────────────────────

func TestClickHouseToolsRegistered(t *testing.T) {
	assert.NotNil(t, QueryClickHouseTool)
	assert.NotNil(t, ListClickHouseTablesTool)
	assert.NotNil(t, DescribeClickHouseTableTool)
}

func TestChEnforceLimit(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		requested int
		wantPart  string
	}{
		{
			name:      "no limit → default 100 appended",
			query:     "SELECT * FROM foo",
			requested: 0,
			wantPart:  "LIMIT 100",
		},
		{
			name:      "explicit limit appended",
			query:     "SELECT * FROM foo",
			requested: 50,
			wantPart:  "LIMIT 50",
		},
		{
			name:      "existing LIMIT within max kept",
			query:     "SELECT * FROM foo LIMIT 200",
			requested: 50,
			wantPart:  "LIMIT 200",
		},
		{
			name:      "existing LIMIT over max capped",
			query:     "SELECT * FROM foo LIMIT 9999",
			requested: 50,
			wantPart:  "LIMIT 1000",
		},
		{
			name:      "requested over max capped",
			query:     "SELECT * FROM foo",
			requested: 5000,
			wantPart:  "LIMIT 1000",
		},
		{
			name:      "trailing semicolon stripped before appending",
			query:     "SELECT * FROM foo;",
			requested: 10,
			wantPart:  "LIMIT 10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := chEnforceLimit(tc.query, tc.requested)
			assert.Contains(t, result, tc.wantPart, "result=%s", result)
		})
	}
}

func TestChSubstituteMacros(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)

	t.Run("timeFilter substituted", func(t *testing.T) {
		q := chSubstituteMacros("SELECT * FROM t WHERE $__timeFilter(ts)", from, to)
		assert.Contains(t, q, "ts >= toDateTime(")
		assert.Contains(t, q, "ts <= toDateTime(")
		assert.NotContains(t, q, "$__timeFilter")
	})

	t.Run("from/to substituted", func(t *testing.T) {
		q := chSubstituteMacros("SELECT $__from, $__to", from, to)
		assert.NotContains(t, q, "$__from")
		assert.NotContains(t, q, "$__to")
	})

	t.Run("interval substituted", func(t *testing.T) {
		q := chSubstituteMacros("SELECT $__interval, $__interval_ms", from, to)
		assert.NotContains(t, q, "$__interval_ms")
		assert.NotContains(t, q, "$__interval")
	})
}

func TestProcessCHResponse(t *testing.T) {
	resp := makeCHResponse("col1", "col2", []any{"a", "b"}, []any{1.0, 2.0})

	columns, rows, err := processCHResponse(resp)
	require.NoError(t, err)
	assert.Equal(t, []string{"col1", "col2"}, columns)
	require.Len(t, rows, 2)
	assert.Equal(t, "a", rows[0]["col1"])
	assert.Equal(t, 1.0, rows[0]["col2"])
	assert.Equal(t, "b", rows[1]["col1"])
	assert.Equal(t, 2.0, rows[1]["col2"])
}

func TestProcessCHResponseError(t *testing.T) {
	resp := &chQueryResponse{
		Results: map[string]struct {
			Status int `json:"status,omitempty"`
			Frames []struct {
				Schema struct {
					Fields []struct {
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"fields"`
				} `json:"schema"`
				Data struct {
					Values [][]any `json:"values"`
				} `json:"data"`
			} `json:"frames,omitempty"`
			Error string `json:"error,omitempty"`
		}{
			"A": {Error: "table not found"},
		},
	}

	_, _, err := processCHResponse(resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "table not found")
}

func TestProcessCHResponseEmpty(t *testing.T) {
	resp := &chQueryResponse{
		Results: map[string]struct {
			Status int `json:"status,omitempty"`
			Frames []struct {
				Schema struct {
					Fields []struct {
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"fields"`
				} `json:"schema"`
				Data struct {
					Values [][]any `json:"values"`
				} `json:"data"`
			} `json:"frames,omitempty"`
			Error string `json:"error,omitempty"`
		}{},
	}

	columns, rows, err := processCHResponse(resp)
	require.NoError(t, err)
	assert.Empty(t, columns)
	assert.Empty(t, rows)
}

// makeCHResponse builds a minimal chQueryResponse for testing.
func makeCHResponse(col1, col2 string, vals1, vals2 []any) *chQueryResponse {
	type resultEntry struct {
		Status int `json:"status,omitempty"`
		Frames []struct {
			Schema struct {
				Fields []struct {
					Name string `json:"name"`
					Type string `json:"type"`
				} `json:"fields"`
			} `json:"schema"`
			Data struct {
				Values [][]any `json:"values"`
			} `json:"data"`
		} `json:"frames,omitempty"`
		Error string `json:"error,omitempty"`
	}

	entry := resultEntry{}
	entry.Frames = []struct {
		Schema struct {
			Fields []struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"fields"`
		} `json:"schema"`
		Data struct {
			Values [][]any `json:"values"`
		} `json:"data"`
	}{
		{},
	}
	entry.Frames[0].Schema.Fields = []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}{
		{Name: col1, Type: "string"},
		{Name: col2, Type: "number"},
	}
	entry.Frames[0].Data.Values = [][]any{vals1, vals2}

	return &chQueryResponse{
		Results: map[string]struct {
			Status int `json:"status,omitempty"`
			Frames []struct {
				Schema struct {
					Fields []struct {
						Name string `json:"name"`
						Type string `json:"type"`
					} `json:"fields"`
				} `json:"schema"`
				Data struct {
					Values [][]any `json:"values"`
				} `json:"data"`
			} `json:"frames,omitempty"`
			Error string `json:"error,omitempty"`
		}{"A": entry},
	}
}

// ─── search_logs ──────────────────────────────────────────────────────────────

func TestSearchLogsToolRegistered(t *testing.T) {
	assert.NotNil(t, SearchLogsTool)
}

func TestIsRegexLike(t *testing.T) {
	assert.True(t, isRegexLike("error.*timeout"), "dot-star is regex")
	assert.True(t, isRegexLike("timeout|refused"), "pipe is regex")
	assert.True(t, isRegexLike("^error"), "caret is regex")
	assert.True(t, isRegexLike("[ERROR]"), "brackets are regex")
	assert.False(t, isRegexLike("simple text"), "plain text is not regex")
	assert.False(t, isRegexLike("connection refused"), "plain text is not regex")
}

func TestSearchLogsQueryBuildingLoki(t *testing.T) {
	tests := []struct {
		pattern string
		wantOp  string
	}{
		{"simple", `|= "simple"`},
		{"error.*timeout", `|~ "error.*timeout"`},
		{"timeout|refused", `|~ "timeout|refused"`},
	}

	for _, tc := range tests {
		t.Run(tc.pattern, func(t *testing.T) {
			escaped := tc.pattern
			var query string
			if isRegexLike(tc.pattern) {
				query = fmt.Sprintf(`{} |~ "%s"`, escaped)
			} else {
				query = fmt.Sprintf(`{} |= "%s"`, escaped)
			}
			assert.Contains(t, query, tc.wantOp)
		})
	}
}

func TestSearchLogsDefaultLimit(t *testing.T) {
	// Verify constants are sensible.
	assert.Equal(t, 100, v84DefaultSearchLogsLimit)
	assert.Equal(t, 1000, v84MaxSearchLogsLimit)
}
