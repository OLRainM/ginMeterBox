# API 接口文档

## 基础信息

- **基础路径**: `http://localhost:8080/api/v1`
- **数据格式**: JSON
- **字符编码**: UTF-8

---

## 目录

- [基础操作](#基础操作)
- [图片生成](#图片生成)
- [自动延续](#自动延续)
- [批量操作](#批量操作)
- [系统接口](#系统接口)

---

## 基础操作

### 获取所有记录

**请求**
```
GET /billing
```

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| sortBy | string | 否 | 排序字段（room） |
| order | string | 否 | 排序方向（asc/desc） |

**响应示例**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "roomNumber": "101",
      "billingMonth": "2025-11",
      "waterUsage": 24.0,
      "electricUsage": 406.0,
      "totalCost": 417.52,
      "extraFees": [
        {
          "name": "水管维修费",
          "amount": 50.00
        }
      ]
    }
  ]
}
```

---

### 获取单条记录

**请求**
```
GET /billing/:id
```

**路径参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 记录ID |

**响应示例**
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
    "extraFees": [],
    "totalCost": 417.52,
    "billingMonth": "2025-11"
  }
}
```

---

### 创建记录

**请求**
```
POST /billing
```

**请求体**
```json
{
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
  "electricPrice": 0.72,
  "extraFees": [
    {
      "name": "水管维修费",
      "amount": 50.00
    }
  ]
}
```

**响应示例**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "roomNumber": "101",
    "totalCost": 467.52
  }
}
```

---

### 更新记录

**请求**
```
PUT /billing/:id
```

**路径参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 记录ID |

**请求体**（同创建记录）

---

### 删除记录

**请求**
```
DELETE /billing/:id
```

**响应示例**
```json
{
  "success": true,
  "message": "删除成功"
}
```

---

### 按月份查询

**请求**
```
GET /billing/month?month=2025-11
```

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| month | string | 是 | 月份（YYYY-MM格式） |

---

## 图片生成

### 生成批量报表

**请求**
```
GET /billing/report/generate
```

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| month | string | 二选一 | 月份（YYYY-MM格式） |
| ids | string | 二选一 | ID列表（逗号分隔） |
| sortBy | string | 否 | 排序字段（room） |
| order | string | 否 | 排序方向（asc/desc） |

**示例**
```
GET /billing/report/generate?month=2025-11&sortBy=room&order=asc
GET /billing/report/generate?ids=1,2,3&sortBy=room&order=desc
```

**响应示例**
```json
{
  "success": true,
  "data": {
    "filename": "reports/billing_2025-11_20251129010209.png",
    "count": 5,
    "message": "报表生成成功"
  }
}
```

---

### 生成单个卡片

**请求**
```
GET /billing/card/:id
```

**响应示例**
```json
{
  "success": true,
  "data": {
    "filename": "reports/card_101_20251129010227.png",
    "message": "卡片生成成功"
  }
}
```

---

## 自动延续

### 从上月创建新记录

**请求**
```
POST /billing/continue
```

**请求体**
```json
{
  "roomNumber": "101",
  "newMonth": "2025-12"
}
```

**响应示例**
```json
{
  "success": true,
  "message": "成功从上月数据创建新记录",
  "data": {
    "id": 2,
    "roomNumber": "101",
    "billingMonth": "2025-12"
  }
}
```

---

### 获取最新记录

**请求**
```
GET /billing/latest/:room
```

**路径参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| room | string | 是 | 房间号 |

---

## 批量操作

### 批量导入

**请求**
```
POST /billing/import
```

**请求体**
```json
[
  {
    "roomNumber": "101",
    "billingMonth": "2025-11",
    "currentWater": 2893,
    "previousWater": 2870,
    "waterPrice": 4.3,
    "currentElectric": 7660,
    "previousElectric": 7264,
    "electricPrice": 0.72,
    "managementFee": 22
  }
]
```

**响应示例**
```json
{
  "success": true,
  "message": "成功导入记录",
  "count": 10
}
```

---

### 导出数据

**请求**
```
GET /billing/export?filename=exports/backup.json
```

**查询参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| filename | string | 否 | 导出文件名 |

**响应示例**
```json
{
  "success": true,
  "message": "导出成功",
  "file": "exports/billing_export.json"
}
```

---

### 批量设置额外费用

**请求**
```
POST /billing/batch-extra-fee
```

**请求体**
```json
{
  "ids": [1, 2, 3],
  "extraFees": [
    {
      "name": "公共区域维护费",
      "amount": 50.00
    },
    {
      "name": "垃圾清运费",
      "amount": 20.00
    }
  ],
  "mode": "append"
}
```

**参数说明**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| ids | array | 是 | 要设置的记录ID列表 |
| extraFees | array | 是 | 额外费用列表 |
| mode | string | 否 | 操作模式：append（追加）或 replace（替换），默认append |

**响应示例**
```json
{
  "success": true,
  "message": "批量设置成功",
  "count": 3
}
```

---

## 系统接口

### 健康检查

**请求**
```
GET /health
```

**响应示例**
```json
{
  "status": "ok",
  "message": "Water and Electric Billing System is running"
}
```

---

## 错误响应

所有接口在失败时返回统一格式：

```json
{
  "success": false,
  "error": "错误描述信息"
}
```

**常见错误码**

| HTTP状态码 | 说明 |
|-----------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 数据类型

### ExtraFee（额外费用）

```typescript
{
  name: string;    // 费用名称
  amount: number;  // 费用金额
}
```

### BillingRecord（账单记录）

```typescript
{
  id: number;
  roomNumber: string;
  currentWater: number;
  previousWater: number;
  waterAdjustment: number;
  waterUsage: number;        // 自动计算
  currentElectric: number;
  previousElectric: number;
  electricAdjustment: number;
  electricUsage: number;     // 自动计算
  managementFee: number;
  waterPrice: number;
  electricPrice: number;
  totalWaterCost: number;    // 自动计算
  totalElectricCost: number; // 自动计算
  extraFees: ExtraFee[];     // 可选
  totalCost: number;         // 自动计算
  billingMonth: string;      // YYYY-MM格式
  createdAt: string;         // ISO 8601格式
  updatedAt: string;         // ISO 8601格式
}
```

---

## 使用建议

1. **创建记录时**：系统会自动计算 `waterUsage`、`electricUsage`、`totalWaterCost`、`totalElectricCost` 和 `totalCost`，无需手动提供

2. **额外费用**：
   - 可以为空数组 `[]` 或完全省略该字段
   - 额外费用会自动计入 `totalCost`
   - 建议费用名称不超过10个字符

3. **排序功能**：
   - 支持按房号排序
   - 应用于列表查询和报表生成
   - 排序参数可选，不影响其他功能

4. **批量操作**：
   - 批量导入会自动分配ID
   - 批量设置额外费用支持追加和替换两种模式
   - 操作前建议先导出备份

---

## 测试工具

推荐使用以下工具测试API：

- **PowerShell**: 使用 `Invoke-RestMethod`
- **curl**: 命令行工具
- **Postman**: 图形化API测试工具
- **前端界面**: http://localhost:8080

详细测试脚本请参考 `tests/scripts/` 目录。
