# 批量详细报表生成测试
Write-Host "`n=== 批量详细报表生成测试 ===" -ForegroundColor Cyan

# 1. 获取所有记录
Write-Host "`n1. 获取数据库中的所有记录..." -ForegroundColor Yellow
$allRecords = (Invoke-RestMethod -Uri "http://localhost:8080/api/v1/billing" -Method GET).data
Write-Host "   共找到 $($allRecords.Count) 条记录" -ForegroundColor Green

# 2. 按月份分组
Write-Host "`n2. 按月份分组统计:" -ForegroundColor Yellow
$grouped = $allRecords | Group-Object billingMonth
foreach ($group in $grouped) {
    Write-Host "   月份: $($group.Name) - $($group.Count) 条记录" -ForegroundColor Cyan
}

# 3. 生成批量详细报表（按月份）
Write-Host "`n3. 生成2025-11月份的批量详细报表..." -ForegroundColor Yellow
try {
    $result = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/billing/report/generate?month=2025-11" -Method GET
    if ($result.success) {
        Write-Host "   ✓ 报表生成成功！" -ForegroundColor Green
        Write-Host "   文件名: $($result.data.filename)" -ForegroundColor Cyan
        Write-Host "   记录数: $($result.data.count)" -ForegroundColor Cyan
        Write-Host "   消息: $($result.data.message)" -ForegroundColor Cyan
        
        # 检查文件
        if (Test-Path $result.data.filename) {
            $fileInfo = Get-Item $result.data.filename
            Write-Host "   文件大小: $([math]::Round($fileInfo.Length / 1KB, 2)) KB" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "   ✗ 生成失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 4. 生成指定ID的批量报表
Write-Host "`n4. 生成指定ID的批量详细报表 (ID: 1,12)..." -ForegroundColor Yellow
try {
    $result2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/billing/report/generate?ids=1,12" -Method GET
    if ($result2.success) {
        Write-Host "   ✓ 报表生成成功！" -ForegroundColor Green
        Write-Host "   文件名: $($result2.data.filename)" -ForegroundColor Cyan
        Write-Host "   记录数: $($result2.data.count)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "   ✗ 生成失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Cyan
Write-Host "`n详细卡片报表特点:" -ForegroundColor Yellow
Write-Host "  • 每个账单以完整的详细卡片形式展示" -ForegroundColor White
Write-Host "  • 包含完整的水费、电费、管理费详情" -ForegroundColor White
Write-Host "  • 美观的渐变背景和分区设计" -ForegroundColor White
Write-Host "  • 每行最多显示2个卡片，自动换行" -ForegroundColor White
Write-Host "  • 适合打印和分享" -ForegroundColor White
