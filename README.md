# 💧⚡ 水电费计算管理系统

一个基于 Go 语言开发的前后端分离水电费计算管理系统，提供简洁美观的 Web 界面进行水电费的计算、记录和管理。

## ✨ 功能特点

### 核心功能
- 📊 **数据管理**：完整的CRUD操作（创建、读取、更新、删除）
- 🧮 **自动计算**：智能计算水电费用
  - 用水量 = 本月读数 - 上月读数 + 补差
  - 用电量 = 本月读数 - 上月读数 + 补差
  - 水费 = 用水量 × 水单价
  - 电费 = 用电量 × 电单价
  - 总费用 = 管理费 + 水费 + 电费 + 额外费用
- 🔍 **数据筛选**：按月份筛选查看记录
- 📈 **统计汇总**：实时显示总记录数、总费用、总水费、总电费
- 💾 **数据持久化**：使用 JSON 文件存储数据

### 高级功能
- 🖼️ **图片报表生成**：生成精美的水电费账单图片
  - 单用户详细卡片
  - 多用户批量报表
  - 支持自由选择用户
  - 动态显示额外费用
- 🔄 **自动延续功能**：从上月数据自动创建新月份记录
  - 自动继承水电表读数
  - 自动继承价格配置
  - 无需手动输入初始值
- 💵 **额外费用管理**：灵活的额外费用系统
  - 单个记录添加多项额外费用
  - 批量设置额外费用（追加/替换模式）
  - 自动计入总费用
  - 图片报表智能显示
- 🔼🔽 **房号排序**：按房号升序/降序排列
- 📥📤 **批量导入导出**：JSON格式批量数据处理

### 界面特性
- 🎨 **现代UI**：响应式设计，支持桌面和移动端
- ⚡ **快速计算器**：独立计算器功能，快速计算费用
- 🎯 **批量操作**：支持批量选择和批量设置
- 💡 **智能提示**：友好的操作提示和错误处理

## 📋 技术栈

### 后端
- **Go 1.21+**
- **Gin** - Web 框架
- **Gin-CORS** - 跨域支持
- **fogleman/gg** - 图形绘制库
- **golang/freetype** - 字体渲染

### 前端
- **HTML5/CSS3**
- **JavaScript (ES6+)**
- **原生 Fetch API** - HTTP 请求

## 🚀 快速开始

### 前置要求

- Go 1.21 或更高版本
- 现代浏览器（Chrome、Firefox、Safari、Edge）

### 安装步骤

1. **克隆或下载项目**

```bash
cd e:\project\go-ele
```

2. **安装依赖**

```bash
go mod tidy
```

3. **运行服务器**

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动

4. **访问应用**

在浏览器中打开：`http://localhost:8080`

## 📁 项目结构

```
go-ele/
├── main.go                      # 主程序入口
├── go.mod                       # Go 模块配置
├── go.sum                       # 依赖版本锁定
│
├── 📂 models/                   # 数据模型
│   └── billing.go              # 账单数据模型、额外费用模型
│
├── 📂 handlers/                 # HTTP 请求处理器
│   └── billing_handler.go      # 账单相关API处理器
│
├── 📂 services/                 # 业务服务层
│   └── image_generator.go      # 图片生成服务
│
├── 📂 storage/                  # 数据存储层
│   └── storage.go              # JSON文件存储实现
│
├── 📂 static/                   # 前端静态资源
│   ├── index.html              # 前端主页面
│   ├── app.js                  # 前端JavaScript逻辑
│   └── style.css               # 前端样式表
│
├── 📂 data/                     # 数据文件
│   ├── billing_records.json    # 账单记录数据
│   └── .gitkeep
│
├── 📂 reports/                  # 生成的报表图片
│   ├── billing_*.png           # 批量报表
│   ├── card_*.png              # 单个卡片
│   └── .gitkeep
│
├── 📂 docs/                     # 项目文档
│   ├── 项目结构说明.md
│   ├── api/                    # API文档
│   ├── features/               # 功能说明
│   │   ├── 新功能说明.md
│   │   ├── 前端功能说明.md
│   │   ├── 批量设置额外费用说明.md
│   │   └── 导入导出说明.md
│   └── user-guide/             # 用户指南
│
├── 📂 tests/                    # 测试文件
│   ├── scripts/                # 测试脚本
│   │   ├── test_new_features.ps1
│   │   ├── test_batch_extra_fee.ps1
│   │   └── test_batch_report.ps1
│   └── data/                   # 测试数据
│       ├── test_extra_fees.json
│       └── sample_import.json
│
├── .gitignore                   # Git忽略规则
└── README.md                    # 项目说明
```

## 🔌 API 接口

### 基础路径
```
http://localhost:8080/api/v1
```

### 接口列表

