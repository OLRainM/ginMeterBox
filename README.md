# ginMeterBox

基于 Go + Gin 的水电费计费管理系统，支持账单管理、报表生成、智能水表匹配等功能。

## 功能

- 账单 CRUD、按月份筛选、自动计算费用
- 自动延续（从上月数据创建新月份记录）
- 批量操作（删除、补差、额外费用）
- 智能水表匹配（最小用水量原则自动分配读数）
- 图片报表生成（单卡片 / 批量报表）
- 导入导出（JSON / Excel）
- 总表管理

## 技术栈

- Go 1.24 + Gin
- SQLite 运行时存储（JSON 仅用于首次迁移与人工回滚）
- 图片生成：fogleman/gg
- Excel 导出：xuri/excelize

## 快速开始

```bash
go mod tidy
go run ./cmd/server
```

默认启动在 `http://localhost:8080`。

## 配置

复制 `config.example.json` 为 `config.json` 自定义配置：

```json
{
  "server": { "port": ":8080" },
  "data": {
    "billingFile": "data/billing_records.json",
    "totalMeterFile": "data/total_meter_records.json"
  },
  "export": { "dir": "exports" },
  "report": { "dir": "reports" },
  "font": {
    "bold": "C:\\Windows\\Fonts\\msyhbd.ttc",
    "regular": "C:\\Windows\\Fonts\\msyh.ttc"
  },
  "security": {
    "adminPasswordHash": "$2a$12$replace-with-a-real-bcrypt-hash",
    "sessionCookieSecure": false,
    "allowedOrigins": []
  }
}
```

安全配置说明：

- `adminPasswordHash` 必须是管理员密码的 bcrypt 哈希；服务会拒绝使用空值或无效哈希启动，不能将明文密码写入配置。
- 可在安装 Apache 工具后使用 `htpasswd -bnBC 12 "" "你的管理员密码" | tr -d ':\n'` 生成 bcrypt 哈希；将输出完整复制到 `adminPasswordHash`。
- `sessionCookieSecure` 在 HTTPS 部署时必须设为 `true`；本机 `http://localhost` 调试可设为 `false`。
- `allowedOrigins` 默认空数组，即仅允许同源浏览器访问。只有确实存在独立前端域名时才加入精确来源，例如 `["https://billing.example.com"]`；不要填 `*`。
- 会话使用 HttpOnly、SameSite=Strict Cookie，默认有效期为 8 小时；服务重启后会话会失效，需要重新登录。
- 安全加固版本要求 `security.adminPasswordHash`，因此首次运行前必须复制并配置 `config.json`。

## 项目结构

```text
go-ele/
├── cmd/
│   ├── server/                # HTTP 服务入口、路由与中间件
│   └── migrate/               # JSON 到 SQLite 的独立迁移工具
├── internal/
│   ├── authentication/        # 登录、会话与访问控制
│   ├── config/                # 配置结构与加载
│   ├── dto/                   # API 输入结构与校验
│   ├── handlers/              # HTTP 处理与错误映射
│   ├── models/                # 领域模型与费用计算
│   ├── platform/              # 共享错误和统一响应
│   ├── repository/            # SQLite/JSON 仓储与迁移
│   └── services/              # 业务编排、报表与文件管理
├── web/
│   ├── pages/                 # HTML 页面
│   └── assets/                # CSS 与浏览器 JavaScript
├── docs/project/              # 当前系统维护文档
├── tests/                     # PowerShell 接口与安全测试脚本
├── diagram/                   # 架构图资源
├── memory/                    # 设计决策和未来规划
├── data/                      # 本地数据库与历史数据（不提交）
├── reports/                   # 生成的报表（不提交）
└── exports/                   # 导出文件（不提交）
```

## API

基础路径：`/api/v1`

### 认证

除 `POST /auth/login` 与 `GET /health` 外，所有 API 均要求有效的管理员会话 Cookie。浏览器访问首页后会显示登录层；脚本客户端需先调用 `POST /auth/login` 并保存服务器下发的 Cookie。

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /auth/login | 使用 JSON `{ "password": "..." }` 建立会话 |
| POST | /auth/logout | 注销当前会话 |
| GET | /auth/session | 校验当前会话 |

### 账单

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /billing | 获取所有（支持 sortBy=room&order=asc） |
| GET | /billing/:id | 根据 ID 获取 |
| POST | /billing | 创建 |
| PUT | /billing/:id | 更新 |
| DELETE | /billing/:id | 删除 |
| GET | /billing/month?month=2025-01 | 按月份查询 |
| POST | /billing/calculate | 计算费用（不保存） |
| POST | /billing/continue | 自动延续 |
| POST | /billing/batch-continue | 批量延续 |
| GET | /billing/latest/:room | 获取住户最新记录 |
| POST | /billing/batch-delete | 批量删除 |
| POST | /billing/batch-adjustment | 批量补差 |
| POST | /billing/batch-extra-fee | 批量额外费用 |
| POST | /billing/import | 批量导入 |
| GET | /billing/export | 由服务端生成 JSON，并返回受限下载 URL |
| GET | /billing/export/download?file=generated_... | 下载服务端生成的 JSON/Excel |
| POST | /billing/export-excel | 由服务端生成 Excel，并返回受限下载 URL |
| POST | /billing/report/generate | 生成报表图片 |
| POST | /billing/card/:id | 生成单卡片 |
| GET | /billing/download?file=xxx | 下载图片 |
| POST | /billing/smart-water-match | 智能水表匹配 |

### 总表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /total-meter | 获取所有 |
| GET | /total-meter/month?month=2025-01 | 按月份查询 |
| POST | /total-meter | 创建 |
| PUT | /total-meter/:month | 更新 |
| DELETE | /total-meter/:month | 删除 |

## 费用计算公式

```
用水量 = 本月水表 - 上月水表 + 水补差
用电量 = 本月电表 - 上月电表 + 电补差
水费 = 用水量 × 水单价
电费 = 用电量 × 电单价
总费用 = 管理费 + 水费 + 电费 + 额外费用之和
```


## 许可证

本项目采用 [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/) 许可证。

允许自由使用、修改和分享，但 **禁止商业用途**。
