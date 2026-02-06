use tickets_hunter;
CREATE TABLE `user` (
    `id` bigint(20) UNSIGNED NOT NULL COMMENT '主键ID',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `delete_time` datetime DEFAULT NULL COMMENT '删除时间(NULL表示未删除)',
    `mobile` char(11) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '手机号',
    `password` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT '密码(加密存储)',
    `nickname` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '昵称',
    `sex` tinyint(1) NOT NULL DEFAULT '0' COMMENT '性别 0:未知, 1:男, 2:女',
    `avatar` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '头像URL',
    `info` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL COMMENT '简介',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_mobile` (`mobile`) USING BTREE COMMENT '手机号唯一索引',
    KEY `idx_create_time` (`create_time`) USING BTREE COMMENT '创建时间索引(用于统计)'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户表';