# MCP Server Makefile
#
# 使用方式:
#   make init           # 初始化（首次使用）
#   make build-web      # 构建前端
#   make up-middleware  # 启动中间件（MySQL）
#   make up             # 启动所有服务（先中间件后应用）
#   make down           # 停止所有服务
#   make down-middleware # 停止中间件
#   make restart        # 重启所有服务（重新加载最新代码）
#   make restart-server # 仅重启 Go 服务
#   make restart-web    # 重启 Web 服务
#   make restart-app    # 重启应用服务（Go + Web）
#   make logs           # 查看日志
#   make ps             # 查看状态
#   make clean           # 清理

.PHONY: help init build-web build-web-clean up up-middleware down down-middleware restart restart-server restart-web restart-app logs logs-app ps ps-middleware clean

help:
	@echo "MCP Server 管理脚本"
	@echo ""
	@echo "  make init              - 初始化（首次使用，包含数据库和前端构建）"
	@echo "  make build-web         - 构建前端（本地构建，速度快）"
	@echo "  make build-web-clean   - 清理并重新构建前端（修复依赖问题时使用）"
	@echo "  make up-middleware     - 启动中间件（MySQL + Redis）"
	@echo "  make up                - 启动所有服务（先中间件后应用）"
	@echo "  make down              - 停止所有服务"
	@echo "  make down-middleware    - 停止中间件"
	@echo "  make restart            - 重启所有服务（中间件+应用）"
	@echo "  make restart-server     - 仅重启 Go 服务"
	@echo "  make restart-web        - 重启 Web 服务"
	@echo "  make restart-app        - 重启应用服务（Go + Web）"
	@echo "  make logs              - 查看应用服务日志"
	@echo "  make logs-app          - 查看应用服务日志（Go + Web）"
	@echo "  make ps                - 查看应用服务状态"
	@echo "  make ps-middleware      - 查看中间件状态"
	@echo "  make clean             - 清理容器和数据"

init:
	@echo "开始初始化..."
	bash scripts/init.sh

build-web:
	@echo "构建前端..."
	cd web/admin && npm run build
	@echo "前端构建完成！"

build-web-clean:
	@echo "清理并重新构建前端..."
	cd web/admin && rm -rf node_modules package-lock.json && npm install && npm run build
	@echo "前端构建完成！"

up-middleware:
	@echo "启动中间件（MySQL + Redis）..."
	docker-compose -f docker-compose.middleware.yml up -d
	@echo ""
	@echo "中间件已启动:"
	@echo "  MySQL:  localhost:3306"
	@echo "  Redis:  localhost:6379"

up:
	@echo "启动中间件（MySQL + Redis）..."
	docker-compose -f docker-compose.middleware.yml up -d
	@echo "启动应用服务..."
	docker-compose up -d
	@echo ""
	@echo "服务已启动:"
	@echo "  MCP Server: http://localhost:18080"
	@echo "  MySQL:     localhost:3306"
	@echo "  Redis:     localhost:6379"
	@echo "  Web Admin: http://localhost:17000"

down:
	@echo "停止应用服务..."
	docker-compose down
	@echo "停止中间件..."
	docker-compose -f docker-compose.middleware.yml down

down-middleware:
	@echo "停止中间件..."
	docker-compose -f docker-compose.middleware.yml down

restart:
	@echo "重启所有服务（中间件+应用）..."
	docker-compose -f docker-compose.middleware.yml restart
	docker-compose restart

restart-server:
	@echo "重启 Go 服务..."
	docker-compose restart mcp-server
	@echo "Go 服务已重启！"

restart-web:
	@echo "构建前端..."
	cd web/admin && npm run build
	@echo "重启 Web 容器..."
	docker-compose restart web
	@echo "Web 服务已更新！"

restart-app:
	@echo "重启应用服务（Go + Web）..."
	$(MAKE) restart-web
	$(MAKE) restart-server
	@echo "应用服务已重启！"

logs:
	docker-compose logs -f

logs-app:
	docker-compose logs -f mcp-server web

logs-middleware:
	docker-compose -f docker-compose.middleware.yml logs -f

ps:
	docker-compose ps

ps-middleware:
	docker-compose -f docker-compose.middleware.yml ps

clean:
	@echo "清理应用容器和数据..."
	docker-compose down -v
	@echo "清理完成"

db-connect:
	docker-compose -f docker-compose.middleware.yml exec mysql mysql -uroot -p1234qwer mcp_server
