import sys
import psycopg2
from scapy.all import *
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

data = b""
if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 import_pcap.py <pcapng file>")
        exit(1)

    pcap = rdpcap(sys.argv[1])
    print("[#] importing {} packets".format(len(pcap)))
    # for each packet, print tcp payload
    for pkt in pcap:
        if pkt.haslayer(TCP) and len(pkt[TCP].payload) > 0:
            payload = bytes(pkt[TCP].payload)
            data += payload
    print("[#] parsing {} bytes".format(len(data)))
    offset = 0
    packet_size = 0
    packet_id = 0
    packet_buffer = b""
    packets = 0
    while offset < len(data):
        packet_size = data[offset+0] << 8 | data[offset+1]
        packet_index = data[offset + 5] << 8 | data[offset + 6]
        if len(packet_buffer) == 0:
            packet_id = data[offset+3] << 8 | data[offset+4]
            print("[+] new packet detected, id:", packet_id)
        packet_buffer += data[offset + 7:offset + 7 + packet_size - 5]
        offset += packet_size + 2
        if len(packet_buffer) == packet_size - 5:
            print(f"[+] complete SC_{packet_id} packet received, inserting...")
            try:
                cursor.execute("INSERT INTO debugs (packet_size, packet_id, data) VALUES (%s, %s, %s)", (packet_size, packet_id, psycopg2.Binary(packet_buffer)))
            except:
                traceback.print_exc()
                print("[!] error inserting packet, skipping...")
            packet_buffer = b""
            packets += 1
        else:
            print(f"[-] incomplete packet received, waiting for more data... ({len(packet_buffer)}/{packet_size}) [SC_{packet_id}]")

    print("[#] committing changes..", end=" ", flush=True)
    db.commit()
    print("done! ({} packets inserted)".format(packets))