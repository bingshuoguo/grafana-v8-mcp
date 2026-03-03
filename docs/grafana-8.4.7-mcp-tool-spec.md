# Grafana 8.4.7 MCP 接口清单与 Tool Schema（Breaking 版）

## 1. 目标与决策

本文件定义 **Grafana 8.4.7** 的 MCP 工具契约（默认契约）。

本轮已确认的决策：

1. **替换默认契约（允许 breaking）**。
2. **OpenAPI client 固定到 8.4.7 对应规格**。
3. **不引入兼容过渡期（无 alias、无双写）**。

---

## 2. 版本与规格来源

- Grafana 版本：`8.4.7`（`2022-04-20`）
- 规格来源：
  - `public/api-spec.json`
  - `public/api-merged.json`
- 参考：
  - <https://grafana.com/grafana/download/8.4.7>
  - <https://raw.githubusercontent.com/grafana/grafana/v8.4.7/public/api-spec.json>
  - <https://raw.githubusercontent.com/grafana/grafana/v8.4.7/public/api-merged.json>

---

## 3. 契约与实现约定

### 3.1 API 基本约定

- Base URL：`http(s)://<grafana>/api`
- 鉴权：
  - Basic Auth
  - `Authorization: Bearer <service_account_token>`

### 3.2 Datasource 兼容策略（8.4.7）

- 统一采用 **ID-first**：`id > uid > name`
- datasource proxy/query 优先基于 `datasource_id`
- 不把 `proxy/uid` 作为 8.4.7 稳定前提

### 3.3 MCP tool schema 生成约定

- `inputSchema` 由 Go 请求 struct + `jsonschema` tag 生成（`MustTool/ConvertTool`）
- `outputSchema` 在本文件中作为 **契约文档与测试基线**，不要求 MCP runtime 直接注册
- 错误处理：
  - 可恢复业务错误：`CallToolResult.IsError=true`
  - 鉴权/配置硬错误：`HardError`（协议级）

### 3.4 统一错误模型（文档契约）

```json
{
  "type": "object",
  "required": ["statusCode", "message"],
  "properties": {
    "statusCode": { "type": "integer" },
    "message": { "type": "string" },
    "status": { "type": "string" },
    "detail": { "type": "string" },
    "upstream": { "type": "object" }
  }
}
```

---

## 4. 工具清单

## 4.1 MVP（默认开启）

| Tool | Grafana API | R/W | 说明 |
| --- | --- | --- | --- |
| `get_health` | `GET /api/health` | R | 健康与版本信息 |
| `get_current_user` | `GET /api/user` | R | 当前用户 |
| `get_current_org` | `GET /api/org` | R | 当前组织 |
| `search_dashboards` | `GET /api/search` | R | 搜索 dashboard |
| `get_dashboard_by_uid` | `GET /api/dashboards/uid/{uid}` | R | dashboard 详情 |
| `upsert_dashboard` | `POST /api/dashboards/db` | W | 创建/更新 dashboard |
| `list_folders` | `GET /api/folders` | R | folder 列表 |
| `create_folder` | `POST /api/folders` | W | 创建 folder |
| `update_folder` | `PUT /api/folders/{folder_uid}` | W | 更新 folder |
| `list_datasources` | `GET /api/datasources` | R | datasource 列表 |
| `get_datasource` | `GET by id/uid/name` | R | 单 datasource 详情 |
| `resolve_datasource_ref` | 组合接口 | R | 统一解析 datasource 引用 |
| `query_datasource` | `POST /api/tsdb/query` | R | 通用查询（主链路） |
| `query_prometheus` | `GET /api/datasources/proxy/{id}/api/v1/query(_range)` | R | PromQL 查询（range/instant） |
| `list_prometheus_label_values` | `GET /api/datasources/proxy/{id}/api/v1/label/{label}/values` | R | 查询标签值（模板变量解析） |
| `list_prometheus_metric_names` | `GET /api/datasources/proxy/{id}/api/v1/label/__name__/values` | R | 指标名发现（regex + 分页） |
| `get_annotations` | `GET /api/annotations` | R | 注解查询 |
| `create_annotation` | `POST /api/annotations` | W | 创建注解 |
| `patch_annotation` | `PATCH /api/annotations/{id}` | W | 局部更新注解 |
| `list_legacy_alerts` | `GET /api/alerts` | R | legacy alerts |
| `list_legacy_notification_channels` | `GET /api/alert-notifications` | R | legacy 通知渠道 |
| `list_org_users` | `GET /api/org/users` | R | 当前组织用户 |
| `list_teams` | `GET /api/teams/search` | R | 团队列表 |

## 4.2 第二阶段（默认关闭）

