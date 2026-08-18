-- 交易游戏建表脚本
-- 注意：本脚本会先删除已存在的表再重建，仅应在 init 命令（必要时 --force）中执行

DROP TABLE IF EXISTS daily_prices;
DROP TABLE IF EXISTS warehouse_items;
DROP TABLE IF EXISTS warehouse_purchases;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS crit_events;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS reg_keys;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS settings;

-- 用户表：剩余金钱、仓库大小对应的天数等元数据
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL COMMENT '用户名',
  password CHAR(32) NOT NULL COMMENT '密码 md5',
  money DOUBLE NOT NULL DEFAULT 0 COMMENT '剩余金钱',
  `day` INT NOT NULL DEFAULT 1 COMMENT '当前天数',
  is_admin TINYINT NOT NULL DEFAULT 0 COMMENT '是否管理员',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 启用 0 禁用（禁用后无法登录）',
  key_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '注册时使用的密钥 ID',
  created_at DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 注册密钥表
CREATE TABLE reg_keys (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `key` VARCHAR(10) NOT NULL COMMENT '10 位随机密钥',
  max_uses INT NOT NULL DEFAULT 1 COMMENT '最大使用次数',
  used_count INT NOT NULL DEFAULT 0 COMMENT '已使用次数',
  created_by VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建者（管理员）',
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1 启用 0 禁用',
  created_at DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_reg_keys_key (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='注册密钥表';

-- 商品元数据表
CREATE TABLE products (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL COMMENT '商品名',
  category VARCHAR(32) NOT NULL DEFAULT '' COMMENT '商品种类（食物/材料/科技等）',
  storage_type TINYINT NOT NULL DEFAULT 1 COMMENT '1 普通仓库 2 冷藏仓库 3 两者皆可',
  min_price DOUBLE NOT NULL DEFAULT 0 COMMENT '最低价',
  max_price DOUBLE NOT NULL DEFAULT 0 COMMENT '最高价',
  size BIGINT NOT NULL DEFAULT 1 COMMENT '单位商品占用仓库大小',
  normal_expire_days INT NOT NULL DEFAULT -1 COMMENT '普通仓库过期天数，-1 永久',
  cold_expire_days INT NOT NULL DEFAULT -1 COMMENT '冷藏仓库过期天数，-1 永久',
  crit_min DOUBLE NOT NULL DEFAULT 0 COMMENT '暴击涨幅最小值(%)',
  crit_max DOUBLE NOT NULL DEFAULT 0 COMMENT '暴击涨幅最大值(%)',
  PRIMARY KEY (id),
  UNIQUE KEY uk_products_name (name),
  KEY idx_products_category (category)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品元数据表';

-- 暴击事件元数据表
CREATE TABLE crit_events (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  description VARCHAR(256) NOT NULL DEFAULT '' COMMENT '事件说明',
  target_type VARCHAR(16) NOT NULL DEFAULT 'category' COMMENT 'category 针对种类 / product 针对单一商品',
  target_value VARCHAR(64) NOT NULL DEFAULT '' COMMENT '种类名或商品名',
  probability INT NOT NULL DEFAULT 0 COMMENT '发生概率(%)',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='暴击事件元数据表';

-- 每日商品价格表（保证同一天价格一致性）
CREATE TABLE daily_prices (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  `day` INT NOT NULL DEFAULT 1 COMMENT '天数',
  product_id BIGINT UNSIGNED NOT NULL,
  price DOUBLE NOT NULL DEFAULT 0 COMMENT '当日随机价格',
  crit_applied TINYINT NOT NULL DEFAULT 0 COMMENT '当日价格是否由暴击事件加成: 1 是 0 否',
  PRIMARY KEY (id),
  UNIQUE KEY uk_daily_prices (user_id, `day`, product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='每日商品价格表';

-- 仓库存储表
CREATE TABLE warehouse_items (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  product_id BIGINT UNSIGNED NOT NULL,
  storage_type TINYINT NOT NULL DEFAULT 1 COMMENT '1 普通仓库 2 冷藏仓库',
  quantity INT NOT NULL DEFAULT 0 COMMENT '数量',
  total_cost DOUBLE NOT NULL DEFAULT 0 COMMENT '总价值',
  avg_price DOUBLE NOT NULL DEFAULT 0 COMMENT '均价',
  buy_day INT NOT NULL DEFAULT 1 COMMENT '首次买入天数',
  expire_day INT NOT NULL DEFAULT -1 COMMENT '过期天数，-1 永久',
  updated_at DATETIME NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_warehouse_items (user_id, product_id, storage_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仓库存储表';

-- 交易记录表
CREATE TABLE transactions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  `day` INT NOT NULL DEFAULT 1 COMMENT '发生天数',
  type VARCHAR(16) NOT NULL DEFAULT 'buy' COMMENT 'buy/sell/expire/warehouse',
  direction TINYINT NOT NULL DEFAULT 1 COMMENT '1 支出 2 收入',
  product_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  product_name VARCHAR(64) NOT NULL DEFAULT '',
  quantity INT NOT NULL DEFAULT 0,
  unit_price DOUBLE NOT NULL DEFAULT 0,
  amount DOUBLE NOT NULL DEFAULT 0 COMMENT '金额',
  balance_after DOUBLE NOT NULL DEFAULT 0 COMMENT '交易后余额',
  created_at DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_transactions_user (user_id, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='交易记录表';

-- 特殊商城仓库空间购买记录表
CREATE TABLE warehouse_purchases (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  storage_type TINYINT NOT NULL DEFAULT 1 COMMENT '1 普通仓库 2 冷藏仓库',
  size BIGINT NOT NULL DEFAULT 0 COMMENT '购买的空间大小',
  months INT NOT NULL DEFAULT 1 COMMENT '购买月数（1-12，每月按 30 天）',
  unit_price DOUBLE NOT NULL DEFAULT 0 COMMENT '每单位空间每月价格',
  start_day INT NOT NULL DEFAULT 1 COMMENT '开始天数',
  expire_day INT NOT NULL DEFAULT 1 COMMENT '到期天数',
  created_at DATETIME NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_warehouse_purchases_user (user_id, expire_day)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仓库空间购买记录表';

-- 系统配置表
CREATE TABLE settings (
  k VARCHAR(64) NOT NULL,
  v VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (k)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';
