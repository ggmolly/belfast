-- Commander buff queries

-- name: ListCommanderBuffs :many
SELECT commander_id, buff_id, expires_at
FROM commander_buffs
WHERE commander_id = $1
ORDER BY buff_id ASC;

-- name: ListCommanderActiveBuffs :many
SELECT commander_id, buff_id, expires_at
FROM commander_buffs
WHERE commander_id = $1
  AND expires_at > $2;

-- name: UpsertCommanderBuff :exec
INSERT INTO commander_buffs (
  commander_id,
  buff_id,
  expires_at
) VALUES (
  $1, $2, $3
)
ON CONFLICT (commander_id, buff_id)
DO UPDATE SET expires_at = EXCLUDED.expires_at;
