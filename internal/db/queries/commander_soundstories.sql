-- Commander sound story queries

-- name: ListCommanderSoundStories :many
SELECT story_id
FROM commander_sound_stories
WHERE commander_id = $1
ORDER BY story_id ASC;

-- name: HasCommanderSoundStory :one
SELECT EXISTS(
  SELECT 1
  FROM commander_sound_stories
  WHERE commander_id = $1
    AND story_id = $2
)::bool;

-- name: CreateCommanderSoundStory :exec
INSERT INTO commander_sound_stories (commander_id, story_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
