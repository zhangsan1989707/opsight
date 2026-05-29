#!/usr/bin/env bash
# ============================================
# Opsight - AIOps 监控平台一键部署脚本 (Linux/macOS)
# ============================================
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$PROJECT_DIR"

# ---------- 色彩输出 ----------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }
log_step()  { echo -e "${CYAN}[STEP]${NC}  $*"; }

# ---------- 检查前置条件 ----------
check_prerequisites() {
    log_step "检查运行环境..."

    if ! command -v docker &>/dev/null; then
        log_error "未检测到 Docker，请先安装 Docker"
        log_info "安装指南: https://docs.docker.com/engine/install/"
        exit 1
    fi
    log_info "Docker ✓"

    if docker compose version &>/dev/null; then
        DOCKER_COMPOSE="docker compose"
    elif command -v docker-compose &>/dev/null; then
        DOCKER_COMPOSE="docker-compose"
    else
        log_error "未检测到 Docker Compose，请先安装"
        exit 1
    fi
    log_info "Docker Compose ✓"
}

# ---------- 准备环境配置 ----------
prepare_env() {
    log_step "准备环境配置..."

    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            cp .env.example .env
            log_info "已从 .env.example 创建 .env，请根据实际环境编辑"
        else
            log_warn ".env.example 不存在，将使用 docker-compose.yml 中的默认值"
        fi
    else
        log_info ".env 已存在，跳过"
    fi
}

# ---------- 拉取/构建并启动 ----------
deploy() {
    log_step "构建并启动服务..."

    $DOCKER_COMPOSE up -d --build

    log_info "等待服务就绪..."
    sleep 3

    # 健康检查
    if curl -sf http://localhost:${NGINX_HTTP_PORT:-80}/healthz &>/dev/null; then
        log_info "服务启动成功!"
    else
        log_warn "服务可能仍在启动中，请稍后检查"
        log_info "查看日志: $DOCKER_COMPOSE logs -f"
    fi
}

# ---------- 显示状态 ----------
show_status() {
    echo ""
    echo "============================================"
    echo "  Opsight 部署完成"
    echo "============================================"
    echo "  访问地址:   http://localhost:${NGINX_HTTP_PORT:-80}"
    echo "  查看日志:   $DOCKER_COMPOSE logs -f"
    echo "  停止服务:   $DOCKER_COMPOSE down"
    echo "============================================"
    echo ""
}

# ---------- 主流程 ----------
main() {
    echo ""
    echo "============================================"
    echo "  Opsight - AIOps 监控平台部署"
    echo "============================================"
    echo ""

    check_prerequisites
    prepare_env
    deploy
    show_status
}

main "$@"
