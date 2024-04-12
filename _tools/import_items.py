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
    f.write(requests.get("https://raw.githubusercontent.com/ggmolly/belfast-data/main/EN/item_data_statistics.json").content)
    f.seek(0)
    data = json.load(f)

print("[#] inserting items...")
for item_id in data:
    item = data[item_id]
    id = item["id"]
    name = item["name"]
    rarity = item["rarity"]
    shop_id = item.get("shop_id", -2)
    type = item["type"]
    virtual_type = item["virtual_type"]
    try:
        cursor.execute("""
        insert into items
        (id, name, rarity, shop_id, type, virtual_type)
        values (%s, %s, %s, %s, %s, %s) ON CONFLICT (id) DO UPDATE SET name = %s, rarity = %s, shop_id = %s, type = %s, virtual_type = %s
        """, (id, name, rarity, shop_id, type, virtual_type, name, rarity, shop_id, type, virtual_type))
    except Exception as e:
        print(name)
        print(e)
        exit()

db.commit()
