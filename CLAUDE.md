# CLAUDE.md — aperture-phoenixd

Standalone Go module: Phoenixd challenger for Aperture L402 auth. Implements Aperture's `mint.Challenger` and `auth.InvoiceChecker` interfaces using a Phoenixd Lightning node instead of LND.

## Commands

```bash
go build ./...         # Build
go test ./...          # Run all tests
go test -race ./...    # Run with race detector
go vet ./...           # Lint
```

## Structure

```
client.go              # Phoenixd HTTP client (createinvoice, getpayment)
challenger.go          # PhoenixdChallenger (NewChallenge, VerifyInvoiceStatus)
doc.go                 # Package documentation
cmd/echo-server/       # Minimal demo API for Aperture to proxy
```

## Architecture

`Client` wraps the Phoenixd HTTP API (two endpoints: `POST /createinvoice` and `GET /payments/incoming/{hash}`). `PhoenixdChallenger` uses this client to create invoices and verify payment status. Only `strictVerify=false` is supported (the Aperture default). All HTTP requests use a 10-second timeout context.

## Integration

Import as a Go module:
```go
import phoenixd "github.com/forgesworn/aperture-phoenixd"

challenger := phoenixd.NewChallenger("http://localhost:9740", "phoenixd-password")
```

Wire `challenger` into Aperture's configuration as a custom `Challenger` implementation.

## Conventions

- **British English** — colour, initialise, behaviour, licence
- **Go standard layout** — cmd/ for binaries
- **Git:** commit messages use `type: description` format
- **Git:** Do NOT include `Co-Authored-By` lines
- **testify/require** for all test assertions
- **golangci-lint** with Aperture-compatible linter set
