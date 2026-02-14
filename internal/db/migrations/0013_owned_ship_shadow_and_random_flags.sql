-- 0013_owned_ship_shadow_and_random_flags.sql

CREATE TABLE IF NOT EXISTS owned_ship_shadow_skins (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_id bigint NOT NULL REFERENCES owned_ships(id) ON DELETE CASCADE,
  shadow_id bigint NOT NULL,
  skin_id bigint NOT NULL,
  PRIMARY KEY (commander_id, ship_id, shadow_id)
);

CREATE TABLE IF NOT EXISTS random_flag_ships (
  commander_id bigint NOT NULL REFERENCES commanders(commander_id) ON DELETE CASCADE,
  ship_id bigint NOT NULL REFERENCES owned_ships(id) ON DELETE CASCADE,
  phantom_id bigint NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  PRIMARY KEY (commander_id, ship_id, phantom_id)
);
