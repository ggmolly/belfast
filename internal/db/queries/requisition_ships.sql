-- Requisition ship queries

-- name: ListRequisitionShipIDs :many
SELECT ship_id
FROM requisition_ships
ORDER BY ship_id ASC;

-- name: GetRandomRequisitionShipByRarity :one
SELECT s.template_id, s.name, s.english_name, s.rarity_id, s.star, s.type, s.nationality, s.build_time, s.pool_id
FROM ships s
JOIN requisition_ships rs ON rs.ship_id = s.template_id
WHERE s.rarity_id = $1
ORDER BY RANDOM()
LIMIT 1;
