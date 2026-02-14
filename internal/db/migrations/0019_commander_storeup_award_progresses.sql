-- 0019_commander_storeup_award_progresses.sql

CREATE TABLE IF NOT EXISTS commander_storeup_award_progresses (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  storeup_id bigint NOT NULL,
  last_award_index bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (commander_id, storeup_id)
);
