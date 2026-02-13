-- Owned ship shadow skin queries

-- name: UpsertOwnedShipShadowSkin :exec
INSERT INTO owned_ship_shadow_skins (
  commander_id,
  ship_id,
  shadow_id,
  skin_id
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (commander_id, ship_id, shadow_id)
DO UPDATE SET skin_id = EXCLUDED.skin_id;

-- name: ListOwnedShipShadowSkinsByCommander :many
SELECT commander_id, ship_id, shadow_id, skin_id
FROM owned_ship_shadow_skins
WHERE commander_id = $1
ORDER BY ship_id ASC, shadow_id ASC;

-- name: ListOwnedShipShadowSkinsByCommanderAndShips :many
SELECT commander_id, ship_id, shadow_id, skin_id
FROM owned_ship_shadow_skins
WHERE commander_id = $1
  AND ship_id = ANY($2::bigint[])
ORDER BY ship_id ASC, shadow_id ASC;
