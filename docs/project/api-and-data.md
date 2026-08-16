# 接口与数据

## 1. API 基础约定

基础前缀：`/api/v1`。

除登录外，以下接口均要求 `ginmeterbox_session` Cookie。统一 JSON 响应：

```json
{
  "success": true,
  "data": {},
  "message": "可选成功消息",
  "error": "可选错误消息"
}
```

常用状态码：200 查询/更新成功，201 创建成功，400 参数或业务前置条件错误，401 未认证，404 资源不存在，429 登录锁定，500 内部错误。

## 2. 接口索引

### 认证

| 方法 | 路径 | 输入 | 输出/作用 |
|---|---|---|---|
| POST | `/auth/login` | `{password}` | 创建内存会话并设置 Cookie |
| POST | `/auth/logout` | Cookie | 删除会话并清理 Cookie |
| GET | `/auth/session` | Cookie | 返回 `authenticated=true` |

### 账单核心

| 方法 | 路径 | 输入 | 输出/作用 |
|---|---|---|---|
| GET | `/billing` | 可选排序、分页查询参数 | 账单数组或分页对象 |
| GET | `/billing/:id` | 正整数 ID | 单条账单 |
| POST | `/billing` | `BillingRecordRequest` | 创建并返回账单 |
| PUT | `/billing/:id` | ID + `BillingRecordRequest` | 更新并返回账单 |
| DELETE | `/billing/:id` | ID | 删除账单和额外费用 |
| GET | `/billing/month` | `month=YYYY-MM` | 指定月账单数组 |
| GET | `/billing/latest/:room` | 房号 | 房号最新账单 |
| POST | `/billing/calculate` | `BillingRecordRequest` | 试算，不保存 |
| POST | `/billing/continue` | `{roomNumber,newMonth}` | 延续单户账单 |
| POST | `/billing/batch-continue` | `{roomNumbers,newMonth}` | 原子延续多户账单 |

### 批量与匹配

| 方法 | 路径 | 输入 | 输出/作用 |
|---|---|---|---|
| POST | `/billing/batch-delete` | `{ids}` | 原子批量删除 |
| POST | `/billing/batch-adjustment` | `{ids,waterAdjustment?,electricAdjustment?}` | 批量补差并重算 |
| POST | `/billing/batch-extra-fee` | `{ids,extraFees,mode}` | 追加/替换额外费用 |
| POST | `/billing/smart-match` | `{ids,waterReadings}` | 匹配读数并原子更新 |

### 导入、导出与报表

| 方法 | 路径 | 输入 | 输出/作用 |
|---|---|---|---|
| POST | `/billing/import` | 账单请求数组 | 原子批量导入 |
| GET | `/billing/export` | `format=json|excel` | 返回生成文件 URL |
| GET | `/billing/export/download` | `file=basename` | 下载 JSON/XLSX |
| POST | `/billing/generate-report` | `{ids?,month?,sortBy?,order?}` | 生成 PNG 报表 URL |
| GET | `/billing/download` | `file=basename` | 下载 PNG |

### 总表

| 方法 | 路径 | 输入 | 输出/作用 |
|---|---|---|---|
| GET | `/total-meter` | 无 | 全部总表记录 |
| GET | `/total-meter/month` | `month=YYYY-MM` | 指定月份总表 |
| POST | `/total-meter` | `{month,waterReading,electricReading}` | 创建总表记录 |
| PUT | `/total-meter/:month` | 路径月份 + 两项读数 | 更新该月总表 |
| DELETE | `/total-meter/:month` | 月份 | 删除该月总表 |

## 3. 核心数据模型

### BillingRecord

| 字段 | 含义 | 来源 |
|---|---|---|
| `id` | 账单主键 | SQLite |
| `roomNumber` | 房号；与账期组成唯一键 | 输入 |
| `currentWater` / `previousWater` | 本期/上期水表读数 | 输入或延续 |
| `waterAdjustment` | 水补差 | 输入/批量设置 |
| `waterUsage` | 水用量 | 服务端计算 |
| `currentElectric` / `previousElectric` | 本期/上期电表读数 | 输入或延续 |
| `electricAdjustment` | 电补差 | 输入/批量设置 |
| `electricUsage` | 电用量 | 服务端计算 |
| `managementFee` | 管理费 | 输入 |
| `waterPrice` / `electricPrice` | 水电单价 | 输入 |
| `extraFees` | 最多 20 项附加费用 | 输入/批量设置 |
| `totalWaterCost` / `totalElectricCost` | 水费/电费 | 服务端计算 |
| `totalCost` | 全部费用合计 | 服务端计算 |
| `billingMonth` | `YYYY-MM` 账期 | 输入 |
| `createdAt` / `updatedAt` | 创建/更新时间 | 仓储 |

