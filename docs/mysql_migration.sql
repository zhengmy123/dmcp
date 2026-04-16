-- ============================================
-- 数据库初始化和迁移脚本
-- 表名与 GORM AutoMigrate 保持一致 (mcp_ 前缀)
-- ============================================

SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;
SET character_set_connection=utf8mb4;

USE `mcp_server`;

-- ============================================
-- 步骤1: 创建所有表（使用 IF NOT EXISTS）
-- ============================================

-- 创建 mcp_users 表
CREATE TABLE IF NOT EXISTS `mcp_users` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `username` VARCHAR(64) NOT NULL COMMENT '用户名',
    `password_hash` VARCHAR(256) NOT NULL COMMENT '密码哈希',
    `name` VARCHAR(128) DEFAULT '' COMMENT '显示名称',
    `email` VARCHAR(256) DEFAULT '' COMMENT '邮箱',
    `role` VARCHAR(32) NOT NULL DEFAULT 'user' COMMENT '角色: admin, user',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `last_login_at` DATETIME DEFAULT NULL COMMENT '最后登录时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_username` (`username`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 创建 mcp_auth_keys 表
CREATE TABLE IF NOT EXISTS `mcp_auth_keys` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `key_id` VARCHAR(64) NOT NULL COMMENT 'Key ID',
    `token` VARCHAR(128) NOT NULL COMMENT '访问令牌',
    `secret` VARCHAR(256) NOT NULL COMMENT '密钥',
    `name` VARCHAR(128) DEFAULT '' COMMENT '名称/描述',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '启用状态',
    `last_used_at` DATETIME DEFAULT NULL COMMENT '最后使用时间',
    `expires_at` DATETIME DEFAULT NULL COMMENT '过期时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_key_id` (`key_id`),
    UNIQUE KEY `uk_token` (`token`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证密钥表';

-- 创建 mcp_tool_definitions 表
CREATE TABLE IF NOT EXISTS `mcp_tool_definitions` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `vauth_key` VARCHAR(128) NOT NULL COMMENT '认证密钥',
    `server_desc` VARCHAR(512) DEFAULT '' COMMENT '服务描述',
    `name` VARCHAR(128) NOT NULL COMMENT '工具名称',
    `description` TEXT COMMENT '工具描述',
    `enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用',
    `parameters` TEXT COMMENT '参数定义 JSON',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_vauth_key` (`vauth_key`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工具定义表';

-- 创建 mcp_http_services 表
CREATE TABLE IF NOT EXISTS `mcp_http_services` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` VARCHAR(128) NOT NULL COMMENT '服务名称',
    `description` VARCHAR(512) DEFAULT '' COMMENT '服务描述',
    `target_url` VARCHAR(512) NOT NULL COMMENT '目标URL',
    `method` VARCHAR(16) NOT NULL DEFAULT 'POST' COMMENT 'HTTP方法',
    `body_type` VARCHAR(32) DEFAULT 'JSON' COMMENT '请求体类型',
    `headers` TEXT COMMENT '请求头 JSON',
    `timeout_seconds` INT DEFAULT 30 COMMENT '超时秒数',
    `retry_count` INT DEFAULT 0 COMMENT '重试次数',
    `validation_script` TEXT COMMENT '验证脚本',
    `validation_enabled` TINYINT(1) DEFAULT 0 COMMENT '启用验证',
    `signature_enabled` TINYINT(1) DEFAULT 0 COMMENT '启用签名',
    `signature_algorithm` VARCHAR(32) DEFAULT '' COMMENT '签名算法',
    `signature_key` VARCHAR(256) DEFAULT '' COMMENT '签名密钥',
    `signature_header` VARCHAR(64) DEFAULT '' COMMENT '签名头名称',
    `signature_location` VARCHAR(16) DEFAULT 'header' COMMENT '签名位置: header, url, both',
    `signature_query_param` VARCHAR(64) DEFAULT '' COMMENT '签名URL参数名',
    `signature_script` TEXT COMMENT '动态JS验签脚本',
    `request_transform_script` TEXT COMMENT '请求转换脚本',
    `response_transform_script` TEXT COMMENT '响应转换脚本',
    `input_schema` TEXT COMMENT '入参JSON Schema',
    `output_schema` TEXT COMMENT '出参JSON Schema',
    `enabled` TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='HTTP服务配置表';

-- 创建 mcp_service_mappings 表
CREATE TABLE IF NOT EXISTS `mcp_service_mappings` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `service_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '服务ID',
    `service_name` VARCHAR(128) NOT NULL COMMENT '服务名称',
    `vauth_key` VARCHAR(128) NOT NULL COMMENT '认证密钥',
    `json_schema` TEXT COMMENT 'JSON Schema',
    `schema_hash` VARCHAR(64) DEFAULT '' COMMENT 'Schema哈希',
    `mapping_config` TEXT COMMENT '映射配置',
    `enabled` TINYINT(1) DEFAULT 1 COMMENT '是否启用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_service_vauth` (`service_id`, `vauth_key`),
    INDEX `idx_vauth_key` (`vauth_key`),
    INDEX `idx_enabled` (`enabled`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务映射表';

-- 创建 mcp_service_logs 表
CREATE TABLE IF NOT EXISTS `mcp_service_logs` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `service_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '服务ID',
    `level` VARCHAR(16) NOT NULL DEFAULT 'info' COMMENT '日志级别',
    `message` TEXT COMMENT '日志消息',
    `request_data` TEXT COMMENT '请求数据',
    `response_data` TEXT COMMENT '响应数据',
    `error` TEXT COMMENT '错误信息',
    `duration_ms` INT DEFAULT 0 COMMENT '耗时毫秒',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    INDEX `idx_service_id` (`service_id`),
    INDEX `idx_level` (`level`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='服务日志表';

-- ============================================
-- 完成
-- ============================================
SELECT 'Database tables created successfully!' AS status;
