use tickets_hunter;
-- 3. 订单主表
CREATE TABLE `order_main` (
                              `id` bigint NOT NULL AUTO_INCREMENT,
                              `order_sn` varchar(64) NOT NULL COMMENT '订单流水号(雪花算法生成)',
                              `user_id` bigint NOT NULL COMMENT '用户ID',
                              `event_id` bigint NOT NULL COMMENT '场次ID',
                              `seat_id` bigint NOT NULL COMMENT '座位ID',
                              `section` varchar(32) NOT NULL DEFAULT '' COMMENT '区域',
                              `seat_index` int NOT NULL DEFAULT 0 COMMENT 'Redis Bitmap座位索引',
                              `amount` bigint NOT NULL COMMENT '订单实付金额',
                              `status` tinyint NOT NULL DEFAULT '0' COMMENT '状态: 10待支付, 20已支付(待出票), 30已出票(已完成)，40超时关闭，51已退款',
                              `expire_time` datetime NOT NULL COMMENT '订单支付过期时间(通常为创建时间+15分钟)',
                              `pay_time` datetime DEFAULT NULL COMMENT '实际支付时间',
                              `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                              `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                              PRIMARY KEY (`id`),
                              UNIQUE KEY `idx_order_sn` (`order_sn`),
                              KEY `idx_user_id` (`user_id`),
                              KEY `idx_status_expire` (`status`,`expire_time`,`order_sn`) COMMENT '用于延迟队列或定时任务扫表'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';