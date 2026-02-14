-- Guild chat queries

-- name: CreateGuildChatMessage :one
INSERT INTO guild_chat_messages (
  guild_id,
  sender_id,
  sent_at,
  content
) VALUES (
  $1, $2, $3, $4
)
RETURNING id;

-- name: ListGuildChatMessagesAll :many
SELECT
  m.id,
  m.guild_id,
  m.sender_id,
  m.sent_at,
  m.content,
  c.commander_id AS sender_commander_id,
  c.name AS sender_name,
  c.level AS sender_level,
  c.display_icon_id AS sender_display_icon_id,
  c.display_skin_id AS sender_display_skin_id,
  c.selected_icon_frame_id AS sender_selected_icon_frame_id,
  c.selected_chat_frame_id AS sender_selected_chat_frame_id,
  c.display_icon_theme_id AS sender_display_icon_theme_id
FROM guild_chat_messages m
JOIN commanders c ON c.commander_id = m.sender_id
WHERE m.guild_id = $1
ORDER BY m.sent_at DESC
;

-- name: ListGuildChatMessages :many
SELECT
  m.id,
  m.guild_id,
  m.sender_id,
  m.sent_at,
  m.content,
  c.commander_id AS sender_commander_id,
  c.name AS sender_name,
  c.level AS sender_level,
  c.display_icon_id AS sender_display_icon_id,
  c.display_skin_id AS sender_display_skin_id,
  c.selected_icon_frame_id AS sender_selected_icon_frame_id,
  c.selected_chat_frame_id AS sender_selected_chat_frame_id,
  c.display_icon_theme_id AS sender_display_icon_theme_id
FROM guild_chat_messages m
JOIN commanders c ON c.commander_id = m.sender_id
WHERE m.guild_id = $1
ORDER BY m.sent_at DESC
LIMIT sqlc.arg('limit');
