-- Mail mutation queries

-- name: DeleteMailsByReceiverID :exec
DELETE FROM mails
WHERE receiver_id = $1;

-- name: CreateMail :one
INSERT INTO mails (
  receiver_id,
  read,
  date,
  title,
  body,
  attachments_collected,
  is_important,
  custom_sender,
  is_archived,
  created_at
) VALUES (
  $1, $2, now(), $3, $4, $5, $6, $7, $8, now()
)
RETURNING id, date, created_at;

-- name: CreateMailAttachment :one
INSERT INTO mail_attachments (
  mail_id,
  type,
  item_id,
  quantity
) VALUES (
  $1, $2, $3, $4
)
RETURNING id;

-- name: UpdateMail :execresult
UPDATE mails
SET
  read = $3,
  title = $4,
  body = $5,
  attachments_collected = $6,
  is_important = $7,
  custom_sender = $8,
  is_archived = $9
WHERE receiver_id = $1
  AND id = $2;

-- name: DeleteMail :execresult
DELETE FROM mails
WHERE receiver_id = $1
  AND id = $2;
