-- 0011_month_shop_purchases.sql

CREATE TABLE IF NOT EXISTS month_shop_purchases (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  goods_id bigint NOT NULL,
  month bigint NOT NULL,
  buy_count bigint NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (commander_id, goods_id, month)
);
