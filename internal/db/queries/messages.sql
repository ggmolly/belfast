-- Chat/messages queries

-- name: CreateMessage :one
INSERT INTO messages (
  sender_id,
  room_id,
  sent_at,
  content
) VALUES (
  $1, $2, now(), $3
)
RETURNING id, sent_at;

-- name: UpdateMessageContent :exec
UPDATE messages
SET content = $2
WHERE id = $1;

-- name: DeleteMessageByID :execresult
DELETE FROM messages
WHERE id = $1;

-- name: ListRoomHistory :many
SELECT id, sender_id, room_id, sent_at, content
FROM messages
WHERE room_id = $1
ORDER BY sent_at DESC
LIMIT 50;
