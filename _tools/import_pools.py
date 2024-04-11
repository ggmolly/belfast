import os
import json
import psycopg2
import requests
from bs4 import BeautifulSoup
from dotenv import load_dotenv

load_dotenv("../.env")

db = psycopg2.connect(
    host=os.getenv("POSTGRES_HOST"),
    user=os.getenv("POSTGRES_USER"),
    password=os.getenv("POSTGRES_PASS"),
    database=os.getenv("POSTGRES_DB"),
    port=os.getenv("POSTGRES_PORT"),
)
cur = db.cursor()

URL = "https://azurlane.koumakan.jp/wiki/Building#Permanent_Construction"

resp = requests.get(URL)
assert resp.status_code == 200
soup = BeautifulSoup(resp.text, "html.parser")

name_id = {
    "Light": 2,
    "Heavy": 3,
    "Special": 1,
}

def get_template_id(name: str) -> int:
    name = name.replace("µ", "μ")
    name = name.replace(" (Battleship)", "(BB)")
    try:
        cur.execute("SELECT template_id FROM ships WHERE name = %s ORDER BY template_id ASC LIMIT 1", (name,))
        return cur.fetchone()[0]
    except:
        # print(f"Could not find {name}")
        return None

pools = soup.select("table.azltable.bdpooltbl.mw-collapsible.mw-collapsed.toggle-right > tbody > tr")
entries = []
# print(len(pools))
for row in pools:
    pool_name = row.select_one("th").text.strip()
    pool_id = name_id[pool_name]
    ships = row.select("div.alicapt")
    for ship in ships:
        name = ship.select_one("a").text.strip()
        id = get_template_id(name)
        entries.append((id, pool_id))
cur.execute("UPDATE ships SET pool_id = NULL")
for entry in entries:
    cur.execute("UPDATE ships SET pool_id = %s WHERE template_id = %s", (entry[1], entry[0]))
db.commit()
