package client

import (
	"encoding/json"
	"testing"
)

// NPM releases before v2.12.0 encode database boolean columns as JSON
// numbers (0/1). Newer releases use real booleans. Bool must accept both.
func TestBoolUnmarshalJSON(t *testing.T) {
	cases := []struct {
		in      string
		want    Bool
		wantErr bool
	}{
		{`true`, true, false},
		{`false`, false, false},
		{`1`, true, false},
		{`0`, false, false},
		{`"1"`, true, false},
		{`"0"`, false, false},
		{`"true"`, true, false},
		{`"false"`, false, false},
		{`null`, false, false},
		{`2`, false, true},
		{`"yes"`, false, true},
	}

	for _, tc := range cases {
		var got Bool
		err := json.Unmarshal([]byte(tc.in), &got)
		if tc.wantErr {
			if err == nil {
				t.Errorf("Unmarshal(%s): expected error, got %v", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("Unmarshal(%s): unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Unmarshal(%s) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// Decoding a proxy host payload in the pre-v2.12.0 NPM format (0/1
// integers for booleans) must succeed.
func TestProxyHostDecodeIntBooleans(t *testing.T) {
	payload := `{
		"id": 7,
		"domain_names": ["app.example.com"],
		"forward_host": "192.168.1.10",
		"forward_port": 3000,
		"forward_scheme": "http",
		"certificate_id": 0,
		"ssl_forced": 0,
		"hsts_enabled": 0,
		"hsts_subdomains": 0,
		"http2_support": 1,
		"block_exploits": 1,
		"caching_enabled": 0,
		"allow_websocket_upgrade": 1,
		"enabled": 1,
		"meta": null
	}`

	var host ProxyHost
	if err := json.Unmarshal([]byte(payload), &host); err != nil {
		t.Fatalf("unmarshal proxy host: %v", err)
	}
	if !host.Enabled {
		t.Errorf("Enabled = false, want true")
	}
	if host.SSLForced {
		t.Errorf("SSLForced = true, want false")
	}
	if !host.AllowWebsocketUpgrade {
		t.Errorf("AllowWebsocketUpgrade = false, want true")
	}
}
