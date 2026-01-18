#!/bin/bash
# eino-show 开发环境启动脚本
# 只启动基础设施，应用需要手动在本地运行

# 设置颜色
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # 无颜色

# 获取项目根目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# 日志函数
log_info() {
    printf "%b\n" "${BLUE}[INFO]${NC} $1"
}

log_success() {
    printf "%b\n" "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    printf "%b\n" "${RED}[ERROR]${NC} $1"
}

log_warning() {
    printf "%b\n" "${YELLOW}[WARNING]${NC} $1"
}

# 选择可用的 Docker Compose 命令
DOCKER_COMPOSE_BIN=""
DOCKER_COMPOSE_SUBCMD=""

detect_compose_cmd() {
    if docker compose version &> /dev/null; then
        DOCKER_COMPOSE_BIN="docker"
        DOCKER_COMPOSE_SUBCMD="compose"
        return 0
    fi
    if command -v docker-compose &> /dev/null; then
        if docker-compose version &> /dev/null; then
            DOCKER_COMPOSE_BIN="docker-compose"
            DOCKER_COMPOSE_SUBCMD=""
            return 0
        fi
    fi
    return 1
}

# 显示帮助信息
show_help() {
    printf "%b\n" "${GREEN}eino-show 开发环境脚本${NC}"
    echo "用法: $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  start      启动基础设施服务（postgres, redis）"
    echo "  stop       停止所有服务"
    echo "  restart    重启所有服务"
    echo "  logs       查看服务日志"
    echo "  status     查看服务状态"
    echo "  app        启动后端应用（本地运行）"
    echo "  frontend   启动前端开发服务器（本地运行）"
    echo "  help       显示此帮助信息"
    echo ""
    echo "可选 Profile（用于 start 命令）:"
    echo "  --minio    启动 MinIO 对象存储"
    echo "  --qdrant   启动 Qdrant 向量数据库"
    echo "  --neo4j    启动 Neo4j 图数据库"
    echo "  --jaeger   启动 Jaeger 链路追踪"
    echo "  --full     启动所有可选服务"
    echo ""
    echo "示例："
    echo "  $0 start                    # 启动基础服务"
    echo "  $0 start --qdrant           # 启动基础服务 + Qdrant"
    echo "  $0 start --full             # 启动所有服务"
    echo "  $0 app                      # 在另一个终端启动后端"
    echo "  $0 frontend                 # 在另一个终端启动前端"
}

# 检查 Docker
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "未安装Docker，请先安装Docker"
        return 1
    fi
    
    if ! detect_compose_cmd; then
        log_error "未检测到 Docker Compose"
        return 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker服务未运行"
        return 1
    fi
    
    return 0
}

# 启动基础设施服务
start_services() {
    log_info "启动开发环境基础设施服务..."

    check_docker
    if [ $? -ne 0 ]; then
        return 1
    fi

    cd "$PROJECT_ROOT"

    # 检查 .env 文件
    if [ ! -f ".env" ]; then
        log_warning ".env 文件不存在，从 .env.example 创建..."
        cp .env.example .env
        log_info "已创建 .env 文件，请根据需要修改配置"
    fi

    # 解析 profile 参数
    shift  # 移除 "start" 命令本身
    PROFILES=""
    ENABLED_SERVICES=""

    while [ $# -gt 0 ]; do
        case "$1" in
            --minio)
                PROFILES="$PROFILES --profile minio"
                ENABLED_SERVICES="$ENABLED_SERVICES minio"
                ;;
            --qdrant)
                PROFILES="$PROFILES --profile qdrant"
                ENABLED_SERVICES="$ENABLED_SERVICES qdrant"
                ;;
            --neo4j)
                PROFILES="$PROFILES --profile neo4j"
                ENABLED_SERVICES="$ENABLED_SERVICES neo4j"
                ;;
            --jaeger)
                PROFILES="$PROFILES --profile jaeger"
                ENABLED_SERVICES="$ENABLED_SERVICES jaeger"
                ;;
            --full)
                PROFILES="--profile full"
                ENABLED_SERVICES="minio qdrant neo4j jaeger"
                break
                ;;
            *)
                log_warning "未知参数: $1"
                ;;
        esac
        shift
    done

    # 启动服务
    "$DOCKER_COMPOSE_BIN" $DOCKER_COMPOSE_SUBCMD -f docker-compose.dev.yml $PROFILES up -d

    if [ $? -eq 0 ]; then
        log_success "基础设施服务已启动"
        echo ""
        log_info "服务访问地址:"
        echo "  - PostgreSQL:    localhost:5432"
        echo "  - Redis:         localhost:6379"

        # 根据启用的 profile 显示额外服务
        if [[ "$ENABLED_SERVICES" == *"minio"* ]] || [[ "$PROFILES" == *"full"* ]]; then
            echo "  - MinIO:         localhost:9000 (Console: localhost:9001)"
        fi
        if [[ "$ENABLED_SERVICES" == *"qdrant"* ]] || [[ "$PROFILES" == *"full"* ]]; then
            echo "  - Qdrant:        localhost:6333 (gRPC: localhost:6334)"
        fi
        if [[ "$ENABLED_SERVICES" == *"neo4j"* ]] || [[ "$PROFILES" == *"full"* ]]; then
            echo "  - Neo4j:         localhost:7474 (Bolt: localhost:7687)"
        fi
        if [[ "$ENABLED_SERVICES" == *"jaeger"* ]] || [[ "$PROFILES" == *"full"* ]]; then
            echo "  - Jaeger:        localhost:16686"
        fi

        echo ""
        log_info "接下来的步骤:"
        printf "%b\n" "${YELLOW}1. 在新终端运行后端:${NC} make dev-app"
        printf "%b\n" "${YELLOW}2. 在新终端运行前端:${NC} make dev-frontend"
        return 0
    else
        log_error "服务启动失败"
        return 1
    fi
}

