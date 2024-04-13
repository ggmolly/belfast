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
    password=os.getenv("POSTGRES_PASSWORD"),
    database=os.getenv("POSTGRES_DB"),
    port=os.getenv("POSTGRES_PORT"),
)
cursor = db.cursor()
with tempfile.NamedTemporaryFile() as f:
    f.write(requests.get("https://raw.githubusercontent.com/ggmolly/belfast-data/main/EN/benefit_buff_template.json").content)
    f.seek(0)
    data = json.load(f)

for buff_id in data:
    buff = data[buff_id]
    id = buff["id"]
    name = buff["name"]
    desc = buff["desc"]
    max_time = buff["max_time"]
    benefit_type = buff["benefit_type"]
    cursor.execute("""
    insert into buffs
    (id, name, description, max_time, benefit_type)
    values (%s, %s, %s, %s, %s) ON CONFLICT (id) DO UPDATE SET name = %s, description = %s, max_time = %s, benefit_type = %s
    """, (id, name, desc, max_time, benefit_type, name, desc, max_time, benefit_type))

db.commit()
db.close()
print("[#] done")