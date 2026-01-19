#!/usr/bin/env bash
set -euo pipefail

WG_IFACE="${1:-enp7s0}"
OUT="${2:-wg-traffic.pcap}"

HOSTS=$(awk 'NF {if ($1 == "#") print $3; else print $2}' wg/hosts | sort -u)

FILTER=$(for host in $HOSTS; do getent ahosts "$host" | awk '{print $1}'; done | sort -u | awk 'NF{printf "host %s or ", $1}' | sed 's/ or $//')

exec sudo tcpdump -i "$WG_IFACE" -nn -w "$OUT" -- "$FILTER"
