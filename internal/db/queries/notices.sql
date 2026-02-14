-- Notice queries

-- name: UpsertNotice :exec
INSERT INTO notices (
  id,
  version,
  btn_title,
  title,
  title_image,
  time_desc,
  content,
  tag_type,
  icon,
  track
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (id)
DO UPDATE SET
  version = EXCLUDED.version,
  btn_title = EXCLUDED.btn_title,
  title = EXCLUDED.title,
  title_image = EXCLUDED.title_image,
  time_desc = EXCLUDED.time_desc,
  content = EXCLUDED.content,
  tag_type = EXCLUDED.tag_type,
  icon = EXCLUDED.icon,
  track = EXCLUDED.track;

-- name: GetNotice :one
SELECT id, version, btn_title, title, title_image, time_desc, content, tag_type, icon, track
FROM notices
WHERE id = $1;

-- name: DeleteNotice :execresult
DELETE FROM notices
WHERE id = $1;
