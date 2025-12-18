# 水电费计算系统 - 前端模块化架构

## 📁 目录结构

```
static/js/
├── config.js         # 全局配置和状态管理
├── utils.js          # 通用工具函数
├── api.js            # API 调用封装
├── ui.js             # UI 渲染和显示
├── form.js           # 表单处理
├── filter.js         # 筛选和排序
├── calculator.js     # 计算器功能
├── report.js         # 报表生成
├── continue.js       # 自动延续功能
├── extraFee.js       # 额外费用管理
├── batch.js          # 批量操作
├── main.js           # 主入口文件
└── README.md         # 本文档
```

## 📦 模块说明

### 1. config.js - 配置管理
- **职责**: 全局配置和状态管理
- **导出**:
  - `API_BASE_URL`: API 基础URL
  - `state`: 全局状态对象（记录数据、编辑ID、排序方式等）
  - `resetExtraFeeCounter()`: 重置额外费用计数器
  - `resetBatchExtraFeeCounter()`: 重置批量额外费用计数器

### 2. utils.js - 工具函数
- **职责**: 提供通用工具函数
- **导出**:
  - `showNotification(message, type)`: 显示通知消息
  - `getSelectedIds()`: 获取选中的记录ID列表
  - `clearSelection()`: 清除所有选择

### 3. api.js - API 调用
- **职责**: 封装所有后端API交互
- **导出**:
  - `fetchRecords(sortOrder)`: 获取所有记录
  - `fetchRecordById(id)`: 获取单个记录
  - `createRecord(data)`: 创建新记录
  - `updateRecord(id, data)`: 更新记录
  - `deleteRecord(id)`: 删除记录
  - `batchDeleteRecords(ids)`: 批量删除
  - `exportToExcel(ids)`: 导出Excel
  - `batchSetExtraFees(ids, extraFees, mode)`: 批量设置额外费用
  - `fetchLatestRecord(roomNumber)`: 获取最新记录
  - `continueRecord(roomNumber, newMonth)`: 单户自动延续
  - `batchContinueRecords(roomNumbers, newMonth)`: 批量自动延续

### 4. ui.js - UI 渲染
- **职责**: 页面显示和更新
- **导出**:
  - `displayRecords(records)`: 显示记录列表
  - `updateStatistics(records)`: 更新统计信息
  - `updateSelectedStatistics()`: 更新选中记录统计
  - `populateRoomFilter()`: 填充房号筛选下拉框
  - `toggleSelectAll()`: 全选/取消全选
  - `clearSelectionUI()`: 清除选择UI

### 5. form.js - 表单处理
- **职责**: 表单的显示、编辑和提交
- **导出**:
  - `showAddForm()`: 显示添加表单
  - `editRecord(id)`: 编辑记录
  - `closeModal()`: 关闭模态框
  - `setupFormHandler(onSuccess)`: 设置表单提交处理器

### 6. filter.js - 筛选排序
- **职责**: 数据筛选和排序功能
- **导出**:
  - `applyFilters()`: 应用筛选条件
  - `clearFilter()`: 清除筛选
  - `sortByRoom(order, onSuccess)`: 按房号排序

### 7. calculator.js - 计算器
- **职责**: 快速计算器功能
- **导出**:
  - `showCalculator()`: 显示计算器
  - `closeCalculator()`: 关闭计算器
  - `setupCalculator()`: 设置计算器

### 8. report.js - 报表生成
- **职责**: 报表图片生成
- **导出**:
  - `generateReport()`: 生成报表
  - `generateSelectedReport()`: 生成选中记录报表
  - `generateSingleCard(id)`: 生成单个卡片

### 9. continue.js - 自动延续
- **职责**: 自动延续功能
- **导出**:
  - `showContinueForm()`: 显示自动延续表单
  - `closeContinueModal()`: 关闭模态框
  - `toggleContinueMode()`: 切换单户/批量模式
  - `selectAllRooms()`: 全选房号
  - `deselectAllRooms()`: 取消全选房号
  - `updateSelectedRoomsCount()`: 更新已选择房号数量
  - `previewContinue()`: 预览上月数据
  - `executeContinue(onSuccess)`: 执行自动延续

