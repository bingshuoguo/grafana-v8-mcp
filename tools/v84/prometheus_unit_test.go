//go:build unit

package v84

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePromTimeAt(t *testing.T) {
	now := time.Date(2026, 2, 27, 10, 0, 0, 0, time.UTC)

	t.Run("supports now and relative time", func(t *testing.T) {
		got, err := parsePromTimeAt("now", now)
		require.NoError(t, err)
		assert.Equal(t, now, got)

		got, err = parsePromTimeAt("now-1h30m", now)
		require.NoError(t, err)
		assert.Equal(t, now.Add(-90*time.Minute), got)

		got, err = parsePromTimeAt("now+2h", now)
		require.NoError(t, err)
		assert.Equal(t, now.Add(2*time.Hour), got)
	})

	t.Run("supports RFC3339", func(t *testing.T) {
		got, err := parsePromTimeAt("2026-02-26T13:21:47Z", now)
		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 2, 26, 13, 21, 47, 0, time.UTC), got)
	})

	t.Run("returns error for unsupported format", func(t *testing.T) {
		_, err := parsePromTimeAt("yesterday", now)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported time format")
	})
}

func TestDefaultPromStep(t *testing.T) {
	start := time.Date(2026, 2, 27, 0, 0, 0, 0, time.UTC)

	assert.Equal(t, 30, defaultPromStep(start, start.Add(10*time.Minute)))
	assert.Equal(t, 60, defaultPromStep(start, start.Add(2*time.Hour)))
	assert.Equal(t, 300, defaultPromStep(start, start.Add(12*time.Hour)))
	assert.Equal(t, 3600, defaultPromStep(start, start.Add(48*time.Hour)))
}

