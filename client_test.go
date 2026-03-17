package phoenixd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateInvoice_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/createinvoice", r.URL.Path)

		user, pass, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "", user)
		require.Equal(t, "test-password", pass)

		require.NoError(t, r.ParseForm())
		require.Equal(t, "100", r.FormValue("amountSat"))
		require.Equal(t, "L402", r.FormValue("description"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"paymentHash":"abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789","serialized":"lnbc1..."}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-password")
	inv, err := client.CreateInvoice(t.Context(), 100, "L402")
	require.NoError(t, err)
	require.Equal(t, "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789", inv.PaymentHash)
	require.Equal(t, "lnbc1...", inv.Serialized)
}

func TestCreateInvoice_PhoenixdDown(t *testing.T) {
	client := NewClient("http://127.0.0.1:19999", "test-password")
	_, err := client.CreateInvoice(t.Context(), 100, "L402")
	require.ErrorContains(t, err, "phoenixd: connect:")
}

func TestCreateInvoice_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-password")
	_, err := client.CreateInvoice(t.Context(), 100, "L402")
	require.ErrorContains(t, err, "HTTP 500")
}

func TestCreateInvoice_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-password")
	_, err := client.CreateInvoice(t.Context(), 100, "L402")
	require.ErrorContains(t, err, "phoenixd: invalid response:")
}

const validHash = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

func TestGetPayment_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/payments/incoming/"+validHash, r.URL.Path)

		user, pass, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "", user)
		require.Equal(t, "test-password", pass)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"isPaid":true,"amountSat":100}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-password")
	pmt, err := client.GetPayment(t.Context(), validHash)
	require.NoError(t, err)
	require.True(t, pmt.IsPaid)
	require.Equal(t, int64(100), pmt.AmountSat)
}

func TestGetPayment_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-password")
	_, err := client.GetPayment(t.Context(), validHash)
	require.ErrorContains(t, err, "HTTP 404")
}

func TestGetPayment_InvalidHash(t *testing.T) {
	client := NewClient("http://localhost:9740", "test-password")

	_, err := client.GetPayment(t.Context(), "too-short")
	require.ErrorContains(t, err, "invalid payment hash")

	_, err = client.GetPayment(t.Context(), "../../admin/shutdown")
	require.ErrorContains(t, err, "invalid payment hash")
}
