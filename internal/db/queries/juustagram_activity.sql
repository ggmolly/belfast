-- Juustagram activity queries (templates + per-player state)

-- name: GetJuustagramTemplate :one
SELECT id, group_id, ship_group, name, sculpture, picture_persist, message_persist, is_active, npc_discuss_persist, time, time_persist
FROM juustagram_templates
WHERE id = $1;

-- name: CountJuustagramTemplates :one
SELECT COUNT(*)::bigint FROM juustagram_templates;

-- name: ListJuustagramTemplates :many
SELECT id, group_id, ship_group, name, sculpture, picture_persist, message_persist, is_active, npc_discuss_persist, time, time_persist
FROM juustagram_templates
ORDER BY id ASC
OFFSET $1
LIMIT $2;

-- name: GetJuustagramNpcTemplate :one
SELECT id, ship_group, message_persist, npc_reply_persist, time_persist
FROM juustagram_npc_templates
WHERE id = $1;

-- name: CountJuustagramNpcTemplates :one
SELECT COUNT(*)::bigint FROM juustagram_npc_templates;

-- name: ListJuustagramNpcTemplates :many
SELECT id, ship_group, message_persist, npc_reply_persist, time_persist
FROM juustagram_npc_templates
ORDER BY id ASC
OFFSET $1
LIMIT $2;

-- name: GetJuustagramShipGroupTemplate :one
SELECT ship_group, name, background, sculpture, sculpture_ii, nationality, type
FROM juustagram_ship_group_templates
WHERE ship_group = $1;

-- name: CountJuustagramShipGroupTemplates :one
SELECT COUNT(*)::bigint FROM juustagram_ship_group_templates;

-- name: ListJuustagramShipGroupTemplates :many
SELECT ship_group, name, background, sculpture, sculpture_ii, nationality, type
FROM juustagram_ship_group_templates
ORDER BY ship_group ASC
OFFSET $1
LIMIT $2;

-- name: GetJuustagramLanguage :one
SELECT key, value
FROM juustagram_languages
WHERE key = $1;

-- name: ListJuustagramLanguageByPrefix :many
SELECT key, value
FROM juustagram_languages
WHERE key LIKE $1
ORDER BY key ASC;

-- name: ListJuustagramOpReplies :many
SELECT id, ship_group, message_persist, npc_reply_persist, time_persist
FROM juustagram_npc_templates
WHERE message_persist LIKE $1
ORDER BY id ASC;

-- name: GetJuustagramMessageState :one
SELECT id, commander_id, message_id, is_read, is_good, good_count, updated_at
FROM juustagram_message_states
WHERE commander_id = $1
  AND message_id = $2;

-- name: CreateJuustagramMessageState :one
INSERT INTO juustagram_message_states (
  commander_id,
  message_id,
  is_read,
  is_good,
  good_count,
  updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id;

-- name: UpdateJuustagramMessageState :exec
UPDATE juustagram_message_states
SET is_read = $3,
    is_good = $4,
    good_count = $5,
    updated_at = $6
WHERE commander_id = $1
  AND message_id = $2;

-- name: GetJuustagramPlayerDiscuss :one
SELECT id, commander_id, message_id, discuss_id, option_index, npc_reply_id, comment_time
FROM juustagram_player_discusses
WHERE commander_id = $1
  AND message_id = $2
  AND discuss_id = $3;

-- name: ListJuustagramPlayerDiscuss :many
SELECT id, commander_id, message_id, discuss_id, option_index, npc_reply_id, comment_time
FROM juustagram_player_discusses
WHERE commander_id = $1
  AND message_id = $2
ORDER BY discuss_id ASC;

-- name: UpsertJuustagramPlayerDiscuss :exec
INSERT INTO juustagram_player_discusses (
  commander_id,
  message_id,
  discuss_id,
  option_index,
  npc_reply_id,
  comment_time
) VALUES (
  $1, $2, $3, $4, $5, $6
)
ON CONFLICT (commander_id, message_id, discuss_id)
DO UPDATE SET
  option_index = EXCLUDED.option_index,
  npc_reply_id = EXCLUDED.npc_reply_id,
  comment_time = EXCLUDED.comment_time;
