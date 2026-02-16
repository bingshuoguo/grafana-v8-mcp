# Grafana 8.4.7 MCP 实施蓝图（Breaking 默认契约）

## 1. 已确认的实施前提

本蓝图基于以下已确认决策：

1. **替换默认契约（允许 breaking）**
2. **OpenAPI client 固定到 8.4.7 对应版本/规格**
3. **不引入过渡期（无 alias、无双轨）**

这意味着：

- 旧工具名不再保留为默认能力
- 默认 tool 集直接切换为 `docs/grafana-8.4.7-mcp-tool-spec.md`
- 代码结构以“8.4.7 契约优先”收敛

---

## 2. 目标

在当前仓库骨架上落地可运行的 8.4.7 MCP Server，保证：

- tool 名称、输入参数、返回结构与新契约一致
- datasource 相关能力遵循 `ID-first`
- legacy alerting 能力可用（`/api/alerts`、`/api/alert-notifications`）

---

## 3. 对现有骨架的使用策略

当前骨架保留：

- `cmd/mcp-grafana/main.go`：启动、transport、hooks
- `tools.go`：`MustTool/ConvertTool`、错误语义、trace 注入
- `session/proxy/observability`：基础基础设施

调整策略：

- 保留骨架，不推翻 server 生命周期管理
- 将“默认注册工具集合”切换到 v8.4.7 契约集合
- 旧 categories（oncall/sift/incident 等）从默认注册路径移除

---

## 4. 依赖与版本策略（必须先做）

## 4.1 OpenAPI client 固定到 8.4.7

必须保证客户端模型与 8.4.7 规格一致，推荐二选一：

1. **本地生成并 vendoring（推荐）**
   - 输入：`v8.4.7/public/api-spec.json` + `api-merged.json`
   - 产物：`internal/third_party/grafana-openapi-v847`
   - 在项目中统一 import 该包
2. **依赖 pin 到明确的 v8.4.7 产物版本**
   - `go.mod` 固定到可审计版本
   - 禁止跟随主线最新自动漂移

## 4.2 缺失 endpoint 的处理

如果 8.4.7 生成 client 不覆盖某些 endpoint（常见于 legacy/merged 差异）：

- 在 `internal/compat/v84/rawclient` 提供最小 raw HTTP 封装
- 仅对缺失接口走 raw，其他接口保持 typed client

原则：**不为“追求全 typed”而回退到非 8.4.7 版本 client**。

---

## 5. 包结构建议（最小侵入）

```text
cmd/
  mcp-grafana/
    main.go

internal/
  compat/
    v84/
      profile.go
      datasource_resolver.go
      query_adapter.go
      rawclient/
        legacy_alerts.go
        legacy_notifications.go
      errors.go

tools/
  v84/
    registry.go
    health.go
    user_org.go
    search.go
    dashboard.go
    folder.go
    datasource.go
    query.go
    annotations.go
    legacy_alerting.go
    org_admin.go
```

说明：

- `tools/v84/registry.go` 是唯一默认入口
- 旧 `tools/*.go` 可保留作迁移参考，但不再作为默认注册集合

---

## 6. 默认契约切换方案（无过渡）

## 6.1 默认工具集合

默认仅注册：

- `docs/grafana-8.4.7-mcp-tool-spec.md` 中 MVP 工具

不再默认注册（breaking）：

- 旧命名：`update_dashboard`、`list_users_by_org`、`get_datasource_by_uid`、`get_datasource_by_name`
- 插件/云产品强依赖工具（oncall/sift/incident/asserts 等）

## 6.2 CLI 行为

建议新增并启用 v84 专用开关：

- `--profile=v84`（默认值）
- `--enable-v84-optional-tools`（默认 false）
- `--disable-write`（沿用）

并将 `enabled-tools` 的默认值切换到 `v84` 维度，而不是旧 category 列表。

---

## 7. 核心实现细则

## 7.1 Datasource Resolver（强约束）

统一入口：`ResolveDatasourceRef(ctx, ref)`

- 输入：`id | uid | name`
- 输出：`id, uid, name, type, url`
- 决策：`id > uid > name`

所有 datasource 相关工具必须调用 resolver，禁止分散解析。

## 7.2 Query Adapter

- `query_datasource`：主链路 `POST /api/tsdb/query`
- `query_datasource_expressions`：第二阶段 `POST /api/ds/query`
- 两者保持独立 tool，避免隐式回退导致语义不稳定

## 7.3 Error Mapper

统一转换为：

```json
{ "statusCode": 0, "message": "", "status": "", "detail": "", "upstream": {} }
```

与 `tools.go` 语义对齐：

- 可恢复错误 -> `IsError=true`
- 硬错误（配置/鉴权）-> `HardError`

---

## 8. 分阶段实施计划

## 阶段 A：依赖与骨架切换（1-2 天）

- 固定 OpenAPI client 到 8.4.7
- 建立 `internal/compat/v84` 与 `tools/v84/registry.go`
- 默认注册入口切换到 `AddV84Tools`

交付标准：

- `list_tools` 仅暴露 v84 默认契约工具
- 服务可启动并调用 `get_health`

## 阶段 B：MVP Read（3-5 天）

- user/org/search/dashboard/folder/datasource/query-read
- legacy alerts read
- org users / teams read

交付标准：

- 覆盖 v84 spec 的 read MVP
- 集成测试通过

## 阶段 C：MVP Write（2-4 天）

- `upsert_dashboard`
- `create_folder` / `update_folder`
- `create_annotation` / `patch_annotation`

交付标准：

- `--disable-write` 可正确隐藏写工具
- 写操作错误语义与契约一致

## 阶段 D：Optional（按需）

- `query_datasource_expressions`
- unified alerting 扩展

---

## 9. 测试与验收

## 9.1 环境基线

- Grafana 镜像：`grafana/grafana:8.4.7`
- 预置 datasource/dashboard/test data
- 认证覆盖：Service Account Token + Basic Auth

## 9.2 测试分层

1. 单元测试：resolver、query adapter、error mapper
2. 集成测试：真实 8.4.7 API 回归
3. 合约测试：tool input schema 与文档快照一致

Go 测试命令（仓库规范）：

```bash
GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn spkit run go test ./...
```

## 9.3 验收清单

- 所有 v84 MVP tools 在 8.4.7 可调用
- `id/uid/name` 三种 datasource 输入均可解析
- `--disable-write` 后写工具不暴露
- 默认无旧契约工具名

---

## 10. 风险与规避

风险：

- 8.4.7 返回字段与高版本模型不一致
- legacy alerting 在不同部署模式下行为差异
- query payload 在不同 datasource 类型间差异较大

规避：

- 固定 v8.4.7 规格与 client 产物
- 对 legacy endpoint 使用最小 raw adapter + 回归测试
- 为 query 增加 datasource 类型级测试样例

---

## 11. 建议下一步

1. 先提交依赖固定与默认工具入口切换 PR。
2. 再提交 datasource resolver + query adapter PR。
3. 最后分批接入 legacy alerting 与写工具。
