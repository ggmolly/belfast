-- 0021_fleet_uniqueness.sql

-- +migrate NoTransaction

DELETE FROM fleets older
USING fleets newer
WHERE older.commander_id = newer.commander_id
  AND older.game_id = newer.game_id
  AND older.id < newer.id;

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_fleets_commander_id_game_id_unique
  ON fleets (commander_id, game_id);