- `query_datasource_expressions`（`POST /api/ds/query`）
- `update_annotation`（`PUT /api/annotations/{id}`）
- `delete_annotation`
- `get_annotation_tags`
- `list_ruler_rules`（基于 `api-merged.json`）
- `get_alertmanager_alerts`（基于 `api-merged.json`）

## 4.3 本次 breaking 变更（无过渡）

以下旧契约命名不再作为默认契约的一部分：

- `update_dashboard` -> `upsert_dashboard`
- `list_users_by_org` -> `list_org_users`
- `get_datasource_by_uid` / `get_datasource_by_name` -> `get_datasource`

---

## 5. 共享对象

```json
{
  "DatasourceRef": {
    "type": "object",
    "properties": {
      "id": { "type": "integer" },
      "uid": { "type": "string" },
      "name": { "type": "string" }
    }
  }
}
```

---

## 6. Tool Schema（MVP）

### 6.1 `get_health`

```json
{
  "name": "get_health",
  "inputSchema": { "type": "object", "properties": {} },
  "outputSchema": {
    "type": "object",
    "properties": {
      "database": { "type": "string" },
      "version": { "type": "string" },
      "commit": { "type": "string" }
    }
  }
}
```

### 6.2 `get_current_user`

```json
{
  "name": "get_current_user",
  "inputSchema": { "type": "object", "properties": {} },
  "outputSchema": {
    "type": "object",
    "properties": {
      "id": { "type": "integer" },
      "login": { "type": "string" },
      "email": { "type": "string" },
      "name": { "type": "string" },
      "isGrafanaAdmin": { "type": "boolean" }
    }
  }
}
```

### 6.3 `get_current_org`

```json
{
  "name": "get_current_org",
  "inputSchema": { "type": "object", "properties": {} },
  "outputSchema": {
    "type": "object",
    "properties": {
      "id": { "type": "integer" },
      "name": { "type": "string" }
    }
  }
}
```

### 6.4 `search_dashboards`

```json
{
  "name": "search_dashboards",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": { "type": "string" },
      "limit": { "type": "integer", "default": 50 },
      "page": { "type": "integer", "default": 1 },
      "tag": { "type": "array", "items": { "type": "string" } },
      "folderIds": { "type": "array", "items": { "type": "integer" } },
      "dashboardIds": { "type": "array", "items": { "type": "integer" } },
      "starred": { "type": "boolean" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "items": { "type": "array", "items": { "type": "object" } },
      "page": { "type": "integer" },
      "limit": { "type": "integer" },
      "hasMore": { "type": "boolean" }
    }
  }
}
```

### 6.5 `get_dashboard_by_uid`

```json
{
  "name": "get_dashboard_by_uid",
  "inputSchema": {
    "type": "object",
    "required": ["uid"],
    "properties": { "uid": { "type": "string" } }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "dashboard": { "type": "object" },
      "meta": { "type": "object" }
    }
  }
}
```

### 6.6 `upsert_dashboard`

```json
{
  "name": "upsert_dashboard",
  "inputSchema": {
    "type": "object",
    "required": ["dashboard"],
    "properties": {
      "dashboard": { "type": "object" },
      "folderId": { "type": "integer" },
      "folderUid": { "type": "string" },
      "overwrite": { "type": "boolean", "default": false },
      "message": { "type": "string" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "status": { "type": "string" },
      "id": {
        "oneOf": [
          { "type": "integer" },
          { "type": "string" }
        ]
      },
      "uid": { "type": "string" },
      "title": { "type": "string" },
      "url": { "type": "string" },
      "version": { "type": "integer" }
    }
  }
}
```

### 6.7 `list_folders`

```json
{
  "name": "list_folders",
  "inputSchema": {
    "type": "object",
    "properties": {
      "limit": { "type": "integer", "default": 1000 },
      "page": { "type": "integer", "default": 1 },
      "permission": { "type": "string", "enum": ["View", "Edit"], "default": "View" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": { "items": { "type": "array", "items": { "type": "object" } } }
  }
}
```

### 6.8 `create_folder`

```json
{
  "name": "create_folder",
  "inputSchema": {
    "type": "object",
    "required": ["title"],
    "properties": {
      "title": { "type": "string" },
      "uid": { "type": "string" },
      "parentUid": { "type": "string" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "id": { "type": "integer" },
      "uid": { "type": "string" },
      "title": { "type": "string" },
      "url": { "type": "string" }
    }
  }
}
```

### 6.9 `update_folder`

```json
{
  "name": "update_folder",
  "inputSchema": {
    "type": "object",
    "required": ["folderUid", "title"],
    "properties": {
      "folderUid": { "type": "string" },
      "title": { "type": "string" },
      "description": { "type": "string" },
      "version": { "type": "integer" },
      "overwrite": { "type": "boolean" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "id": { "type": "integer" },
      "uid": { "type": "string" },
      "title": { "type": "string" }
    }
  }
}
```

