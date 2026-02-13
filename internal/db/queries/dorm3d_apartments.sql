-- Dorm3d apartments queries

-- name: GetDorm3dApartment :one
SELECT commander_id,
       daily_vigor_max,
       gifts,
       ships,
       gift_daily,
       gift_permanent,
       furniture_daily,
       furniture_permanent,
       rooms,
       ins
FROM dorm3d_apartments
WHERE commander_id = $1;

-- name: CreateDorm3dApartment :exec
INSERT INTO dorm3d_apartments (commander_id)
VALUES ($1)
ON CONFLICT DO NOTHING;

-- name: UpsertDorm3dApartment :exec
INSERT INTO dorm3d_apartments (
  commander_id,
  daily_vigor_max,
  gifts,
  ships,
  gift_daily,
  gift_permanent,
  furniture_daily,
  furniture_permanent,
  rooms,
  ins
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (commander_id)
DO UPDATE SET
  daily_vigor_max = EXCLUDED.daily_vigor_max,
  gifts = EXCLUDED.gifts,
  ships = EXCLUDED.ships,
  gift_daily = EXCLUDED.gift_daily,
  gift_permanent = EXCLUDED.gift_permanent,
  furniture_daily = EXCLUDED.furniture_daily,
  furniture_permanent = EXCLUDED.furniture_permanent,
  rooms = EXCLUDED.rooms,
  ins = EXCLUDED.ins;
