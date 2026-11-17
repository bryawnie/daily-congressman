package beacon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPulseAt(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Path + "?" + r.URL.RawQuery
		w.Write([]byte(`{"pulse":{"outputValue":"ABCDEF","chainIndex":1,"pulseIndex":9876}}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	p, err := c.PulseAt(context.Background(), time.UnixMilli(1763262000000))
	if err != nil {
		t.Fatal(err)
	}
	if gotQuery != "/pulse?timeGE=1763262000000" {
		t.Fatalf("request = %s", gotQuery)
	}
	if p.RandOut != "ABCDEF" || p.ChainID != "1" || p.PulseID != "9876" {
		t.Fatalf("pulse = %+v", p)
	}
	if got, want := p.URI(srv.URL), srv.URL+"/chain/1/pulse/9876"; got != want {
		t.Fatalf("URI = %s, want %s", got, want)
	}
}

func TestPulseAtErrors(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"status":  func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadGateway) },
		"badjson": func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("not json")) },
		"nooutput": func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(`{"pulse":{"chainIndex":1,"pulseIndex":2}}`))
		},
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(h)
			defer srv.Close()
			if _, err := NewClient(WithBaseURL(srv.URL)).PulseAt(context.Background(), time.Now()); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