### 6.10 `list_datasources`

```json
{
  "name": "list_datasources",
  "inputSchema": {
    "type": "object",
    "properties": {
      "type": { "type": "string" },
      "limit": { "type": "integer", "default": 100 },
      "offset": { "type": "integer", "default": 0 }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "items": { "type": "array", "items": { "type": "object" } },
      "total": { "type": "integer" },
      "hasMore": { "type": "boolean" }
    }
  }
}
```

### 6.11 `get_datasource`

```json
{
  "name": "get_datasource",
  "inputSchema": {
    "type": "object",
    "properties": {
      "id": { "type": "integer" },
      "uid": { "type": "string" },
      "name": { "type": "string" }
    },
    "anyOf": [
      { "required": ["id"] },
      { "required": ["uid"] },
      { "required": ["name"] }
    ]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "resolvedBy": { "type": "string", "enum": ["id", "uid", "name"] },
      "datasource": { "type": "object" }
    }
  }
}
```

### 6.12 `resolve_datasource_ref`

```json
{
  "name": "resolve_datasource_ref",
  "inputSchema": {
    "type": "object",
    "properties": {
      "id": { "type": "integer" },
      "uid": { "type": "string" },
      "name": { "type": "string" }
    },
    "anyOf": [
      { "required": ["id"] },
      { "required": ["uid"] },
      { "required": ["name"] }
    ]
  },
  "outputSchema": {
    "type": "object",
    "required": ["id", "name", "type"],
    "properties": {
      "id": { "type": "integer" },
      "uid": { "type": "string" },
      "name": { "type": "string" },
      "type": { "type": "string" },
      "url": { "type": "string" }
    }
  }
}
```

### 6.13 `query_datasource`

```json
{
  "name": "query_datasource",
  "inputSchema": {
    "type": "object",
    "required": ["from", "to", "queries"],
    "properties": {
      "from": { "type": "string", "example": "now-1h" },
      "to": { "type": "string", "example": "now" },
      "debug": { "type": "boolean" },
      "datasource": {
        "type": "object",
        "properties": {
          "id": { "type": "integer" },
          "uid": { "type": "string" },
          "name": { "type": "string" }
        }
      },
      "queries": { "type": "array", "items": { "type": "object" } }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "raw": { "type": "object" },
      "responses": { "type": "object" },
      "hints": { "type": "array", "items": { "type": "string" } }
    }
  }
}
```

### 6.14 `get_annotations`

```json
{
  "name": "get_annotations",
  "inputSchema": {
    "type": "object",
    "properties": {
      "from": { "type": "integer" },
      "to": { "type": "integer" },
      "userId": { "type": "integer" },
      "alertId": { "type": "integer" },
      "alertUid": { "type": "string" },
      "dashboardId": { "type": "integer" },
      "dashboardUid": { "type": "string" },
      "panelId": { "type": "integer" },
      "limit": { "type": "integer" },
      "tags": { "type": "array", "items": { "type": "string" } },
      "type": { "type": "string", "enum": ["alert", "annotation"] },
      "matchAny": { "type": "boolean" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "items": { "type": "array", "items": { "type": "object" } }
    }
  }
}
```

### 6.15 `create_annotation`

```json
{
  "name": "create_annotation",
  "inputSchema": {
    "type": "object",
    "required": ["text"],
    "properties": {
      "dashboardId": { "type": "integer" },
      "dashboardUid": { "type": "string" },
      "panelId": { "type": "integer" },
      "time": { "type": "integer" },
      "timeEnd": { "type": "integer" },
      "tags": { "type": "array", "items": { "type": "string" } },
      "text": { "type": "string" },
      "data": { "type": "object" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "message": { "type": "string" },
      "id": { "type": "integer" }
    }
  }
}
```

### 6.16 `patch_annotation`

```json
{
  "name": "patch_annotation",
  "inputSchema": {
    "type": "object",
    "required": ["id"],
    "properties": {
      "id": { "type": "integer" },
      "time": { "type": "integer" },
      "timeEnd": { "type": "integer" },
      "tags": { "type": "array", "items": { "type": "string" } },
      "text": { "type": "string" },
      "data": { "type": "object" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": { "message": { "type": "string" } }
  }
}
```

### 6.17 `list_legacy_alerts`

```json
{
  "name": "list_legacy_alerts",
  "inputSchema": {
    "type": "object",
    "properties": {
      "dashboardId": { "type": "integer" },
      "panelId": { "type": "integer" },
      "query": { "type": "string" },
      "state": { "type": "string" },
      "limit": { "type": "integer" },
      "dashboardTag": { "type": "array", "items": { "type": "string" } }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "items": { "type": "array", "items": { "type": "object" } }
    }
  }
}
```

### 6.18 `list_legacy_notification_channels`

