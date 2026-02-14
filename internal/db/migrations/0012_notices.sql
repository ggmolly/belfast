-- 0012_notices.sql
-- Game notices.

CREATE TABLE IF NOT EXISTS notices (
  id bigint PRIMARY KEY,
  version text NOT NULL DEFAULT '1',
  btn_title text NOT NULL,
  title text NOT NULL,
  title_image text NOT NULL,
  time_desc text NOT NULL,
  content text NOT NULL,
  tag_type bigint NOT NULL DEFAULT 1,
  icon bigint NOT NULL DEFAULT 1,
  track text NOT NULL
);
