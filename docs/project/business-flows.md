# 业务链路

本文按“用户入口 -> HTTP -> 处理器 -> 业务/仓储 -> 持久化或输出”描述当前完整业务链路。

## 1. 应用启动与数据迁移

**输入**：`config.json`、SQLite 文件路径、历史账单 JSON、历史总表 JSON。

**调用链**：

```text
main
-> LoadConfig
-> OpenSQLite
-> 创建表/索引并检查房号+账期重复
-> MigrateJSONToSQLiteIfNeeded
-> BackfillLegacyMasterBillsToTotalMeters
-> 构造仓储、服务、处理器
-> 注册路由并监听端口
```

**输出**：可服务的应用和已初始化 SQLite。

**状态流转**：空数据库可从 JSON 整体迁移；非空数据库跳过 JSON 迁移。旧“总表”账单只补充缺失的独立总表月份。迁移和回填都使用事务，失败即回滚并终止启动。

## 2. 管理员登录

**输入**：`POST /api/v1/auth/login`，JSON `{password}`。

**调用链**：

```text
login.html / auth.js
-> authentication.Service.Login
-> 按客户端地址检查锁定状态
-> bcrypt 校验密码哈希
-> 生成 32 字节随机会话令牌
-> 写入内存 sessions
-> 设置 HttpOnly 会话 Cookie
```

**输出**：成功返回 `authenticated=true`；失败返回 400、401 或 429。

**状态流转**：未登录 -> 已登录；连续失败累计 -> 第 5 次后锁定 15 分钟；成功登录清空失败记录。登出删除服务端会话并使 Cookie 过期。会话到期或服务重启后回到未登录。

## 3. 受保护请求

**输入**：业务 API 请求、会话 Cookie、可选 Origin。

**调用链**：

```text
浏览器 api.js(fetch credentials=include)
-> RequireAuth 校验内存会话
-> validateWriteOrigin 校验非只读请求来源
-> 业务 handler
```

**输出**：通过后进入业务；无效会话返回 401，前端跳转登录；来源不允许返回 400。

## 4. 账单列表与筛选

**输入**：

- `GET /billing`，可选 `sortBy=room`、`order=asc|desc`、`page`、`pageSize`。
- `GET /billing/month?month=YYYY-MM`。

**调用链**：前端加载/筛选 -> `BillingHandler` -> `BillingService` -> `SQLiteBillingRepository` -> 查询账单和额外费用 -> 统一响应 -> `state.js` 更新页面状态。

**输出**：全部账单数组，或 `{items,page,pageSize,total}`；按月查询返回数组。

**状态流转**：页面加载中 -> 数据进入 `allRecords` -> 月份筛选/排序 -> 分页切片 -> 渲染；分页大小限定 1 至 100。

## 5. 创建账单

**输入**：`POST /billing` 的 `BillingRecordRequest`。

**调用链**：

```text
form.js 收集表单
-> api.js
-> BillingHandler.Create
-> DTO Validate
-> ToRecord
-> BillingService.Create
-> BillingRecord.CalculateCosts
-> SQLiteBillingRepository.Create
-> 事务插入主表和额外费用
```

**输出**：201 和服务端最终账单。

**状态流转**：表单草稿 -> 校验通过 -> 计算派生字段 -> 已持久化。若同房号同月份已存在、读数回退、费用非法或写入失败，则保持原数据库状态。

**费用公式**：

```text
水用量 = 本期水读数 - 上期水读数 + 水补差
电用量 = 本期电读数 - 上期电读数 + 电补差
水费 = 水用量 * 水单价
电费 = 电用量 * 电单价
总费用 = 水费 + 电费 + 管理费 + 额外费用合计
```

## 6. 编辑账单

**输入**：`PUT /billing/:id` 和完整账单请求。

**调用链**：选择记录 -> 表单回填 -> handler 校验 ID/DTO -> service 重新计算 -> repository 事务更新主表并整体替换额外费用。

**输出**：更新后的账单。

**状态流转**：已存在 -> 编辑中 -> 已更新。目标不存在返回 404；修改房号/账期导致唯一性冲突返回 400；事务失败保留更新前状态。

## 7. 删除账单

**输入**：`DELETE /billing/:id`。

**调用链**：用户确认 -> handler 校验 ID -> service -> repository 事务删除 -> 外键级联删除额外费用。

**输出**：删除成功消息。

**状态流转**：已存在 -> 已删除；不存在返回 404。

## 8. 费用试算

**输入**：`POST /billing/calculate` 和账单请求。

**调用链**：handler DTO 校验 -> `ToRecord` -> `CalculateCosts`。

**输出**：包含派生费用的账单对象，不写数据库。

**说明**：接口已实现，但当前 `calculator.js` 在浏览器本地执行预览计算，没有调用该接口。

## 9. 单户账单延续

**输入**：`POST /billing/continue`，`{roomNumber,newMonth}`。

**调用链**：

```text
continue.js
-> BillingHandler.ContinueFromPrevious
-> DTO 校验房号、月份及排除“总表”
-> BillingService.ContinueFromPrevious
-> GetLatestByRoom
-> 校验目标月晚于最新月且不存在
-> 从历史账单构造新账单
-> repository.Create
```

**输出**：201 和新账单。

**状态流转**：有历史账单 -> 复制固定字段、滚动表底、清零用量/补差/额外费 -> 新月份账单。没有历史记录、月份不递增或目标已存在时不创建。

