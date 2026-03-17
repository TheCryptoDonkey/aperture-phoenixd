package phoenixd

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"
)

// PhoenixdChallenger implements Aperture's mint.Challenger and
// auth.InvoiceChecker interfaces using a Phoenixd Lightning node.
// Only strictVerify=false is supported (the Aperture default).
type PhoenixdChallenger struct {
	client *Client
}

// NewChallenger creates a challenger backed by a Phoenixd instance.
func NewChallenger(phoenixdURL, password string) *PhoenixdChallenger {
	return &PhoenixdChallenger{
		client: NewClient(phoenixdURL, password),
	}
}

// NewChallenge creates a Lightning invoice via Phoenixd and returns the
// BOLT11 payment request and 32-byte payment hash.
func (p *PhoenixdChallenger) NewChallenge(price int64) (string, [32]byte, error) {
	var hash [32]byte

	if price <= 0 {
		return "", hash, fmt.Errorf("phoenixd: price must be positive, got %d", price)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inv, err := p.client.CreateInvoice(ctx, price, "L402")
	if err != nil {
		return "", hash, err
	}

	if inv.PaymentHash == "" {
		return "", hash, fmt.Errorf("phoenixd: empty payment hash in response")
	}

	hashBytes, err := hex.DecodeString(inv.PaymentHash)
	if err != nil {
		return "", hash, fmt.Errorf("phoenixd: invalid payment hash: %w", err)
	}
	if len(hashBytes) != 32 {
		return "", hash, fmt.Errorf("phoenixd: payment hash wrong length: got %d bytes", len(hashBytes))
	}

	copy(hash[:], hashBytes)
	return inv.Serialized, hash, nil
}

// VerifyInvoiceStatus is a no-op. With strictVerify=false (Aperture's
// default), macaroon + preimage verification is the security model.
// strictVerify=true is not supported in this version.
func (p *PhoenixdChallenger) VerifyInvoiceStatus(hash [32]byte, price int64, service string) error {
	return nil
}

// Stop is a no-op. The Phoenixd client is stateless.
func (p *PhoenixdChallenger) Stop() {}
