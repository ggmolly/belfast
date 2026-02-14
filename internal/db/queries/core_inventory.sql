-- Inventory/resource mutation queries (Milestone 4)

-- name: DecrementCommanderItemIfEnough :execresult
UPDATE commander_items
SET count = count - $3
WHERE commander_id = $1
  AND item_id = $2
  AND count >= $3;

-- name: IncrementCommanderItem :exec
INSERT INTO commander_items (
  commander_id,
  item_id,
  count
) VALUES (
  $1, $2, $3
)
ON CONFLICT (commander_id, item_id)
DO UPDATE SET count = commander_items.count + EXCLUDED.count;

-- name: UpsertCommanderItemSet :exec
INSERT INTO commander_items (
  commander_id,
  item_id,
  count
) VALUES (
  $1, $2, $3
)
ON CONFLICT (commander_id, item_id)
DO UPDATE SET count = EXCLUDED.count;

-- name: DecrementOwnedResourceIfEnough :execresult
UPDATE owned_resources
SET amount = amount - $3
WHERE commander_id = $1
  AND resource_id = $2
  AND amount >= $3;

-- name: IncrementOwnedResource :exec
INSERT INTO owned_resources (
  commander_id,
  resource_id,
  amount
) VALUES (
  $1, $2, $3
)
ON CONFLICT (commander_id, resource_id)
DO UPDATE SET amount = owned_resources.amount + EXCLUDED.amount;

-- name: UpsertOwnedResourceSet :exec
INSERT INTO owned_resources (
  commander_id,
  resource_id,
  amount
) VALUES (
  $1, $2, $3
)
ON CONFLICT (commander_id, resource_id)
DO UPDATE SET amount = EXCLUDED.amount;