## 10. 批量账单延续

**输入**：`POST /billing/batch-continue`，1 至 500 个房号和目标月份。

**调用链**：前端选择住户 -> DTO 校验 -> service -> repository 在同一事务内逐户查最新记录、校验并插入。

**输出**：成功数量和空失败列表。

**状态流转**：待处理集合 -> 全部预条件满足 -> 全部创建；任一住户无历史、月份不递增、目标重复或写入失败 -> 全部回滚，响应明确“未创建任何记录”。

## 11. 批量删除

**输入**：`POST /billing/batch-delete`，`{ids}`。

**调用链**：批量选择 -> DTO 校验 ID -> repository 事务确认所有 ID 存在 -> 批量删除。

**输出**：删除数量。

**状态流转**：多条已存在记录 -> 全部删除；任一 ID 缺失 -> 全部保留。

## 12. 批量设置补差

**输入**：`POST /billing/batch-adjustment`，ID 数组及至少一个水/电补差值。

**调用链**：handler -> DTO -> service -> repository 事务读取每条记录 -> 修改指定补差 -> 重新计算费用 -> 更新。

**输出**：更新数量。

**状态流转**：旧补差/费用 -> 新补差 -> 新用量/费用。任一 ID 缺失或更新失败则整体回滚。

## 13. 批量额外费用

**输入**：`POST /billing/batch-extra-fee`，ID 数组、额外费用数组、`append|replace`；模式为空时默认为 `append`。

**调用链**：handler -> DTO -> service -> repository 事务读取 -> 追加或替换 -> 校验总项数 -> 重算总费用 -> 替换从表数据。

**输出**：更新数量。

**状态流转**：旧费用集合 -> 追加/替换后的集合 -> 总费用重算。超出限制、记录缺失或事务失败时全部不变。

## 14. 智能水表匹配

**输入**：`POST /billing/smart-match`，最多 10 个 ID 与等量无房号读数。

**调用链**：

```text
smartMatch.js
-> BillingHandler.SmartWaterMatch
-> 逐 ID 读取账单
-> 回溯枚举读数分配
-> 排除产生负用水量的候选
-> 以总用水量最小选择最佳方案
-> BatchUpdateWaterReadings
-> 事务更新本期水读数并重算费用
```

**输出**：匹配详情和更新数量。

**状态流转**：未匹配读数 -> 有效最优映射 -> 账单读数/水费/总费已更新。无有效方案或任一更新失败时不修改任何账单。

## 15. Excel 导入

**输入**：用户在浏览器选择 `.xlsx/.xls`，前端转换为最多 500 条账单请求，再调用 `POST /billing/import`。

**调用链**：文件选择 -> ExcelJS 解析与列映射 -> 预览 -> JSON 请求 -> handler 逐条 DTO 校验 -> service 批量创建 -> repository 单事务插入主表与额外费用。

**输出**：导入数量，随后刷新账单列表。

**状态流转**：文件未选择 -> 已解析 -> 待确认 -> 已导入。任一记录非法或唯一性冲突时整个导入失败，不产生部分数据。

## 16. JSON/Excel 导出与下载

**输入**：`GET /billing/export?format=json|excel`，随后访问返回的下载 URL。

**调用链**：读取全部账单 -> `GeneratedFileStore.NewExportFile` -> JSON 编码或 Excelize 写入 -> 返回 basename URL -> 下载接口校验 basename/路径/文件类型 -> `c.File`。

**输出**：`{format,count,url}` 和 JSON/XLSX 文件。

**状态流转**：数据库快照 -> 生成文件 -> 可下载。客户端不能指定服务端路径；路径穿越、非法扩展、符号链接或不存在文件均被拒绝。

## 17. PNG 报表生成与下载

**输入**：`POST /billing/generate-report`，提供 ID 数组或月份，以及可选房号排序。

**调用链**：DTO 校验 -> 按 ID 或月份查询 -> 排序 -> `NewReportFile` -> `ImageGenerator.GenerateReport` -> 返回下载 URL -> 受限下载。

**输出**：`{count,url}` 和 PNG。

**状态流转**：账单集合 -> 已排序 -> PNG 已生成 -> 已下载。空集合返回 404；生成失败不暴露无效路径。

## 18. 总表查询与维护

**输入**：

- `GET /total-meter`：全部记录。
- `GET /total-meter/month?month=YYYY-MM`：指定月份。
- `POST /total-meter`：创建。
- `PUT /total-meter/:month`：更新。
- `DELETE /total-meter/:month`：删除。

**调用链**：`total-meter.html` -> `TotalMeterHandler` -> DTO -> `SQLiteTotalMeterRepository` -> `total_meter_records`。

**输出**：总表记录或成功消息。

**状态流转**：月份不存在 -> 创建；已存在 -> 更新；已存在 -> 删除。月份唯一，重复创建返回 400，不存在的查询/更新/删除返回 404。

## 19. 月度总表补差计算

**输入**：选定月份、该月及上月总表读数、该月账单。

**调用链**：页面并行读取两个月总表和当前月账单 -> 排除房号 `总表` -> 计算总表月用量与住户分表合计 -> 计算差额和每户建议值 -> 页面展示。

**输出**：水电总表用量、住户合计、总差额、户数、每户建议补差及住户明细。

**状态流转**：月份未完整 -> 提示补录 -> 数据完整 -> 已计算。结果不写数据库，管理员需要通过账单补差功能另行确认和落库。
