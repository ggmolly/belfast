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
    f.write(requests.get("https://raw.githubusercontent.com/ggmolly/belfast-data/main/EN/player_resource.json").content)
    f.seek(0)
    data = json.load(f)

allowed_items = set()
cursor.execute("SELECT id FROM items")
for row in cursor.fetchall():
    allowed_items.add(row[0])

for resource_id in data:
    # check if resource_id is a number
    if not resource_id.isdigit():
        continue
    resource = data[resource_id]
    id = resource["id"]
    item_id = resource["itemid"]
    if item_id == 0 or item_id not in allowed_items:
        item_id = None
    name = resource["name"]
    cursor.execute(
        """
        INSERT INTO resources (id, item_id, name)
        VALUES (%s, %s, %s)
        ON CONFLICT (id) DO UPDATE SET
            item_id = %s,
            name = %s
        """,
        (id, item_id, name, item_id, name),
    )
db.commit()
db.close()
print("[#] done")