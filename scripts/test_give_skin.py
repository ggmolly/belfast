#!/usr/bin/env python3

import argparse
import sys

import requests


def main() -> int:
    parser = argparse.ArgumentParser(description="Give a skin to a player")
    parser.add_argument(
        "--base-url", default="http://localhost:8000", help="API base URL"
    )
    parser.add_argument("--player-id", type=int, required=True, help="Player ID")
    parser.add_argument("--skin-id", type=int, required=True, help="Skin ID")
    args = parser.parse_args()

    url = f"{args.base_url.rstrip('/')}/api/v1/players/{args.player_id}/give-skin"
    response = requests.post(url, json={"skin_id": args.skin_id}, timeout=10)

    if response.status_code != 204:
        print(f"expected 204, got {response.status_code}")
        try:
            print(response.json())
        except ValueError:
            print(response.text)
        return 1

    print("ok")
    return 0


if __name__ == "__main__":
    sys.exit(main())
