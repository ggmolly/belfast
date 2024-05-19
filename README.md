# ⚓ Belfast

Belfast is a private server reimplementation for the mobile game [Azur Lane](https://en.wikipedia.org/wiki/Azur_Lane), written in [Go](https://go.dev/) using [Gorm](https://gorm.io), [HTMX](https://htmx.org), [Hyperscript](https://hyperscript.org) and [Gofiber](https://gofiber.io). Currently in early development, Belfast works on both iOS and Android without requiring jailbreak or root access, making it a tamper-free alternative to the official game server.

> [!CAUTION]
> Latest game update broke Belfast, working on it

> [!WARNING]
> To bump the version of the game, you need to load the index of Belfast's web UI through a browser or `curl`.

> [!WARNING]
> Protobuf messages **are not** automatically updated. You need to update them manually, yet.

> [!IMPORTANT]
> Some packets have invalid / no names.

> [!TIP]
> The [import_pcap.py](./_tools/import_pcap.py) script can help you import packets from a `pcap` file into Belfast's dissection tool.

# 🌟 Features

Belfast currently has:

- A cool looking web UI to tinker easily with the game.
- A low-level multiplexed TCP server, which allows multiple connections at once.
- The ability of following game updates, along with importing ship, items, ... data automatically (US version).
- A small API that allows you to quickly implement new game messages without head scratching.
- A great dissection tool in which every packet is stored, along with a `protobuf` -> `json` deserializer (available in the web UI).

# 🌠 State

Belfast reimplements these features from the game:

- Custom server list
- Player bans
- Commander's dock (owned ships).
- Commander's depot (owned items).
- Build (you can start / end / edit builds).
- Resources (collection / consuming).
- Mails (along with custom sender, body, title, attachments, read / important states).
- Retire ships.
- Buying / equipping skins.
- Propose ships.
- Game's public chatroom.
- Secretaries (add, remove, moving is buggy but 'works').
- Lock / unlock ships.
- Rename proposed ships (features the 30d cooldown too).
- Custom notices

# 🚀 Roadmap

As I just started opening this project to the public I want to do these things:

1. Clean the code.
2. Refactor the web UI.
3. Add unit tests.
4. Create a repository for the [Belfast's website](https://belfast.mana.rip/) in general.
5. Add a web UI tab for game's notices.

Before continuting the implementation of the game's protocol.

# 📦 Python requirements

To use the dissection tool (`import_pcap.py`), you need to install the following python dependencies:

- `psycopg2`
- `scapy`
- `python-dotenv`

# ⚠️ Note

While I'm proud about the progress made, I can't deny the code quality is less than ideal. This entire thing was hacked in over the course of just five days.

I initially had no plans to make this project public or open-source. It was simply a fun challenge that I undertook in my free time starting in December 2023. While the code quality may not be perfect, my hope is that it can serve as a starting point for others looking to get into Go development or explore Azur Lane's netcode. Let's see where this journey takes us!

# 📧 Contact

You can contact me (Molly) [here](molly+belfast@mana.rip)
