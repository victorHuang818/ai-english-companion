-- 如果已存在数据库，先删除再重建
DROP DATABASE IF EXISTS ai_english_db;
CREATE DATABASE ai_english_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ai_english_db;

-- 1. 用户表 Users
CREATE TABLE `users` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    `email` VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱',
    `password` VARCHAR(255) NOT NULL COMMENT '加密后的密码',
    `daily_free_tokens` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '当天剩余免费Token额度 (单位: 个)',
    `recharge_tokens` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户充值的剩余Token额度 (单位: 个)',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '注册/创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 2. 用户背景档案表 User Profiles
CREATE TABLE `user_profiles` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `user_id` CHAR(36) NOT NULL COMMENT '用户ID',
    `english_level` VARCHAR(50) NOT NULL COMMENT '英语水平级别 (beginner/intermediate/advanced)',
    `learning_target` TEXT NOT NULL COMMENT '学习目标 (例如：雅思口语、西餐点餐、日常会话等)',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 3. 场景配置表 Scenarios
CREATE TABLE `scenarios` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `creator_id` CHAR(36) NOT NULL COMMENT '创建人ID',
    `name` VARCHAR(255) NOT NULL COMMENT '场景名称 (例如：Ordering Food)',
    `description` TEXT COMMENT '场景上下文描述/系统引导 Prompt',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 4. 练习会话表 Practice Sessions
CREATE TABLE `practice_sessions` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `user_id` CHAR(36) NOT NULL COMMENT '用户ID',
    `user_profile_id` CHAR(36) NOT NULL COMMENT '用户档案ID',
    `scenario_id` CHAR(36) NOT NULL COMMENT '场景ID',
    `status` ENUM('pending', 'in_progress', 'completed', 'cancelled') DEFAULT 'pending' COMMENT '状态',
    `evaluation_report` JSON DEFAULT NULL COMMENT '口语练习最终评估报告',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 5. 对话记录表 Dialogues
CREATE TABLE `dialogues` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `practice_session_id` CHAR(36) NOT NULL COMMENT '练习会话ID',
    `role` ENUM('user', 'interviewer', 'assistant') NOT NULL COMMENT '角色 (保持与老代码兼容使用这三种枚举)',
    `audio_url` VARCHAR(1024) DEFAULT NULL COMMENT '录音音频URL',
    `content` TEXT NOT NULL COMMENT '对话文本内容',
    `evaluation` TEXT DEFAULT NULL COMMENT 'AI对本轮回答的评估打分(JSON格式文本)',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_session_time` (`practice_session_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 插入4个公共预设场景
INSERT INTO `scenarios` (`id`, `creator_id`, `name`, `description`, `created_at`) VALUES
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000000', 'Job Interview', 'Simulate a professional interview at a multinational corporation. Focus on project experience, career goals, and behavioral questions.', CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000000', 'Ordering Food', 'Practice ordering food, asking about the menu, and handling payments in a dining or restaurant context.', CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000000', 'IELTS Speaking', 'Simulate IELTS speaking test sections (Part 1, Part 2, Part 3) under standardized constraints.', CURRENT_TIMESTAMP),
('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000000', 'Business Meeting', 'Practice presenting an idea, reporting project status, or discussing proposals in a business meeting.', CURRENT_TIMESTAMP);