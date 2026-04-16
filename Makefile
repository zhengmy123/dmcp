# MCP Server Makefile
#
# 使用方式:
#   make init       # 初始化（首次使用）
#   make up         # 启动服务
#   make down       # 停止服务
#   make restart    # 重启服务（重新加载最新代码）
#   make logs       # 查看日志
#   make ps         # 查看状态
#   make clean      # 清理

.PHONY: help init up down restart logs ps clean

help:
	@echo "MCP Server 管理脚本"
	@echo ""
	@echo "  make init      - 初始化（首次使用，包含数据库和前端构建）"
	@echo "  make up        - 启动所有服务"
	@echo "  make down      - 停止所有服务"
	@echo "  make restart   - 重启所有服务"
	@echo "  make logs      - 查看日志"
	@echo "  make ps        - 查看运行状态"
	@echo "  make clean     - 清理容器和数据"

init:
	@echo "开始初始化..."
	bash scripts/init.sh

up:
	@echo "启动服务..."
	docker-compose up -d
	@echo ""
	@echo "服务已启动:"
	@echo "  MCP Server: http://localhost:18080"
	@echo "  MySQL:     localhost:3306"
	@echo "  Web Admin: http://localhost:17000"

down:
	@echo "停止服务..."
	docker-compose down

restart:
	@echo "重启服务（重新加载最新代码）..."
	docker-compose restart

logs:
	docker-compose logs -f

ps:
	docker-compose ps

clean:
	@echo "清理容器和数据..."
	docker-compose down -v
	@echo "清理完成"

db-connect:
	docker-compose exec mcp-mysql mysql -uroot -p1234qwer mcp_server
