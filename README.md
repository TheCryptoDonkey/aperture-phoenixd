# aperture-phoenixd

**Nostr:** [`npub1mgvlrnf5hm9yf0n5mf9nqmvarhvxkc6remu5ec3vf8r0txqkuk7su0e7q2`](https://njump.me/npub1mgvlrnf5hm9yf0n5mf9nqmvarhvxkc6remu5ec3vf8r0txqkuk7su0e7q2)

[![MIT licence](https://img.shields.io/badge/licence-MIT-blue.svg)](./LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8)](https://golang.org/)
[![GitHub Sponsors](https://img.shields.io/github/sponsors/TheCryptoDonkey?logo=githubsponsors&color=ea4aaa&label=Sponsor)](https://github.com/sponsors/TheCryptoDonkey)

Use [Phoenixd](https://phoenix.acinq.co/server) as the Lightning backend for [Aperture](https://github.com/lightninglabs/aperture) — no LND required.

Implements Aperture's `mint.Challenger` and `auth.InvoiceChecker` interfaces against Phoenixd's REST API. Drop-in replacement for LND with `strictVerify=false` (Aperture's default).

## Quick start

```bash
go get github.com/forgesworn/aperture-phoenixd
```

```go
import "github.com/forgesworn/aperture-phoenixd"

challenger := phoenixd.NewChallenger("http://localhost:9740", "your-phoenixd-password")

// challenger.NewChallenge(priceSats) — creates a Lightning invoice via Phoenixd
// challenger.VerifyInvoiceStatus(...) — no-op for strictVerify=false
// challenger.Stop() — no-op (stateless HTTP client)
```

## Integrating with Aperture

See `testdata/aperture-patch.diff` for the ~20-line diff to wire this into Aperture's `aperture.go`. Adds `PhoenixdURL` and `PhoenixdPassword` config fields alongside the existing `LndHost` and `Passphrase` options.

## Demo

The included echo server provides a minimal API to proxy through Aperture:

```bash
go run ./cmd/echo-server
# Listens on :8080, returns request body at /v1/echo
```

## How it works

1. Aperture receives an HTTP request and needs to create an L402 challenge
2. `PhoenixdChallenger.NewChallenge(price)` calls Phoenixd's `POST /createinvoice`
3. Phoenixd creates a Lightning invoice and returns the BOLT11 string + payment hash
4. Aperture wraps this in a macaroon and returns it as the L402 challenge
5. The client pays the invoice and presents the macaroon + preimage
6. Aperture verifies the preimage against the payment hash — access granted

## Limitations

- **`strictVerify=true` is not supported.** Phoenixd's WebSocket does not emit invoice cancellation events required for full invoice status tracking. With `strictVerify=false` (the default), macaroon + preimage verification is the security model.

## Ecosystem

| Project | Role |
|---------|------|
| [aperture](https://github.com/lightninglabs/aperture) | L402 reverse proxy (what this adapter plugs into) |
| [aperture-announce](https://github.com/forgesworn/aperture-announce) | Announces Aperture services on Nostr for discovery |
| [402-mcp](https://github.com/forgesworn/402-mcp) | MCP client that discovers and pays for L402 APIs |

## Licence

[MIT](LICENSE)
