# 功能模块

## 1. 配置与应用装配

**职责**：读取运行配置，初始化 SQLite、迁移历史数据，构造仓储/服务/处理器，注册中间件和路由。

**依赖**：`config` -> `repository` -> `services` -> `handlers` -> Gin。

**输入**：`config.json` 中的端口、数据文件、SQLite 路径、导出/报表目录、密码哈希、Cookie 与 CORS 配置。

**输出**：可运行的 HTTP 服务、已初始化数据库连接和完整依赖图。

**状态**：未启动 -> 配置有效 -> 数据库就绪 -> 迁移/回填完成 -> 路由监听；任一步失败均终止启动。

**核心文件**：`cmd/server/main.go`、`internal/config/config.go`、`config.example.json`。

## 2. 认证与访问控制

**职责**：校验管理员密码、限制暴力尝试、创建和销毁会话、保护业务 API。

**依赖**：bcrypt、Gin、`internal/platform/response`；写请求还依赖 `cmd/server/security_middleware.go` 的来源校验。

**输入**：登录密码、会话 Cookie、客户端地址、请求 Origin。

**输出**：认证状态 JSON、会话 Cookie，或 400/401/429 响应。

**状态**：未认证 -> 密码通过 -> 会话有效 -> 过期/登出；失败计数达到 5 次 -> 锁定 15 分钟 -> 自动解锁。

**核心文件**：`internal/authentication/auth.go`、`internal/authentication/auth_test.go`、`cmd/server/security_middleware.go`、`web/pages/index.html`、`web/assets/js/auth.js`。

## 3. 账单核心

**职责**：账单查询、创建、编辑、删除、按月筛选、分页排序和不落库费用试算。

**依赖**：`BillingHandler` -> `BillingService` -> `BillingRepository` -> SQLite；输入通过 `dto.BillingRecordRequest` 校验，费用由 `models.BillingRecord` 计算。

**输入**：房号、账期、水电本期/上期读数、补差、单价、管理费和额外费用。

**输出**：完整账单，包含用量、水费、电费和总费用等派生字段。

**关键规则**：

- 房号和月份必填，月份格式为 `YYYY-MM`。
- 数值必须非负且有限；补差后本期读数不得小于上期读数。
- 单条账单最多 20 个额外费用项。
- 同一房号、同一账期唯一。
- 创建和更新时服务端重新计算所有派生金额。

**状态**：不存在 -> 已创建 -> 已更新 -> 已删除。额外费用随账单在同一事务内替换或级联删除。

**核心文件**：`internal/handlers/billing_handler.go`、`internal/services/billing_service.go`、`internal/dto/billing.go`、`internal/models/billing.go`、`internal/repository/billing_sqlite.go`、`web/assets/js/form.js`、`web/assets/js/main.js`。

## 4. 账单延续

**职责**：从房间最新历史账单生成目标月份的新账单，支持单个和批量房间。

**依赖**：`BillingHandler` -> `BillingService` -> 仓储的最新记录查询与创建/批量创建。

**输入**：房号或房号数组、目标月份。

**输出**：单条新账单，或批量创建数量。

**复制与重置规则**：

- 上期“本期读数”变为新账单“上期读数”。
- 新账单“本期读数”初始等于上期读数，因此初始用量为 0。
- 继承管理费和水电单价。
- 补差和额外费用清空。
- 目标月份必须晚于该房号最新账期，且不能已存在。
- `总表` 被明确禁止参与延续。

**状态**：历史账单存在 -> 构造待创建账单 -> 唯一性/月份检查 -> 新账单已创建；批量操作任一房号失败则整体回滚。

**核心文件**：`internal/handlers/billing_handler.go`、`internal/services/billing_service.go`、`internal/repository/billing_sqlite.go`、`web/assets/js/continue.js`。

## 5. 批量账单操作

**职责**：批量删除、设置补差、追加或替换额外费用。

**依赖**：`BatchHandler` 方法 -> `BillingService` -> SQLite 批量事务。

**输入**：1 至 500 个不重复正整数 ID；补差指针或额外费用及模式。

**输出**：成功处理数量和消息。

**状态**：选择记录 -> DTO 校验 -> 事务内确认所有记录存在 -> 更新并重算/删除 -> 提交。任一 ID 不存在或写入失败时回滚，不产生部分成功。

**核心文件**：`internal/handlers/batch_handler.go`、`internal/dto/billing.go`、`internal/repository/billing_sqlite.go`、`web/assets/js/batch.js`。

## 6. 智能水表匹配

