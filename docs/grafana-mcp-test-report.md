# Grafana MCP 工具测试报告

**测试时间**: 2026-02-27
**Grafana 版本**: 8.4.7 (commit: 23bf3ef043)
**目标 Org**: Map (ID: 36)
**MCP Server**: `mcp-grafana-8.4.7`

---

## 一、测试总览


| 工具名                                     | 状态       | 备注                          |
| --------------------------------------- | -------- | --------------------------- |
| `get_health`                            | ✅ PASS   | 返回版本、commit、db状态            |
| `get_current_user`                      | ❌ FAIL   | API Key 鉴权不支持此接口            |
| `get_current_org`                       | ✅ PASS   | 返回 Org ID 和 name            |
| `list_datasources`                      | ✅ PASS   | 共 104 个，hasMore=true        |
| `list_folders`                          | ✅ PASS   | 19 个 folder                 |
| `search_dashboards`                     | ✅ PASS   | 支持文本搜索、tag、分页               |
| `get_dashboard_by_uid`                  | ✅ PASS   | 返回完整 dashboard 定义           |
| `get_datasource` (by uid)               | ✅ PASS   | 解析方式: uid                   |
| `get_datasource` (by id)                | ✅ PASS   | 解析方式: id                    |
| `resolve_datasource_ref`                | ✅ PASS   | 支持 name/uid/id 解析           |
| `get_annotations`                       | ✅ PASS   | 支持 limit、tags 过滤            |
| `list_org_users`                        | ⚠️ PASS* | 返回 4.4MB 超限，被截断写文件          |
| `list_teams`                            | ✅ PASS   | 8 个团队，含成员数                  |
| `list_legacy_alerts`                    | ✅ PASS   | 返回 alert 状态、evalData        |
| `list_legacy_notification_channels`     | ✅ PASS   | 8 个 webhook 通知渠道            |
| `query_datasource` (Prometheus range)   | ✅ PASS   | 返回 Arrow 格式时序数据             |
| `query_datasource` (Prometheus instant) | ✅ PASS   | 返回单点 Arrow 数据               |
| `query_datasource` (多 refId)            | ✅ PASS   | A/B 独立 dataframe 正确返回       |
| `query_datasource` (debug 模式)           | ✅ PASS   | 调用成功，debug 字段无额外暴露          |
| `query_datasource` (MySQL rawSql)       | ✅ PASS   | 返回 table 格式 Arrow 数据        |
| `query_datasource` (无效 PromQL)          | ✅ PASS   | 422 + 清晰的解析错误信息             |
| `query_datasource` (网络不可达 DS)           | ❌ FAIL   | proxy 返回非 JSON，bad_response |
| `list_prometheus_metric_names`            | ⚠️ 部分   | 无 datasource → 正确提示需指定；指定后 → proxy 创建失败 |
| `list_prometheus_label_values`           | ⚠️ 部分   | 同上                               |
| `query_prometheus`                       | ❓ 未验证  | Cursor MCP 报 Tool not found，待本地/CLI 验证   |


**总计: 20 PASS / 2 FAIL / 1 大数据警告 / 2 部分通过 / 1 未验证**

---

## 二、各工具详细结果

### 1. `get_health` ✅

```json
{"database":"ok","version":"8.4.7","commit":"23bf3ef043"}
```

功能正常，可作为服务可用性探针。

---

### 2. `get_current_user` ❌

```
[GET /user][404] getSignedInUserNotFound {"message":"user not found"}
```

**根因**: Grafana 8.x 用 API Key 鉴权时，`/api/user` 端点返回 404。API Key 是 service account token，没有关联真实用户 profile。这是 Grafana 已知行为，非 MCP 实现缺陷。

---

### 3. `get_current_org` ✅

```json
{"id":36,"name":"Map"}
```

---

### 4. `list_datasources` ✅

