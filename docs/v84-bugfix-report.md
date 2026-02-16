# Grafana 8.4.7 MCP Tools — Bug 修复与规格对齐报告

> 生成时间：2026-02-16  
> 对标文档：`grafana-8.4.7-mcp-tool-spec.md` · `grafana-8.4.7-mcp-go-struct-mapping.md` · `grafana-8.4.7-mcp-implementation-blueprint.md`

---

## 一、Bug 修复（3 项）

### Bug 1：`upsert_dashboard` 响应中 Title 字段映射错误

| 项目 | 内容 |
|------|------|
| **文件** | `tools/v84/dashboard.go` |
| **严重等级** | 🔴 高 — 导致返回空 Title |
| **根因** | 生成的 OpenAPI model `PostDashboardOKBody` 中 Go 字段名为 `Slug`，但 JSON tag 为 `json:"title"`。而 Grafana 8.4.7 实际 API 返回的 JSON key 是 `"slug"` 而非 `"title"`，导致反序列化时该字段为 nil。此外 `ID` 被定义为 `*string`，但实际 API 返回整数。 |

**修复前代码：**

```go
// 使用生成的 OpenAPI client，字段映射错误
resp, err := gc.Dashboards.PostDashboard(cmd)
payload := resp.Payload
if payload.Slug != nil {
    out.Title = *payload.Slug  // Slug 的 json tag 是 "title"，但 API 返回 "slug"
}
if payload.Verion != nil {     // 字段名拼写错误（Verion vs Version）
    out.Version = *payload.Verion
}
```

**修复后代码：**

```go
// 改用 raw HTTP 请求 + 自定义 response struct，确保 JSON tag 与实际 API 一致
type upsertDashboardRawResponse struct {
    Status  string      `json:"status"`
    ID      FlexibleID  `json:"id"`      // 支持 int 或 string
    UID     string      `json:"uid"`
    Slug    string      `json:"slug"`    // 正确映射到 API 返回的 "slug" 字段
    URL     string      `json:"url"`
    Version json.Number `json:"version"` // 正确解析数字
}

// 发送原始 HTTP 请求
respBody, statusCode, err := doAPIRequest(ctx, "POST", "/dashboards/db", nil, requestBody)

// Title 优先从输入 dashboard payload 获取，回退到 API 返回的 slug
title := raw.Slug
if t, ok := args.Dashboard["title"].(string); ok && t != "" {
    title = t
}
```

**核心改动：**
1. 弃用生成的 OpenAPI client 调用，改用 `doAPIRequest` 发送 raw HTTP
2. 自定义 `upsertDashboardRawResponse` struct，JSON tag 与 Grafana 8.4.7 实际返回一致
3. `ID` 使用 `FlexibleID`（`json.RawMessage`）处理类型不确定性
4. `Version` 使用 `json.Number` 避免类型断言问题
5. `Title` 优先从 `args.Dashboard["title"]` 取值（即用户提交的标题），slug 作为兜底

---

### Bug 2：`create_folder` 声明了不可用的 ParentUID 参数

| 项目 | 内容 |
|------|------|
| **文件** | `tools/v84/folder.go` |
| **严重等级** | 🟡 中 — 参数被忽略，用户设置无效 |
| **根因** | `CreateFolderRequest` 声明了 `ParentUID` 字段，但 Grafana 8.4.7 的 `CreateFolderCommand` model 仅支持 `Title` + `UID`。嵌套文件夹是 Grafana 9.x 引入的功能。 |

**修复前代码：**

```go
type CreateFolderRequest struct {
    Title     string `json:"title" jsonschema:"required,description=Folder title"`
    UID       string `json:"uid,omitempty" jsonschema:"description=Optional custom folder UID"`
    ParentUID string `json:"parentUid,omitempty" jsonschema:"description=Optional parent folder UID"`
    // ↑ 该字段在 createFolder 函数中从未使用
}

func createFolder(ctx context.Context, args CreateFolderRequest) (*models.Folder, error) {
    cmd := &models.CreateFolderCommand{
        Title: args.Title,
        UID:   args.UID,
        // ParentUID 未传入，用户设置后无任何效果
    }
    // ...
}
```

**修复后代码：**

```go
type CreateFolderRequest struct {
    Title string `json:"title" jsonschema:"required,description=Folder title"`
    UID   string `json:"uid,omitempty" jsonschema:"description=Optional custom folder UID"`
    // ParentUID 已移除 — Grafana 8.4.7 不支持嵌套文件夹
}
```

**核心改动：**
1. 从 `CreateFolderRequest` 中移除 `ParentUID` 字段，避免误导用户
2. 同时将返回类型从 `*models.Folder` 改为自定义 `*CreateFolderResponse`（见下方规格对齐部分）

---

### Bug 3：`list_folders` 声明了不可用的 Permission 参数

