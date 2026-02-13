-- Juustagram state queries

-- name: CountJuustagramGroupsByCommander :one
SELECT COUNT(*)::bigint
FROM juustagram_groups
WHERE commander_id = $1;

-- name: ListJuustagramGroupsByCommanderPaged :many
SELECT id, commander_id, group_id, skin_id, favorite, cur_chat_group
FROM juustagram_groups
WHERE commander_id = $1
ORDER BY group_id ASC
OFFSET $2
LIMIT $3;

-- name: GetJuustagramGroupByCommanderAndGroupID :one
SELECT id, commander_id, group_id, skin_id, favorite, cur_chat_group
FROM juustagram_groups
WHERE commander_id = $1
  AND group_id = $2;

-- name: UpdateJuustagramGroup :execresult
UPDATE juustagram_groups
SET
  skin_id = COALESCE(sqlc.narg('skin_id'), skin_id),
  favorite = COALESCE(sqlc.narg('favorite'), favorite),
  cur_chat_group = COALESCE(sqlc.narg('cur_chat_group'), cur_chat_group)
WHERE commander_id = sqlc.arg('commander_id')
  AND group_id = sqlc.arg('group_id');

-- name: ListJuustagramGroups :many
SELECT id, commander_id, group_id, skin_id, favorite, cur_chat_group
FROM juustagram_groups
WHERE commander_id = $1
ORDER BY group_id ASC;

-- name: ListJuustagramChatGroupsByGroupRecordIDs :many
SELECT id, commander_id, group_record_id, chat_group_id, op_time, read_flag
FROM juustagram_chat_groups
WHERE commander_id = $1
  AND group_record_id = ANY($2::bigint[])
ORDER BY chat_group_id ASC;

-- name: ListJuustagramRepliesByChatGroupRecordIDs :many
SELECT id, chat_group_record_id, sequence, key, value
FROM juustagram_replies
WHERE chat_group_record_id = ANY($1::bigint[])
ORDER BY chat_group_record_id ASC, sequence ASC;

-- name: CreateJuustagramGroup :one
INSERT INTO juustagram_groups (
  commander_id,
  group_id,
  skin_id,
  favorite,
  cur_chat_group
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id;

-- name: CreateJuustagramChatGroup :one
INSERT INTO juustagram_chat_groups (
  commander_id,
  group_record_id,
  chat_group_id,
  op_time,
  read_flag
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING id;

-- name: GetJuustagramChatGroupByCommanderAndChatGroupID :one
SELECT id, commander_id, group_record_id, chat_group_id, op_time, read_flag
FROM juustagram_chat_groups
WHERE commander_id = $1
  AND chat_group_id = $2;

-- name: UpdateJuustagramChatGroupOpTimeReadFlag :exec
UPDATE juustagram_chat_groups
SET op_time = $3,
    read_flag = $4
WHERE id = $1
  AND commander_id = $2;

-- name: GetMaxJuustagramReplySequence :one
SELECT COALESCE(MAX(sequence), 0)::bigint
FROM juustagram_replies
WHERE chat_group_record_id = $1;

-- name: CreateJuustagramReply :one
INSERT INTO juustagram_replies (
  chat_group_record_id,
  sequence,
  key,
  value
) VALUES (
  $1, $2, $3, $4
)
RETURNING id;

-- name: UpdateJuustagramGroupCurrentChatGroupByID :exec
UPDATE juustagram_groups
SET cur_chat_group = $3
WHERE id = $1
  AND commander_id = $2;

-- name: MarkAllJuustagramChatGroupsRead :exec
UPDATE juustagram_chat_groups
SET read_flag = 1
WHERE commander_id = $1;

-- name: MarkJuustagramChatGroupsRead :exec
UPDATE juustagram_chat_groups
SET read_flag = 1
WHERE commander_id = $1
  AND chat_group_id = ANY($2::bigint[]);
