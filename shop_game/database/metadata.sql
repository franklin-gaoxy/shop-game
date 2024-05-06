--
-- CREATE DATABASE IF NOT EXISTS `trading_game` CHARACTER SET utf8;
-- store : goods metadata information
DROP TABLE IF EXISTS `trading_game`.`store`;
CREATE TABLE IF NOT EXISTS `trading_game`.`store`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `name` varchar(255) NOT NULL COMMENT '商品名称',
  `start_price` bigint NOT NULL COMMENT '起始价格',
  `end_price` bigint NOT NULL COMMENT '结束价格',
  `max_number` bigint NOT NULL COMMENT '每天可以有的最大数量',
  `mini_number` bigint NOT NULL COMMENT '每天可以有的最少数量',
  `commodity_size` bigint NOT NULL COMMENT '商品占用仓库大小',
  `storage_date` bigint NOT NULL COMMENT '该商品最大存储时间 过期时间 -1为不会过期',
  `refrigerate` int NOT NULL COMMENT '该商品是否为需要冷藏商品 0为不需要',
  `max_price` BIGINT NOT NULL COMMENT '该商品触发暴击最高价格',
  `mini_price` BIGINT NOT NULL COMMENT '该商品触发暴击最低价格',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `name_index`(`name`) USING BTREE COMMENT 'name字段索引'
);
-- transaction
DROP TABLE IF EXISTS `trading_game`.`transaction_record`;
CREATE TABLE IF NOT EXISTS `trading_game`.`transaction_record`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(255) NOT NULL COMMENT '商品名称',
  `shop_price` bigint NOT NULL COMMENT '购买/卖出 时的价格',
  `number` bigint NOT NULL COMMENT '该商品购买/卖出 时的数量',
  `type` int NOT null COMMENT '商品类型 0:冷冻 1:普通',
  `shop_day` bigint NOT NULL COMMENT '商品位于那一天买入/卖出',
  `op_type` int NOT null COMMENT '操作类型 卖出/买入 买入0 卖出1',
  PRIMARY KEY (`id`),
  -- UNIQUE INDEX `name_index`(`name`) USING BTREE
  INDEX `name_index`(`name`) USING BTREE
);
-- warehouse
DROP TABLE IF EXISTS `trading_game`.`warehouse`;
CREATE TABLE IF NOT EXISTS `trading_game`.`warehouse`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(255) NOT NULL COMMENT '商品名称',
  `day` bigint NOT NULL COMMENT '那一天入库',
  `avg_price` bigint NOT NULL COMMENT '平均价格',
  `number` bigint NOT NULL COMMENT '拥有该商品的数量',
  `storage_date` bigint NOT NULL COMMENT '该商品的过期时间',
  PRIMARY KEY (`id`),
  -- UNIQUE INDEX `name_index`(`name`) USING BTREE
  INDEX `name_index`(`name`) USING BTREE
);
-- tmp day
DROP TABLE IF EXISTS `trading_game`.`tmp_day`;
CREATE TABLE IF NOT EXISTS `trading_game`.`tmp_day`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(255) NOT NULL COMMENT '商品名称',
  `price` bigint NOT NULL COMMENT '该商品当天的价格',
  `size` bigint NOT NULL COMMENT '该商品需要占用仓库的大小',
  `purchase_number` bigint NOT NULL COMMENT '该商品当日最大可购买的数量',
  `selling_number` bigint NOT NULL COMMENT '该商品当日最大可卖出的数量',
  `refrigerate` int NOT NULL COMMENT '该商品是否为冷冻商品 1普通 0冷冻',
  `critical_strike` int NOT NULL COMMENT '是否触发暴击效果 0无暴击 1暴涨 2暴跌',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `name_index`(`name`) USING BTREE
);
-- user information
DROP TABLE IF EXISTS `trading_game`.`user`;
CREATE TABLE IF NOT EXISTS `trading_game`.`user`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `username` varchar(255) NOT NULL COMMENT '用户名',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `money` bigint NOT NULL COMMENT '金币数量',
  `warehouse` bigint NOT NULL COMMENT '默认普通仓库大小',
  `refrigerate_warehouse` bigint NOT NULL COMMENT '默认冷藏仓库大小',
  `day` bigint NOT NULL COMMENT '用户处于那一天',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `name_index`(`username`) USING BTREE
);
--
DROP TABLE IF EXISTS `trading_game`.`opportunity`;
CREATE TABLE IF NOT EXISTS `trading_game`.`opportunity`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `events_name` varchar(255) NOT NULL COMMENT '事件名称',
  `critical_chance` int NOT NULL COMMENT '该事件触发暴击的几率',
  `max_time` int NOT NULL COMMENT '暴击持续最长天数',
  `mini_time` int NOT NULL COMMENT '暴击持续最短天数',
  PRIMARY KEY (`id`)
);
--
DROP TABLE IF EXISTS `trading_game`.`special_goods`;
CREATE TABLE IF NOT EXISTS `trading_game`.`special_goods`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `special_name` varchar(255) NOT NULL COMMENT '特殊商品名称',
  `special_price` int NOT NULL COMMENT '特殊商品的价格',
  `special_content` json NOT NULL COMMENT '特殊商品的描述信息等,也可以用于记录特殊内容',
  PRIMARY KEY (`id`),
  UNIQUE INDEX `name_index`(`special_name`) USING BTREE
);

