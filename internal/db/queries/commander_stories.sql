-- Commander stories queries

-- name: ListCommanderStories :many
SELECT story_id
FROM commander_stories
WHERE commander_id = $1
ORDER BY story_id ASC;

-- name: CreateCommanderStory :exec
INSERT INTO commander_stories (commander_id, story_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
