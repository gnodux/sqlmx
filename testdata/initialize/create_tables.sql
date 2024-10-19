
CREATE TABLE IF NOT EXISTS `tenant`
(
    `id`   BIGINT PRIMARY KEY AUTO_INCREMENT NOT NULL,
    `name` VARCHAR(255)                      NOT NULL
);
CREATE TABLE IF NOT EXISTS `user`
(
    `id`        bigint primary key auto_increment NOT NULL,
    `tenant_id` BIGINT                            NOT NULL,
    `name`      varchar(128)                      NOT NULL default '',
    `password`  varchar(32)                       NOT NULL,
    `birthday`  datetime,
    `address`   varchar(255),
    `role`      varchar(128),
    `is_deleted` BOOLEAN DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS `role`
(
    `id`   bigint primary key auto_increment NOT NULL,
    `name` varchar(128)                      NOT NULL,
    `desc` varchar(255)                      not NULL,
    `is_deleted` BOOLEAN DEFAULT FALSE
);
CREATE TABLE IF NOT EXISTS `account_book`
(
    `id`        bigint primary key auto_increment NOT NULL,
    `tenant_id` bigint                            NOT NULL COMMENT '租户ID',
    `create_by` bigint                            NOT NULL COMMENT '创建人',
    `owner`     BIGINT                            NOT NULL COMMENT '账本所有人',
    `name`      varchar(128)                      NOT NULL COMMENT '账本名称',
    `balance`   decimal(10, 2)                    NOT NULL DEFAULT 0 COMMENT '账户余额',
    `desc`      varchar(255) COMMENT '账本描述',
    `is_deleted` BOOLEAN DEFAULT FALSE
) COMMENT '账本表';
CREATE TABLE IF NOT EXISTS `transaction`
(
    `id`              BIGINT PRIMARY KEY AUTO_INCREMENT    NOT NULL,
    `tenant_id`       BIGINT                               NOT NULL COMMENT '租户ID',
    `account_book_id` BIGINT                               NOT NULL COMMENT '账本ID',
    `create_by`       BIGINT                               NOT NULL COMMENT '创建人',
    `create_time`     BIGINT COMMENT ' 创建时间 ',
    `amount`          decimal(10, 2)                       NOT NULL COMMENT ' 交易金额 ',
    `type`            enum ('Income','Expense')            NOT NULL COMMENT ' 交易类型 ',
    `desc`            varchar(255) COMMENT ' 交易描述 ',
    `status`          enum (' Draft ',' Done ',' Cancel ') NOT NULL COMMENT ' 交易状态 ',
    `is_deleted` BOOLEAN DEFAULT FALSE
) COMMENT ' 交易表 ';