| 项目 | 内容 |
|------|------|
| **文件** | `tools/v84/folder.go` |
| **严重等级** | 🟡 中 — 参数被忽略，用户设置无效 |
| **根因** | `ListFoldersRequest` 声明了 `Permission` 字段，但 Grafana 8.4.7 的 `GetFoldersParams` 仅支持 `Limit` + `Page`。Permission 过滤是后续版本引入的。 |

**修复前代码：**

```go
type ListFoldersRequest struct {
    Limit      *int64 `json:"limit,omitempty" jsonschema:"description=Max items\\, default 1000"`
    Page       *int64 `json:"page,omitempty" jsonschema:"description=Page number\\, default 1"`
    Permission string `json:"permission,omitempty" jsonschema:"description=View or Edit"`
    // ↑ 该字段在 listFolders 函数中从未使用
}
```

**修复后代码：**

```go
type ListFoldersRequest struct {
    Limit *int64 `json:"limit,omitempty" jsonschema:"description=Max items\\, default 1000"`
    Page  *int64 `json:"page,omitempty" jsonschema:"description=Page number\\, default 1"`
    // Permission 已移除 — Grafana 8.4.7 GetFolders API 不支持该参数
}
```

---

## 二、响应类型规格对齐（7 项）

所有工具的返回类型从 OpenAPI 生成的模型 / `any` / `map[string]any` 统一为 `go-struct-mapping.md` 定义的合约类型。

### 2.1 新增合约类型文件

**新文件：`tools/v84/types.go`**

定义了以下合约类型：

| 类型 | 说明 |
|------|------|
| `APIError` | MCP 统一错误模型，含 `StatusCode`、`Message`、`Detail`、`Upstream` |
| `SearchHit` | `/api/search` 返回对象 |
| `FolderItem` | 文件夹列表项 |
| `DatasourceModel` | 数据源详情 |
| `AnnotationItem` | 注解对象 |
| `LegacyAlertItem` | 旧版告警对象 |
| `LegacyNotificationChannel` | 旧版通知渠道对象 |
| `OrgUserItem` | 组织用户对象 |
| `TeamItem` | 团队对象 |
| `FlexibleID` | 弹性 ID（`json.RawMessage`），支持 int 或 string |

### 2.2 各工具变更明细

| 工具 | 文件 | 修复前返回类型 | 修复后返回类型 |
|------|------|---------------|---------------|
| `get_current_user` | `health_user_org.go` | `any` | `*GetCurrentUserResponse` |
| `get_current_org` | `health_user_org.go` | `any` | `*GetCurrentOrgResponse` |
| `search_dashboards` | `search.go` | `[]*models.Hit` | `[]SearchHit` |
| `get_dashboard_by_uid` | `dashboard.go` | `Dashboard any, Meta any` | `Dashboard map[string]any, Meta map[string]any` |
| `list_folders` | `folder.go` | `[]*models.FolderSearchHit` | `[]FolderItem` |
| `create_folder` | `folder.go` | `*models.Folder` | `*CreateFolderResponse` |
| `update_folder` | `folder.go` | `*models.Folder` | `*UpdateFolderResponse` |
| `list_datasources` | `datasource.go` | `[]*models.DataSourceListItemDTO` | `[]DatasourceModel` |
| `get_datasource` | `datasource.go` | `*models.DataSource` | `DatasourceModel` |
| `get_annotations` | `annotations.go` | `[]*models.ItemDTO` | `[]AnnotationItem` |
| `list_legacy_alerts` | `legacy_alerting.go` | `[]map[string]any` | `[]LegacyAlertItem` |
| `list_legacy_notification_channels` | `legacy_alerting.go` | `[]map[string]any` | `[]LegacyNotificationChannel` |
| `list_org_users` | `org_admin.go` | `[]*models.OrgUserDTO` | `[]OrgUserItem` |
| `list_teams` | `org_admin.go` | `[]*models.TeamDTO` | `[]TeamItem` |

**每个工具的转换均通过显式字段映射（非类型断言），从 OpenAPI model 到合约类型，确保字段可控：**

```go
// 示例：datasource_resolver.go 中的转换函数
func dataSourceToModel(ds *models.DataSource) DatasourceModel {
    return DatasourceModel{
        ID:   ds.ID,
        UID:  ds.UID,
        Name: ds.Name,
        Type: ds.Type,
        // ... 逐字段显式映射
    }
}
```

---

## 三、缺失功能补充（2 项）

### 3.1 统一 `APIError` 错误模型与错误映射器

**文件：`tools/v84/types.go` + `tools/v84/common.go`**

新增内容：

```go
// types.go — 错误模型
type APIError struct {
    StatusCode int            `json:"statusCode"`
    Message    string         `json:"message"`
    Status     string         `json:"status,omitempty"`
    Detail     string         `json:"detail,omitempty"`
    Upstream   map[string]any `json:"upstream,omitempty"`
}

func (e *APIError) Error() string { ... }

// common.go — 错误映射工具函数
func newAPIError(statusCode int, message string, upstream map[string]any) *APIError { ... }
func wrapAPIError(statusCode int, respBody []byte, fallbackMsg string) *APIError { ... }
```

