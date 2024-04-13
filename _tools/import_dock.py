import os
import json
import psycopg2
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


ships_id = []

with open("CommanderFleetA (SC_12010)_15.json", "r") as f:
    data = json.load(f)
    for ship in data["ship_list"]:
        ships_id.append(
            (
                ship["template_id"],
                ship["level"],
                ship["max_level"],
                ship["energy"],
                ship["intimacy"],
                ship["is_locked"] == "true",
                ship["propose"] == "true",
                ship["common_flag"] == "true",
                ship["blue_print_flag"] == "true",
                ship["proficiency"] == "true",
                ship["activity_npc"] or 0,
                ship["name"],
                ship["change_name_timestamp"],
                ship["create_time"],
                ship["skin_id"]
            )
        )

with open("CommanderFleet (SC_12001)_14.json", "r") as f:
    data = json.load(f)
    for ship in data["shiplist"]:
        ships_id.append(
            (
                ship["template_id"],
                ship["level"],
                ship["max_level"],
                ship["energy"],
                ship["intimacy"],
                ship["is_locked"] == "true",
                ship["propose"] > 0,
                ship["common_flag"] == "true",
                ship["blue_print_flag"] == "true",
                ship["proficiency"] == "true",
                ship["activity_npc"] or 0,
                ship["name"],
                ship["change_name_timestamp"],
                ship["create_time"],
                ship["skin_id"]
            )
        )

COMMANDER_ID = 5640350

for ship in ships_id:
    cursor.execute("""insert into owned_ships (owner_id, ship_id, level, max_level, energy, intimacy, is_locked, propose, common_flag,
        blueprint_flag, proficiency, activity_npc, custom_name, change_name_timestamp, create_time, skin_id)
        values (%s, %s, %s, %s, %s, %s, %s::boolean, %s::boolean, %s::boolean, %s::boolean, %s::boolean, %s, %s, to_timestamp(%s), to_timestamp(%s), %s)
    """, (COMMANDER_ID, *ship))
db.commit()