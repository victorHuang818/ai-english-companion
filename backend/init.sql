-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS ai_interview_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ai_interview_db;

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

-- 2. 简历表 Resumes 
CREATE TABLE `resumes` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `user_id` CHAR(36) NOT NULL COMMENT '用户ID',
    `content` TEXT NOT NULL COMMENT '简历内容(Markdown格式)',
    `pdf_url` VARCHAR(1024) DEFAULT NULL COMMENT 'pdf地址',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_resume_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 3. 岗位表 Job Profiles
CREATE TABLE `job_profiles` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `creator_id` char(36) not null comment '创建人ID',
    `name` VARCHAR(255) NOT NULL COMMENT '岗位名称',
    `description` TEXT COMMENT '岗位描述',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 4. 面试会话表 Interview Sessions (已关联简历表)
CREATE TABLE `interview_sessions` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `user_id` CHAR(36) NOT NULL COMMENT '用户ID',
    `resume_id` CHAR(36) NOT NULL COMMENT '简历ID',
    `job_profile_id` CHAR(36) NOT NULL COMMENT '岗位ID',
    `status` ENUM('pending', 'in_progress', 'completed', 'cancelled') DEFAULT 'pending' COMMENT '状态',
    `evaluation_report` JSON DEFAULT NULL COMMENT '面试评估报告',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_session_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_session_resume` FOREIGN KEY (`resume_id`) REFERENCES `resumes` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_session_job` FOREIGN KEY (`job_profile_id`) REFERENCES `job_profiles` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 5. 对话记录表 Interview Dialogues
CREATE TABLE `interview_dialogues` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `interview_session_id` CHAR(36) NOT NULL COMMENT '面试会话ID',
    `role` ENUM('user', 'interviewer', 'assistant') NOT NULL COMMENT '角色',
    `audio_url` VARCHAR(1024) DEFAULT NULL COMMENT '音频URL(optional)',
    `content` TEXT NOT NULL COMMENT '对话内容',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_session_time` (`interview_session_id`, `created_at`), -- 联合索引优化时间轴查询
    CONSTRAINT `fk_dialogue_session` FOREIGN KEY (`interview_session_id`) REFERENCES `interview_sessions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;