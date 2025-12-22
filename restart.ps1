# 重启服务器脚本

Write-Host "正在停止旧服务器..." -ForegroundColor Yellow

# 查找并停止Go进程
$processes = Get-Process | Where-Object {$_.ProcessName -like "*go_build*" -or $_.ProcessName -like "*main*"}

if ($processes) {
    foreach ($proc in $processes) {
        Write-Host "  停止进程: $($proc.ProcessName) (PID: $($proc.Id))" -ForegroundColor Cyan
        Stop-Process -Id $proc.Id -Force
    }
    Start-Sleep -Seconds 1
}

Write-Host "正在启动新服务器..." -ForegroundColor Green
Write-Host "服务器地址: http://localhost:8080" -ForegroundColor Cyan
Write-Host "按 Ctrl+C 停止服务器" -ForegroundColor Yellow
Write-Host ""

# 启动服务器
go run main.go
