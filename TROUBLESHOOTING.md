# 🔧 故障排查指南

## 问题：数据不显示

如果你看到"暂无数据，请添加记录"，但实际上数据库中有数据，请按以下步骤排查：

### 步骤1：检查服务器是否运行

```powershell
# 检查进程
Get-Process | Where-Object {$_.ProcessName -like "*go*"}
```

如果没有看到 `___go_build_go_ele` 进程，请启动服务器：

```powershell
go run main.go
```

### 步骤2：测试API连接

运行测试脚本：

```powershell
.\test_api.ps1
```

应该看到：
- ✓ API连接成功
- ✓ 数据获取成功
- 记录数: XX

### 步骤3：使用诊断页面

在浏览器中打开：`http://localhost:8080/debug.html`

这个页面会自动测试：
1. API健康检查
2. 数据获取
3. JavaScript模块加载
4. 浏览器兼容性

### 步骤4：检查浏览器控制台

1. 打开主页面：`http://localhost:8080`
2. 按 `F12` 打开开发者工具
3. 切换到 `Console` 标签
4. 查看是否有红色错误信息

常见错误：
- **CORS错误**：跨域请求被阻止
- **模块加载失败**：JavaScript文件路径错误
- **网络错误**：无法连接到服务器

### 步骤5：清除浏览器缓存

有时浏览器缓存会导致问题：

1. 按 `Ctrl + Shift + Delete`
2. 选择"缓存的图片和文件"
3. 点击"清除数据"
4. 刷新页面 (`Ctrl + F5`)

### 步骤6：检查数据文件

```powershell
# 查看数据文件是否存在
Test-Path data/billing_records.json

# 查看文件大小
(Get-Item data/billing_records.json).Length

# 查看前几行
Get-Content data/billing_records.json | Select-Object -First 20
```

### 步骤7：重启服务器

```powershell
# 停止当前服务器（Ctrl+C）
# 然后重新启动
go run main.go
```

## 常见问题解决方案

### 问题1：端口被占用

**错误信息：** `bind: address already in use`

**解决方法：**
```powershell
# 查找占用8080端口的进程
netstat -ano | findstr :8080

# 结束进程（替换PID为实际进程ID）
taskkill /PID <PID> /F
```

### 问题2：数据文件损坏

**症状：** API返回错误，无法读取数据

**解决方法：**
```powershell
# 备份当前文件
Copy-Item data/billing_records.json data/billing_records.json.backup

# 检查JSON格式
Get-Content data/billing_records.json | ConvertFrom-Json
```

如果JSON格式错误，可以从备份恢复或手动修复。

### 问题3：JavaScript模块加载失败

**错误信息：** `Failed to load module script`

**可能原因：**
- 文件路径错误
- 服务器没有正确配置静态文件服务
- 浏览器不支持ES6模块

**解决方法：**
1. 确保使用现代浏览器（Chrome、Firefox、Edge最新版）
2. 检查文件路径是否正确
3. 确保服务器正在运行

### 问题4：CORS错误

**错误信息：** `Access to fetch at 'http://localhost:8080' from origin 'null' has been blocked by CORS policy`

**解决方法：**
- 确保通过 `http://localhost:8080` 访问，而不是直接打开HTML文件
- 检查服务器CORS配置

## 快速诊断命令

```powershell
# 一键诊断
Write-Host "=== 系统诊断 ===" -ForegroundColor Cyan

# 1. 检查服务器
Write-Host "`n1. 检查服务器进程..." -ForegroundColor Yellow
Get-Process | Where-Object {$_.ProcessName -like "*go*"} | Select-Object ProcessName, Id

# 2. 检查数据文件
Write-Host "`n2. 检查数据文件..." -ForegroundColor Yellow
if (Test-Path data/billing_records.json) {
    $size = (Get-Item data/billing_records.json).Length
    Write-Host "  文件存在，大小: $size 字节" -ForegroundColor Green
} else {
    Write-Host "  文件不存在！" -ForegroundColor Red
}

# 3. 测试API
Write-Host "`n3. 测试API连接..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -TimeoutSec 3
    Write-Host "  API正常: $($response.status)" -ForegroundColor Green
} catch {
    Write-Host "  API连接失败！" -ForegroundColor Red
}

# 4. 测试数据获取
Write-Host "`n4. 测试数据获取..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/billing" -TimeoutSec 3
    Write-Host "  记录数: $($response.data.Count)" -ForegroundColor Green
} catch {
    Write-Host "  数据获取失败！" -ForegroundColor Red
}

Write-Host "`n=== 诊断完成 ===" -ForegroundColor Cyan
```

## 联系支持

如果以上方法都无法解决问题，请提供以下信息：

1. 浏览器控制台的错误信息（截图）
2. 服务器终端的输出
3. `test_api.ps1` 的运行结果
4. `debug.html` 页面的显示内容

---

**最后更新：** 2025-12-18  
**版本：** v2.4.2
