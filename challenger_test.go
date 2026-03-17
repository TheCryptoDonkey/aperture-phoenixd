package phoenixd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

const testHash = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

func TestNewChallenge_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Invoice{
			PaymentHash: testHash,
			Serialized:  "lnbc1...",
		})
	}))
	defer srv.Close()

	challenger := &PhoenixdChallenger{client: NewClient(srv.URL, "pw")}
	bolt11, hash, err := challenger.NewChallenge(100)
	require.NoError(t, err)
	require.Equal(t, "lnbc1...", bolt11)

	// Verify the 32-byte hash matches the hex constant.
	var expected [32]byte
	b := make([]byte, 32)
	_, _ = (&[64]byte{})[0:0], b // suppress unused
	for i := 0; i < 32; i++ {
		var v byte
		hi := testHash[i*2]
		lo := testHash[i*2+1]
		for _, c := range []byte{hi} {
			if c >= '0' && c <= '9' {
				v = (c - '0') << 4
			} else {
				v = (c - 'a' + 10) << 4
			}
		}
		for _, c := range []byte{lo} {
			if c >= '0' && c <= '9' {
				v |= c - '0'
			} else {
				v |= c - 'a' + 10
			}
		}
		expected[i] = v
	}
	require.Equal(t, expected, hash)
}

func TestNewChallenge_EmptyHash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Invoice{
			PaymentHash: "",
			Serialized:  "lnbc1...",
		})
	}))
	defer srv.Close()

	challenger := &PhoenixdChallenger{client: NewClient(srv.URL, "pw")}
	_, _, err := challenger.NewChallenge(100)
	require.ErrorContains(t, err, "empty payment hash")
}

func TestNewChallenge_InvalidHex(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Invoice{
			PaymentHash: "not-hex",
			Serialized:  "lnbc1...",
		})
	}))
	defer srv.Close()

	challenger := &PhoenixdChallenger{client: NewClient(srv.URL, "pw")}
	_, _, err := challenger.NewChallenge(100)
	require.ErrorContains(t, err, "invalid payment hash")
}

func TestNewChallenge_WrongHashLength(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// "abcdef" = 3 bytes, not 32
		_ = json.NewEncoder(w).Encode(Invoice{
			PaymentHash: "abcdef",
			Serialized:  "lnbc1...",
		})
	}))
	defer srv.Close()

	challenger := &PhoenixdChallenger{client: NewClient(srv.URL, "pw")}
	_, _, err := challenger.NewChallenge(100)
	require.ErrorContains(t, err, "wrong length")
}

func TestNewChallenge_ZeroPrice(t *testing.T) {
	challenger := NewChallenger("http://localhost:9740", "pw")
	_, _, err := challenger.NewChallenge(0)
	require.ErrorContains(t, err, "price must be positive")
}

func TestNewChallenge_NegativePrice(t *testing.T) {
	challenger := NewChallenger("http://localhost:9740", "pw")
	_, _, err := challenger.NewChallenge(-1)
	require.ErrorContains(t, err, "price must be positive")
}

func TestVerifyInvoiceStatus_NoOp(t *testing.T) {
	challenger := NewChallenger("http://localhost:9740", "pw")
	err := challenger.VerifyInvoiceStatus([32]byte{}, 100, "test")
	require.NoError(t, err)
}

func TestStop_NoOp(t *testing.T) {
	challenger := NewChallenger("http://localhost:9740", "pw")
	require.NotPanics(t, func() {
		challenger.Stop()
	})
}