# 停止服务
stop_services() {
    log_info "停止开发环境服务..."

    check_docker
    if [ $? -ne 0 ]; then
        return 1
    fi

    cd "$PROJECT_ROOT"
    "$DOCKER_COMPOSE_BIN" $DOCKER_COMPOSE_SUBCMD -f docker-compose.dev.yml down

    if [ $? -eq 0 ]; then
        log_success "所有服务已停止"
        return 0
    else
        log_error "服务停止失败"
        return 1
    fi
}

# 重启服务
restart_services() {
    stop_services
    sleep 2
    start_services
}

# 查看日志
show_logs() {
    detect_compose_cmd || return 1
    cd "$PROJECT_ROOT"
    "$DOCKER_COMPOSE_BIN" $DOCKER_COMPOSE_SUBCMD -f docker-compose.dev.yml logs -f
}

# 查看状态
show_status() {
    detect_compose_cmd || return 1
    cd "$PROJECT_ROOT"
    "$DOCKER_COMPOSE_BIN" $DOCKER_COMPOSE_SUBCMD -f docker-compose.dev.yml ps
}

# 启动后端应用（本地）
start_app() {
    log_info "启动后端应用（本地开发模式）..."

    cd "$PROJECT_ROOT"

    # 检查 Go 是否安装
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装"
        return 1
    fi

    # 加载环境变量
    if [ -f ".env" ]; then
        log_info "加载 .env 文件..."
        set -a
        source .env
        set +a
    else
        log_error ".env 文件不存在，请先运行 make dev-start 创建配置文件"
        return 1
    fi

    # 设置本地开发环境变量（覆盖 Docker 容器地址）
    export DB_HOST=localhost
    export REDIS_ADDR=localhost:6379

    # 设置 HTTP 代理（用于 web_search 工具访问外网）
    export HTTP_PROXY=http://127.0.0.1:7897
    export HTTPS_PROXY=http://127.0.0.1:7897

    # 自动 kill 8080 端口上的进程（多次尝试确保释放）
    for i in 1 2 3; do
        if lsof -i :8080 &> /dev/null; then
            log_warning "端口 8080 被占用，正在释放 (尝试 $i)..."
            lsof -ti :8080 | xargs kill -9 2>/dev/null
            sleep 2
        else
            break
        fi
    done
    
    if lsof -i :8080 &> /dev/null; then
        log_error "无法释放端口 8080，请手动处理"
        return 1
    fi
    log_success "端口 8080 已就绪"

    log_info "环境变量已设置，启动应用..."
    log_info "配置文件: configs/mb-apiserver.yaml"

    # 运行应用
    go run ./cmd/mb-apiserver/main.go --config=configs/mb-apiserver.yaml
}

# 启动前端（本地）
start_frontend() {
    log_info "启动前端开发服务器..."

    cd "$PROJECT_ROOT/frontend"

    # 检查 npm 是否安装
    if ! command -v npm &> /dev/null; then
        log_error "npm 未安装"
        return 1
    fi

    # 检查依赖是否已安装
    if [ ! -d "node_modules" ]; then
        log_warning "node_modules 不存在，正在安装依赖..."
        npm install
    fi

    log_info "启动 Vite 开发服务器..."
    log_info "前端将运行在 http://localhost:5173"

    # 运行开发服务器
    npm run dev
}

# 解析命令
CMD="${1:-help}"
case "$CMD" in
    start)
        start_services "$@"
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    logs)
        show_logs
        ;;
    status)
        show_status
        ;;
    app)
        start_app
        ;;
    frontend)
        start_frontend
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "未知命令: $CMD"
        show_help
        exit 1
        ;;
esac

exit 0
