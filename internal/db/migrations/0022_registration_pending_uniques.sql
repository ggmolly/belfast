-- 0022_registration_pending_uniques.sql

WITH ranked_by_commander AS (
  SELECT ctid,
         row_number() OVER (PARTITION BY commander_id ORDER BY created_at DESC, id DESC) AS rn
  FROM user_registration_challenges
  WHERE status = 'pending'
)
DELETE FROM user_registration_challenges target
USING ranked_by_commander ranked
WHERE target.ctid = ranked.ctid
  AND ranked.rn > 1;

WITH ranked_by_pin AS (
  SELECT ctid,
         row_number() OVER (PARTITION BY pin ORDER BY created_at DESC, id DESC) AS rn
  FROM user_registration_challenges
  WHERE status = 'pending'
)
DELETE FROM user_registration_challenges target
USING ranked_by_pin ranked
WHERE target.ctid = ranked.ctid
  AND ranked.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_registration_challenges_pending_commander_unique
  ON user_registration_challenges (commander_id)
  WHERE status = 'pending';

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_registration_challenges_pending_pin_unique
  ON user_registration_challenges (pin)
  WHERE status = 'pending';
