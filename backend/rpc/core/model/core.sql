CREATE TABLE `resumes` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `user_id` CHAR(36) NOT NULL COMMENT '用户ID',
    `content` TEXT NOT NULL COMMENT '简历内容(Markdown或纯文本格式)',
    `pdf_url` VARCHAR(1024) DEFAULT NULL COMMENT 'pdf地址',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_resume_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `job_profiles` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `creator_id` char(36) not null comment '创建人ID',
    `name` VARCHAR(255) NOT NULL COMMENT '岗位名称',
    `description` TEXT COMMENT '岗位描述',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
    INDEX `idx_session_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `interview_dialogues` (
    `id` CHAR(36) NOT NULL COMMENT '唯一标识 (UUID)',
    `interview_session_id` CHAR(36) NOT NULL COMMENT '面试会话ID',
    `role` ENUM('user', 'interviewer', 'assistant') NOT NULL COMMENT '角色',
    `audio_url` VARCHAR(1024) DEFAULT NULL COMMENT '音频URL(optional)',
    `content` TEXT NOT NULL COMMENT '对话内容',
    `evaluation` JSON DEFAULT NULL COMMENT '此轮回答的 AI 评估报告 (JSON 格式文本)',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    INDEX `idx_session_time` (`interview_session_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
