-- ============================================
-- 数据库迁移脚本（完整版）
-- 用于初始化 mcp_server 数据库
-- ============================================

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
SET character_set_connection=utf8mb4;

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS `mcp_server` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `mcp_server`;

-- ============================================
-- 步骤1: 创建 mcp_auth_keys 表（认证密钥表）
-- ============================================
CREATE TABLE IF NOT EXISTS `mcp_auth_keys` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `key` VARCHAR(256) NOT NULL COMMENT '认证密钥',
    `name` VARCHAR(128) NOT NULL COMMENT '密钥名称',
    `description` VARCHAR(512) DEFAULT '' COMMENT '密钥描述',
    `state` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 1-正常 0-删除',
    `last_used_at` DATETIME DEFAULT NULL COMMENT '最后使用时间',
    `expires_at` DATETIME DEFAULT NULL COMMENT '过期时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key` (`key`),
    INDEX `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证密钥表';

-- ============================================
-- 步骤2: 创建 mcp_servers 表（MCP服务信息）
-- ============================================
CREATE TABLE IF NOT EXISTS `mcp_servers` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `vauth_key` VARCHAR(128) NOT NULL COMMENT '认证密钥',
    `name` VARCHAR(128) NOT NULL COMMENT '服务名称',
    `description` VARCHAR(512) DEFAULT '' COMMENT '服务描述',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `state` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 1-正常 0-删除',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_vauth_key` (`vauth_key`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MCP服务表';

-- ============================================
-- 步骤3: 创建 mcp_http_services 表（HTTP服务配置）
-- ============================================
CREATE TABLE IF NOT EXISTS `mcp_http_services` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` VARCHAR(128) NOT NULL COMMENT '服务名称',
    `description` VARCHAR(512) DEFAULT '' COMMENT '服务描述',
    `target_url` VARCHAR(512) NOT NULL COMMENT '目标URL',
    `method` VARCHAR(16) NOT NULL DEFAULT 'POST' COMMENT 'HTTP方法',
    `body_type` VARCHAR(32) DEFAULT 'JSON' COMMENT '请求体类型',
    `headers` TEXT COMMENT '请求头JSON',
    `timeout_seconds` INT NOT NULL DEFAULT 30 COMMENT '超时秒数',
    `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    `validation_script` TEXT COMMENT '验证脚本',
    `validation_enabled` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否启用验证',
    `request_transform_script` TEXT COMMENT '请求转换脚本',
    `response_transform_script` TEXT COMMENT '响应转换脚本',
    `input_schema` TEXT COMMENT '入参JSON Schema',
    `output_schema` TEXT COMMENT '出参JSON Schema',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `state` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 1-正常 0-删除',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_name` (`name`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HTTP服务配置表';

-- ============================================
-- 步骤4: 创建 mcp_tool_definitions 表（工具定义）
-- ============================================
CREATE TABLE IF NOT EXISTS `mcp_tool_definitions` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` VARCHAR(128) NOT NULL COMMENT '工具名称',
    `description` TEXT COMMENT '工具描述',
    `service_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联的HTTP服务ID',
    `parameters` TEXT COMMENT '参数定义JSON',
    `input_mapping` TEXT COMMENT '入参映射配置',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `output_mapping` TEXT COMMENT '出参映射配置',
    `state` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 1-正常 0-删除',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_name` (`name`),
    INDEX `idx_service_id` (`service_id`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工具定义表';

-- ============================================
-- 步骤5: 创建 tool_mcp_server_bindings 表（工具与MCP Server绑定）
-- ============================================
CREATE TABLE IF NOT EXISTS `tool_mcp_server_bindings` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `tool_id` INT UNSIGNED NOT NULL COMMENT '工具ID',
    `server_id` INT UNSIGNED NOT NULL COMMENT '服务ID',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `state` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 1-正常 0-删除',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_tool_server` (`tool_id`, `server_id`),
    INDEX `idx_tool_id` (`tool_id`),
    INDEX `idx_server_id` (`server_id`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工具与MCP Server绑定表';

-- ============================================
-- 步骤6: 创建 server_build_info 表（MCP Server 构建信息表）
-- ============================================
CREATE TABLE IF NOT EXISTS `server_build_info` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `server_id` BIGINT UNSIGNED NOT NULL COMMENT '关联 mcp_servers.id',
    `version` INT UNSIGNED NOT NULL DEFAULT 1 COMMENT '版本号',
    `build_uuid` VARCHAR(36) NOT NULL COMMENT '构建UUID',
    `hash` VARCHAR(64) NOT NULL COMMENT 'SHA256',
    `build_data` TEXT COMMENT 'JSON: 工具和HTTP服务快照合并',
    `state` INT NOT NULL DEFAULT 1 COMMENT '状态 1-有效 0-失效',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY `uk_build_uuid` (`build_uuid`),
    INDEX `idx_hash` (`hash`),
    INDEX `idx_server_state` (`server_id`, `state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MCP Server 构建信息表';

-- ============================================
-- 步骤7: 创建 users 表（用户表）
-- ============================================
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `username` VARCHAR(64) NOT NULL COMMENT '用户名',
    `password` VARCHAR(256) NOT NULL COMMENT '密码（bcrypt加密）',
    `email` VARCHAR(128) DEFAULT '' COMMENT '邮箱',
    `nickname` VARCHAR(128) DEFAULT '' COMMENT '昵称',
    `role` VARCHAR(32) NOT NULL DEFAULT 'user' COMMENT '角色',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `state` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态 1-正常 0-删除',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    INDEX `idx_role` (`role`),
    INDEX `idx_enabled` (`enabled`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ============================================
-- 完成
-- ============================================
SELECT 'Migration completed successfully!' AS status;