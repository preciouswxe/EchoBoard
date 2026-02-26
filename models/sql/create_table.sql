
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
) ENGINE=InnoDB DEFAULT CHARSET = utf8mb4 COLLATE=utf8mb4_general_ci;

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



DROP TABLE IF EXISTS `post`;

# PRIMARY KEY：主键，唯一标识一行记录，自动建立唯一索引
# UNIQUE KEY：唯一索引，列的值必须唯一（可用于非主键约束）
# KEY（或 INDEX）：普通索引，加快某列的查询速度

CREATE TABLE `post` (
                        `id` BIGINT(20) NOT NULL AUTO_INCREMENT,
                        `post_id` BIGINT(20) NOT NULL COMMENT '帖子id',
                        `title` VARCHAR(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',
                        `content` VARCHAR(8192) COLLATE utf8mb4_general_ci NOT NULL COMMENT '内容',
                        `author_id` BIGINT(20) NOT NULL COMMENT '作者的用户id',
                        `community_id` BIGINT(20) NOT NULL COMMENT '所属社区',
                        `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '帖子状态',
                        `create_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                        `update_time` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                        PRIMARY KEY (`id`),
                        UNIQUE KEY `idx_post_id` (`post_id`),
                        KEY `idx_author_id` (`author_id`),
                        KEY `idx_community_id` (`community_id`)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;


-- ================================================
-- 1. post 表新增计数字段
-- ================================================
ALTER TABLE `post`
    ADD COLUMN `like_num` int NOT NULL DEFAULT 0 COMMENT '点赞数',
	ADD COLUMN `collect_num` int NOT NULL DEFAULT 0 COMMENT '收藏数',
	ADD COLUMN `comment_num` int NOT NULL DEFAULT 0 COMMENT '评论数';


-- ================================================
-- 2. 点赞记录表
-- ================================================
CREATE TABLE `post_like`(
                            `id` bigint not null auto_increment COMMENT '记录ID',
                            `user_id` bigint not null COMMENT '用户ID',
                            `post_id` bigint not null COMMENT '帖子ID',
                            `create_time` TIMESTAMP not null DEFAULT CURRENT_TIMESTAMP,

                            PRIMARY KEY (`id`),
                            UNIQUE KEY `uk_user_post` (`user_id`, `post_id`),
                            KEY `idx_user_id` (`user_id`),
                            KEY `idx_post_id` (`post_id`)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT='帖子点赞记录';


-- ================================================
-- 3. 收藏记录表
-- ================================================
CREATE TABLE `post_collect`(
                               `id` bigint not null auto_increment COMMENT '记录ID',
                               `user_id` bigint not null COMMENT '用户ID',
                               `post_id` bigint not null COMMENT '帖子ID',
                               `create_time` TIMESTAMP not null DEFAULT CURRENT_TIMESTAMP,

                               PRIMARY KEY (`id`),
                               UNIQUE KEY `uk_user_post` (`user_id`, `post_id`),
                               KEY `idx_user_id` (`user_id`),
                               KEY `idx_post_id` (`post_id`)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT='帖子收藏记录';


-- ================================================
-- 4. 评论表
-- ================================================
CREATE TABLE `post_comment` (
                                `id` bigint not null auto_increment COMMENT '记录ID',
                                `comment_id` bigint not null COMMENT '雪花算法评论id',
                                `post_id` bigint not null COMMENT '帖子id',
                                `user_id` bigint not null COMMENT '评论者的user_id',
                                `parent_id` bigint null DEFAULT NULL COMMENT 'NULL表示顶级评论（没有上一级评论），否则为回复的评论id',
                                `content` VARCHAR(1024) NOT NULL COMMENT '评论内容',
                                `status` tinyint not null DEFAULT 1 COMMENT '状态 1正常 0删除（软删除）',
                                `create_time` TIMESTAMP not null DEFAULT CURRENT_TIMESTAMP,
                                `update_time` TIMESTAMP not null DEFAULT CURRENT_TIMESTAMP on UPDATE CURRENT_TIMESTAMP,

                                PRIMARY KEY (`id`),
                                UNIQUE KEY `uk_comment_id` (`comment_id`),
                                KEY `idx_post_id` (`post_id`),
                                KEY `idx_user_id` (`user_id`),
                                KEY `idx_parent_id` (`parent_id`)
)ENGINE=INNODB DEFAULT CHARSET=utf8mb4 COMMENT='帖子评论表';
