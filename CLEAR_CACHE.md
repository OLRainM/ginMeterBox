# 🔄 清除浏览器缓存指南

## 问题症状

如果你看到以下错误：
```
Uncaught SyntaxError: The requested module './batch.js' does not provide an export named 'closeBatchAdjustmentModal'
```

这是因为浏览器缓存了旧版本的JavaScript文件。

## 解决方法

### 方法1：硬刷新（推荐）⚡

**Windows/Linux:**
- `Ctrl + F5`
- 或 `Ctrl + Shift + R`

**Mac:**
- `Cmd + Shift + R`

### 方法2：清除浏览器缓存 🧹

#### Chrome/Edge
1. 按 `Ctrl + Shift + Delete`
2. 选择"时间范围"：**全部时间**
3. 勾选"缓存的图片和文件"
4. 点击"清除数据"
5. 刷新页面

#### Firefox
1. 按 `Ctrl + Shift + Delete`
2. 选择"时间范围"：**全部**
3. 勾选"缓存"
4. 点击"立即清除"
5. 刷新页面

### 方法3：使用开发者工具 🛠️

1. 按 `F12` 打开开发者工具
2. 右键点击刷新按钮
3. 选择"清空缓存并硬性重新加载"

### 方法4：禁用缓存（开发时使用）

1. 按 `F12` 打开开发者工具
2. 点击 `Network` 标签
3. 勾选 `Disable cache`
4. 保持开发者工具打开状态

## 验证是否成功

清除缓存后，打开浏览器控制台（F12 → Console），应该：
- ✅ 没有红色错误信息
- ✅ 看到数据正常显示
- ✅ 所有功能正常工作

## 为什么会出现这个问题？

当我们添加新功能时：
1. 服务器上的文件已更新
2. 但浏览器仍使用缓存的旧文件
3. 新旧文件不匹配导致错误

## 预防措施

我已经在HTML文件中添加了版本号：
```html
<script type="module" src="/static/js/main.js?v=2.4.2"></script>
```

这样每次更新版本号后，浏览器会自动加载新文件。

## 快速测试

清除缓存后，在浏览器控制台运行：
```javascript
// 测试新功能是否可用
console.log(typeof window.billingApp.showBatchAdjustmentModal);
// 应该输出: "function"
```

## 仍然有问题？

如果清除缓存后仍有问题：

1. **检查服务器是否重启**
   ```powershell
   # 停止服务器 (Ctrl+C)
   # 重新启动
   go run main.go
   ```

2. **检查文件是否正确保存**
   ```powershell
   # 查看batch.js文件大小
   (Get-Item static/js/batch.js).Length
   
   # 应该大于5000字节
   ```

3. **使用诊断页面**
   访问：`http://localhost:8080/debug.html`

4. **尝试其他浏览器**
   如果Chrome有问题，试试Firefox或Edge

---

**提示：** 开发时建议保持开发者工具打开并禁用缓存，避免此类问题。