func TestConvertPromSample(t *testing.T) {
	sample, err := convertPromSample([]any{float64(1700000000.5), "0.95"})
	require.NoError(t, err)
	assert.Equal(t, "2023-11-14T22:13:20.5Z", sample.Time)
	assert.Equal(t, "0.95", sample.Value)

	_, err = convertPromSample([]any{float64(1)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly 2 fields")
}

func TestQueryPrometheus(t *testing.T) {
	t.Run("range query with datasource ref", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/datasources":
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"id": 7, "uid": "prom-main", "name": "Prom Main", "type": "prometheus", "isDefault": true},
				})
			case "/api/datasources/7":
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":   7,
					"uid":  "prom-main",
					"name": "Prom Main",
					"type": "prometheus",
				})
			case "/api/datasources/proxy/7/api/v1/query_range":
				assert.Equal(t, "sum(rate(http_requests_total[5m]))", r.URL.Query().Get("query"))
				assert.Equal(t, "30", r.URL.Query().Get("step"))
				assert.Equal(t, "1709164800", r.URL.Query().Get("start"))
				assert.Equal(t, "1709168400", r.URL.Query().Get("end"))

				_, _ = w.Write([]byte(`{
					"status":"success",
					"data":{
						"resultType":"matrix",
						"result":[
							{
								"metric":{"job":"api"},
								"values":[[1709193600,"1.0"],[1709193630,"1.2"]]
							}
						]
					}
				}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		got, err := queryPrometheus(newV84TestContext(server), QueryPrometheusRequest{
			Datasource: &DatasourceRef{UID: "prom-main"},
			Expr:       "sum(rate(http_requests_total[5m]))",
			Start:      "2024-02-29T00:00:00Z",
			End:        "2024-02-29T01:00:00Z",
			Step:       "30s",
			QueryType:  "range",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "matrix", got.ResultType)
		require.Len(t, got.Result, 1)
		assert.Equal(t, "api", got.Result[0].Metric["job"])
		require.Len(t, got.Result[0].Values, 2)
		assert.Equal(t, "1.0", got.Result[0].Values[0].Value)
		assert.Empty(t, got.Hints)
	})

	t.Run("instant query uses default datasource and emits hints for empty result", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/datasources":
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"id": 5, "uid": "prom-default", "name": "Prom Default", "type": "prometheus", "isDefault": true},
					{"id": 6, "uid": "loki-main", "name": "Loki Main", "type": "loki"},
				})
			case "/api/datasources/proxy/5/api/v1/query":
				assert.Equal(t, "up", r.URL.Query().Get("query"))
				assert.NotEmpty(t, r.URL.Query().Get("time"))
				_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		got, err := queryPrometheus(newV84TestContext(server), QueryPrometheusRequest{
			Expr:      "up",
			Start:     "now-5m",
			QueryType: "instant",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "vector", got.ResultType)
		assert.Len(t, got.Result, 0)
		assert.NotEmpty(t, got.Hints)
	})

	t.Run("rejects non-prometheus datasource", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/datasources":
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"id": 9, "uid": "loki-main", "name": "Loki Main", "type": "loki"},
				})
			case "/api/datasources/9":
				_ = json.NewEncoder(w).Encode(map[string]any{
					"id":   9,
					"uid":  "loki-main",
					"name": "Loki Main",
					"type": "loki",
				})
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		_, err := queryPrometheus(newV84TestContext(server), QueryPrometheusRequest{
			Datasource: &DatasourceRef{UID: "loki-main"},
			Expr:       "up",
			Start:      "now-5m",
			QueryType:  "instant",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected prometheus")
	})

	t.Run("supports scalar result type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/datasources":
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"id": 5, "uid": "prom-default", "name": "Prom Default", "type": "prometheus", "isDefault": true},
				})
			case "/api/datasources/proxy/5/api/v1/query":
				_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"scalar","result":[1709193600.5,"42"]}}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		got, err := queryPrometheus(newV84TestContext(server), QueryPrometheusRequest{
			Expr:      "scalar(up)",
			Start:     "now-5m",
			QueryType: "instant",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "scalar", got.ResultType)
		require.Len(t, got.Result, 1)
		assert.Empty(t, got.Result[0].Metric)
		require.NotNil(t, got.Result[0].Value)
		assert.Equal(t, "42", got.Result[0].Value.Value)
		assert.Empty(t, got.Hints)
	})

	t.Run("supports string result type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/datasources":
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"id": 5, "uid": "prom-default", "name": "Prom Default", "type": "prometheus", "isDefault": true},
				})
			case "/api/datasources/proxy/5/api/v1/query":
				_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"string","result":[1709193600,"stale"]}}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		got, err := queryPrometheus(newV84TestContext(server), QueryPrometheusRequest{
			Expr:      `label_replace("stale","x","y","z","w")`,
			Start:     "now-5m",
			QueryType: "instant",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "string", got.ResultType)
		require.Len(t, got.Result, 1)
		assert.Empty(t, got.Result[0].Metric)
		require.NotNil(t, got.Result[0].Value)
		assert.Equal(t, "stale", got.Result[0].Value.Value)
		assert.Empty(t, got.Hints)
	})

	t.Run("returns prometheus error with errorType", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/datasources":
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"id": 5, "uid": "prom-default", "name": "Prom Default", "type": "prometheus", "isDefault": true},
				})
			case "/api/datasources/proxy/5/api/v1/query_range":
				_, _ = w.Write([]byte(`{"status":"error","errorType":"bad_data","error":"invalid parameter \"query\": parse error"}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		_, err := queryPrometheus(newV84TestContext(server), QueryPrometheusRequest{
			Expr:      "sum(rate(http_requests_total[5m]))",
			Start:     "2024-02-29T00:00:00Z",
			End:       "2024-02-29T01:00:00Z",
			QueryType: "range",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid parameter")
		assert.Contains(t, err.Error(), "bad_data")
	})
}

func TestListPrometheusLabelValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/datasources":
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{"id": 11, "uid": "prom-default", "name": "Prom Default", "type": "prometheus", "isDefault": true},
			})
		case "/api/datasources/proxy/11/api/v1/label/instance/values":
			assert.NotEmpty(t, r.URL.Query().Get("start"))
			assert.NotEmpty(t, r.URL.Query().Get("end"))
			_, _ = w.Write([]byte(`{"status":"success","data":["i-1","i-2","i-3"]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	got, err := listPrometheusLabelValues(newV84TestContext(server), ListPrometheusLabelValuesRequest{
		LabelName: "instance",
		Start:     "2024-02-29T00:00:00Z",
		End:       "2024-02-29T01:00:00Z",
		Limit:     2,
	})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "instance", got.LabelName)
	assert.Equal(t, []string{"i-1", "i-2"}, got.Values)
	assert.Equal(t, 3, got.Total)
	assert.True(t, got.Truncated)
}

func TestListPrometheusMetricNames(t *testing.T) {
	t.Run("regex and pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/datasources":
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"id": 12, "uid": "prom-default", "name": "Prom Default", "type": "prometheus", "isDefault": true},
				})
			case "/api/datasources/proxy/12/api/v1/label/__name__/values":
				_, _ = w.Write([]byte(`{"status":"success","data":["up","process_cpu_seconds_total","http_requests_total"]}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		got, err := listPrometheusMetricNames(newV84TestContext(server), ListPrometheusMetricNamesRequest{
			Regex: "^http_.*|^up$",
			Limit: 1,
			Page:  2,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, []string{"up"}, got.Metrics)
		assert.Equal(t, 2, got.Total)
		assert.Equal(t, 2, got.Page)
		assert.False(t, got.HasMore)
	})

	t.Run("invalid regex returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/api/datasources":
				_ = json.NewEncoder(w).Encode([]map[string]any{
					{"id": 12, "uid": "prom-default", "name": "Prom Default", "type": "prometheus", "isDefault": true},
				})
			case "/api/datasources/proxy/12/api/v1/label/__name__/values":
				_, _ = w.Write([]byte(`{"status":"success","data":["up"]}`))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		_, err := listPrometheusMetricNames(newV84TestContext(server), ListPrometheusMetricNamesRequest{
			Regex: "(",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex")
	})
}