**职责**：将一组无房号水表读数分配给选中的账单，使每户用水量非负且总用水量最小，并原子更新账单。

**依赖**：`BillingHandler.SmartWaterMatch` -> 读取账单 -> handler 内回溯算法 -> `BatchUpdateWaterReadings` -> SQLite。

**输入**：相同数量的账单 ID 和水表读数，最多 10 户。

**输出**：匹配数量和每户房号、读数、上期读数、用水量。

**状态**：选择账单/输入读数 -> 枚举有效排列并剪枝 -> 得到最小方案 -> 事务更新本期水读数及费用；无非负方案或更新失败时不修改记录。

**核心文件**：`internal/handlers/match_handler.go`、`internal/dto/billing.go`、`internal/repository/billing_sqlite.go`、`web/assets/js/smartMatch.js`。`internal/services/match_service.go` 当前未接入。

## 7. 数据导入导出

**职责**：批量导入账单，导出全部账单为 JSON 或 Excel，并安全下载生成文件。

**依赖**：`ExportHandler` -> DTO/模型校验 -> `BillingService`/仓储；Excel 使用 Excelize；文件路径由 `GeneratedFileStore` 管理。

**输入**：

- 导入：1 至 500 条 `BillingRecordRequest` JSON；浏览器可先解析 Excel 并转为 JSON。
- 导出：格式参数 `json` 或 `excel`。
- 下载：服务端生成的 basename。

**输出**：导入数量；导出下载 URL；JSON/XLSX 文件流。

**状态**：导入文件选择 -> 客户端解析预览 -> 服务端逐条校验 -> 事务批量创建 -> 刷新列表。导出为读取快照 -> 创建随机文件 -> 写入 -> 返回受限 URL -> 鉴权下载。

**核心文件**：`internal/handlers/export_handler.go`、`internal/services/generated_file_store.go`、`web/assets/js/excel-import.js`、`web/assets/js/api.js`。

## 8. 报表生成

**职责**：按记录 ID 或账期选择账单，排序后生成 PNG 账单报表并通过受限接口下载。

**依赖**：`ReportHandler` -> `BillingService` -> `ImageGenerator` -> `GeneratedFileStore`。

**输入**：记录 ID 数组或月份，可选按房号升/降序。

**输出**：生成数量、随机 basename 对应的下载 URL和 PNG 文件。

**状态**：选择范围 -> DTO 校验 -> 查询账单 -> 可选排序 -> 分配文件 -> 绘图 -> 返回 URL -> 鉴权下载。没有记录时返回 404；生成失败不返回可下载结果。

**核心文件**：`internal/handlers/report_handler.go`、`internal/services/image_generator.go`、`internal/services/generated_file_store.go`、`web/assets/js/report.js`。

## 9. 总表管理与补差

**职责**：按月份维护独立水/电总表读数，并在前端计算总表用量、住户分表合计、总差额和每户建议补差。

**依赖**：`TotalMeterHandler` -> `TotalMeterRepo` -> SQLite；补差计算还依赖账单按月查询。

**输入**：月份、水总表读数、电总表读数；补差计算需要当前月、上月总表及当前账期住户账单。

**输出**：总表记录；总表用量、住户用量合计、差额和每户建议分摊。

**关键规则**：

- 月份唯一，读数必须为非负有限值。
- 更新 URL 的月份覆盖请求体月份。
- 计算时排除房号为 `总表` 的旧账单记录。
- 每户建议补差 = `(总表当月用量 - 住户分表用量合计) / 当月住户数`。
- 建议补差只在页面展示，不自动写回账单。

**状态**：月份无记录 -> 创建；已有记录 -> 更新；删除后不存在。启动回填只补缺失月份，手工记录优先。

**核心文件**：`internal/handlers/total_meter_handler.go`、`internal/models/total_meter.go`、`internal/repository/total_meter_sqlite.go`、`web/pages/total-meter.html`、`web/assets/js/api.js`。

## 10. 统一响应与前端状态

**职责**：统一 API 包装，维护前端列表、筛选、分页和选择状态，处理会话失效。

**依赖**：后端 `internal/platform/response`；前端 `api.js` 和 `state.js`。

**输入**：HTTP 状态、后端响应体、用户筛选/选择操作。

**输出**：`success/data/message/error` JSON；页面列表和交互反馈。

**状态**：加载中 -> 成功渲染/错误提示；收到 401 -> 清理交互流程并跳转登录。

**核心文件**：`internal/platform/response/response.go`、`web/assets/js/api.js`、`web/assets/js/state.js`、`web/assets/js/ui.js`。