`wrapAPIError` 尝试解析 HTTP 错误响应的 JSON body，提取 `message` 字段；解析失败时使用 `fallbackMsg` 兜底。已在 `upsert_dashboard` 中应用。

### 3.2 补全 `jsonschema` 描述标签

以下 Request Struct 的字段原先缺少 `jsonschema:"description=..."` 标签，已全部补齐：

| 文件 | Struct | 补充字段数 |
|------|--------|-----------|
| `annotations.go` | `GetAnnotationsRequest` | 12 个字段全部补齐 |
| `annotations.go` | `CreateAnnotationRequest` | 7 个字段补齐 |
| `annotations.go` | `PatchAnnotationRequest` | 5 个字段补齐 |
| `legacy_alerting.go` | `ListLegacyAlertsRequest` | 6 个字段补齐 |
| `org_admin.go` | `ListTeamsRequest` | 3 个字段补齐 |
| `search.go` | `SearchDashboardsRequest` | 部分描述优化 |
| `datasource.go` | `ListDatasourcesRequest` | 修正转义语法 |

---

## 四、单元测试补充

**文件：`tools/v84/v84_unit_test.go`**

| 分类 | 新增测试函数 | 测试用例数 |
|------|-------------|-----------|
| Health / User / Org | `TestGetHealth`, `TestGetCurrentUser`, `TestGetCurrentOrg` | 3 |
| Search | `TestSearchDashboards` | 2（分页 + 空结果） |
| Dashboard CRUD | `TestGetDashboardByUID`, `TestUpsertDashboard` | 5（验证 + 正常 + slug 回退） |
| Folder CRUD | `TestListFolders`, `TestCreateFolder`, `TestUpdateFolder` | 5（验证 + 成功） |
| Annotations | `TestGetAnnotations`, `TestCreateAnnotation`, `TestPatchAnnotation` | 4（验证 + 成功） |
| Org Users / Teams | `TestListOrgUsers`, `TestListTeams` | 3（成功 + 空结果） |
| APIError | `TestAPIErrorModel` | 5（Error() + wrapAPIError 3 种场景） |
| **合计** | **15 个新函数** | **27 个新用例** |

**测试覆盖改善：**

| 工具 | 修复前 | 修复后 |
|------|--------|--------|
| `get_health` | ❌ | ✅ |
| `get_current_user` | ❌ | ✅ |
| `get_current_org` | ❌ | ✅ |
| `search_dashboards` | ❌ | ✅ |
| `get_dashboard_by_uid` | ❌ | ✅ |
| `upsert_dashboard` | ❌ | ✅ |
| `list_folders` | ❌ | ✅ |
| `create_folder` | ❌ | ✅ |
| `update_folder` | ❌ | ✅ |
| `get_annotations` | ❌ | ✅ |
| `create_annotation` | ❌ | ✅ |
| `patch_annotation` | ❌ | ✅ |
| `list_org_users` | ❌ | ✅ |
| `list_teams` | ❌ | ✅ |
| `APIError` | ❌ | ✅ |

---

## 五、变更文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `tools/v84/types.go` | **新增** | 合约类型 + APIError 模型 |
| `tools/v84/common.go` | 修改 | 新增 `newAPIError` / `wrapAPIError` 错误映射函数 |
| `tools/v84/dashboard.go` | 修改 | Bug 1 修复：raw HTTP + 正确 JSON mapping |
| `tools/v84/folder.go` | 修改 | Bug 2+3 修复：移除不可用参数 + 合约返回类型 |
| `tools/v84/health_user_org.go` | 修改 | `any` → 类型化响应 |
| `tools/v84/search.go` | 修改 | `[]*models.Hit` → `[]SearchHit` |
| `tools/v84/datasource.go` | 修改 | OpenAPI model → `DatasourceModel` |
| `tools/v84/datasource_resolver.go` | 修改 | `*models.DataSource` → `DatasourceModel` |
| `tools/v84/annotations.go` | 修改 | jsonschema 补全 + `[]AnnotationItem` |
| `tools/v84/legacy_alerting.go` | 修改 | `[]map[string]any` → 类型化 struct |
| `tools/v84/org_admin.go` | 修改 | OpenAPI model → 合约类型 |
| `tools/v84/v84_unit_test.go` | 修改 | 新增 15 个测试函数、27 个测试用例 |

---

## 六、验证结果

```
$ GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn go build ./...
# 编译通过，无错误

$ GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn go test -tags unit ./tools/v84/ -v
# 30 个测试全部通过 (PASS)
# ok  github.com/grafana/mcp-grafana/tools/v84  1.148s
```
