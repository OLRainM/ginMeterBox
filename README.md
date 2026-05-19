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
- JSON 文件存储（Repository 接口化，可扩展为数据库）
- 图片生成：fogleman/gg
- Excel 导出：xuri/excelize

## 快速开始

```bash
go mod tidy
go run main.go
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
  }
}
```

不提供 `config.json` 时使用内置默认值。

## 项目结构

```
go-ele/
├── main.go                    # 入口
├── config/config.go           # 配置管理
├── handlers/                  # HTTP 处理器
│   ├── billing_handler.go     # 账单 CRUD + 自动延续
│   ├── batch_handler.go       # 批量操作
│   ├── export_handler.go      # 导入导出
│   ├── report_handler.go      # 报表图片
│   ├── match_handler.go       # 智能匹配
│   └── total_meter_handler.go # 总表管理
├── services/                  # 业务逻辑
│   ├── billing_service.go
│   ├── match_service.go
│   └── image_generator.go
├── repository/                # 数据访问（接口 + JSON 实现）
│   ├── interface.go
│   ├── billing_json.go
│   └── total_meter_json.go
├── models/                    # 数据模型
├── pkg/
│   ├── response/              # 统一 JSON 响应
│   └── errors/                # 业务错误定义
├── static/                    # 前端页面
├── data/                      # 数据文件（gitignore）
├── reports/                   # 生成的报表（gitignore）
└── exports/                   # 导出文件（gitignore）
```

## API

基础路径：`/api/v1`

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
| GET | /billing/export | 导出 JSON |
| POST | /billing/export-excel | 导出 Excel |
| GET | /billing/report/generate | 生成报表图片 |
| GET | /billing/card/:id | 生成单卡片 |
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
