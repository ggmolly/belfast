-- Month shop purchase queries

-- name: ListMonthShopPurchasesByCommanderAndMonth :many
SELECT commander_id, goods_id, month, buy_count, updated_at
FROM month_shop_purchases
WHERE commander_id = $1
  AND month = $2;

-- name: GetMonthShopPurchase :one
SELECT commander_id, goods_id, month, buy_count, updated_at
FROM month_shop_purchases
WHERE commander_id = $1
  AND goods_id = $2
  AND month = $3;

-- name: IncrementMonthShopPurchase :exec
INSERT INTO month_shop_purchases (commander_id, goods_id, month, buy_count)
VALUES ($1, $2, $3, $4)
ON CONFLICT (commander_id, goods_id, month)
DO UPDATE SET
  buy_count = month_shop_purchases.buy_count + EXCLUDED.buy_count,
  updated_at = CURRENT_TIMESTAMP;
