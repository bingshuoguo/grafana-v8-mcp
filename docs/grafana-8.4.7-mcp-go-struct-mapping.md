# Grafana 8.4.7 MCP Go Struct 字段定义（Breaking 版）

## 1. 说明

本文件是 `docs/grafana-8.4.7-mcp-tool-spec.md` 的 Go 类型化版本，目标：

- 直接作为 `MustTool/ConvertTool` 的请求/响应 struct 基线
- 与 **v8.4.7 breaking 默认契约** 保持一致
- 与 `ID-first` datasource 策略保持一致

说明：

- 请求 struct 建议带 `jsonschema` tag（用于 MCP input schema 生成）
- `jsonschema` 描述中若有逗号，需写成 `\\,`（见仓库 linter 约束）

---

## 2. 共享类型（建议放 `types.go`）

```go
package v84contract

import (
	"encoding/json"
	"time"
)

// APIError is the MCP-facing normalized error payload.
type APIError struct {
	StatusCode int            `json:"statusCode"`
	Message    string         `json:"message"`
	Status     string         `json:"status,omitempty"`
	Detail     string         `json:"detail,omitempty"`
	Upstream   map[string]any `json:"upstream,omitempty"`
}

// DatasourceRef identifies datasource by id/uid/name.
// ID-first resolution order: id > uid > name.
type DatasourceRef struct {
	ID   *int64 `json:"id,omitempty"`
	UID  string `json:"uid,omitempty"`
	Name string `json:"name,omitempty"`
}

// SearchHit maps /api/search hit object.
type SearchHit struct {
	ID          int64    `json:"id,omitempty"`
	UID         string   `json:"uid,omitempty"`
	Title       string   `json:"title,omitempty"`
	Type        string   `json:"type,omitempty"` // dash-db | dash-folder
	URL         string   `json:"url,omitempty"`
	URI         string   `json:"uri,omitempty"`
	Slug        string   `json:"slug,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	FolderID    int64    `json:"folderId,omitempty"`
	FolderUID   string   `json:"folderUid,omitempty"`
	FolderTitle string   `json:"folderTitle,omitempty"`
	FolderURL   string   `json:"folderUrl,omitempty"`
	IsStarred   bool     `json:"isStarred,omitempty"`
}

type FolderItem struct {
	ID    int64  `json:"id,omitempty"`
	UID   string `json:"uid,omitempty"`
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}

type DatasourceModel struct {
	ID               int64           `json:"id,omitempty"`
	UID              string          `json:"uid,omitempty"`
	Name             string          `json:"name,omitempty"`
	Type             string          `json:"type,omitempty"`
	TypeName         string          `json:"typeName,omitempty"`
	TypeLogoURL      string          `json:"typeLogoUrl,omitempty"`
	URL              string          `json:"url,omitempty"`
	Access           string          `json:"access,omitempty"`
	Database         string          `json:"database,omitempty"`
	User             string          `json:"user,omitempty"`
	OrgID            int64           `json:"orgId,omitempty"`
	IsDefault        bool            `json:"isDefault,omitempty"`
	BasicAuth        bool            `json:"basicAuth,omitempty"`
	WithCredentials  bool            `json:"withCredentials,omitempty"`
	ReadOnly         bool            `json:"readOnly,omitempty"`
	Version          int64           `json:"version,omitempty"`
	JSONData         map[string]any  `json:"jsonData,omitempty"`
	SecureJSONFields map[string]bool `json:"secureJsonFields,omitempty"`
}