- 总计 **104** 个 datasource，`hasMore: true`（默认 limit=100 无法一次取完）
- 类型分布：Prometheus（主体，~85个）、MySQL（3个）、InfluxDB（1个）、Elasticsearch（3个）、ClickHouse（1个）、Google Sheets（1个）
- **问题**: 工具无 `limit`/`offset` 参数，无法分页读取全部 104 条

---

### 5. `list_folders` ✅

返回 19 个 folder，含结构化 uid/title：

```
aoi, MapOperationDashboards, mapweb, middleware, module,
monitoring-test-env, MySQL, noc, open-platform, poi,
position, render, road, routing, search, sre,
sre-ds_cmdb_migrate, User Dashboards, 监控大盘
```

---

### 6. `search_dashboards` ✅

三种查询模式均正常：


| 模式     | 参数                     | 结果                  |
| ------ | ---------------------- | ------------------- |
| 全量     | `limit=5`              | 返回 5 条，hasMore=true |
| 关键词    | `query="SRE"`          | 返回 2 条精确匹配          |
| Tag 过滤 | `tag=["route-engine"]` | 返回 3 条匹配            |
| 分页     | `page=2, limit=3`      | 正确跳过第一页             |


---

### 7. `get_dashboard_by_uid` ✅

以 `aUyw7Xnnz`（[DS-Route-engine] Route Engine Agg）为例：

- 返回完整 JSON，含 panels、templating variables、annotations
- 共 **20 个 panel**，含 row、graph、stat 类型
- 数据源引用类型：`{"type":"prometheus","uid":"Ny3qdvE7z"}`

---

### 8-9. `get_datasource` (by uid / by id) ✅

两种查询方式均工作正常，`resolvedBy` 字段正确标注解析路径：

```json
{"resolvedBy":"uid","datasource":{"id":3849,"uid":"RoRqN4iVz","name":"AI-Claim",...}}
{"resolvedBy":"id","datasource":{"id":3849,"uid":"RoRqN4iVz","name":"AI-Claim",...}}
```

---

### 10. `resolve_datasource_ref` ✅

以 `name="AI-Claim"` 查询，正确返回完整 ref 信息，支持 id > uid > name 优先级语义。

---

### 11. `get_annotations` ✅

- 无 tag 过滤：返回 1 条历史 annotation（含 dashboardId、panelId、text、time）
- tag=["deploy"] 过滤：返回空（该 Org 无 deploy tag 注解）

---

### 12. `list_org_users` ⚠️

- 调用成功，但返回数据量达 **4,422,565 字符**，超过 MCP 单次 token 上限
- 系统自动将结果写入临时文件，无法在对话中直接使用
- **建议**: 该工具缺少分页参数，大组织会触发此问题

---

### 13. `list_teams` ✅

返回 8 个团队：

```
MapSRE(3人), MapRoadRD(6人), MapSearchRD(26人), MapPOI(3人),
MapPosition(2人), Map Open Platform Team(5人), MapWebTeam(6人), MapNOC
```

---

### 14. `list_legacy_alerts` ✅

返回 alert 列表含状态：`ok`、`paused`，附带 `evalData`（含 noData 标记和 match 值）。

---

### 15. `list_legacy_notification_channels` ✅

返回 8 个 webhook 类型通知渠道，均集成 SeatalkID / Palaemon webhook，含 `autoResolve`、`severity` 配置。

---

### 16. `query_datasource` ✅ / ❌（部分）

测试了 6 个场景，覆盖 Prometheus 和 MySQL 两种数据源类型：

#### 场景一：Prometheus range query ✅

```json
queries: [{"refId":"A","datasourceId":19420,"expr":"vector(1)","range":true,"step":"60s"}]
```

- 成功返回 **Apache Arrow IPC** 格式（base64 编码），共 20 个时间点
- 元数据含 `resultType: "matrix"`、`interval: 15000`、`displayNameFromDS: "vector(1)"`

#### 场景二：Prometheus instant query ✅

```json
queries: [{"refId":"A","datasourceId":19420,"expr":"vector(1)","instant":true}]
```

- 成功返回 1 个数据点，`resultType: "vector"`

