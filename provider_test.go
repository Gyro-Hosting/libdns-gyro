package gyro

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/libdns/libdns"
)

func TestProvider(t *testing.T) {
	var mu sync.Mutex
	vals := map[string]bool{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(401)
			w.Write([]byte(`{"error":{"code":"invalid_token","message":"Invalid API key."}}`))
			return
		}
		mu.Lock()
		defer mu.Unlock()
		name, value := r.URL.Query().Get("name"), r.URL.Query().Get("value")
		if r.Method == http.MethodPost {
			var b map[string]string
			json.NewDecoder(r.Body).Decode(&b)
			name, value = b["name"], b["value"]
		}
		if name != "_acme-challenge.example.com" {
			w.WriteHeader(404)
			w.Write([]byte(`{"error":{"code":"zone_not_found","message":"no zone"}}`))
			return
		}
		switch r.Method {
		case http.MethodPost:
			vals[value] = true
		case http.MethodDelete:
			delete(vals, value)
		}
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"ready": vals[value]}})
	}))
	defer srv.Close()
	p := &Provider{APIToken: "tok", APIBase: srv.URL + "/"}
	ctx := context.Background()
	recs := []libdns.Record{libdns.TXT{Name: "_acme-challenge", Text: "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"}}
	if _, err := p.AppendRecords(ctx, "example.com.", recs); err != nil {
		t.Fatal(err)
	}
	if !vals["abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"] {
		t.Fatal("value not added")
	}
	if _, err := p.DeleteRecords(ctx, "example.com.", recs); err != nil {
		t.Fatal(err)
	}
	if len(vals) != 0 {
		t.Fatal("value not removed")
	}
	if _, err := p.AppendRecords(ctx, "example.com.", []libdns.Record{libdns.TXT{Name: "www", Text: "x"}}); err == nil {
		t.Fatal("non _acme-challenge name accepted")
	}
	if _, err := p.AppendRecords(ctx, "example.com.", []libdns.Record{libdns.Address{Name: "_acme-challenge"}}); err == nil {
		t.Fatal("non TXT accepted")
	}
	bad := &Provider{APIToken: "nope", APIBase: srv.URL}
	if _, err := bad.AppendRecords(ctx, "example.com.", recs); err == nil {
		t.Fatal("bad token accepted")
	}
}
