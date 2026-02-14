-- Arena shop

-- name: GetArenaShopState :one
SELECT commander_id, flash_count, last_refresh_time, next_flash_time
FROM arena_shop_states
WHERE commander_id = $1;

-- name: CreateArenaShopState :exec
INSERT INTO arena_shop_states (commander_id, flash_count, last_refresh_time, next_flash_time)
VALUES ($1, $2, $3, $4);

-- name: UpdateArenaShopState :exec
UPDATE arena_shop_states
SET flash_count = $2,
    last_refresh_time = $3,
    next_flash_time = $4
WHERE commander_id = $1;

-- Shopping street

-- name: GetShoppingStreetState :one
SELECT commander_id, level, next_flash_time, level_up_time, flash_count
FROM shopping_street_states
WHERE commander_id = $1;

-- name: CreateShoppingStreetState :exec
INSERT INTO shopping_street_states (commander_id, level, next_flash_time, level_up_time, flash_count)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdateShoppingStreetStateRefresh :exec
UPDATE shopping_street_states
SET next_flash_time = $2,
    flash_count = $3
WHERE commander_id = $1;

-- name: ListShoppingStreetGoodsByCommander :many
SELECT commander_id, goods_id, discount, buy_count
FROM shopping_street_goods
WHERE commander_id = $1
ORDER BY goods_id ASC;

-- name: DeleteShoppingStreetGoodsByCommander :exec
DELETE FROM shopping_street_goods
WHERE commander_id = $1;

-- name: CreateShoppingStreetGood :exec
INSERT INTO shopping_street_goods (commander_id, goods_id, discount, buy_count)
VALUES ($1, $2, $3, $4);

-- name: ListShopOffersByGenre :many
SELECT id, effects, effect_args, number, resource_number, resource_id, type, genre, discount
FROM shop_offers
WHERE genre = $1;

-- name: ListShopOffersByIDsAndGenre :many
SELECT id, effects, effect_args, number, resource_number, resource_id, type, genre, discount
FROM shop_offers
WHERE genre = $1
  AND id = ANY($2::bigint[]);

-- Medal shop

-- name: GetMedalShopState :one
SELECT commander_id, next_refresh_time
FROM medal_shop_states
WHERE commander_id = $1;

-- name: CreateMedalShopState :exec
INSERT INTO medal_shop_states (commander_id, next_refresh_time)
VALUES ($1, $2);

-- name: UpdateMedalShopState :exec
UPDATE medal_shop_states
SET next_refresh_time = $2
WHERE commander_id = $1;

-- name: ListMedalShopGoodsByCommander :many
SELECT commander_id, index, goods_id, count
FROM medal_shop_goods
WHERE commander_id = $1
ORDER BY index ASC;

-- name: DeleteMedalShopGoodsByCommander :exec
DELETE FROM medal_shop_goods
WHERE commander_id = $1;

-- name: CreateMedalShopGood :exec
INSERT INTO medal_shop_goods (commander_id, index, goods_id, count)
VALUES ($1, $2, $3, $4);

-- Mini-game shop

-- name: GetMiniGameShopState :one
SELECT commander_id, next_refresh_time
FROM mini_game_shop_states
WHERE commander_id = $1;

-- name: CreateMiniGameShopState :exec
INSERT INTO mini_game_shop_states (commander_id, next_refresh_time)
VALUES ($1, $2);

-- name: UpdateMiniGameShopState :exec
UPDATE mini_game_shop_states
SET next_refresh_time = $2
WHERE commander_id = $1;

-- name: ListMiniGameShopGoodsByCommander :many
SELECT commander_id, goods_id, count
FROM mini_game_shop_goods
WHERE commander_id = $1
ORDER BY goods_id ASC;

-- name: DeleteMiniGameShopGoodsByCommander :exec
DELETE FROM mini_game_shop_goods
WHERE commander_id = $1;

-- name: CreateMiniGameShopGood :exec
INSERT INTO mini_game_shop_goods (commander_id, goods_id, count)
VALUES ($1, $2, $3);

-- Guild shop

-- name: GetGuildShopState :one
SELECT commander_id, refresh_count, next_refresh_time
FROM guild_shop_states
WHERE commander_id = $1;

-- name: CreateGuildShopState :exec
INSERT INTO guild_shop_states (commander_id, refresh_count, next_refresh_time)
VALUES ($1, $2, $3);

-- name: UpdateGuildShopState :exec
UPDATE guild_shop_states
SET refresh_count = $2,
    next_refresh_time = $3
WHERE commander_id = $1;

-- name: ListGuildShopGoodsByCommander :many
SELECT commander_id, index, goods_id, count
FROM guild_shop_goods
WHERE commander_id = $1
ORDER BY index ASC;

-- name: DeleteGuildShopGoodsByCommander :exec
DELETE FROM guild_shop_goods
WHERE commander_id = $1;

-- name: CreateGuildShopGood :exec
INSERT INTO guild_shop_goods (commander_id, index, goods_id, count)
VALUES ($1, $2, $3, $4);

-- Debug packet capture

-- name: CreateDebugPacket :exec
INSERT INTO debugs (packet_size, packet_id, data)
VALUES ($1, $2, $3);