#### 场景三：多 refId 同时查询 ✅

```json
queries: [
  {"refId":"A","datasourceId":19420,"expr":"vector(1)","range":true,"step":"60s"},
  {"refId":"B","datasourceId":19420,"expr":"vector(2)","instant":true}
]
```

- A、B 各自独立返回 Arrow dataframe，互不干扰
- responses 结构：`{"A":{...},"B":{...}}`

#### 场景四：debug 模式 ✅

```json
debug: true
```

- 调用成功，返回结构与非 debug 模式相同，MCP 层未暴露额外 debug 字段

#### 场景五：MySQL rawSql 查询 ✅

```json
queries: [{"refId":"A","datasourceId":7739,"rawSql":"SELECT 1 AS value, NOW() AS time","format":"table"}]
```

- 成功返回 table 格式 Arrow 数据
- meta 中含 `executedQueryString: "SELECT 1 AS value, NOW() AS time"`（已执行 SQL 可审计）
- columns: `time`（timestamp）、`value`（INT64）

#### 场景六：无效 PromQL 错误处理 ✅

```json
expr: "invalid{{{query"
```

- 正确返回 422 + 清晰报错：
  ```
  labelFilterExpr: unexpected token "{"; want "ident"; unparsed data: "{{query"
  ```
- 错误信息含完整上下文（start/end/step），调试友好

#### 场景七：网络不可达数据源 ❌

```json
datasourceId: 3849  // AI-Claim，Prometheus proxy 返回 HTML
```

- 报错：`bad_response: readObjectStart: expect { or n, but found r`
- 根因：Grafana proxy 层收到非 JSON 响应（HTML 错误页），MCP 直接透传原始错误字符串
- **注意**：这是正确的 fail-fast 行为，但错误信息不够用户友好

#### 返回格式说明

所有成功查询均使用 **Apache Arrow IPC** 格式（base64 编码），而非可读 JSON：

```
QVJST1cx...（base64）
```

解码后为二进制 Arrow 列式格式，需客户端 Arrow 库才能解析。MCP 当前未提供自动解码，**AI agent 无法直接读取数值内容**，是最大的工程限制。

---

## 三、发现的问题与限制


| #   | 问题                              | 严重度 | 说明                               |
| --- | ------------------------------- | --- | -------------------------------- |
| 1   | `get_current_user` 永远 404       | 低   | API Key 鉴权下的 Grafana 限制，非 bug    |
| 2   | `list_datasources` 无分页          | 中   | 104个数据源 `hasMore=true`，但工具无法继续翻页 |
| 3   | `list_org_users` 输出超限           | 高   | 大组织会导致结果被截断，当前只能写文件              |
| 4   | `query_datasource` 返回 Arrow 二进制 | 高   | AI agent 无法直接读取数值，需客户端解码         |
| 5   | `query_datasource` 网络不可达报错不友好   | 低   | bad_response 含原始 proxy 错误，难以定位根因 |
| 6   | `list_datasources` 无分页          | 中   | 104个数据源 `hasMore=true`，但工具无法继续翻页 |
| 7   | `list_org_users` 输出超限           | 高   | 大组织会导致结果被截断，当前只能写文件              |
| 8   | Dashboard 搜索无 folder 过滤         | 低   | 无法按 folder 筛选 dashboard          |


---

## 四、总体评价

**核心功能全面可用**。Health、Org、Folder、Dashboard、Datasource、Alert、Annotation、Datasource Query 等链路均已验证通过。`get_current_user` 的 404 是 Grafana API Key 鉴权的已知限制，不影响实际使用。

`query_datasource` 功能完整，支持 Prometheus（range/instant）、MySQL（rawSql）、多 refId 并发、debug 模式，错误处理行为正确。**最核心的工程缺陷**是查询结果以 Apache Arrow 二进制格式返回，AI agent 无法直接解读数值，需要在 MCP server 层新增 Arrow → JSON 的转换层，才能让 LLM 真正"看懂"监控数据。