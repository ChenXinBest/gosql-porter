@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

echo ========================================
echo     GoSQL-Porter Cross-Platform Build
echo ========================================
echo.

:: 设置变量
set APP_NAME=gosql-porter
set VERSION=1.0.0
set BUILD_DIR=build
set MAIN_PATH=.

:: 清理旧的构建目录
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
mkdir %BUILD_DIR%

:: 获取当前 git commit hash（如果有 git）
for /f "delims=" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=dev

:: 构建参数
set LDFLAGS=-s -w -X main.Version=%VERSION% -X main.GitCommit=%GIT_COMMIT%

echo 开始编译...
echo.

:: Linux AMD64
echo [1/3] 编译 Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%/%APP_NAME%-linux-amd64 %MAIN_PATH%
if !errorlevel! neq 0 (
    echo Linux AMD64 编译失败！
    exit /b 1
)
echo     完成: %BUILD_DIR%/%APP_NAME%-linux-amd64

:: Linux ARM64
echo [2/3] 编译 Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%/%APP_NAME%-linux-arm64 %MAIN_PATH%
if !errorlevel! neq 0 (
    echo Linux ARM64 编译失败！
    exit /b 1
)
echo     完成: %BUILD_DIR%/%APP_NAME%-linux-arm64

:: Windows AMD64 (附带)
echo [3/3] 编译 Windows AMD64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "%LDFLAGS%" -o %BUILD_DIR%/%APP_NAME%-windows-amd64.exe %MAIN_PATH%
if !errorlevel! neq 0 (
    echo Windows AMD64 编译失败！
    exit /b 1
)
echo     完成: %BUILD_DIR%/%APP_NAME%-windows-amd64.exe

echo.
echo ========================================
echo 编译完成！输出目录: %BUILD_DIR%
echo ========================================
dir /b %BUILD_DIR%
echo.
pause
