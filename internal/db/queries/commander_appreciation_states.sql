-- Commander appreciation state queries

-- name: GetCommanderAppreciationStateByCommanderID :one
SELECT
  commander_id,
  music_no,
  music_mode,
  cartoon_read_mark,
  cartoon_collect_mark,
  gallery_unlocks,
  gallery_favor_ids,
  music_favor_ids
FROM commander_appreciation_states
WHERE commander_id = $1;

-- name: CreateCommanderAppreciationState :exec
INSERT INTO commander_appreciation_states (
  commander_id,
  music_no,
  music_mode,
  cartoon_read_mark,
  cartoon_collect_mark,
  gallery_unlocks,
  gallery_favor_ids,
  music_favor_ids
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: UpdateCommanderAppreciationState :exec
UPDATE commander_appreciation_states
SET
  music_no = $2,
  music_mode = $3,
  cartoon_read_mark = $4,
  cartoon_collect_mark = $5,
  gallery_unlocks = $6,
  gallery_favor_ids = $7,
  music_favor_ids = $8
WHERE commander_id = $1;
