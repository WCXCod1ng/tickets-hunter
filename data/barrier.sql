CREATE DATABASE if not exists `dtm_barrier`;
USE `dtm_barrier`;

CREATE TABLE if not exists `barrier` (
    `id` bigint(22) NOT NULL AUTO_INCREMENT,
    `trans_type` varchar(45) NOT NULL DEFAULT '',
    `gid` varchar(128) NOT NULL DEFAULT '',
    `branch_id` varchar(128) NOT NULL DEFAULT '',
    `op` varchar(45) NOT NULL DEFAULT '',
    `barrier_id` varchar(45) NOT NULL DEFAULT '',
    `reason` varchar(45) DEFAULT '' COMMENT 'the branch type who insert this record',
    `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
    `update_time` datetime DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_uniq_branch` (`trans_type`,`gid`,`branch_id`,`op`,`barrier_id`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;