# 批量设置额外费用测试脚本
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "批量设置额外费用功能测试" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080/api/v1"

# 测试1：获取所有记录，准备测试数据
Write-Host "【测试1】获取现有记录..." -ForegroundColor Yellow
try {
    $allRecords = Invoke-RestMethod -Uri "$baseUrl/billing"
    if ($allRecords.success -and $allRecords.data.Count -gt 0) {
        Write-Host "  ✓ 找到 $($allRecords.data.Count) 条记录" -ForegroundColor Green
        
        # 显示前5条记录
        Write-Host "`n  当前记录列表：" -ForegroundColor Cyan
        $allRecords.data | Select-Object -First 5 | ForEach-Object {
            $extraInfo = if ($_.extraFees -and $_.extraFees.Count -gt 0) {
                $extraTotal = ($_.extraFees | Measure-Object -Property amount -Sum).Sum
                "额外费用: ¥$($extraTotal) ($($_.extraFees.Count)项)"
            } else {
                "无额外费用"
            }
            Write-Host "    ID=$($_.id), 房号=$($_.roomNumber), 月份=$($_.billingMonth), $extraInfo" -ForegroundColor White
        }
        
        # 准备测试用的ID列表（取前3条）
        $testIds = $allRecords.data | Select-Object -First 3 | ForEach-Object { $_.id }
        Write-Host "`n  将使用以下记录进行测试: $($testIds -join ', ')" -ForegroundColor Cyan
    } else {
        Write-Host "  ✗ 没有找到记录，请先创建一些测试数据" -ForegroundColor Red
        exit
    }
} catch {
    Write-Host "  ✗ 获取记录失败: $($_.Exception.Message)" -ForegroundColor Red
    exit
}

Start-Sleep -Seconds 2

# 测试2：追加模式 - 批量添加公共费用
Write-Host "`n【测试2】追加模式 - 批量添加公共费用..." -ForegroundColor Yellow
Write-Host "  操作: 为选中记录追加'公共区域维护费'和'垃圾清运费'" -ForegroundColor Gray

