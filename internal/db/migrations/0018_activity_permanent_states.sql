-- 0018_activity_permanent_states.sql

CREATE TABLE IF NOT EXISTS activity_permanent_states (
  commander_id bigint PRIMARY KEY,
  current_activity_id bigint NOT NULL DEFAULT 0,
  finished_activity_ids jsonb NOT NULL DEFAULT '[]'::jsonb
);

DELETE FROM activity_permanent_states aps
WHERE NOT EXISTS (
  SELECT 1
  FROM commanders c
  WHERE c.commander_id = aps.commander_id
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_activity_permanent_states_commander'
      AND conrelid = 'activity_permanent_states'::regclass
  ) THEN
    ALTER TABLE activity_permanent_states
      ADD CONSTRAINT fk_activity_permanent_states_commander
      FOREIGN KEY (commander_id)
      REFERENCES commanders(commander_id)
      ON DELETE CASCADE;
  END IF;
END $$;
