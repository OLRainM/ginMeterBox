# 系统架构

## 1. 系统定位

本项目是一个 Go 单体水电计费应用。浏览器使用原生 HTML/CSS/JavaScript，后端使用 Gin 提供静态页面和 JSON API，业务数据保存到 SQLite，报表和导出结果保存为受控本地文件。

## 2. 技术栈

| 层次 | 技术 | 用途 |
|---|---|---|
| 前端 | 原生 HTML/CSS/JavaScript | 账单、批量操作、导入导出、报表、总表页面 |
| HTTP | Gin 1.9 | 路由、中间件、JSON 绑定、静态资源和文件下载 |
| 业务 | Go service/handler | 费用计算、账单延续、批量事务、智能匹配 |
| 数据 | `database/sql` + `modernc.org/sqlite` | 账单、额外费用、总表持久化 |
| 文件 | Excelize、fogleman/gg | Excel 导入导出和 PNG 报表生成 |
| 安全 | bcrypt、随机会话 Cookie | 管理员认证、登录限流和会话保护 |

## 3. 运行结构

```text
浏览器页面
  -> web/assets/js/*
  -> Gin 路由与中间件
  -> internal/handlers/* 或 internal/authentication/*
  -> internal/dto/* 输入校验
  -> internal/services/* 业务编排
  -> internal/repository/* 仓储接口
  -> SQLite / 生成文件目录
```

总表模块目前是例外：`TotalMeterHandler` 直接依赖 `TotalMeterRepo`，未经过 service 层。智能匹配算法也直接位于 handler 中，`internal/services/match_service.go` 当前未接入运行路由。

## 4. 分层职责与依赖方向

| 层 | 职责 | 允许依赖 |
|---|---|---|
| `web/` | 页面状态、表单交互、API 调用和结果展示 | HTTP API |
| `cmd/server/` | 组装配置、数据库、仓储、服务、处理器和路由 | 所有装配对象 |
| `internal/authentication/` | 登录、登出、会话校验和登录锁定 | `internal/platform/response` |
| `internal/handlers/` | HTTP 参数解析、DTO 校验、错误到状态码映射 | `dto`、`services`、`repository` 接口/错误 |
| `internal/dto/` | 请求结构和业务输入约束 | `models` |
| `internal/services/` | 费用计算、账单延续、批量业务和文件路径控制 | `models`、`repository` |
| `internal/repository/` | 查询、事务、唯一性和 SQLite 映射 | `models`、SQLite |
| `internal/models/` | 核心实体及费用计算 | 无上层依赖 |
| `internal/platform/response/` | 统一 JSON 响应 | Gin |

正常依赖方向是外层指向内层。仓储实现由 `cmd/server/main.go` 注入服务，业务层不直接创建数据库连接。

## 5. 启动链路

```text
main
  -> config.LoadConfig
  -> repository.OpenSQLite
  -> MigrateJSONToSQLiteIfNeeded
  -> BackfillLegacyMasterBillsToTotalMeters
  -> NewSQLiteBillingRepository / NewSQLiteTotalMeterRepository
  -> NewBillingService / NewGeneratedFileStore / NewImageGenerator
  -> NewBillingHandler / NewTotalMeterHandler / authentication.NewService
  -> 注册中间件、静态资源和 API 路由
  -> router.Run
```

启动关键状态：

1. 配置缺失、密码哈希为空、SQLite 打开失败或迁移失败时，进程直接终止。
2. SQLite 初始化创建三张表和索引，并检查重复的“房号 + 账期”。发现历史重复数据时拒绝启动。
3. 仅当账单表和总表表都为空时，才从 JSON 迁移；JSON 原文件不修改。
4. 旧账单中 `room_number = '总表'` 的月份只回填到缺失的独立总表记录，不覆盖手工维护数据。

核心文件：`cmd/server/main.go`、`internal/config/config.go`、`internal/repository/sqlite.go`、`internal/repository/sqlite_migration.go`。

## 6. 数据存储

### SQLite

- `billing_records`：账单主表。
- `billing_extra_fees`：账单额外费用，从表通过外键级联删除。
- `total_meter_records`：按月份唯一的总表读数。

连接策略：最大连接数和空闲连接数均为 1；启用外键、5 秒 busy timeout 和 WAL。

### JSON

`data_file` 与 `total_meter_file` 仅作为首次迁移来源和人工回滚依据。运行时仓储使用 SQLite。`internal/repository/billing_json.go`、`internal/repository/total_meter_json.go` 是保留实现，不由当前 `cmd/server/main.go` 装配。

### 生成文件

- 导出目录：JSON、XLSX。
- 报表目录：PNG。
- 文件名由服务端时间戳和 128 位随机值生成。
- 下载时校验 basename、扩展名、普通文件类型、符号链接和最终路径边界。

核心文件：`internal/services/generated_file_store.go`、`internal/handlers/export_handler.go`、`internal/handlers/report_handler.go`。

## 7. HTTP 与安全边界

全局中间件顺序包括日志、恢复、安全响应头、CORS。API 分为：

- 公开：`POST /api/v1/auth/login`。
- 受保护：登出、会话检查及全部业务 API，统一经过 `RequireAuth` 和写请求 Origin 校验。

会话状态仅保存在进程内存中，默认 8 小时；服务重启后全部失效。Cookie 为 HttpOnly、SameSite=Strict，可由配置启用 Secure。单个客户端连续失败 5 次后锁定 15 分钟。

浏览器 GET/HEAD/OPTIONS 不做来源拦截；其他带 Origin 的请求必须同源或在配置白名单中。无 Origin 的非浏览器客户端仍需有效会话。

核心文件：`internal/authentication/auth.go`、`cmd/server/security_middleware.go`、`cmd/server/main.go`。

## 8. 前端结构

- `web/pages/index.html`：账单主页面，包含登录遮罩。
- `web/pages/total-meter.html`：总表维护和补差计算页面，业务脚本内嵌。
- `web/assets/js/api.js`：统一 API 请求封装，处理会话失效。
- `web/assets/js/state.js`：账单列表、筛选、选择和分页状态。
- `web/assets/js/form.js`：账单创建/编辑。
- `web/assets/js/continue.js`：账单延续。
- `web/assets/js/batch.js`：批量选择和批量更新。
- `web/assets/js/excel-import.js`：Excel 文件解析、预览和导入。
- `web/assets/js/smartMatch.js`：智能匹配交互。
- `web/assets/js/report.js`：报表生成和下载。
- `web/assets/js/calculator.js`：客户端费用预览。

前端客户端计算用于交互反馈，最终持久化金额以后端 `BillingRecord.CalculateCosts()` 为准。

## 9. 已知架构边界

- `internal/services/match_service.go` 未被生产调用；实际算法在 `internal/handlers/match_handler.go`。
- `POST /billing/calculate` 已提供，但当前前端计算器未调用该接口。
- 总表补差计算主要在 `total-meter.html` 前端完成，后端仅保存总表原始读数。
- 当前没有账单“已缴/未缴”等持久化生命周期状态；账单状态仅表现为不存在、已创建、已更新或已删除。
- 会话为单进程内存状态，不适合多实例共享。
