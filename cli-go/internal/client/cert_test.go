package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCertificateIssuanceOutlivesTheDefaultTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(300 * time.Millisecond):
		case <-r.Context().Done():
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7}`))
	}))
	defer srv.Close()
	c := NewClientWithToken(srv.URL, false)
	c.HTTPClient.Timeout = 50 * time.Millisecond

	cert, err := c.CreateCertificate(context.Background(), map[string]interface{}{"provider": "letsencrypt"})
	if err != nil || cert.ID != 7 {
		t.Fatalf("create: %v %+v", err, cert)
	}
	if err := c.RenewCertificate(context.Background(), 7); err != nil {
		t.Fatalf("renew: %v", err)
	}
	if _, err := c.GetCertificate(context.Background(), 7); err == nil {
		t.Fatal("other requests must keep the client timeout")
	}
}

func TestTestHTTPUsesTheLongerTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(300 * time.Millisecond):
		case <-r.Context().Done():
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"example.com":"ok"}`))
	}))
	defer srv.Close()
	c := NewClientWithToken(srv.URL, false)
	c.HTTPClient.Timeout = 50 * time.Millisecond

	got, err := c.TestHTTP(context.Background(), map[string]interface{}{"domains": []string{"example.com"}})
	if err != nil || got["example.com"] != "ok" {
		t.Fatalf("test-http: %v %v", err, got)
	}
}