```json
{
  "name": "list_legacy_notification_channels",
  "inputSchema": {
    "type": "object",
    "properties": {
      "name": { "type": "string" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "items": { "type": "array", "items": { "type": "object" } }
    }
  }
}
```

### 6.19 `list_org_users`

```json
{
  "name": "list_org_users",
  "inputSchema": { "type": "object", "properties": {} },
  "outputSchema": {
    "type": "object",
    "properties": {
      "items": { "type": "array", "items": { "type": "object" } }
    }
  }
}
```

### 6.20 `list_teams`

```json
{
  "name": "list_teams",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": { "type": "string" },
      "page": { "type": "integer", "default": 1 },
      "perPage": { "type": "integer", "default": 100 }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "totalCount": { "type": "integer" },
      "page": { "type": "integer" },
      "perPage": { "type": "integer" },
      "teams": { "type": "array", "items": { "type": "object" } }
    }
  }
}
```

### 6.21 `query_prometheus`

```json
{
  "name": "query_prometheus",
  "inputSchema": {
    "type": "object",
    "required": ["expr", "start"],
    "properties": {
      "datasource": {
        "type": "object",
        "properties": {
          "id": { "type": "integer" },
          "uid": { "type": "string" },
          "name": { "type": "string" }
        }
      },
      "expr": { "type": "string", "description": "PromQL expression" },
      "start": { "type": "string", "description": "Start time: 'now-1h', 'now', or RFC3339" },
      "end": { "type": "string", "description": "End time (default: now)" },
      "step": { "type": "string", "description": "Step interval (default: auto). E.g. '30s', '1m', '1d'" },
      "queryType": { "type": "string", "enum": ["range", "instant"], "default": "range" }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "resultType": { "type": "string", "enum": ["matrix", "vector", "scalar", "string"] },
      "result": {
        "type": "array",
        "items": {
          "type": "object",
          "properties": {
            "metric": {
              "type": "object",
              "additionalProperties": { "type": "string" }
            },
            "values": {
              "type": "array",
              "items": {
                "type": "object",
                "properties": {
                  "time": { "type": "string" },
                  "value": { "type": "string" }
                }
              }
            },
            "value": {
              "type": "object",
              "properties": {
                "time": { "type": "string" },
                "value": { "type": "string" }
              }
            }
          }
        }
      },
      "hints": { "type": "array", "items": { "type": "string" } }
    }
  }
}
```

### 6.22 `list_prometheus_label_values`

```json
{
  "name": "list_prometheus_label_values",
  "inputSchema": {
    "type": "object",
    "required": ["labelName"],
    "properties": {
      "datasource": {
        "type": "object",
        "properties": {
          "id": { "type": "integer" },
          "uid": { "type": "string" },
          "name": { "type": "string" }
        }
      },
      "labelName": { "type": "string" },
      "start": { "type": "string" },
      "end": { "type": "string" },
      "limit": { "type": "integer", "default": 100 }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "labelName": { "type": "string" },
      "values": { "type": "array", "items": { "type": "string" } },
      "total": { "type": "integer" },
      "truncated": { "type": "boolean" }
    }
  }
}
```

### 6.23 `list_prometheus_metric_names`

```json
{
  "name": "list_prometheus_metric_names",
  "inputSchema": {
    "type": "object",
    "properties": {
      "datasource": {
        "type": "object",
        "properties": {
          "id": { "type": "integer" },
          "uid": { "type": "string" },
          "name": { "type": "string" }
        }
      },
      "regex": { "type": "string" },
      "limit": { "type": "integer", "default": 50 },
      "page": { "type": "integer", "default": 1 }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "metrics": { "type": "array", "items": { "type": "string" } },
      "total": { "type": "integer" },
      "page": { "type": "integer" },
      "hasMore": { "type": "boolean" }
    }
  }
}
```

---

## 7. 落地顺序建议

1. 先落地 `health/user/org/search/dashboard/folder/datasource`。
2. 加入 `resolve_datasource_ref`，统一 datasource 解析。
3. 实现 `query_datasource`（`/api/tsdb/query`）。
4. 实现 Prometheus 工具链：`list_prometheus_metric_names` -> `list_prometheus_label_values` -> `query_prometheus`。
5. 实现 annotations。
6. 实现 legacy alerting。
7. 第二阶段再接 `query_datasource_expressions` 与 unified alerting 扩展。

---

## 8. 兼容性备注

- 8.4.7 下不假设所有 `uid` 路由都可用，datasource 统一走 ID-first。
- Prometheus 查询统一走 datasource proxy：`/api/datasources/proxy/{id}/api/v1/*`。
- `query_datasource` 与 `query_datasource_expressions` 分离，避免隐式回退导致行为不确定。
- 本文档是 **v8.4.7 默认契约**；非本文档工具名不再视为默认可用。
