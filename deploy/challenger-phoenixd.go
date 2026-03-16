// This file is designed to be copied into lightninglabs/aperture/challenger/
// during the Docker build. It uses Aperture's actual types.
package challenger

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
)

// phoenixdInvoice holds the response from Phoenixd's createinvoice endpoint.
type phoenixdInvoice struct {
	PaymentHash string `json:"paymentHash"`
	Serialized  string `json:"serialized"`
}

// PhoenixdChallenger implements the Challenger interface using a Phoenixd
// Lightning node instead of LND. Only strictVerify=false is supported.
type PhoenixdChallenger struct {
	baseURL  string
	password string
	client   *http.Client
}

// NewPhoenixdChallenger creates a challenger backed by a Phoenixd instance.
func NewPhoenixdChallenger(phoenixdURL,
	password string) (*PhoenixdChallenger, error) {

	return &PhoenixdChallenger{
		baseURL:  strings.TrimRight(phoenixdURL, "/"),
		password: password,
		client:   &http.Client{},
	}, nil
}

// NewChallenge creates a Lightning invoice via Phoenixd and returns the
// BOLT11 payment request and payment hash.
func (p *PhoenixdChallenger) NewChallenge(price int64) (string,
	lntypes.Hash, error) {

	var hash lntypes.Hash

	form := url.Values{
		"amountSat":   {strconv.FormatInt(price, 10)},
		"description": {"L402"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost,
		p.baseURL+"/createinvoice",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", hash, fmt.Errorf("phoenixd: build request: %w",
			err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("", p.password)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", hash, fmt.Errorf("phoenixd: connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", hash, fmt.Errorf("phoenixd: createinvoice: "+
			"HTTP %d", resp.StatusCode)
	}

	var inv phoenixdInvoice
	if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
		return "", hash, fmt.Errorf("phoenixd: invalid response: "+
			"%w", err)
	}

	if inv.PaymentHash == "" {
		return "", hash, fmt.Errorf("phoenixd: empty payment hash " +
			"in response")
	}

	hashBytes, err := hex.DecodeString(inv.PaymentHash)
	if err != nil {
		return "", hash, fmt.Errorf("phoenixd: invalid payment "+
			"hash: %w", err)
	}
	if len(hashBytes) != 32 {
		return "", hash, fmt.Errorf("phoenixd: payment hash wrong "+
			"length: got %d bytes", len(hashBytes))
	}

	copy(hash[:], hashBytes)
	return inv.Serialized, hash, nil
}

// VerifyInvoiceStatus is a no-op. With strictVerify=false (Aperture's
// default), macaroon + preimage verification is the security model.
func (p *PhoenixdChallenger) VerifyInvoiceStatus(_ lntypes.Hash,
	_ lnrpc.Invoice_InvoiceState, _ time.Duration) error {

	return nil
}

// Stop is a no-op. The Phoenixd client is stateless.
func (p *PhoenixdChallenger) Stop() {}
