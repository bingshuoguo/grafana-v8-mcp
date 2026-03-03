package v84

import "github.com/mark3labs/mcp-go/server"

// AddV84Tools registers Grafana 8.4.7 tool contracts.
func AddV84Tools(m *server.MCPServer, enableWriteTools, enableOptionalTools bool) {
	registerV84CoreReadTools(m)
	registerV84CoreReadOnlyExtensions(m)
	registerV84P1DataTools(m)

	if enableWriteTools {
		registerV84CoreWriteTools(m)
	}

	if enableOptionalTools {
		registerV84P2OptionalTools(m, enableWriteTools)
	}
}

func registerV84CoreReadTools(m *server.MCPServer) {
	GetHealthTool.Register(m)
	GetCurrentUserTool.Register(m)
	GetCurrentOrgTool.Register(m)
	SearchDashboardsTool.Register(m)
	GetDashboardByUIDTool.Register(m)
	ListFoldersTool.Register(m)
	ListDatasourcesTool.Register(m)
	GetDatasourceTool.Register(m)
	ResolveDatasourceRefTool.Register(m)
	QueryDatasourceTool.Register(m)
	QueryDatasourceExpressionsTool.Register(m)
	QueryPrometheusTool.Register(m)
	ListPrometheusLabelValuesTool.Register(m)
	ListPrometheusMetricNamesTool.Register(m)
	GetAnnotationsTool.Register(m)
	ListLegacyAlertsTool.Register(m)
	ListLegacyNotificationChannelsTool.Register(m)
	ListOrgUsersTool.Register(m)
	ListTeamsTool.Register(m)
}

// registerV84CoreReadOnlyExtensions registers P0 read-only extension tools.
func registerV84CoreReadOnlyExtensions(m *server.MCPServer) {
	// Compat aliases
	GetDatasourceByUIDTool.Register(m)
	GetDatasourceByNameTool.Register(m)
	ListUsersByOrgTool.Register(m)

	// Folder search
	SearchFoldersTool.Register(m)

	// Dashboard helpers
	GetDashboardPanelQueriesTool.Register(m)
	GetDashboardPropertyTool.Register(m)
	GetDashboardSummaryTool.Register(m)
	GetDashboardVersionsTool.Register(m)

	// Annotation extras (read)
	GetAnnotationTagsTool.Register(m)

	// Navigation & examples
	GenerateDeeplinkTool.Register(m)
	GetQueryExamplesTool.Register(m)
}

// registerV84P1DataTools registers P1 datasource-specific tools.
func registerV84P1DataTools(m *server.MCPServer) {
	// Prometheus extras
	ListPrometheusLabelNamesTool.Register(m)
	ListPrometheusMetricMetadataTool.Register(m)
	QueryPrometheusHistogramTool.Register(m)

	// Loki (5 tools, ID-based proxy)
	ListLokiLabelNamesTool.Register(m)
	ListLokiLabelValuesTool.Register(m)
	QueryLokiLogsTool.Register(m)
	QueryLokiStatsTool.Register(m)
	QueryLokiPatternsTool.Register(m)

	// ClickHouse (3 tools)
	QueryClickHouseTool.Register(m)
	ListClickHouseTablesTool.Register(m)
	DescribeClickHouseTableTool.Register(m)

	// Cross-datasource log search
	SearchLogsTool.Register(m)
}

// registerV84P2OptionalTools registers optional P2 tools gated by enableOptionalTools.
// These require Unified Alerting (Ruler API) or image-renderer to be available.
func registerV84P2OptionalTools(m *server.MCPServer, enableWriteTools bool) {
	// Unified Alerting read tools
	ListAlertRulesTool.Register(m)
	GetAlertRuleByUIDTool.Register(m)
	ListContactPointsTool.Register(m)
	GetFiringAlertsTool.Register(m)
	GetAlertRulesWithStateTool.Register(m)

	// Unified Alerting write tools
	if enableWriteTools {
		CreateAlertRuleTool.Register(m)
		UpdateAlertRuleTool.Register(m)
		DeleteAlertRuleTool.Register(m)
	}

	// Rendering
	GetPanelImageTool.Register(m)
}

func registerV84CoreWriteTools(m *server.MCPServer) {
	UpsertDashboardTool.Register(m)
	UpdateDashboardTool.Register(m) // compat alias
	CreateFolderTool.Register(m)
	UpdateFolderTool.Register(m)
	CreateAnnotationTool.Register(m)
	PatchAnnotationTool.Register(m)
	UpdateAnnotationTool.Register(m)
	CreateGraphiteAnnotationTool.Register(m)
	DeleteAnnotationTool.Register(m)
}
