@echo off
REM ============================================
REM Opsight - AIOps 监控平台一键部署脚本 (Windows)
REM ============================================
setlocal enabledelayedexpansion

cd /d "%~dp0"

echo.
echo ============================================
echo   Opsight - AIOps 监控平台部署
echo ============================================
echo.

REM ---------- 检查 Docker ----------
echo [STEP] 检查运行环境...

where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] 未检测到 Docker，请先安装 Docker Desktop
    echo         下载地址: https://www.docker.com/products/docker-desktop/
    pause
    exit /b 1
)
echo [INFO]  Docker ✓

where docker >nul 2>&1
docker compose version >nul 2>&1
if %errorlevel% equ 0 (
    set "DOCKER_COMPOSE=docker compose"
) else (
    where docker-compose >nul 2>&1
    if %errorlevel% equ 0 (
        set "DOCKER_COMPOSE=docker-compose"
    ) else (
        echo [ERROR] 未检测到 Docker Compose
        pause
        exit /b 1
    )
)
echo [INFO]  Docker Compose ✓

REM ---------- 准备环境配置 ----------
echo [STEP] 准备环境配置...

if not exist ".env" (
    if exist ".env.example" (
        copy /Y ".env.example" ".env" >nul
        echo [INFO]  已从 .env.example 创建 .env，请根据实际环境编辑
    ) else (
        echo [WARN]  .env.example 不存在，将使用默认配置
    )
) else (
    echo [INFO]  .env 已存在，跳过
)

REM ---------- 构建并启动 ----------
echo [STEP] 构建并启动服务...

%DOCKER_COMPOSE% up -d --build

if %errorlevel% neq 0 (
    echo [ERROR] 启动失败，请检查上方错误信息
    pause
    exit /b 1
)

echo [INFO]  服务启动成功!

echo.
echo ============================================
echo   Opsight 部署完成
echo ============================================
echo   访问地址:   http://localhost:80
echo   查看日志:   %DOCKER_COMPOSE% logs -f
echo   停止服务:   %DOCKER_COMPOSE% down
echo ============================================
echo.

pause
