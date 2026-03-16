use `tickets_hunter`;
-- 为了模拟支付功能，创建一个用户钱包表，包含用户ID和余额字段
CREATE TABLE if not exists `user_wallet` (
    `id` bigint(22) NOT NULL AUTO_INCREMENT,
    `user_id` bigint(22) NOT NULL DEFAULT 0 COMMENT '用户ID',
    `balance` bigint NOT NULL DEFAULT 0 COMMENT '余额，单位为分',
    `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
    `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- 查询用户表，给每个用户创建一个钱包记录，初始余额为1000000分（即10000元）
INSERT INTO `user_wallet` (`user_id`, `balance`)
SELECT `id`, 1000000 FROM `user`;