#### 基础操作
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/billing` | 获取所有记录（支持排序） |
| GET | `/billing/:id` | 根据ID获取记录 |
| POST | `/billing` | 创建新记录 |
| PUT | `/billing/:id` | 更新记录 |
| DELETE | `/billing/:id` | 删除记录 |
| GET | `/billing/month?month=2025-11` | 按月份查询 |
| POST | `/billing/calculate` | 计算费用（不保存） |

#### 高级功能
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/billing/report/generate` | 生成报表图片 |
| GET | `/billing/card/:id` | 生成单个卡片 |
| GET | `/billing/download` | 下载图片 |
| POST | `/billing/continue` | 自动延续（从上月创建新记录） |
| GET | `/billing/latest/:room` | 获取指定房号的最新记录 |
| POST | `/billing/import` | 批量导入JSON |
| GET | `/billing/export` | 导出为JSON |
| POST | `/billing/batch-extra-fee` | 批量设置额外费用 |

#### 系统
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |

#### 查询参数说明

**获取所有记录（带排序）**
```
GET /billing?sortBy=room&order=asc
```
- `sortBy`: 排序字段（目前支持 `room`）
- `order`: 排序方向（`asc` 升序，`desc` 降序）

**生成报表**
```
GET /billing/report/generate?month=2025-11&sortBy=room&order=asc
GET /billing/report/generate?ids=1,2,3&sortBy=room&order=desc
```
- `month`: 按月份生成
- `ids`: 按ID列表生成
- `sortBy`, `order`: 排序参数

### 请求示例

**创建记录**
```bash
curl -X POST http://localhost:8080/api/v1/billing \
  -H "Content-Type: application/json" \
  -d '{
    "roomNumber": "101",
    "billingMonth": "2025-11",
    "currentWater": 2893,
    "previousWater": 2870,
    "waterAdjustment": 1,
    "currentElectric": 7660,
    "previousElectric": 7264,
    "electricAdjustment": 10,
    "managementFee": 22,
    "waterPrice": 4.3,
    "electricPrice": 0.72
  }'
```

**获取所有记录**
```bash
curl http://localhost:8080/api/v1/billing
```

**按月份查询**
```bash
curl "http://localhost:8080/api/v1/billing/month?month=2025-11"
```

### 响应格式

**成功响应**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "roomNumber": "101",
    "currentWater": 2893,
    "previousWater": 2870,
    "waterAdjustment": 1,
    "waterUsage": 24,
    "currentElectric": 7660,
    "previousElectric": 7264,
    "electricAdjustment": 10,
    "electricUsage": 406,
    "managementFee": 22,
    "waterPrice": 4.3,
    "electricPrice": 0.72,
    "totalWaterCost": 103.2,
    "totalElectricCost": 292.32,
    "totalCost": 417.52,
    "billingMonth": "2025-11",
    "createdAt": "2025-11-29T00:00:00Z",
    "updatedAt": "2025-11-29T00:00:00Z"
  }
}
```

**错误响应**
```json
{
  "success": false,
  "error": "错误信息"
}
```

## 💡 使用说明

### 1. 配置价格

在页面左上角的价格配置区域设置：
- **水价**：默认 4.3 元/吨
- **电价**：默认 0.72 元/度

### 2. 添加记录

1. 点击"新增记录"按钮
2. 填写住户编号和缴费月份
3. 输入水表和电表的本月、上月读数
4. 如有需要，填写补差和管理费
5. 点击"保存"

系统会自动计算：
- 用水量和用电量
- 水费和电费
- 总费用

### 3. 编辑记录

点击记录行的"编辑"按钮，修改信息后保存

### 4. 删除记录

点击记录行的"删除"按钮，确认后删除

### 5. 筛选查询

使用月份筛选器选择特定月份，查看该月的所有记录

### 6. 快速计算器

点击"快速计算"按钮，打开计算器：
- 输入用水量、用电量和管理费
- 实时计算并显示各项费用
- 不会保存到数据库

## 🔧 开发指南

### 修改端口

编辑 `main.go` 文件，修改端口号：

```go
if err := r.Run(":8080"); err != nil { // 修改为其他端口，如 ":3000"
    log.Fatal("Failed to start server:", err)
}
```

同时修改 `static/app.js` 中的 API 地址：

```javascript
const API_BASE_URL = 'http://localhost:8080/api/v1'; // 修改端口
```

### 添加新字段

1. 在 `models/billing.go` 中添加字段
2. 更新 `CalculateCosts()` 方法的计算逻辑
3. 修改前端表单和显示逻辑

### 更换数据存储

当前使用 JSON 文件存储，可以替换为：
- SQLite
- MySQL
- PostgreSQL

修改 `storage/storage.go` 实现新的存储接口即可

## 📊 数据结构

### BillingRecord (账单记录)

```go
type ExtraFee struct {
    Name   string  `json:"name"`   // 费用名称（如"水管维修费"）
    Amount float64 `json:"amount"` // 费用金额
}

