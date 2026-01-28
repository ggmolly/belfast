# ⚓ Belfast

Belfast is a private server reimplementation for the mobile game [Azur Lane](https://en.wikipedia.org/wiki/Azur_Lane), written in [Go](https://go.dev/) using [Iris](https://www.iris-go.com/) and [Gorm](https://gorm.io). It targets iOS and Android clients without requiring jailbreak or root access.

Belfast is in a very unstable state and the server is not complete at all.

> [!TIP]
> Use `cmd/pcap_decode/main.go` to decode packets from `pcap` files into JSON.

# 📊 Packet Progress

![Packet progress](https://cdn.molly.sh/belfast/implem.png)

# 🌟 Features

Belfast currently has:

- A low-level multiplexed TCP server, which allows multiple connections at once.
- The ability of following game updates, along with importing ship, items, ... data automatically (US version).
- A small API that allows you to quickly implement new game messages without head scratching.
- A great dissection tool in which every packet is stored, along with a `protobuf` -> `json` deserializer.
- A REST API with Swagger docs and admin endpoints for server tooling.
- A web UI in development: https://github.com/ggmolly/belfast-web.
- Config-driven packet response hydration for rapid prototyping.
- Packet progress tooling and webhook-based status updates.
- Runtime config toggles (maintenance mode, host/port overrides).

# ⚙️ Config

- `cmd/belfast` defaults to `server.toml` (game server config).
- `cmd/gateway` defaults to `gateway.toml` (gateway config).
- Gateway server list is defined in `[[servers]]`; server names come from each game server's `/api/v1/server/status`.
- To embed the git commit in status, build with `-ldflags "-X github.com/ggmolly/belfast/internal/buildinfo.Commit=$(git rev-parse --short HEAD)"`.

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
- Arena shop.
- Medal shop.
- Minigame shop.
- Guild shop.
- Juustagram activity + chat operations.
- Educate/TB flows and state.
- Compensations (notifications + reward claims).
- Commander buffs.
- Shopping street shop.
- Dorm3D apartment state (persisted on reconnect).
- Build queue snapshot.
- Random flagship selection updates.
- Remaster tickets, progress, and rewards.

# 🚀 Roadmap

As I just started opening this project to the public I want to do these things:

1. Add unit tests.
2. Make a roadmap.

Before continuting the implementation of the game's protocol.

# 📧 Contact

You can contact me (Molly) [here](mailto:molly@molly.sh).
