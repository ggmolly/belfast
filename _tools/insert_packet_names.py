import os
import psycopg2
import re
from glob import glob
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

packets = glob("../protobuf/CS_*.go")
packets.extend(glob("../protobuf/SC_*.go"))
id_regex = re.compile(r"(CS|SC)_(\d+)\.pb\.go")
for packet in packets:
    packet_id = id_regex.search(packet).group(2)
    packet_name = os.path.basename(packet).split(".")[0]
    cursor.execute("INSERT INTO debug_names (id, name) VALUES (%s, %s) ON CONFLICT (id) DO UPDATE SET name = %s", (packet_id, packet_name, packet_name))

db.commit()