-- Commander surveys queries

-- name: HasCommanderSurvey :one
SELECT EXISTS(
  SELECT 1
  FROM commander_surveys
  WHERE commander_id = $1
    AND survey_id = $2
)::bool;

-- name: UpsertCommanderSurvey :exec
INSERT INTO commander_surveys (commander_id, survey_id, completed_at)
VALUES ($1, $2, $3)
ON CONFLICT (commander_id, survey_id)
DO UPDATE SET completed_at = EXCLUDED.completed_at;
