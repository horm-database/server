CREATE TABLE `tbl_access_db` (
                                 `id` int NOT NULL AUTO_INCREMENT,
                                 `appid` bigint NOT NULL DEFAULT '0' COMMENT '应用appid',
                                 `db` int NOT NULL DEFAULT '0' COMMENT '数据库id',
                                 `privilege` varchar(32) NOT NULL DEFAULT '' COMMENT '库权限： 1-表权限（拥有之后可以对库下所有表都增删改查） 2-查，3-增/改 4-删  99-超级权限，在表权限之外，还可以 CREATE、DROP 表',
                                 `status` tinyint NOT NULL DEFAULT '3' COMMENT '状态：1-正常 2-下线 3-审核中 4-审核撤回 5-拒绝',
                                 `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                 `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `appid` (`appid`,`db`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='应用访问数据库权限'

CREATE TABLE `tbl_access_table` (
                                    `id` int NOT NULL AUTO_INCREMENT,
                                    `appid` bigint NOT NULL DEFAULT '0' COMMENT '应用appid',
                                    `table_id` int NOT NULL DEFAULT '0' COMMENT '表id',
                                    `privilege` varchar(256) NOT NULL DEFAULT '2' COMMENT '权限 1-表所有权限（增删改查），2-查，3-增/改 4-删',
                                    `status` tinyint NOT NULL DEFAULT '3' COMMENT '状态：1-正常 2-下线 3-审核中 4-审核撤回 5-拒绝',
                                    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                    PRIMARY KEY (`id`),
                                    UNIQUE KEY `appid` (`appid`,`table_id`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='应用访问表权限'

CREATE TABLE `tbl_app_info` (
                                `appid` bigint NOT NULL DEFAULT '' COMMENT '应用appid',
                                `name` varchar(64) NOT NULL DEFAULT '' COMMENT '应用名称',
                                `secret` varchar(64) NOT NULL DEFAULT '' COMMENT '应用秘钥',
                                `intro` varchar(512) NOT NULL DEFAULT '' COMMENT '简介',
                                `creator` bigint NOT NULL DEFAULT '0' COMMENT 'creator',
                                `manager` varchar(1025) NOT NULL DEFAULT '' COMMENT '管理员，多个逗号分隔',
                                `status` tinyint NOT NULL DEFAULT '1' COMMENT '1-正常 2-下线',
                                `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                PRIMARY KEY (`id`),
                                UNIQUE KEY `appid` (`appid`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='应用信息'

CREATE TABLE `tbl_collect_table` (
                                     `id` int NOT NULL AUTO_INCREMENT,
                                     `userid` bigint NOT NULL DEFAULT '0' COMMENT '用户id',
                                     `table_id` int NOT NULL COMMENT '表id',
                                     `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                     `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                     PRIMARY KEY (`id`),
                                     UNIQUE KEY `user_table` (`userid`,`table_id`),
                                     KEY `table_id` (`table_id`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='收藏的表'

CREATE TABLE `tbl_db` (
                          `id` int NOT NULL AUTO_INCREMENT,
                          `name` varchar(64) NOT NULL DEFAULT '' COMMENT '数据库名称',
                          `intro` varchar(256) NOT NULL COMMENT '简介',
                          `desc` varchar(512) NOT NULL DEFAULT '' COMMENT '详细介绍',
                          `product_id` int NOT NULL DEFAULT '0' COMMENT '产品id',
                          `type` int NOT NULL DEFAULT '3' COMMENT '数据库类型 0-nil（仅执行拦截器） 1-elastic 2-mongo 3-redis 10-mysql 11-postgresql 12-clickhouse 13-oracle 14-DB2 15-sqlite',
                          `version` varchar(16) NOT NULL DEFAULT '' COMMENT '数据库版本，比如elastic v6，v7',
                          `network` varchar(64) NOT NULL DEFAULT '' COMMENT 'network',
                          `address` varchar(4096) NOT NULL DEFAULT '' COMMENT 'address',
                          `bak_address` varchar(4096) NOT NULL DEFAULT '' COMMENT 'backup address',
                          `password` varchar(256) NOT NULL DEFAULT '' COMMENT 'password',
                          `write_timeout` int DEFAULT NULL COMMENT '写超时（毫秒）',
                          `read_timeout` int NOT NULL DEFAULT '0' COMMENT '读超时（毫秒）',
                          `warn_timeout` int NOT NULL DEFAULT '200' COMMENT '告警超时（ms），如果请求耗时超过这个时间，就会打 warning 日志',
                          `omit_error` tinyint NOT NULL DEFAULT '0' COMMENT '是否忽略 error 日志，0-否 1-是',
                          `debug` tinyint NOT NULL DEFAULT '0' COMMENT '是否开启 debug 日志，正常的数据库请求也会被打印到日志，0-否 1-是，会造成海量日志，慎重开启',
                          `creator` bigint NOT NULL DEFAULT '0' COMMENT 'creator',
                          `manager` varchar(1025) NOT NULL DEFAULT '' COMMENT '管理员，多个逗号分隔',
                          `status` tinyint NOT NULL DEFAULT '1' COMMENT '1-正常 2-下线',
                          `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                          `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                          PRIMARY KEY (`id`),
                          UNIQUE KEY `name` (`name`),
                          KEY `productid` (`product_id`)
) ENGINE=InnoDB AUTO_INCREMENT=31 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='数据库表'

CREATE TABLE `tbl_filter` (
                              `id` int NOT NULL AUTO_INCREMENT,
                              `name` varchar(128) NOT NULL DEFAULT '' COMMENT '插件名称',
                              `version` varchar(1024) NOT NULL DEFAULT '0' COMMENT '所有支持的插件版本，逗号分开',
                              `func` varchar(128) NOT NULL COMMENT '插件注册函数名',
                              `type` tinyint NOT NULL DEFAULT '1' COMMENT '插件类型 1-前置插件 2-后置插件',
                              `online` tinyint NOT NULL DEFAULT '1' COMMENT '状态 1-上线 2-下线',
                              `source` tinyint NOT NULL DEFAULT '1' COMMENT '来源：1-官方插件 2-第三方插件 3-个人插件',
                              `desc` varchar(512) NOT NULL DEFAULT '' COMMENT '插件插件介绍',
                              `creator` bigint NOT NULL DEFAULT '0' COMMENT 'creator',
                              `manager` varchar(512) NOT NULL DEFAULT '' COMMENT '管理员，多个逗号分隔',
                              `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                              `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                              PRIMARY KEY (`id`),
                              UNIQUE KEY `name` (`name`),
                              UNIQUE KEY `func_name` (`func`)
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='插件信息'

CREATE TABLE `tbl_filter_config` (
                                     `id` int NOT NULL AUTO_INCREMENT,
                                     `filter_id` int NOT NULL COMMENT '插件id',
                                     `filter_version` int NOT NULL DEFAULT '0' COMMENT 'filter版本',
                                     `key` varchar(64) NOT NULL DEFAULT '' COMMENT '插件配置 key ',
                                     `name` varchar(64) NOT NULL DEFAULT '' COMMENT '插件配置名',
                                     `type` tinyint NOT NULL DEFAULT '1' COMMENT '配置类型 1-bool、2-string、3-int、4-uint、5-float、6-枚举、7-时间、8-array、9-map、10-multi-conf',
                                     `not_null` tinyint NOT NULL DEFAULT '1' COMMENT '是否必输 1-是 2-否',
                                     `more_info` varchar(4096) NOT NULL DEFAULT '' COMMENT '更多细节，例如单选、多选、时间、array、map、multi-conf 等',
                                     `default` varchar(512) NOT NULL DEFAULT '' COMMENT '默认值，仅用于预填充配置值。',
                                     `desc` varchar(512) NOT NULL DEFAULT '' COMMENT '配置描述',
                                     `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                     `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                     PRIMARY KEY (`id`),
                                     UNIQUE KEY `filter_config` (`filter_id`,`filter_version`,`key`)
) ENGINE=InnoDB AUTO_INCREMENT=26 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='插件配置'

CREATE TABLE `tbl_product` (
                               `id` int NOT NULL AUTO_INCREMENT,
                               `name` varchar(64) NOT NULL DEFAULT '' COMMENT '产品名称',
                               `intro` varchar(512) NOT NULL DEFAULT '' COMMENT '简介',
                               `creator` bigint NOT NULL DEFAULT '0' COMMENT 'creator',
                               `manager` varchar(1025) NOT NULL DEFAULT '' COMMENT '管理员，多个逗号分隔',
                               `status` tinyint NOT NULL DEFAULT '1' COMMENT '1-正常 2-下线',
                               `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                               `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                               PRIMARY KEY (`id`),
                               UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=73 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='数据库表'

CREATE TABLE `tbl_product_member` (
                                      `id` int NOT NULL AUTO_INCREMENT,
                                      `product_id` int NOT NULL DEFAULT '0' COMMENT 'product id',
                                      `userid` bigint NOT NULL DEFAULT '0' COMMENT '用户id',
                                      `role` tinyint NOT NULL DEFAULT '2' COMMENT '1-管理员（实际通过 tbl_product 表 manager 字段决定） 2-开发者 3-运营者',
                                      `status` tinyint NOT NULL DEFAULT '3' COMMENT '1-待审批 2-续期审批 3-角色变更审批 4-已加入 5-审批拒绝  6-已退出',
                                      `join_time` int DEFAULT '0' COMMENT '申请/加入时间',
                                      `expire_type` tinyint NOT NULL DEFAULT '0' COMMENT '0: 永久 1: 一个月 2: 三个月 3: 半年 4: 一年',
                                      `expire_time` int DEFAULT '0' COMMENT '过期时间',
                                      `out_time` int DEFAULT '0' COMMENT '退出时间',
                                      `change_role` tinyint NOT NULL DEFAULT '0' COMMENT '变更为目标角色 0-无 2-开发者 3-运营者',
                                      `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                      `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                      PRIMARY KEY (`id`),
                                      UNIQUE KEY `product_user` (`product_id`,`userid`),
                                      KEY `userid` (`userid`)
) ENGINE=InnoDB AUTO_INCREMENT=46 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='产品成员'

CREATE TABLE `tbl_search_keyword` (
                                      `id` int NOT NULL AUTO_INCREMENT,
                                      `type` tinyint NOT NULL DEFAULT '1' COMMENT '1-product 2-db 3-table',
                                      `sid` int NOT NULL DEFAULT '0' COMMENT '检索id',
                                      `sname` varchar(256) NOT NULL DEFAULT '' COMMENT '检索名',
                                      `field` varchar(64) NOT NULL DEFAULT '' COMMENT '字段',
                                      `skey` varchar(64) NOT NULL DEFAULT '0' COMMENT '检索key',
                                      `scontent` varchar(512) NOT NULL DEFAULT '' COMMENT '检索内容',
                                      `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                      `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                      PRIMARY KEY (`id`),
                                      UNIQUE KEY `ukey` (`type`,`sid`,`field`,`skey`),
                                      KEY `scontent` (`scontent`),
                                      KEY `skey` (`skey`)
) ENGINE=InnoDB AUTO_INCREMENT=84 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='关键字检索'

CREATE TABLE `tbl_sequence` (
                                `id` bigint NOT NULL AUTO_INCREMENT,
                                `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=13832 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='自增序列'

CREATE TABLE `tbl_table` (
                             `id` int NOT NULL AUTO_INCREMENT,
                             `name` varchar(128) NOT NULL DEFAULT '' COMMENT '名称',
                             `intro` varchar(64) NOT NULL DEFAULT '' COMMENT '简介',
                             `desc` varchar(512) NOT NULL DEFAULT '' COMMENT '详细描述',
                             `table_verify` varchar(256) NOT NULL DEFAULT '' COMMENT '表校验，为空时不校验，默认同 name，即只允许访问 name 表/索引',
                             `db` int NOT NULL DEFAULT '0' COMMENT '所属数据库',
                             `create` varchar(4098) NOT NULL DEFAULT '' COMMENT '建表语句',
                             `table_fields` varchar(4098) NOT NULL DEFAULT '' COMMENT '表字段',
                             `table_indexs` varchar(2048) NOT NULL DEFAULT '' COMMENT '表索引',
                             `status` tinyint NOT NULL DEFAULT '1' COMMENT '1-正常 2-下线',
                             `creator` bigint NOT NULL DEFAULT '0' COMMENT '创建者',
                             `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                             `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `name` (`name`,`db`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='表配置'

CREATE TABLE `tbl_table_filter` (
                                    `id` int NOT NULL AUTO_INCREMENT,
                                    `table_id` int NOT NULL COMMENT '表id',
                                    `filter_id` int NOT NULL COMMENT '插件id',
                                    `filter_version` int NOT NULL DEFAULT '0' COMMENT 'filter版本',
                                    `seq` int NOT NULL DEFAULT '1' COMMENT '插件执行顺序',
                                    `schedule_config` longtext COMMENT '插件调度配置，是一个json，内容是 map[string]interface{}',
                                    `config` longtext COMMENT '插件配置，是一个json，内容是 map[string]interface{}',
                                    `desc` varchar(512) NOT NULL DEFAULT '' COMMENT '描述',
                                    `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态 1-启用 2-停用',
                                    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                    PRIMARY KEY (`id`),
                                    KEY `table_id` (`table_id`)
) ENGINE=InnoDB AUTO_INCREMENT=28 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='表插件配置'

CREATE TABLE `tbl_user` (
                            `id` bigint NOT NULL COMMENT '用户id',
                            `account` varchar(128) NOT NULL DEFAULT '' COMMENT '账号，可以是 admin，邮箱等。。。',
                            `nickname` varchar(128) NOT NULL DEFAULT '' COMMENT '昵称',
                            `password` varchar(64) NOT NULL DEFAULT '' COMMENT '密码',
                            `mobile` varchar(64) NOT NULL DEFAULT '' COMMENT '手机号',
                            `token` varchar(64) NOT NULL DEFAULT '' COMMENT 'token',
                            `avatar_url` varchar(512) NOT NULL DEFAULT '' COMMENT '头像',
                            `gender` tinyint NOT NULL DEFAULT '1' COMMENT '性别 1-男  2-女',
                            `company` varchar(128) NOT NULL DEFAULT '' COMMENT '公司',
                            `department` varchar(128) NOT NULL DEFAULT '' COMMENT '部门',
                            `city` varchar(128) NOT NULL DEFAULT '' COMMENT '城市',
                            `province` varchar(128) NOT NULL DEFAULT '' COMMENT '省份',
                            `country` varchar(128) NOT NULL DEFAULT '' COMMENT '国家',
                            `last_login_time` int NOT NULL DEFAULT '0' COMMENT '上次登录时间',
                            `last_login_ip` varchar(32) NOT NULL DEFAULT '' COMMENT '上次登录ip',
                            `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                            `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                            PRIMARY KEY (`id`),
                            UNIQUE KEY `account` (`account`,`mobile`),
                            KEY `mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='账号信息'

CREATE TABLE `tbl_workspace` (
                                 `id` int NOT NULL AUTO_INCREMENT,
                                 `workspace` varchar(64) NOT NULL DEFAULT '' COMMENT 'workspace',
                                 `name` varchar(128) NOT NULL DEFAULT '' COMMENT '名称',
                                 `intro` varchar(256) NOT NULL DEFAULT '' COMMENT '简介',
                                 `company` varchar(128) DEFAULT '' COMMENT '公司',
                                 `department` varchar(256) DEFAULT '' COMMENT '部门',
                                 `token` varchar(64) NOT NULL DEFAULT '' COMMENT 'token',
                                 `enforce_sign` tinyint NOT NULL DEFAULT '0' COMMENT '是否强制签名 0-否 1-是（请求数据必须得签名或者加密）',
                                 `creator` bigint NOT NULL DEFAULT '0' COMMENT 'creator',
                                 `manager` varchar(1025) NOT NULL DEFAULT '' COMMENT '管理员，多个逗号分隔',
                                 `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                 `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `workspace` (`workspace`)
) ENGINE=InnoDB AUTO_INCREMENT=32 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='workspace 信息'

CREATE TABLE `tbl_workspace_member` (
                                        `id` int NOT NULL AUTO_INCREMENT,
                                        `workspace_id` int NOT NULL DEFAULT '0' COMMENT 'workspace id',
                                        `userid` bigint NOT NULL DEFAULT '0' COMMENT '用户id',
                                        `status` tinyint NOT NULL DEFAULT '3' COMMENT '1-待审批 2-续期审批 3-暂未申请 4-已加入 5-审批拒绝  6-已退出',
                                        `join_time` int DEFAULT '0' COMMENT '申请/加入时间',
                                        `expire_type` tinyint NOT NULL DEFAULT '0' COMMENT '0: 永久 1: 一个月 2: 三个月 3: 半年 4: 一年',
                                        `expire_time` int DEFAULT '0' COMMENT '过期时间',
                                        `out_time` int DEFAULT '0' COMMENT '退出时间',
                                        `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
                                        `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录最后修改时间',
                                        PRIMARY KEY (`id`),
                                        UNIQUE KEY `userid_workspace` (`userid`,`workspace_id`)
) ENGINE=InnoDB AUTO_INCREMENT=43 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='workspace 用户信息'

