# 测试API连接
Write-Host "测试API连接..." -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method Get -TimeoutSec 5
    Write-Host "✓ API连接成功" -ForegroundColor Green
    Write-Host "响应: $($response | ConvertTo-Json)" -ForegroundColor White
} catch {
    Write-Host "✗ API连接失败" -ForegroundColor Red
    Write-Host "错误: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "测试获取数据..." -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/billing" -Method Get -TimeoutSec 5
    if ($response.success) {
        Write-Host "✓ 数据获取成功" -ForegroundColor Green
        Write-Host "记录数: $($response.data.Count)" -ForegroundColor White
        if ($response.data.Count -gt 0) {
            Write-Host "第一条记录: 房号=$($response.data[0].roomNumber), 月份=$($response.data[0].billingMonth)" -ForegroundColor White
        }
    } else {
        Write-Host "✗ 数据获取失败" -ForegroundColor Red
        Write-Host "错误: $($response.error)" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ 请求失败" -ForegroundColor Red
    Write-Host "错误: $_" -ForegroundColor Red
}
