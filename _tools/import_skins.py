import os
import psycopg2
import json
from dotenv import load_dotenv
from tqdm import tqdm

load_dotenv("../.env")
db = psycopg2.connect(
    host=os.getenv("POSTGRES_HOST"),
    user=os.getenv("POSTGRES_USER"),
    password=os.getenv("POSTGRES_PASS"),
    database=os.getenv("POSTGRES_DB"),
    port=os.getenv("POSTGRES_PORT"),
)
cursor = db.cursor()

SKIN_STATS_PATH = "/home/molly/Documents/al-zero/AzurLaneData/EN/ShareCfg/ship_skin_template.json"

print("[#] loading skin data")
with open(SKIN_STATS_PATH, "r") as f:
    skin_stats = json.load(f)

print("[#] inserting skin data")
for skin_id in tqdm(skin_stats, desc="inserting skin data", total=len(skin_stats)):
    skin = skin_stats[skin_id]
    id = skin["id"]
    name = skin["name"]
    ship_group = skin["ship_group"]
    cursor.execute("""
        INSERT INTO skins (id, name, ship_group)
        VALUES (%s, %s, %s) ON CONFLICT (id) DO UPDATE SET name = %s, ship_group = %s
        """, (id, name, ship_group, name, ship_group)
    )

print("[#] commiting changes")
db.commit()