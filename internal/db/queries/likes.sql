-- Likes queries

-- name: CreateLike :exec
INSERT INTO likes (group_id, liker_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