type AnnotationItem struct {
	ID           int64          `json:"id,omitempty"`
	AlertID      int64          `json:"alertId,omitempty"`
	AlertUID     string         `json:"alertUID,omitempty"`
	AlertName    string         `json:"alertName,omitempty"`
	DashboardID  int64          `json:"dashboardId,omitempty"`
	DashboardUID string         `json:"dashboardUID,omitempty"`
	PanelID      int64          `json:"panelId,omitempty"`
	UserID       int64          `json:"userId,omitempty"`
	Login        string         `json:"login,omitempty"`
	Email        string         `json:"email,omitempty"`
	AvatarURL    string         `json:"avatarUrl,omitempty"`
	NewState     string         `json:"newState,omitempty"`
	PrevState    string         `json:"prevState,omitempty"`
	Text         string         `json:"text,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
	Time         int64          `json:"time,omitempty"`
	TimeEnd      int64          `json:"timeEnd,omitempty"`
	Created      int64          `json:"created,omitempty"`
	Updated      int64          `json:"updated,omitempty"`
	Data         map[string]any `json:"data,omitempty"`
}

type LegacyAlertItem struct {
	ID             int64          `json:"id,omitempty"`
	DashboardID    int64          `json:"dashboardId,omitempty"`
	DashboardUID   string         `json:"dashboardUid,omitempty"`
	DashboardSlug  string         `json:"dashboardSlug,omitempty"`
	PanelID        int64          `json:"panelId,omitempty"`
	Name           string         `json:"name,omitempty"`
	State          string         `json:"state,omitempty"`
	NewStateDate   time.Time      `json:"newStateDate,omitempty"`
	EvalDate       time.Time      `json:"evalDate,omitempty"`
	ExecutionError string         `json:"executionError,omitempty"`
	URL            string         `json:"url,omitempty"`
	EvalData       map[string]any `json:"evalData,omitempty"`
}

type LegacyNotificationChannel struct {
	ID                    int64           `json:"id,omitempty"`
	UID                   string          `json:"uid,omitempty"`
	Name                  string          `json:"name,omitempty"`
	Type                  string          `json:"type,omitempty"`
	IsDefault             bool            `json:"isDefault,omitempty"`
	SendReminder          bool            `json:"sendReminder,omitempty"`
	DisableResolveMessage bool            `json:"disableResolveMessage,omitempty"`
	Frequency             string          `json:"frequency,omitempty"`
	Created               time.Time       `json:"created,omitempty"`
	Updated               time.Time       `json:"updated,omitempty"`
	SecureFields          map[string]bool `json:"secureFields,omitempty"`
	Settings              map[string]any  `json:"settings,omitempty"`
}

type OrgUserItem struct {
	UserID        int64            `json:"userId,omitempty"`
	OrgID         int64            `json:"orgId,omitempty"`
	Login         string           `json:"login,omitempty"`
	Name          string           `json:"name,omitempty"`
	Email         string           `json:"email,omitempty"`
	AvatarURL     string           `json:"avatarUrl,omitempty"`
	Role          string           `json:"role,omitempty"`
	LastSeenAt    time.Time        `json:"lastSeenAt,omitempty"`
	LastSeenAtAge string           `json:"lastSeenAtAge,omitempty"`
	AccessControl map[string]bool  `json:"accessControl,omitempty"`
}

type TeamItem struct {
	ID            int64           `json:"id,omitempty"`
	OrgID         int64           `json:"orgId,omitempty"`
	Name          string          `json:"name,omitempty"`
	Email         string          `json:"email,omitempty"`
	AvatarURL     string          `json:"avatarUrl,omitempty"`
	MemberCount   int64           `json:"memberCount,omitempty"`
	Permission    int64           `json:"permission,omitempty"`
	AccessControl map[string]bool `json:"accessControl,omitempty"`
}

// ID in upsert response can be either int or string.
// Keep raw value and normalize in caller if needed.
type FlexibleID = json.RawMessage
```

---

## 3. Tool 级 Request / Response 定义（MVP）

### 3.1 `get_health` (`GET /api/health`)

```go
type GetHealthRequest struct{}

type GetHealthResponse struct {
	Database string `json:"database,omitempty"`
	Version  string `json:"version,omitempty"`
	Commit   string `json:"commit,omitempty"`
}
```

### 3.2 `get_current_user` (`GET /api/user`)

```go
type GetCurrentUserRequest struct{}

type GetCurrentUserResponse struct {
	ID             int64           `json:"id,omitempty"`
	Login          string          `json:"login,omitempty"`
	Email          string          `json:"email,omitempty"`
	Name           string          `json:"name,omitempty"`
	AvatarURL      string          `json:"avatarUrl,omitempty"`
	IsGrafanaAdmin bool            `json:"isGrafanaAdmin,omitempty"`
	IsExternal     bool            `json:"isExternal,omitempty"`
	IsDisabled     bool            `json:"isDisabled,omitempty"`
	OrgID          int64           `json:"orgId,omitempty"`
	Theme          string          `json:"theme,omitempty"`
	AuthLabels     []string        `json:"authLabels,omitempty"`
	AccessControl  map[string]bool `json:"accessControl,omitempty"`
}
```

### 3.3 `get_current_org` (`GET /api/org`)

```go
type GetCurrentOrgRequest struct{}

type GetCurrentOrgResponse struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
```

### 3.4 `search_dashboards` (`GET /api/search`)

```go
type SearchDashboardsRequest struct {
	Query        string   `json:"query,omitempty" jsonschema:"description=Search text for dashboard titles and metadata"`
	Limit        *int64   `json:"limit,omitempty" jsonschema:"description=Max results per page\, default 50"`
	Page         *int64   `json:"page,omitempty" jsonschema:"description=Page number\, 1-indexed"`
	Tag          []string `json:"tag,omitempty" jsonschema:"description=Filter by tags"`
	DashboardIDs []int64  `json:"dashboardIds,omitempty" jsonschema:"description=Filter by dashboard IDs"`
	FolderIDs    []int64  `json:"folderIds,omitempty" jsonschema:"description=Filter by folder IDs"`
	Starred      *bool    `json:"starred,omitempty" jsonschema:"description=Only starred dashboards when true"`
}

type SearchDashboardsResponse struct {
	Items   []SearchHit `json:"items"`
	Page    int64       `json:"page"`
	Limit   int64       `json:"limit"`
	HasMore bool        `json:"hasMore"`
}
```

### 3.5 `get_dashboard_by_uid` (`GET /api/dashboards/uid/{uid}`)

```go
type GetDashboardByUIDRequest struct {
	UID string `json:"uid" jsonschema:"required,description=Dashboard UID"`
}

type GetDashboardByUIDResponse struct {
	Dashboard map[string]any `json:"dashboard"`
	Meta      map[string]any `json:"meta,omitempty"`
}
```

### 3.6 `upsert_dashboard` (`POST /api/dashboards/db`)

```go
type UpsertDashboardRequest struct {
	Dashboard map[string]any `json:"dashboard" jsonschema:"required,description=Full dashboard JSON payload"`
	FolderID  *int64         `json:"folderId,omitempty" jsonschema:"description=Numeric folder ID"`
	FolderUID string         `json:"folderUid,omitempty" jsonschema:"description=Folder UID"`
	Overwrite *bool          `json:"overwrite,omitempty" jsonschema:"description=Overwrite existing dashboard"`
	Message   string         `json:"message,omitempty" jsonschema:"description=Version history message"`
}

type UpsertDashboardResponse struct {
	Status  string     `json:"status,omitempty"`
	ID      FlexibleID `json:"id,omitempty"`
	UID     string     `json:"uid,omitempty"`
	Title   string     `json:"title,omitempty"`
	URL     string     `json:"url,omitempty"`
	Version int64      `json:"version,omitempty"`
}
```

### 3.7 `list_folders` (`GET /api/folders`)

```go
type ListFoldersRequest struct {
	Limit      *int64  `json:"limit,omitempty" jsonschema:"description=Max items\, default 1000"`
	Page       *int64  `json:"page,omitempty" jsonschema:"description=Page number\, default 1"`
	Permission string  `json:"permission,omitempty" jsonschema:"description=View or Edit"`
}

type ListFoldersResponse struct {
	Items []FolderItem `json:"items"`
}
```

### 3.8 `create_folder` (`POST /api/folders`)

```go
type CreateFolderRequest struct {
	Title     string `json:"title" jsonschema:"required,description=Folder title"`
	UID       string `json:"uid,omitempty" jsonschema:"description=Custom folder UID"`
	ParentUID string `json:"parentUid,omitempty" jsonschema:"description=Parent folder UID"`
}

type CreateFolderResponse struct {
	ID    int64  `json:"id,omitempty"`
	UID   string `json:"uid,omitempty"`
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}
```

### 3.9 `update_folder` (`PUT /api/folders/{folder_uid}`)

```go
type UpdateFolderRequest struct {
	FolderUID   string `json:"folderUid" jsonschema:"required,description=Folder UID"`
	Title       string `json:"title" jsonschema:"required,description=New folder title"`
	Description string `json:"description,omitempty" jsonschema:"description=New folder description"`
	Version     *int64 `json:"version,omitempty" jsonschema:"description=Legacy folder version guard"`
	Overwrite   *bool  `json:"overwrite,omitempty" jsonschema:"description=Legacy folder overwrite switch"`
}

type UpdateFolderResponse struct {
	ID    int64  `json:"id,omitempty"`
	UID   string `json:"uid,omitempty"`
	Title string `json:"title,omitempty"`
}
```

### 3.10 `list_datasources` (`GET /api/datasources`)

```go
type ListDatasourcesRequest struct {
	Type   string `json:"type,omitempty" jsonschema:"description=Datasource type filter"`
	Limit  *int64 `json:"limit,omitempty" jsonschema:"description=Max result size\, default 100"`
	Offset *int64 `json:"offset,omitempty" jsonschema:"description=Pagination offset"`
}

type ListDatasourcesResponse struct {
	Items   []DatasourceModel `json:"items"`
	Total   int64             `json:"total"`
	HasMore bool              `json:"hasMore"`
}
```

### 3.11 `get_datasource` (`GET by id/name/uid`)

```go
type GetDatasourceRequest struct {
	ID   *int64 `json:"id,omitempty" jsonschema:"description=Datasource numeric ID"`
	UID  string `json:"uid,omitempty" jsonschema:"description=Datasource UID"`
	Name string `json:"name,omitempty" jsonschema:"description=Datasource name"`
}

type GetDatasourceResponse struct {
	ResolvedBy string          `json:"resolvedBy,omitempty"` // id | uid | name
	Datasource DatasourceModel `json:"datasource"`
}
```

### 3.12 `resolve_datasource_ref`（组合工具）

```go
type ResolveDatasourceRefRequest struct {
	ID   *int64 `json:"id,omitempty" jsonschema:"description=Datasource numeric ID"`
	UID  string `json:"uid,omitempty" jsonschema:"description=Datasource UID"`
	Name string `json:"name,omitempty" jsonschema:"description=Datasource name"`
}

type ResolveDatasourceRefResponse struct {
	ID   int64  `json:"id"`
	UID  string `json:"uid,omitempty"`
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url,omitempty"`
}
```

### 3.13 `query_datasource` (`POST /api/tsdb/query`)

```go
type QueryDatasourceRequest struct {
	From       string           `json:"from" jsonschema:"required,description=From time\, e.g. now-1h"`
	To         string           `json:"to" jsonschema:"required,description=To time\, e.g. now"`
	Debug      *bool            `json:"debug,omitempty" jsonschema:"description=Enable query debug mode"`
	Datasource *DatasourceRef   `json:"datasource,omitempty" jsonschema:"description=Optional datasource reference"`
	Queries    []map[string]any `json:"queries" jsonschema:"required,description=Datasource query payload list"`
}

