import os
import json
import psycopg2
import requests
import tempfile
from dotenv import load_dotenv

load_dotenv("../.env")
db = psycopg2.connect(
    host=os.getenv("POSTGRES_HOST"),
    user=os.getenv("POSTGRES_USER"),
    password=os.getenv("POSTGRES_PASS"),
    database=os.getenv("POSTGRES_DB"),
    port=os.getenv("POSTGRES_PORT"),
)
cursor = db.cursor()

with tempfile.NamedTemporaryFile() as f:
    f.write(requests.get("https://raw.githubusercontent.com/ggmolly/belfast-data/main/EN/shop_template.json").content)
    f.seek(0)
    data = json.load(f)

print("[#] inserting shop offers...")

"""
create table shop_offers
(
    id              bigserial
        primary key,
    effects         integer[],
    number          bigint,
    resource_number bigint,
    resource_id     bigint
        constraint fk_shop_offers_resource
            references resources,
    type            bigint
);
"""

RESERVED_STRINGS = { # this map is used to convert some strings to an arbitrary number
    "count": 2890000,
    "equip_bag_size": 2890001,
    "ship_bag_size": 2890002,
    "dorm_exp_pos": 2890003,
    "dorm_food_max": 2890004,
    "tradingport_level": 2890005,
    "oilfield_level": 2890006,
    "shop_street_level": 2890007,
    "shop_street_flash": 2890008,
    "dorm_fix_pos": 2890009,
    "dorm_floor": 2890010,
    "class_room_level": 2890011,
    "skill_room_pos": 2890012,
    "commander_bag_size": 2890013,
    "spweapon_bag_size": 2890014,
}

for shop_offer_id in data:
    shop_offer = data[shop_offer_id]
    id = shop_offer["id"]
    effects = shop_offer["effect_args"]
    number = shop_offer["num"]
    if isinstance(effects, str):
        effects = [RESERVED_STRINGS[effects]]
    for i, effect in enumerate(effects):
        if isinstance(effect, str):
            effects[i] = RESERVED_STRINGS[effect]
    resource_number = shop_offer["resource_num"]
    resource_id = shop_offer["resource_type"]
    type = shop_offer["type"]
    cursor.execute("""
    insert into shop_offers
    (id, effects, number, resource_number, resource_id, type)
    values (%s, %s, %s, %s, %s, %s) ON CONFLICT (id) DO UPDATE SET effects = %s, number = %s, resource_number = %s, resource_id = %s, type = %s
    """, (id, effects, number, resource_number, resource_id, type, effects, number, resource_number, resource_id, type))

db.commit()