package v84

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	mcpgrafana "github.com/grafana/mcp-grafana"
)

type QueryDatasourceRequest struct {
	From       string           `json:"from" jsonschema:"required,description=From time\\, e.g. now-1h"`
	To         string           `json:"to" jsonschema:"required,description=To time\\, e.g. now"`
	Debug      *bool            `json:"debug,omitempty" jsonschema:"description=Enable query debug mode"`
	Datasource *DatasourceRef   `json:"datasource,omitempty" jsonschema:"description=Optional datasource reference"`
	Queries    []map[string]any `json:"queries" jsonschema:"required,description=Datasource query payload list"`
}

type QueryDatasourceResponse struct {
	Raw       json.RawMessage `json:"raw,omitempty"`
	Responses map[string]any  `json:"responses,omitempty"`
	Hints     []string        `json:"hints,omitempty"`
}

func queryDatasource(ctx context.Context, args QueryDatasourceRequest) (*QueryDatasourceResponse, error) {
	if args.From == "" || args.To == "" {
		return nil, fmt.Errorf("from and to are required")
	}
	if len(args.Queries) == 0 {
		return nil, fmt.Errorf("queries is required")
	}

	var resolvedID *int64
	if args.Datasource != nil {
		resolved, err := resolveDatasourceRef(ctx, *args.Datasource)
		if err != nil {
			return nil, fmt.Errorf("resolve datasource: %w", err)
		}
		resolvedID = &resolved.Datasource.ID
	}

	normalizedQueries := make([]map[string]any, 0, len(args.Queries))
	for _, q := range args.Queries {
		copyQ := make(map[string]any, len(q)+1)
		for k, v := range q {
			copyQ[k] = v
		}
		if resolvedID != nil {
			if _, exists := copyQ["datasourceId"]; !exists {
				copyQ["datasourceId"] = *resolvedID
			}
		}
		normalizedQueries = append(normalizedQueries, copyQ)
	}

	requestBody := map[string]any{
		"from":    args.From,
		"to":      args.To,
		"queries": normalizedQueries,
	}
	if args.Debug != nil {
		requestBody["debug"] = *args.Debug
	}

	respBody, statusCode, err := doAPIRequest(ctx, "POST", "/tsdb/query", nil, requestBody)
	if err != nil {
		return nil, fmt.Errorf("query datasource: %w", wrapRawAPIError(statusCode, respBody, err))
	}

	response := &QueryDatasourceResponse{Raw: json.RawMessage(respBody)}

	var parsed map[string]any
	if err := json.Unmarshal(respBody, &parsed); err == nil {
		if results, ok := parsed["results"].(map[string]any); ok {
			response.Responses = results
		} else {
			response.Responses = parsed
		}

		hints := make([]string, 0)
		if message, ok := parsed["message"].(string); ok && message != "" {
			hints = append(hints, message)
		}
		if len(hints) > 0 {
			response.Hints = hints
		}
	}

	return response, nil
}

var QueryDatasourceTool = mcpgrafana.MustTool(
	"query_datasource",
	"Query datasource metrics via Grafana /api/tsdb/query.",
	queryDatasource,
	mcp.WithTitleAnnotation("Query datasource"),
	mcp.WithReadOnlyHintAnnotation(true),
	mcp.WithIdempotentHintAnnotation(true),
)
