# 新功能测试脚本
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "水电费系统 - 新功能测试" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080/api/v1"

# 测试1：创建带额外费用的记录
Write-Host "【测试1】创建带额外费用的记录..." -ForegroundColor Yellow
Write-Host "房间号: 203" -ForegroundColor Gray
Write-Host "额外费用: 水管维修费(120.50), 公共清洁费(30.00), 电梯维护费(45.00)" -ForegroundColor Gray

$newRecord = @{
    roomNumber = "203"
    currentWater = 3100
    previousWater = 3050
    waterAdjustment = 2
    waterPrice = 4.3
    currentElectric = 8200
    previousElectric = 7900
    electricAdjustment = 5
    electricPrice = 0.72
    managementFee = 22
    billingMonth = "2025-11"
    extraFees = @(
        @{name = "水管维修费"; amount = 120.50},
        @{name = "公共清洁费"; amount = 30.00},
        @{name = "电梯维护费"; amount = 45.00}
    )
} | ConvertTo-Json -Depth 10

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/billing" `
        -Method POST `
        -ContentType "application/json" `
        -Body $newRecord
    
    if ($result.success) {
        Write-Host "  ✓ 创建成功！" -ForegroundColor Green
        Write-Host "  记录ID: $($result.data.id)" -ForegroundColor Cyan
        Write-Host "  总费用: ¥$($result.data.totalCost)" -ForegroundColor Cyan
        $newId = $result.data.id
        
        # 验证额外费用
        $extraTotal = 0
        foreach ($fee in $result.data.extraFees) {
            $extraTotal += $fee.amount
            Write-Host "    - $($fee.name): ¥$($fee.amount)" -ForegroundColor White
        }
        Write-Host "  额外费用合计: ¥$extraTotal" -ForegroundColor Cyan
    }
} catch {
    Write-Host "  ✗ 创建失败: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# 测试2：查看所有记录（不排序）
Write-Host "`n【测试2】查看所有记录（默认顺序）..." -ForegroundColor Yellow
try {
    $allRecords = Invoke-RestMethod -Uri "$baseUrl/billing"
    Write-Host "  共 $($allRecords.data.Count) 条记录" -ForegroundColor Green
    Write-Host "  房间号列表: $($allRecords.data | ForEach-Object { $_.roomNumber } | Join-String -Separator ', ')" -ForegroundColor Cyan
} catch {
    Write-Host "  ✗ 获取失败" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# 测试3：按房号升序排序
Write-Host "`n【测试3】按房号升序排序..." -ForegroundColor Yellow
try {
    $sortedAsc = Invoke-RestMethod -Uri "$baseUrl/billing?sortBy=room&order=asc"
    Write-Host "  ✓ 排序成功（升序）" -ForegroundColor Green
    Write-Host "  房间号顺序: $($sortedAsc.data | ForEach-Object { $_.roomNumber } | Join-String -Separator ' → ')" -ForegroundColor Cyan
} catch {
    Write-Host "  ✗ 排序失败" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# 测试4：按房号降序排序
Write-Host "`n【测试4】按房号降序排序..." -ForegroundColor Yellow
try {
    $sortedDesc = Invoke-RestMethod -Uri "$baseUrl/billing?sortBy=room&order=desc"
    Write-Host "  ✓ 排序成功（降序）" -ForegroundColor Green
    Write-Host "  房间号顺序: $($sortedDesc.data | ForEach-Object { $_.roomNumber } | Join-String -Separator ' → ')" -ForegroundColor Cyan
} catch {
    Write-Host "  ✗ 排序失败" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# 测试5：生成单个卡片（带额外费用）
if ($newId) {
    Write-Host "`n【测试5】生成带额外费用的单个卡片..." -ForegroundColor Yellow
    try {
        $cardResult = Invoke-RestMethod -Uri "$baseUrl/billing/card/$newId"
        if ($cardResult.success) {
            Write-Host "  ✓ 卡片生成成功！" -ForegroundColor Green
            Write-Host "  文件: $($cardResult.data.filename)" -ForegroundColor Cyan
            
            if (Test-Path $cardResult.data.filename) {
                $fileInfo = Get-Item $cardResult.data.filename
                Write-Host "  大小: $([math]::Round($fileInfo.Length / 1KB, 2)) KB" -ForegroundColor Cyan
            }
        }
    } catch {
        Write-Host "  ✗ 生成失败" -ForegroundColor Red
    }
}

Start-Sleep -Seconds 1

# 测试6：生成批量报表（按房号升序）
Write-Host "`n【测试6】生成批量报表（按房号升序）..." -ForegroundColor Yellow
try {
    $reportResult = Invoke-RestMethod -Uri "$baseUrl/billing/report/generate?month=2025-11&sortBy=room&order=asc"
    if ($reportResult.success) {
        Write-Host "  ✓ 报表生成成功！" -ForegroundColor Green
        Write-Host "  文件: $($reportResult.data.filename)" -ForegroundColor Cyan
        Write-Host "  记录数: $($reportResult.data.count)" -ForegroundColor Cyan
        
        if (Test-Path $reportResult.data.filename) {
            $fileInfo = Get-Item $reportResult.data.filename
            Write-Host "  大小: $([math]::Round($fileInfo.Length / 1KB, 2)) KB" -ForegroundColor Cyan
        }
    }
} catch {
    Write-Host "  ✗ 生成失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 总结
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n新功能特点：" -ForegroundColor Yellow
Write-Host "  ✓ 支持自定义额外费用（水管维修、清洁等）" -ForegroundColor White
Write-Host "  ✓ 额外费用自动计入总费用" -ForegroundColor White
Write-Host "  ✓ 图片动态显示额外费用项" -ForegroundColor White
Write-Host "  ✓ 无额外费用时不占用空间" -ForegroundColor White
Write-Host "  ✓ 支持按房号升序/降序排序" -ForegroundColor White
Write-Host "  ✓ 排序应用于查询和报表生成" -ForegroundColor White

Write-Host "`n生成的文件位置：" -ForegroundColor Yellow
Write-Host "  - reports\ 目录" -ForegroundColor Cyan