type QueryDatasourceResponse struct {
	Raw       json.RawMessage `json:"raw,omitempty"`
	Responses map[string]any  `json:"responses,omitempty"`
	Hints     []string        `json:"hints,omitempty"`
}
```

### 3.14 `get_annotations` (`GET /api/annotations`)

```go
type GetAnnotationsRequest struct {
	From         *int64   `json:"from,omitempty"`
	To           *int64   `json:"to,omitempty"`
	UserID       *int64   `json:"userId,omitempty"`
	AlertID      *int64   `json:"alertId,omitempty"`
	AlertUID     *string  `json:"alertUid,omitempty"`
	DashboardID  *int64   `json:"dashboardId,omitempty"`
	DashboardUID *string  `json:"dashboardUid,omitempty"`
	PanelID      *int64   `json:"panelId,omitempty"`
	Limit        *int64   `json:"limit,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Type         string   `json:"type,omitempty"` // alert | annotation
	MatchAny     *bool    `json:"matchAny,omitempty"`
}

type GetAnnotationsResponse struct {
	Items []AnnotationItem `json:"items"`
}
```

### 3.15 `create_annotation` (`POST /api/annotations`)

```go
type CreateAnnotationRequest struct {
	DashboardID  *int64         `json:"dashboardId,omitempty"`
	DashboardUID string         `json:"dashboardUid,omitempty"`
	PanelID      *int64         `json:"panelId,omitempty"`
	Time         *int64         `json:"time,omitempty"`
	TimeEnd      *int64         `json:"timeEnd,omitempty"`
	Tags         []string       `json:"tags,omitempty"`
	Text         string         `json:"text" jsonschema:"required,description=Annotation text"`
	Data         map[string]any `json:"data,omitempty"`
}

type CreateAnnotationResponse struct {
	Message string `json:"message,omitempty"`
	ID      int64  `json:"id,omitempty"`
}
```

### 3.16 `patch_annotation` (`PATCH /api/annotations/{annotation_id}`)

```go
type PatchAnnotationRequest struct {
	ID      int64          `json:"id" jsonschema:"required,description=Annotation ID"`
	Time    *int64         `json:"time,omitempty"`
	TimeEnd *int64         `json:"timeEnd,omitempty"`
	Tags    []string       `json:"tags,omitempty"`
	Text    *string        `json:"text,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
}

