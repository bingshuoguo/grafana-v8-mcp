package v84

import "github.com/mark3labs/mcp-go/server"

// AddV84Tools registers Grafana 8.4.7 tool contracts.
func AddV84Tools(m *server.MCPServer, enableWriteTools, enableOptionalTools bool) {
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
	QueryPrometheusTool.Register(m)
	ListPrometheusLabelValuesTool.Register(m)
	ListPrometheusMetricNamesTool.Register(m)
	GetAnnotationsTool.Register(m)
	ListLegacyAlertsTool.Register(m)
	ListLegacyNotificationChannelsTool.Register(m)
	ListOrgUsersTool.Register(m)
	ListTeamsTool.Register(m)

	if enableWriteTools {
		UpsertDashboardTool.Register(m)
		CreateFolderTool.Register(m)
		UpdateFolderTool.Register(m)
		CreateAnnotationTool.Register(m)
		PatchAnnotationTool.Register(m)
	}

	_ = enableOptionalTools
}
