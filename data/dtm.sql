CREATE DATABASE IF NOT EXISTS `dtm` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `dtm`;

CREATE TABLE IF NOT EXISTS `trans_global` (
                                              `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `gid` varchar(128) NOT NULL,
    `trans_type` varchar(45) NOT NULL,
    `status` varchar(45) NOT NULL,
    `query_prepared` varchar(1024) NOT NULL,
    `protocol` varchar(45) NOT NULL,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `finish_time` datetime DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `gid` (`gid`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `trans_branch` (
                                              `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `gid` varchar(128) NOT NULL,
    `branch_id` varchar(128) NOT NULL,
    `trans_type` varchar(45) NOT NULL,
    `status` varchar(45) NOT NULL,
    `data` longtext,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `gid_branch_id` (`gid`, `branch_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `kv` (
                                    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `cat` varchar(45) NOT NULL,
    `k` varchar(45) NOT NULL,
    `v` longtext,
    PRIMARY KEY (`id`),
    UNIQUE KEY `cat_k` (`cat`, `k`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;