type BillingRecord struct {
    ID                int         // 记录ID
    RoomNumber        string      // 住户编号
    CurrentWater      float64     // 本月水表读数
    PreviousWater     float64     // 上月水表读数
    WaterAdjustment   float64     // 水表补差
    WaterUsage        float64     // 用水量（自动计算）
    CurrentElectric   float64     // 本月电表读数
    PreviousElectric  float64     // 上月电表读数
    ElectricAdjustment float64    // 电表补差
    ElectricUsage     float64     // 用电量（自动计算）
    ManagementFee     float64     // 管理费
    WaterPrice        float64     // 水单价
    ElectricPrice     float64     // 电单价
    TotalWaterCost    float64     // 水费（自动计算）
    TotalElectricCost float64     // 电费（自动计算）
    ExtraFees         []ExtraFee  // 额外费用列表（可选）
    TotalCost         float64     // 总费用（自动计算，包含额外费用）
    BillingMonth      string      // 缴费月份
    CreatedAt         time.Time   // 创建时间
    UpdatedAt         time.Time   // 更新时间
}
```

### 计算逻辑

```go
总费用 = 管理费 + 水费 + 电费 + 所有额外费用之和
```

## 🐛 常见问题

### 1. 端口已被占用

错误信息：`bind: address already in use`

解决方法：
- 修改端口号
- 或终止占用 8080 端口的程序

### 2. 跨域问题

如果前后端分离部署，确保 CORS 配置正确：

```go
r.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://your-frontend-domain.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
    AllowCredentials: true,
}))
```

### 3. 数据文件权限问题

确保程序对 `data/` 目录有读写权限

## 📝 更新日志

### v2.3.0 (2025-11-29)
- ✨ **新增**：Excel导出功能
  - 选中记录批量导出为Excel
  - 精美的表格样式（紫色表头）
  - 包含20个数据列的完整信息
  - 自动列宽优化
  - 时间戳文件命名
  - 即时浏览器下载
- 🔌 **API**：新增Excel导出接口
- 📦 **依赖**：添加excelize/v2库
- 📁 **目录**：新建exports目录
- 📚 **文档**：Excel导出功能说明

### v2.2.1 (2025-11-29)
- ✨ **新增**：批量删除记录功能
  - 支持多选批量删除
  - 二次确认防止误操作
  - 不可撤销明确警告
  - 删除后自动清除选择状态
- 🎨 **优化**：筛选控件样式升级
  - 房号下拉框圆角优化
  - 自定义紫色下拉箭头
  - 聚焦时边框高亮+阴影
  - 0.3s平滑过渡动画
- 🔌 **API**：新增批量删除接口
- 📚 **文档**：批量删除和样式优化说明

### v2.2.0 (2025-11-29)
- ✨ **新增**：按房号筛选功能
  - 房号下拉框自动填充和排序
  - 支持月份+房号组合筛选
  - 筛选状态保持
- ✨ **新增**：批量自动延续功能
  - 单户/批量双模式切换
  - 可视化房号多选列表
  - 全选/取消全选快捷操作
  - 部分成功处理和失败反馈
- 🔌 **API**：新增批量延续接口
- 📚 **文档**：筛选和批量延续功能说明

### v2.1.0 (2025-11-29)
- 🎨 **优化**：批量报表分辨率自适应
  - 智能布局策略（1-2条/3-6条/7-12条/13+条）
  - 自适应字体和元素缩放
  - 高度限制保护（最大30000px）
  - 完美支持1-100+条记录
- 📚 **文档**：批量报表分辨率优化说明

### v2.0.0 (2025-11-29)
- 💵 **新增**：额外费用管理系统
  - 单个记录添加多项额外费用
  - 批量设置额外费用（追加/替换模式）
  - 图片报表智能显示额外费用
- 🔼🔽 **新增**：房号排序功能（升序/降序）
- 📥📤 **新增**：批量导入导出JSON功能
- 🖼️ **优化**：图片生成功能
  - 修复emoji显示问题
  - 批量报表优化为详细卡片布局
  - 动态高度支持
- 🎨 **优化**：前端界面改进
  - 新增批量设置模态框
  - 额外费用显示优化
  - 操作提示优化
- 📚 **文档**：完善项目文档
  - 功能说明文档
  - API文档
  - 测试脚本
- 🔧 **修复**：编译错误和代码重构

### v1.1.0 (2025-11-29)
- 🖼️ **新增**：图片报表生成功能
  - 单用户详细卡片
  - 多用户批量报表
  - 美观的渐变设计
- 🔄 **新增**：自动延续功能
  - 从上月数据创建新记录
  - 自动继承水电表读数
- 📊 **新增**：图片下载功能

### v1.0.0 (2025-11-29)
- ✨ 初始版本发布
- 📊 完整的CRUD功能
- 🧮 自动计算功能
- 🎨 现代化UI设计
- 💾 JSON文件存储
- ⚡ 快速计算器
- 🔍 月份筛选功能

## 📄 许可证

MIT License

## 👨‍💻 作者

由 Windsurf Cascade AI 协助开发

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

如有问题，请通过 GitHub Issues 联系。