-- -- 以下为用于测试的数据
-- INSERT INTO `trading_game`.`store` (`name`, `start_price`, `end_price`, `max_number` ,`mini_number` , `commodity_size`, `storage_date`, `refrigerate`, `max_price`, `mini_price`) VALUES ('苹果', 2, 6,3000, 1000, 1, 30, 0, 12, 0);
-- INSERT INTO `trading_game`.`store` (`name`, `start_price`, `end_price`, `max_number` ,`mini_number` , `commodity_size`, `storage_date`, `refrigerate`, `max_price`, `mini_price`) VALUES ('西红柿', 3, 5, 1000, 200, 1, 20, 0, 18, 0);
-- INSERT INTO `trading_game`.`store` (`name`, `start_price`, `end_price`, `max_number` ,`mini_number` , `commodity_size`, `storage_date`, `refrigerate`, `max_price`, `mini_price`) VALUES ('铁', 20, 60, 1000, 100, 10, 100, 0, 200, 5);
-- -- 插入用户
-- INSERT INTO `trading_game`.`user` (`username`, `password`, `money`, `warehouse`, `refrigerate_warehouse`, `day`) VALUES ('frank', '92d7ddd2a010c59511dc2905b7e14f64', 100000, 300, 100, 0);
-- -- 插入机遇
-- INSERT INTO `trading_game`.`opportunity` (`events_name`, `critical_chance`, `max_time`, `mini_time`) VALUES ('因为战争原因,某件商品价格', 10, 30, 5);
-- -- 插入tmp_day的随机数据
-- INSERT INTO `trading_game`.`tmp_day` (`name`, `price`, `size`, `purchase_number`, `selling_number`,`refrigerate` , `critical_strike`) VALUES ('苹果', 10, 1, 1000, 1000, 0, 0);
-- INSERT INTO `trading_game`.`tmp_day` (`name`, `price`, `size`, `purchase_number`, `selling_number`,`refrigerate` , `critical_strike`) VALUES ('西红柿', 2, 1, 800, 100, 0, 1);
-- INSERT INTO `trading_game`.`tmp_day` (`name`, `price`, `size`, `purchase_number`, `selling_number`,`refrigerate` , `critical_strike`) VALUES ('铁', 1000, 100, 10, 100, 1, 2);
-- 插入特殊商品
-- INSERT INTO `trading_game`.`special_goods` (`special_name`,`special_price`,`special_content`) values ("普通仓库")