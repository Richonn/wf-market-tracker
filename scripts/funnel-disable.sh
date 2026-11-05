#!/bin/sh
set -e

# Emergency stop — immediately kills public access.

docker exec tailscale tailscale funnel --https=443 off
echo "Funnel disabled."
