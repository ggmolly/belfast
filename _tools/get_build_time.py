import sqlite3
import sys
import json
import requests
import re
import psycopg2
import tempfile
import os
from bs4 import BeautifulSoup
from tqdm import tqdm
from dotenv import load_dotenv
load_dotenv("../.env")
db = psycopg2.connect(
    host=os.getenv("POSTGRES_HOST"),
    user=os.getenv("POSTGRES_USER"),
    password=os.getenv("POSTGRES_PASSWORD"),
    database=os.getenv("POSTGRES_DB"),
    port=os.getenv("POSTGRES_PORT"),
)
cur = db.cursor()


TIME_REGEX = re.compile(r"(\d+):(\d+):(\d+)")
conn = sqlite3.connect("build_times.db")
cursor = conn.cursor()

cursor.execute("""
    CREATE TABLE IF NOT EXISTS build_times (
        template_id INTEGER PRIMARY KEY,
        name TEXT NOT NULL,
        time INTEGER NOT NULL DEFAULT 0
    )
""")
cursor.execute("""
    CREATE INDEX IF NOT EXISTS name_index ON build_times (name)
""")
conn.commit()

with tempfile.NamedTemporaryFile() as f:
    f.write(requests.get("https://raw.githubusercontent.com/ggmolly/belfast-data/main/EN/ship_data_statistics.json").content)
    f.seek(0)
    ship_stats = json.load(f)

def get_build_time(name: str) -> int:
    url = f"https://azurlane.koumakan.jp/wiki/{name}"
    r = requests.get(url)
    soup = BeautifulSoup(r.text, "html.parser")
    try:
        info = soup.select("table.card-info-tbl > tbody > tr:first-child > td:nth-child(2)")[0].text
        info = TIME_REGEX.match(info).groups()
        info = int(info[0]) * 3600 + int(info[1]) * 60 + int(info[2])
        return info
    except:
        return 0

def build_table():
    registered_ships = set()
    for key in tqdm(ship_stats, desc="getting ship build times", total=len(ship_stats)):
        ship = ship_stats[key]
        # skip if we already have the ship in the database, they never change
        cur.execute("SELECT COUNT(*) FROM ships WHERE name = %s", (ship["name"],))
        if cur.fetchone()[0] > 0:
            continue
        id = ship["id"]
        name = ship["name"].strip()
        if name in registered_ships: # some ships have multiple entries for some reason
            continue
        registered_ships.add(name)
        seconds = get_build_time(name)
        cursor.execute("""
            INSERT INTO build_times (template_id, name, time)
            VALUES (?, ?, ?) ON CONFLICT (template_id) DO UPDATE SET name = ?, time = ?
        """, (id, name, seconds, name, seconds))
    conn.commit()

def query_table(name: str) -> int:
    cursor.execute("""
        SELECT time FROM build_times WHERE name = ?;
    """, (name,))
    return cursor.fetchone()[0]

def update_ships():
    cur.execute("SELECT template_id, name FROM ships")
    for template_id, name in tqdm(cur.fetchall(), desc="updating ships"):
        seconds = query_table(name)
        cur.execute("UPDATE ships SET build_time = %s WHERE template_id = %s", (seconds, template_id))
    db.commit()

if __name__ == "__main__":
    if len(sys.argv) == 1:
        build_table()
    update_ships()