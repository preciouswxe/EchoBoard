
CREATE TABLE `user` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `user_id` bigint(20) NOT NULL,
    `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL ,
    `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL ,
    `email` varchar(64) COLLATE utf8mb4_general_ci,
    `gender` tinyint(4) NOT NULL DEFAULT '0',
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_username` (`username`) USING BTREE,
    UNIQUE KEY `idx_user_od` (`user_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET = utf8m4 COLLATE=utf8mb4_general_ci;

# 防止相同表
DROP TABLE IF EXISTS `community`;

CREATE TABLE `community` (
                             `id` int(11) NOT NULL AUTO_INCREMENT,
                             `community_id` int(10) UNSIGNED NOT NULL,
                             `community_name` VARCHAR(128) COLLATE utf8mb4_general_ci NOT NULL,
                             `introduction` VARCHAR(256) COLLATE utf8mb4_general_ci NOT NULL,
                             `create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                             `update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `idx_community_id` (`community_id`),
                             UNIQUE KEY `idx_community_name` (`community_name`)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

INSERT INTO `community` VALUES ('1','1','Go','Golang','2024-11-01 08:10:10','2024-11-01 08:10:10');
INSERT INTO `community` VALUES ('2','2','Leetcode','刷题刷题','2025-01-01 08:01:10','2025-01-01 08:01:10');
INSERT INTO `community` VALUES ('3','3','CS:Go','Rush B...','2018-08-07 20:30:10','2018-08-07 20:30:10');
INSERT INTO `community` VALUES ('4','4','LOL','欢迎来到英雄联盟！','2016-01-01 08:00:00','2016-01-01 08:00:00');


