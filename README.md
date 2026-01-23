# ⚓ Belfast

Belfast is a private server reimplementation for the mobile game [Azur Lane](https://en.wikipedia.org/wiki/Azur_Lane), written in [Go](https://go.dev/) using [Gorm](https://gorm.io). It targets iOS and Android clients without requiring jailbreak or root access.

Belfast is in a very unstable state and the server is not complete at all.

> [!TIP]
> Use `cmd/pcap_decode/main.go` to decode packets from `pcap` files into JSON.

# 🌟 Features

Belfast currently has:

- A low-level multiplexed TCP server, which allows multiple connections at once.
- The ability of following game updates, along with importing ship, items, ... data automatically (US version).
- A small API that allows you to quickly implement new game messages without head scratching.
- A great dissection tool in which every packet is stored, along with a `protobuf` -> `json` deserializer.

# 📊 Packet Progress

![Packet progress](https://cdn.molly.sh/belfast/implem.png)

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
- Fleets management (add / remove / move ships & rename)

# 🚀 Roadmap

As I just started opening this project to the public I want to do these things:

1. Add unit tests.
2. Make a roadmap.

Before continuting the implementation of the game's protocol.

# 📧 Contact

You can contact me (Molly) [here](mailto:molly@molly.sh).
