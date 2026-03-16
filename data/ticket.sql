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
create table `ticket_seat`
(
    id          bigint not null auto_increment,
    event_id    bigint                             not null comment '所属场次ID',
    seat_type   tinyint  default 1                 not null comment '座位类型: 1普通, 2VIP, 3内场',
    section     varchar(32)                        not null comment '区域(如A区)',
    seat_index  int                                not null comment '区域内相对索引(从0开始, 映射Redis BitMap)',
    row_no      int                                not null comment '排号',
    seat_no     int                                not null comment '座位号',
    price       bigint                             not null comment '座位价格，以分为单位',
    status      tinyint  default 0                 not null comment '状态: 0可选, 1锁定(未支付), 2已售(已出票)',
    version     int      default 0                 not null comment '乐观锁版本号(防超卖核心)',
    create_time datetime default CURRENT_TIMESTAMP not null,
    update_time datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),

    UNIQUE KEY `idx_event_section_index` (`event_id`, `section`, `seat_index`),

    UNIQUE KEY `idx_event_seat` (`event_id`,`section`,`row_no`,`seat_no`)
) comment '座位表';

