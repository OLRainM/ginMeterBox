@echo off
echo ============================================
echo   水电费计算管理系统 - 构建脚本
echo   Building Water Electric Billing System
echo ============================================
echo.

echo [1/3] 安装依赖...
go mod tidy
if %errorlevel% neq 0 (
    echo 错误：安装依赖失败
    pause
    exit /b 1
)

echo.
echo [2/3] 编译程序...
go build -o bin\go-ele.exe main.go
if %errorlevel% neq 0 (
    echo 错误：编译失败
    pause
    exit /b 1
)

echo.
echo [3/3] 复制静态文件...
if not exist "bin\static" mkdir bin\static
xcopy /E /I /Y static bin\static > nul

echo.
echo ============================================
echo   构建完成！
echo   可执行文件: bin\go-ele.exe
echo   运行命令: cd bin && go-ele.exe
echo ============================================
echo.

pause
