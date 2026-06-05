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
