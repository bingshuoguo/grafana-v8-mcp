package tools

import "context"

// ProbeUnifiedAlerting checks if the Grafana instance has Unified Alerting enabled.
// It probes the Ruler API; returns true if the endpoint exists (200 or 403),
// false if it responds with 404/405 (route missing → legacy-only alerting).
func ProbeUnifiedAlerting(ctx context.Context) bool {
	_, status, _ := doAPIRequest(ctx, "GET", "/ruler/grafana/api/v1/rules", nil, nil)
	return status != 0 && status != 404 && status != 405
}

// ProbeImageRenderer checks if the Grafana Image Renderer plugin is installed.
func ProbeImageRenderer(ctx context.Context) bool {
	_, status, _ := doAPIRequest(ctx, "GET", "/plugins/grafana-image-renderer/settings", nil, nil)
	return status == 200
}
