#!/bin/bash
# Announce the Aperture demo service on Nostr.
# Run this after docker-compose is up and the service is reachable.
#
# Prerequisites:
#   - aperture-announce installed: go install github.com/TheCryptoDonkey/aperture-announce/cmd/aperture-announce@latest
#   - PUBLIC_URL set to your domain (e.g. https://demo.example.com:8081)
#   - ANNOUNCE_RELAYS set (e.g. wss://relay.damus.io,wss://nos.lol)

set -euo pipefail

: "${PUBLIC_URL:?Set PUBLIC_URL to your public endpoint}"
: "${ANNOUNCE_RELAYS:?Set ANNOUNCE_RELAYS to comma-separated relay URLs}"

echo "Dry run first..."
aperture-announce \
  --config aperture.yaml \
  --public-urls "$PUBLIC_URL" \
  --dry-run

echo ""
read -rp "Publish to relays? [y/N] " confirm
if [[ "$confirm" =~ ^[Yy]$ ]]; then
  aperture-announce \
    --config aperture.yaml \
    --public-urls "$PUBLIC_URL" \
    --relays "$ANNOUNCE_RELAYS" \
    --topics "demo,phoenixd" \
    --verbose
fi
