-- Commander medal display queries

-- name: ListCommanderMedalDisplay :many
SELECT position, medal_id
FROM commander_medal_displays
WHERE commander_id = $1
ORDER BY position ASC;

-- name: DeleteCommanderMedalDisplayByCommanderID :exec
DELETE FROM commander_medal_displays
WHERE commander_id = $1;

-- name: CreateCommanderMedalDisplayRow :exec
INSERT INTO commander_medal_displays (commander_id, position, medal_id)
VALUES ($1, $2, $3);
