.PHONY: help build run test clean fmt lint deps wire dev-start dev-stop dev-restart dev-logs dev-status dev-app

# Show help
help:
	@echo "eino-show Makefile 帮助"
	@echo ""
	@echo "基础命令:"
	@echo "  build             构建应用"
	@echo "  run               运行应用"
	@echo "  test              运行测试"
	@echo "  clean             清理构建文件"
	@echo ""
	@echo "开发工具:"
	@echo "  fmt               格式化代码"
	@echo "  lint              代码检查"
	@echo "  deps              安装依赖"
	@echo "  wire              运行 wire 生成依赖注入代码"
	@echo ""
	@echo "开发模式（推荐）:"
	@echo "  dev-start         启动开发环境基础设施（PostgreSQL + Redis）"
	@echo "  dev-stop          停止开发环境"
	@echo "  dev-restart       重启开发环境"
	@echo "  dev-logs          查看开发环境日志"
	@echo "  dev-status        查看开发环境状态"
	@echo "  dev-app           启动后端应用（本地运行，需先运行 dev-start）"

# Go 相关变量
BINARY_NAME=es-apiserver
MAIN_PATH=./cmd/mb-apiserver

# 构建应用
build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)/main.go

# 运行应用
run: build
	./bin/$(BINARY_NAME)

# 运行测试
test:
	go test -v ./...

# 清理构建产物
clean:
	go clean
	rm -f bin/$(BINARY_NAME)

# 格式化代码
fmt:
	go fmt ./...
	goimports -w .

# 代码检查
lint:
	golangci-lint run

# 安装依赖
deps:
	go mod download
	go mod tidy

# 运行 wire 生成依赖注入代码
wire:
	cd $(MAIN_PATH) && wire
	cd internal/apiserver && wire

# 开发环境 - 启动基础设施服务
dev-start:
	./scripts/dev.sh start

# 开发环境 - 停止服务
dev-stop:
	./scripts/dev.sh stop

# 开发环境 - 重启服务
dev-restart:
	./scripts/dev.sh restart

# 开发环境 - 查看日志
dev-logs:
	./scripts/dev.sh logs

# 开发环境 - 查看状态
dev-status:
	./scripts/dev.sh status

# 开发环境 - 启动后端应用
dev-app:
	./scripts/dev.sh app