### 10. extraFee.js - 额外费用管理
- **职责**: 额外费用的添加、删除和管理
- **导出**:
  - `addExtraFeeInput(name, amount)`: 添加额外费用输入框
  - `removeExtraFeeInput(id)`: 删除额外费用输入框
  - `clearExtraFeeInputs()`: 清空额外费用输入框
  - `getExtraFees()`: 获取所有额外费用
  - `loadExtraFees(extraFees)`: 加载额外费用到表单
  - `addBatchExtraFeeInput(name, amount)`: 添加批量额外费用输入框
  - `removeBatchExtraFeeInput(id)`: 删除批量额外费用输入框
  - `clearBatchExtraFeeInputs()`: 清空批量额外费用输入框
  - `getBatchExtraFees()`: 获取批量额外费用

### 11. batch.js - 批量操作
- **职责**: 批量操作功能
- **导出**:
  - `handleBatchDelete(onSuccess)`: 批量删除处理
  - `handleExportToExcel()`: 导出Excel处理
  - `showBatchExtraFeeModal()`: 显示批量额外费用模态框
  - `closeBatchExtraFeeModal()`: 关闭模态框
  - `executeBatchExtraFee(onSuccess)`: 执行批量设置额外费用

### 12. main.js - 主入口
- **职责**: 应用初始化和模块协调
- **导出**:
  - `BillingApp`: 应用主类
  - 通过 `window.billingApp` 暴露所有功能给HTML

## 🔄 数据流

```
用户操作 (HTML)
    ↓
window.billingApp.xxx()
    ↓
main.js (BillingApp)
    ↓
具体功能模块 (form.js, filter.js, etc.)
    ↓
api.js (后端通信)
    ↓
ui.js (更新显示)
```

## 🎯 设计原则

1. **单一职责**: 每个模块负责特定功能
2. **低耦合**: 模块间通过明确的接口通信
3. **高内聚**: 相关功能集中在同一模块
4. **可维护性**: 代码结构清晰，易于理解和修改
5. **可扩展性**: 新功能可以独立添加新模块

## 🔧 如何添加新功能

1. **创建新模块文件**: 在 `static/js/` 目录下创建新文件
2. **导入依赖**: 使用 ES6 import 导入需要的模块
3. **实现功能**: 编写功能代码并导出
4. **注册到 main.js**: 在 `exposeGlobalMethods()` 中暴露方法
5. **更新 HTML**: 在需要的地方调用 `window.billingApp.newFunction()`

## 📝 示例：添加新功能

```javascript
// 1. 创建 newFeature.js
export function doSomething() {
    // 实现功能
}

// 2. 在 main.js 中导入
import { doSomething } from './newFeature.js';

// 3. 暴露给全局
window.billingApp = {
    ...
    doSomething
};

// 4. 在 HTML 中调用
<button onclick="window.billingApp.doSomething()">新功能</button>
```

## 🚀 优势

- ✅ **模块化**: 代码按功能分离，易于管理
- ✅ **可维护**: 修改某个功能不影响其他模块
- ✅ **可测试**: 每个模块可以独立测试
- ✅ **可复用**: 模块可以在其他项目中复用
- ✅ **团队协作**: 不同开发者可以并行开发不同模块

## 📚 技术栈

- **ES6 Modules**: 模块化系统
- **Async/Await**: 异步操作
- **Fetch API**: HTTP 请求
- **DOM API**: 页面操作

## 🔍 迁移说明

从原来的 `app.js` (1117行) 重构为 12 个模块文件，每个文件平均约 100-200 行，更易于维护和理解。

原 `app.js` 的所有功能都已保留，只是按职责重新组织到不同模块中。