type PatchAnnotationResponse struct {
	Message string `json:"message,omitempty"`
}
```

### 3.17 `list_legacy_alerts` (`GET /api/alerts`)

```go
type ListLegacyAlertsRequest struct {
	DashboardID  *int64   `json:"dashboardId,omitempty"`
	PanelID      *int64   `json:"panelId,omitempty"`
	Query        string   `json:"query,omitempty"`
	State        string   `json:"state,omitempty"`
	Limit        *int64   `json:"limit,omitempty"`
	DashboardTag []string `json:"dashboardTag,omitempty"`
}

type ListLegacyAlertsResponse struct {
	Items []LegacyAlertItem `json:"items"`
}
```

### 3.18 `list_legacy_notification_channels` (`GET /api/alert-notifications`)

```go
type ListLegacyNotificationChannelsRequest struct {
	Name string `json:"name,omitempty"`
}

type ListLegacyNotificationChannelsResponse struct {
	Items []LegacyNotificationChannel `json:"items"`
}
```

### 3.19 `list_org_users` (`GET /api/org/users`)

```go
type ListOrgUsersRequest struct{}

type ListOrgUsersResponse struct {
	Items []OrgUserItem `json:"items"`
}
```

### 3.20 `list_teams` (`GET /api/teams/search`)

```go
type ListTeamsRequest struct {
	Query   string `json:"query,omitempty"`
	Page    *int64 `json:"page,omitempty"`
	PerPage *int64 `json:"perPage,omitempty"`
}

type ListTeamsResponse struct {
	TotalCount int64      `json:"totalCount,omitempty"`
	Page       int64      `json:"page,omitempty"`
	PerPage    int64      `json:"perPage,omitempty"`
	Teams      []TeamItem `json:"teams"`
}
```

---

## 4. 第二阶段（默认关闭）建议命名

- `QueryDatasourceExpressionsRequest / QueryDatasourceExpressionsResponse`
- `UpdateAnnotationRequest / UpdateAnnotationResponse`
- `DeleteAnnotationRequest / DeleteAnnotationResponse`
- `GetAnnotationTagsRequest / GetAnnotationTagsResponse`
- `ListRulerRulesRequest / ListRulerRulesResponse`
- `GetAlertmanagerAlertsRequest / GetAlertmanagerAlertsResponse`

---

## 5. 实现注意事项

- `id` 字段在不同接口可能是 string/int，`upsert_dashboard` 使用 `json.RawMessage` 处理。
- optional query 参数建议使用指针类型（`*bool`, `*int64`, `*string`）区分“未传”与零值。
- datasource 相关工具内部应统一走 `resolve_datasource_ref`，避免重复分支。
- 本文档按 breaking 契约编写，不包含旧工具名兼容层。
