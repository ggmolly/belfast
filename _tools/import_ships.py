import os
import psycopg2
import json
import sqlite3
import requests
import tempfile
import subprocess
from dotenv import load_dotenv
from tqdm import tqdm
from typing import Dict, List, Tuple

# execute the get_build_time.py script
# TODO: move this to the belfast-data repo
# if not os.getcwd().endswith("_tools"):
#     os.chdir("_tools")
# subprocess.run(["python3", "get_build_time.py"])

build_time_db = sqlite3.connect("build_times.db")
build_time_cursor = build_time_db.cursor()

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
    f.write(requests.get("https://raw.githubusercontent.com/ggmolly/belfast-data/main/EN/ship_data_statistics.json").content)
    f.seek(0)
    ship_stats = json.load(f)

rarities = set()
ships = []
print("[#] inserting ship data")
for ship_id in tqdm(ship_stats, desc="inserting ship data", total=len(ship_stats)):
    ship = ship_stats[ship_id]
    id = ship["id"]
    name = ship["name"].strip()
    nationality = ship["nationality"]
    rarity = ship["rarity"]
    star = ship["star"]
    type = ship["type"]
    skin_id = ship["skin_id"]
    build_time_cursor.execute("""
        SELECT time FROM build_times WHERE name = ?;
    """, (name,))
    try:
        build_time = build_time_cursor.fetchone()[0]
    except:
        print("[!] no build time found for", name)
        build_time = 0
    rarities.add(rarity)
    ships.append((id, name, nationality, rarity, star, type, build_time, skin_id))

print("[#] inserting ships")
KNOWN_INVALID_IDS = [
    900197,
    900198,
    900029,
]

groups: Dict[int, Tuple[List[int], List[str]]] = {}
ship_data: Dict[int, tuple] = {}
for ship in ships:
    ship_group = ship[0] // 10
    if ship_group not in groups:
        groups[ship_group] = ([], [])
    groups[ship_group][0].append(ship[0])
    groups[ship_group][1].append(ship[1])
    ship_data[ship[0]] = ship

# Remove all groups that have less than 2 ships or less
for group in list(groups.keys()):
    if len(groups[group][0]) < 2:
        print("[!] invalid group", groups[group])
        del groups[group]
        continue
    if len(set(groups[group][1])) != 1: # if there are different names in the group
        print("[!] invalid group", groups[group])
        del groups[group]

for group in list(groups.keys()):
    ships = groups[group][0]
    for ship in ships:
        data = ship_data[ship]
        cursor.execute("""
            INSERT INTO ships (template_id, name, nationality, rarity_id, star, type, build_time)
            VALUES (%s, %s, %s, %s, %s, %s, %s) ON CONFLICT DO NOTHING;
            """, (data[0], data[1], data[2], data[3], data[4], data[5], data[6])
        )
db.commit()