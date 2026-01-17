import os
import requests
import datetime
from bs4 import BeautifulSoup

URL = "https://azurlane.koumakan.jp/wiki/List_of_Ships_by_Image"
OUTPUT_PATH = "../web/static/al_faces"

if os.path.exists(".icon_last_update"):
    with open(".icon_last_update", "r") as f:
        last_update = int(f.read())
else:
    last_update = 0

req = requests.get(URL)
soup = BeautifulSoup(req.text, "html.parser")

last_edit = soup.select_one("li#footer-info-lastmod").text.split(" on ")[1].replace(".", "")
last_edit = datetime.datetime.strptime(last_edit, "%d %B %Y, at %H:%M")
last_edit = last_edit.timestamp()

if last_edit > last_update:
    print("[+] updating ship icons")
else:
    print("[!] ship icons are up to date")
    exit()

for ship_car in soup.select("div.azl-shipcard"):
    name = ship_car.select_one("div.alc-top.truncate > a").text
    img_src = ship_car.select_one("div.alc-img > span > a > img").get("src")
    extension = img_src.split(".")[-1].split("?")[0]
    ship_path = os.path.join(OUTPUT_PATH, f"{name}.{extension}").replace("µ", "μ")
    if os.path.exists(ship_path):
        continue
    print(f"[+] downloading {name}")
    with requests.get(img_src, stream=True) as r:
        with open(ship_path, "wb") as f:
            for chunk in r.iter_content(chunk_size=8192):
                f.write(chunk)

with open(".icon_last_update", "w") as f:
    f.write(str(int(last_edit)))