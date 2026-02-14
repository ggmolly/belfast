-- Owned ship mutation queries

-- name: CreateOwnedShip :one
INSERT INTO owned_ships (
  owner_id,
  ship_id
) VALUES (
  $1, $2
)
RETURNING id, create_time, change_name_timestamp;

-- name: UpsertOwnedShipEquipmentDefault :exec
INSERT INTO owned_ship_equipments (
  owner_id,
  ship_id,
  pos,
  equip_id,
  skin_id
) VALUES (
  $1, $2, $3, $4, $5
)
ON CONFLICT (owner_id, ship_id, pos)
DO NOTHING;

-- name: DeleteOwnedShipsByOwnerAndIDs :exec
DELETE FROM owned_ships
WHERE owner_id = $1
  AND id = ANY($2::bigint[]);

-- name: ClearOwnedShipSecretariesByOwnerID :exec
UPDATE owned_ships
SET is_secretary = false,
    secretary_position = NULL,
    secretary_phantom_id = 0
WHERE owner_id = $1
  AND deleted_at IS NULL
  AND is_secretary = true;

-- name: SetOwnedShipSecretary :exec
UPDATE owned_ships
SET is_secretary = true,
    secretary_position = $3,
    secretary_phantom_id = $4
WHERE owner_id = $1
  AND id = $2
  AND deleted_at IS NULL;

-- name: UpdateOwnedShip :execresult
UPDATE owned_ships
SET
  level = $3,
  exp = $4,
  surplus_exp = $5,
  max_level = $6,
  intimacy = $7,
  is_locked = $8,
  propose = $9,
  common_flag = $10,
  blueprint_flag = $11,
  proficiency = $12,
  activity_npc = $13,
  custom_name = $14,
  change_name_timestamp = $15,
  create_time = $16,
  energy = $17,
  state = $18,
  state_info1 = $19,
  state_info2 = $20,
  state_info3 = $21,
  state_info4 = $22,
  skin_id = $23,
  is_secretary = $24,
  secretary_position = $25,
  secretary_phantom_id = $26,
  deleted_at = $27
WHERE owner_id = $1
  AND id = $2;

-- name: SoftDeleteOwnedShip :execresult
UPDATE owned_ships
SET deleted_at = CURRENT_TIMESTAMP
WHERE owner_id = $1
  AND id = $2
  AND deleted_at IS NULL;
