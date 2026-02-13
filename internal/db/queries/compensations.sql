-- Compensation queries

-- name: CreateCompensation :one
INSERT INTO compensations (
  commander_id,
  title,
  text,
  send_time,
  expires_at,
  attach_flag
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, send_time, created_at;

-- name: UpdateCompensation :exec
UPDATE compensations
SET title = $2,
    text = $3,
    send_time = $4,
    expires_at = $5,
    attach_flag = $6
WHERE id = $1;

-- name: DeleteCompensation :exec
DELETE FROM compensations
WHERE id = $1;

-- name: GetCompensation :one
SELECT id,
       commander_id,
       title,
       text,
       send_time,
       expires_at,
       attach_flag,
       created_at
FROM compensations
WHERE id = $1;

-- name: ListCompensationsByCommander :many
SELECT id,
       commander_id,
       title,
       text,
       send_time,
       expires_at,
       attach_flag,
       created_at
FROM compensations
WHERE commander_id = $1
ORDER BY id ASC;

-- name: CreateCompensationAttachment :exec
INSERT INTO compensation_attachments (
  compensation_id,
  type,
  item_id,
  quantity
) VALUES (
  $1, $2, $3, $4
);

-- name: DeleteCompensationAttachmentsByCompensationID :exec
DELETE FROM compensation_attachments
WHERE compensation_id = $1;

-- name: ListCompensationAttachmentsByCompensationIDs :many
SELECT id,
       compensation_id,
       type,
       item_id,
       quantity
FROM compensation_attachments
WHERE compensation_id = ANY($1::bigint[])
ORDER BY id ASC;
