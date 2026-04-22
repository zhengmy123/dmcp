#!/bin/bash
# MCP Server 初始化脚本

set -e

echo "===================================="
echo "MCP Server 初始化脚本"
echo "===================================="

# 检查 docker-compose 是否运行
if ! docker ps | grep -q mcp-mysql; then
    echo "MySQL 容器未运行，正在启动..."
    docker-compose up -d mysql
    echo "等待 MySQL 启动..."
    sleep 10
fi

# 等待 MySQL 就绪
echo "等待 MySQL 就绪..."
for i in {1..30}; do
    if docker exec mcp-mysql mysqladmin ping -h localhost -uroot -p1234qwer &>/dev/null; then
        echo "MySQL 已就绪"
        break
    fi
    echo "等待中... ($i/30)"
    sleep 2
done

# 执行数据库初始化（使用 --login-path 或 stdin 方式避免密码警告）
echo ""
echo "执行数据库迁移..."
docker exec -i mcp-mysql mysql -uroot -p1234qwer < docs/mysql_migration.sql 2>/dev/null || \
docker exec -i mcp-mysql sh -c 'mysql -uroot -p1234qwer' < docs/mysql_migration.sql
echo "数据库迁移完成"

# 创建管理员用户
echo ""
echo "初始化管理员用户..."
cd scripts/init && go run init_user.go
cd ../..
echo "管理员用户初始化完成"

# 检查 web/admin/dist 是否存在
echo ""
if [ ! -d "web/admin/dist" ]; then
    echo "前端未构建，正在构建..."
    cd web/admin && npm install && npm run build && cd ../..
    echo "前端构建完成"
else
    echo "前端已存在，跳过构建"
fi

echo ""
echo "===================================="
echo "初始化完成!"
echo "===================================="
echo ""
echo "启动服务: make up"
echo "查看状态: make ps"
echo ""
echo "服务地址:"
echo "  MCP Server: http://localhost:18080"
echo "  MySQL:     localhost:3306"
echo "  Web Admin: http://localhost:17000"
echo ""
