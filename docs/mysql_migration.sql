-- ============================================
-- 数据库迁移脚本 V2
-- 新增 mcp_servers 表和 token_mcp_server_bindings 表
-- 修改 mcp_tool_definitions 表结构
-- ============================================

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
SET character_set_connection=utf8mb4;

USE `mcp_server`;

-- ============================================
-- 步骤1: 创建 mcp_servers 表（存储 MCP 服务信息）
-- ============================================
CREATE TABLE IF NOT EXISTS `mcp_servers` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `vauth_key` VARCHAR(128) NOT NULL COMMENT '认证密钥',
    `name` VARCHAR(128) NOT NULL COMMENT '服务名称',
    `description` VARCHAR(512) DEFAULT '' COMMENT '服务描述',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_vauth_key` (`vauth_key`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MCP服务表';

-- ============================================
-- 步骤2: 修改 mcp_tool_definitions 表（新增字段）
-- ============================================
-- 新增 service_id 字段（关联 mcp_servers 表）
ALTER TABLE `mcp_tool_definitions` ADD COLUMN IF NOT EXISTS `service_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联服务ID';

-- 新增 input_extra 字段（输入额外配置）
ALTER TABLE `mcp_tool_definitions` ADD COLUMN IF NOT EXISTS `input_extra` TEXT COMMENT '输入额外配置JSON';

-- 新增 output_mapping 字段（输出映射配置）
ALTER TABLE `mcp_tool_definitions` ADD COLUMN IF NOT EXISTS `output_mapping` TEXT COMMENT '输出映射配置JSON';

-- 新增索引（如果尚未存在）
-- 注意：MySQL 不支持 IF NOT EXISTS 语法添加索引，需要手动判断或使用存储过程
-- 这里仅记录建议的索引创建语句，实际执行时请根据情况
-- CREATE INDEX IF NOT EXISTS `idx_service_id` ON `mcp_tool_definitions` (`service_id`);

-- ============================================
-- 步骤3: 创建 token_mcp_server_bindings 表（Token与服务绑定关系）
-- ============================================
CREATE TABLE IF NOT EXISTS `token_mcp_server_bindings` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `token_id` INT UNSIGNED NOT NULL COMMENT '认证密钥ID（关联 mcp_auth_keys 表）',
    `server_id` INT UNSIGNED NOT NULL COMMENT '服务ID（关联 mcp_servers 表）',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_token_server` (`token_id`, `server_id`),
    INDEX `idx_token_id` (`token_id`),
    INDEX `idx_server_id` (`server_id`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Token与服务绑定表';

-- ============================================
-- 步骤4: 数据迁移 - 将现有的 vauth_key 迁移到 mcp_servers 表
-- ============================================

-- 4.1 从 mcp_tool_definitions 表迁移独立的 vauth_key 到 mcp_servers 表
-- 规则：只迁移有意义的 vauth_key（非空且非默认值）
INSERT IGNORE INTO `mcp_servers` (`vauth_key`, `name`, `description`, `enabled`, `created_at`, `updated_at`)
SELECT DISTINCT
    t.`vauth_key` AS `vauth_key`,
    COALESCE(t.`name`, CONCAT('Service_', t.`vauth_key`)) AS `name`,
    COALESCE(t.`server_desc`, '') AS `description`,
    t.`enabled` AS `enabled`,
    t.`created_at` AS `created_at`,
    t.`updated_at` AS `updated_at`
FROM `mcp_tool_definitions` t
WHERE t.`vauth_key` IS NOT NULL
  AND t.`vauth_key` != ''
  AND t.`vauth_key` != 'default';

-- 4.2 从 mcp_service_mappings 表迁移 vauth_key 到 mcp_servers 表
INSERT IGNORE INTO `mcp_servers` (`vauth_key`, `name`, `description`, `enabled`, `created_at`, `updated_at`)
SELECT DISTINCT
    m.`vauth_key` AS `vauth_key`,
    COALESCE(m.`service_name`, CONCAT('Service_', m.`vauth_key`)) AS `name`,
    '' AS `description`,
    m.`enabled` AS `enabled`,
    m.`created_at` AS `created_at`,
    m.`updated_at` AS `updated_at`
FROM `mcp_service_mappings` m
WHERE m.`vauth_key` IS NOT NULL
  AND m.`vauth_key` != ''
  AND m.`vauth_key` != 'default'
  AND NOT EXISTS (
      SELECT 1 FROM `mcp_servers` s WHERE s.`vauth_key` = m.`vauth_key`
  );

-- 4.3 同步更新 mcp_tool_definitions 表的 service_id 字段
-- 将 vauth_key 关联到对应的 mcp_servers.id
UPDATE `mcp_tool_definitions` t
INNER JOIN `mcp_servers` s ON t.`vauth_key` = s.`vauth_key`
SET t.`service_id` = s.`id`
WHERE t.`vauth_key` IS NOT NULL AND t.`vauth_key` != '';

-- ============================================
-- 完成
-- ============================================
SELECT 'Migration completed successfully!' AS status;
