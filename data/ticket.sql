use tickets_hunter;
-- 1. 演唱会场次表
CREATE TABLE `ticket_event` (
                                `id` bigint NOT NULL AUTO_INCREMENT,
                                `title` varchar(128) NOT NULL COMMENT '演唱会标题',
                                `cover_url` varchar(255) NOT NULL DEFAULT '' COMMENT '海报封面图',
                                `show_time` datetime NOT NULL COMMENT '演出开始时间',
                                `venue` varchar(128) NOT NULL COMMENT '场馆名称',
                                `sale_start_time` datetime NOT NULL COMMENT '开售时间(到达此时间才允许锁座)',
                                `sale_end_time` datetime NOT NULL COMMENT '停售时间',
                                `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态: 0下架, 1上架',
                                `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                                `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                                PRIMARY KEY (`id`),
                                KEY `idx_sale_start` (`sale_start_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='演唱会场次表';

-- 2. 座位表 (核心表)
CREATE TABLE `ticket_seat` (
                               `id` bigint NOT NULL AUTO_INCREMENT,
                               `event_id` bigint NOT NULL COMMENT '所属场次ID',
                               `seat_type` tinyint NOT NULL DEFAULT '1' COMMENT '座位类型: 1普通, 2VIP, 3内场',
                               `section` varchar(32) NOT NULL COMMENT '区域(如A区)',
                               `row_no` int NOT NULL COMMENT '排号',
                               `seat_no` int NOT NULL COMMENT '座位号',
                               `price` decimal(10,2) NOT NULL COMMENT '座位价格',
                               `status` tinyint NOT NULL DEFAULT '0' COMMENT '状态: 0可选, 1锁定(未支付), 2已售(已出票)',
                               `version` int NOT NULL DEFAULT '0' COMMENT '乐观锁版本号(防超卖核心)',
                               `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
                               `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                               PRIMARY KEY (`id`),
                               UNIQUE KEY `idx_event_seat` (`event_id`,`section`,`row_no`,`seat_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='座位表';