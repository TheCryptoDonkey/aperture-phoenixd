package phoenixd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Client communicates with a Phoenixd instance via its REST API.
type Client struct {
	baseURL  string
	password string
	http     *http.Client
}

// Invoice holds the response from Phoenixd's createinvoice endpoint.
type Invoice struct {
	PaymentHash string `json:"paymentHash"`
	Serialized  string `json:"serialized"`
}

// Payment holds the response from Phoenixd's incoming payment endpoint.
type Payment struct {
	IsPaid    bool  `json:"isPaid"`
	AmountSat int64 `json:"amountSat"`
}

// NewClient creates a Phoenixd client. The password is used for HTTP
// Basic Auth with an empty username.
func NewClient(baseURL, password string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		password: password,
		http:     &http.Client{},
	}
}

// CreateInvoice creates a Lightning invoice via Phoenixd.
func (c *Client) CreateInvoice(ctx context.Context, amountSats int64, description string) (*Invoice, error) {
	form := url.Values{
		"amountSat":   {strconv.FormatInt(amountSats, 10)},
		"description": {description},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/createinvoice", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("phoenixd: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("", c.password)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("phoenixd: connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("phoenixd: createinvoice: HTTP %d", resp.StatusCode)
	}

	var inv Invoice
	if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
		return nil, fmt.Errorf("phoenixd: invalid response: %w", err)
	}

	return &inv, nil
}

// GetPayment retrieves an incoming payment by hash. Reserved for
// future strictVerify=true support.
func (c *Client) GetPayment(ctx context.Context, paymentHash string) (*Payment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/payments/incoming/"+paymentHash, nil)
	if err != nil {
		return nil, fmt.Errorf("phoenixd: build request: %w", err)
	}
	req.SetBasicAuth("", c.password)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("phoenixd: connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("phoenixd: getpayment: HTTP %d", resp.StatusCode)
	}

	var pmt Payment
	if err := json.NewDecoder(resp.Body).Decode(&pmt); err != nil {
		return nil, fmt.Errorf("phoenixd: invalid response: %w", err)
	}

	return &pmt, nil
}
