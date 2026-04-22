-- ============================================
-- 数据库表结构更新脚本
-- 删除 enabled 字段，使用 state 字段进行软删除
-- ============================================

SET NAMES utf8mb4;

-- 1. 删除 mcp_servers 表的 enabled 字段
ALTER TABLE `mcp_servers` DROP COLUMN `enabled`;

-- 2. 删除 mcp_http_services 表的 enabled 字段
ALTER TABLE `mcp_http_services` DROP COLUMN `enabled`;

-- 3. 删除 mcp_tool_definitions 表的 enabled 字段
ALTER TABLE `mcp_tool_definitions` DROP COLUMN `enabled`;

-- 4. 删除 tool_mcp_server_bindings 表的 enabled 字段
ALTER TABLE `tool_mcp_server_bindings` DROP COLUMN `enabled`;

-- 5. 删除 mcp_users 表的 enabled 字段
ALTER TABLE `mcp_users` DROP COLUMN `enabled`;

-- ============================================
SELECT 'Table update completed successfully!' AS status;