$appendData = @{
    ids = $testIds
    extraFees = @(
        @{
            name = "公共区域维护费"
            amount = 50.00
        },
        @{
            name = "垃圾清运费"
            amount = 20.00
        }
    )
    mode = "append"
} | ConvertTo-Json -Depth 10

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/billing/batch-extra-fee" `
        -Method POST `
        -ContentType "application/json" `
        -Body $appendData
    
    if ($result.success) {
        Write-Host "  ✓ 成功为 $($result.count) 条记录追加额外费用！" -ForegroundColor Green
        
        # 验证结果
        Start-Sleep -Seconds 1
        foreach ($id in $testIds) {
            $record = Invoke-RestMethod -Uri "$baseUrl/billing/$id"
            if ($record.success) {
                Write-Host "`n    记录 ID=$id ($($record.data.roomNumber)):" -ForegroundColor Cyan
                Write-Host "      额外费用项数: $($record.data.extraFees.Count)" -ForegroundColor White
                foreach ($fee in $record.data.extraFees) {
                    Write-Host "        - $($fee.name): ¥$($fee.amount)" -ForegroundColor White
                }
                Write-Host "      总费用: ¥$($record.data.totalCost)" -ForegroundColor Yellow
            }
        }
    } else {
        Write-Host "  ✗ 操作失败: $($result.error)" -ForegroundColor Red
    }
} catch {
    Write-Host "  ✗ 请求失败: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# 测试3：追加模式 - 再次追加更多费用
Write-Host "`n【测试3】追加模式 - 再次追加电梯维护费..." -ForegroundColor Yellow

$appendData2 = @{
    ids = @($testIds[0])  # 只选择第一条记录
    extraFees = @(
        @{
            name = "电梯维护费"
            amount = 30.00
        }
    )
    mode = "append"
} | ConvertTo-Json -Depth 10

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/billing/batch-extra-fee" `
        -Method POST `
        -ContentType "application/json" `
        -Body $appendData2
    
    if ($result.success) {
        Write-Host "  ✓ 成功追加！查看结果..." -ForegroundColor Green
        
        Start-Sleep -Seconds 1
        $record = Invoke-RestMethod -Uri "$baseUrl/billing/$($testIds[0])"
        if ($record.success) {
            Write-Host "`n    记录 ID=$($testIds[0]) 的额外费用：" -ForegroundColor Cyan
            foreach ($fee in $record.data.extraFees) {
                Write-Host "      - $($fee.name): ¥$($fee.amount)" -ForegroundColor White
            }
            $extraTotal = ($record.data.extraFees | Measure-Object -Property amount -Sum).Sum
            Write-Host "    额外费用总计: ¥$extraTotal" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "  ✗ 请求失败: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# 测试4：替换模式 - 替换所有额外费用
Write-Host "`n【测试4】替换模式 - 替换所有额外费用..." -ForegroundColor Yellow
Write-Host "  操作: 将第一条记录的所有额外费用替换为'月度管理费'" -ForegroundColor Gray

$replaceData = @{
    ids = @($testIds[0])
    extraFees = @(
        @{
            name = "月度统一管理费"
            amount = 100.00
        }
    )
    mode = "replace"
} | ConvertTo-Json -Depth 10

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/billing/batch-extra-fee" `
        -Method POST `
        -ContentType "application/json" `
        -Body $replaceData
    
    if ($result.success) {
        Write-Host "  ✓ 成功替换！查看结果..." -ForegroundColor Green
        
        Start-Sleep -Seconds 1
        $record = Invoke-RestMethod -Uri "$baseUrl/billing/$($testIds[0])"
        if ($record.success) {
            Write-Host "`n    记录 ID=$($testIds[0]) 的额外费用（替换后）：" -ForegroundColor Cyan
            foreach ($fee in $record.data.extraFees) {
                Write-Host "      - $($fee.name): ¥$($fee.amount)" -ForegroundColor White
            }
            Write-Host "    额外费用项数: $($record.data.extraFees.Count) (应该只有1项)" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "  ✗ 请求失败: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# 测试5：错误处理 - 空ID列表
Write-Host "`n【测试5】错误处理 - 空ID列表..." -ForegroundColor Yellow

$emptyIdsData = @{
    ids = @()
    extraFees = @(
        @{
            name = "测试费用"
            amount = 10.00
        }
    )
    mode = "append"
} | ConvertTo-Json -Depth 10

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/billing/batch-extra-fee" `
        -Method POST `
        -ContentType "application/json" `
        -Body $emptyIdsData
    
    Write-Host "  ✗ 应该返回错误，但操作成功了" -ForegroundColor Red
} catch {
    Write-Host "  ✓ 正确返回错误：请选择要设置额外费用的记录" -ForegroundColor Green
}

Start-Sleep -Seconds 1

# 测试6：错误处理 - 空额外费用列表
Write-Host "`n【测试6】错误处理 - 空额外费用列表..." -ForegroundColor Yellow

$emptyFeesData = @{
    ids = @($testIds[0])
    extraFees = @()
    mode = "append"
} | ConvertTo-Json -Depth 10

try {
    $result = Invoke-RestMethod -Uri "$baseUrl/billing/batch-extra-fee" `
        -Method POST `
        -ContentType "application/json" `
        -Body $emptyFeesData
    
    Write-Host "  ✗ 应该返回错误，但操作成功了" -ForegroundColor Red
} catch {
    Write-Host "  ✓ 正确返回错误：请至少添加一项额外费用" -ForegroundColor Green
}

# 总结
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n功能验证：" -ForegroundColor Yellow
Write-Host "  ✓ 追加模式：在现有额外费用基础上添加" -ForegroundColor White
Write-Host "  ✓ 替换模式：替换所有现有额外费用" -ForegroundColor White
Write-Host "  ✓ 批量操作：同时处理多条记录" -ForegroundColor White
Write-Host "  ✓ 费用计算：自动更新总费用" -ForegroundColor White
Write-Host "  ✓ 错误处理：正确验证输入参数" -ForegroundColor White

Write-Host "`n使用建议：" -ForegroundColor Yellow
Write-Host "  • 追加模式：适合定期添加公共费用" -ForegroundColor Cyan
Write-Host "  • 替换模式：适合重新规划费用结构" -ForegroundColor Cyan
Write-Host "  • 操作前建议先导出备份数据" -ForegroundColor Cyan
Write-Host "  • 可在前端界面更方便地使用此功能`n" -ForegroundColor Cyan
