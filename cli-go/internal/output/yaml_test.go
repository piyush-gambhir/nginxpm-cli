package output

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestYAMLMatchesJSONShape(t *testing.T) {
	type host struct {
		DomainNames []string        `json:"domain_names"`
		Items       json.RawMessage `json:"items"`
	}
	var buf bytes.Buffer
	data := []host{{DomainNames: []string{"app.example.com"}, Items: json.RawMessage(`[{"username":"admin"}]`)}}
	if err := (&YAMLFormatter{}).Format(&buf, data, nil); err != nil {
		t.Fatal(err)
	}
	want := "- domain_names:\n    - app.example.com\n  items:\n    - username: admin\n"
	if buf.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
}
