#!/bin/sh
set -e

# Expose the API container publicly via Tailscale Funnel (port 443 → 8080).
# Run this once on the NAS after `docker compose up -d`.
#
# Emergency stop: tailscale funnel off

docker exec tailscale tailscale funnel --bg 8080
echo "Funnel active. Public URL:"
docker exec tailscale tailscale funnel status
