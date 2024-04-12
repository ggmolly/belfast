import os
import psycopg2
import json
import requests
import tempfile
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

with tempfile.NamedTemporaryFile() as f:
    f.write(requests.get("https://raw.githubusercontent.com/ggmolly/belfast-data/main/EN/ship_skin_template.json").content)
    f.seek(0)
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