持久化拆分为 `billing_records` 和 `billing_extra_fees`。删除主记录时从表级联删除。

### ExtraFee

| 字段 | 约束 |
|---|---|
| `name` | 去空白后非空，最长 64 字节 |
| `amount` | 非负有限数值 |

### TotalMeterRecord

| 字段 | 含义/约束 |
|---|---|
| `id` | 主键 |
| `month` | `YYYY-MM`，数据库唯一 |
| `waterReading` | 非负有限水总表读数 |
| `electricReading` | 非负有限电总表读数 |
| `createdAt` / `updatedAt` | 创建/更新时间 |

## 4. 输入校验总表

| 项目 | 规则 |
|---|---|
| 月份 | 严格 `YYYY-MM`，月份 01 至 12 |
| 房号 | 最长 64；首字符为中英文或数字；后续允许空格、`_#-` |
| 账单数值 | 非负且非 NaN/Inf |
| 读数关系 | 本期读数 + 补差不得小于上期读数 |
| 额外费用 | 最多 20 项；名称最长 64；金额非负有限 |
| 批量 ID | 1 至 500 个正整数，不得重复 |
| 批量房号 | 1 至 500 个；逐个校验；排除 `总表` |
| 智能匹配 | 最多 10 户；ID 与读数等长；读数非负有限 |
| 分页 | page >= 1；pageSize 为 1 至 100 |
| 报表 | ID 或月份至少一个；仅支持房号排序和 asc/desc |

## 5. 数据库约束与事务

- 唯一索引：`billing_records(room_number,billing_month)`。
- 唯一约束：`total_meter_records.month`。
- 外键：额外费用关联账单并 `ON DELETE CASCADE`。
- 批量创建、导入、延续、删除、补差、额外费用和智能匹配均以事务实现全成全败。
- 单条创建/更新也在同一事务中处理账单主表和额外费用。
- 费用派生字段在写入前重新计算，不信任请求中的计算值。

## 6. 业务错误映射

| 仓储/业务情况 | HTTP | 对外语义 |
|---|---:|---|
| `ErrRecordNotFound` | 404 | 账单不存在；批量操作不执行任何修改 |
| `ErrBillingPeriodExists` | 400 | 同房号目标月份已有账单 |
| `ErrBillingMonthNotNext` | 400 | 延续月份必须晚于最新账期 |
| `ErrTotalMeterRecordNotFound` | 404 | 指定月份总表不存在 |
| `ErrTotalMeterMonthExists` | 400 | 指定月份总表已存在 |
| 无智能匹配方案 | 400 | 所有分配都会产生负用水量 |
| 非法生成文件标识 | 400 | basename/扩展/路径边界不合法 |
| 生成文件不存在 | 404 | 受限目录内没有对应普通文件 |

## 7. 核心目录索引

| 路径 | 内容 |
|---|---|
| `cmd/server/` | HTTP 服务入口、应用装配和中间件 |
| `cmd/migrate/` | JSON 到 SQLite 的独立迁移工具 |
| `internal/config/` | 配置结构和加载 |
| `internal/authentication/` | 登录、会话和认证中间件 |
| `internal/handlers/` | HTTP 适配、参数和错误映射 |
| `internal/dto/` | API 请求结构与校验 |
| `internal/services/` | 账单业务、图像生成和文件存储 |
| `internal/repository/` | 仓储接口、SQLite/JSON 实现和迁移 |
| `internal/models/` | 账单、额外费用和总表实体 |
| `internal/platform/` | 统一响应与共享错误定义 |
| `web/pages/` | HTML 页面与诊断页面 |
| `web/assets/` | 样式和浏览器业务脚本 |
| `data/` | 本地数据库和历史 JSON，运行数据不应提交 |
| `exports/` | 生成的 JSON/XLSX，不应提交 |
| `reports/` | 生成的 PNG，不应提交 |
| `docs/project/` | 当前系统维护文档 |
| `memory/` | 设计决策和未来架构规划 |
| `diagram/` | 架构图资源 |

## 8. 测试覆盖位置

- `internal/authentication/auth_test.go`：登录、会话、Cookie 和锁定。
- `internal/config/config_test.go`：配置加载和边界。
- `internal/dto/billing_test.go`：月份、ID、账单和批量输入校验。
- `internal/handlers/*_test.go`：当前主要位于仓库根目录的 handler/路由测试。
- `internal/repository/*_test.go`：SQLite 完整性、批量原子性、迁移和总表。
- `internal/services/billing_service_test.go`：延续与业务错误传播。
- `internal/services/generated_file_store_test.go`：生成文件路径安全。
- `security_middleware_test.go`：安全头和写请求来源校验。
