# 智能水表匹配功能测试脚本

$baseUrl = "http://localhost:8080/api/v1"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "智能水表匹配功能测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. 创建测试数据
Write-Host "步骤 1: 创建测试数据..." -ForegroundColor Yellow

$testRecords = @(
    @{
        roomNumber = "101"
        billingMonth = "2025-12"
        previousWater = 2870
        currentWater = 2870
        waterAdjustment = 0
        previousElectric = 7264
        currentElectric = 7264
        electricAdjustment = 0
        managementFee = 22
        waterPrice = 4.3
        electricPrice = 0.72
    },
    @{
        roomNumber = "102"
        billingMonth = "2025-12"
        previousWater = 2850
        currentWater = 2850
        waterAdjustment = 0
        previousElectric = 7100
        currentElectric = 7100
        electricAdjustment = 0
        managementFee = 22
        waterPrice = 4.3
        electricPrice = 0.72
    },
    @{
        roomNumber = "103"
        billingMonth = "2025-12"
        previousWater = 2900
        currentWater = 2900
        waterAdjustment = 0
        previousElectric = 7500
        currentElectric = 7500
        electricAdjustment = 0
        managementFee = 22
        waterPrice = 4.3
        electricPrice = 0.72
    },
    @{
        roomNumber = "104"
        billingMonth = "2025-12"
        previousWater = 2880
        currentWater = 2880
        waterAdjustment = 0
        previousElectric = 7300
        currentElectric = 7300
        electricAdjustment = 0
        managementFee = 22
        waterPrice = 4.3
        electricPrice = 0.72
    }
)

$createdIds = @()

foreach ($record in $testRecords) {
    $json = $record | ConvertTo-Json
    $response = Invoke-RestMethod -Uri "$baseUrl/billing" -Method Post -Body $json -ContentType "application/json"
    if ($response.success) {
        $createdIds += $response.data.id
        Write-Host "  ✓ 创建记录: 房号 $($record.roomNumber), ID: $($response.data.id)" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "创建了 $($createdIds.Count) 条测试记录" -ForegroundColor Green
Write-Host "记录IDs: $($createdIds -join ', ')" -ForegroundColor Cyan
Write-Host ""

# 2. 显示初始状态
Write-Host "步骤 2: 显示初始水表读数..." -ForegroundColor Yellow
Write-Host "房号  | 上月水表 | 本月水表" -ForegroundColor Cyan
Write-Host "------|----------|----------" -ForegroundColor Cyan

foreach ($id in $createdIds) {
    $record = Invoke-RestMethod -Uri "$baseUrl/billing/$id" -Method Get
    if ($record.success) {
        $data = $record.data
        Write-Host "$($data.roomNumber)  |   $($data.previousWater)  |   $($data.currentWater)" -ForegroundColor White
    }
}

Write-Host ""

# 3. 准备水表读数（故意打乱顺序）
Write-Host "步骤 3: 准备水表读数（模拟抄表员记录）..." -ForegroundColor Yellow
$waterReadings = @(2893, 2870, 2920, 2900)
Write-Host "水表读数: $($waterReadings -join ', ')" -ForegroundColor Cyan
Write-Host ""

# 4. 执行智能匹配
Write-Host "步骤 4: 执行智能水表匹配..." -ForegroundColor Yellow

$matchRequest = @{
    ids = $createdIds
    waterReadings = $waterReadings
} | ConvertTo-Json

try {
    $matchResult = Invoke-RestMethod -Uri "$baseUrl/billing/smart-water-match" -Method Post -Body $matchRequest -ContentType "application/json"
    
    if ($matchResult.success) {
        Write-Host "  ✓ 匹配成功！" -ForegroundColor Green
        Write-Host "  更新了 $($matchResult.count) 条记录" -ForegroundColor Green
        Write-Host ""
        
        # 5. 显示匹配结果
        Write-Host "步骤 5: 显示匹配结果..." -ForegroundColor Yellow
        Write-Host "房号  | 上月水表 | 本月水表 | 用水量" -ForegroundColor Cyan
        Write-Host "------|----------|----------|--------" -ForegroundColor Cyan
        
        $totalUsage = 0
        foreach ($match in $matchResult.matches) {
            Write-Host "$($match.roomNumber)  |   $($match.previousWater)  |   $($match.waterReading)  |  $($match.waterUsage) 吨" -ForegroundColor White
            $totalUsage += $match.waterUsage
        }
        
        Write-Host "------|----------|----------|--------" -ForegroundColor Cyan
        Write-Host "总用水量: $totalUsage 吨" -ForegroundColor Green
        Write-Host ""
        
        # 6. 验证数据已更新
        Write-Host "步骤 6: 验证数据库中的数据..." -ForegroundColor Yellow
        foreach ($id in $createdIds) {
            $record = Invoke-RestMethod -Uri "$baseUrl/billing/$id" -Method Get
            if ($record.success) {
                $data = $record.data
                Write-Host "  房号 $($data.roomNumber): 本月水表 = $($data.currentWater), 用水量 = $($data.waterUsage) 吨, 水费 = ¥$($data.totalWaterCost)" -ForegroundColor White
            }
        }
        
    } else {
        Write-Host "  ✗ 匹配失败: $($matchResult.error)" -ForegroundColor Red
    }
} catch {
    Write-Host "  ✗ 请求失败: $_" -ForegroundColor Red
}

Write-Host ""

# 7. 清理测试数据
Write-Host "步骤 7: 清理测试数据..." -ForegroundColor Yellow
$cleanupChoice = Read-Host "是否删除测试数据？(y/n)"

if ($cleanupChoice -eq "y") {
    foreach ($id in $createdIds) {
        $response = Invoke-RestMethod -Uri "$baseUrl/billing/$id" -Method Delete
        if ($response.success) {
            Write-Host "  ✓ 删除记录 ID: $id" -ForegroundColor Green
        }
    }
    Write-Host "测试数据已清理" -ForegroundColor Green
} else {
    Write-Host "保留测试数据，IDs: $($createdIds -join ', ')" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "测试完成！" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
