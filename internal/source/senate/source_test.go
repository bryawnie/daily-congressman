package senate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// Fixture with the structure of the real response of /api/hemicycle (names are
// made up). It keeps the quirks of the real data: padded names and a blank
// second surname.
const membersJSON = `{
  "data": {
    "hemiciclo": [[[1001, 1002]]],
    "parlamentarios": {
      "total": 3,
      "data": [
        {"ID_PARLAMENTARIO": 1001, "NOMBRE": "Ana María", "APELLIDO_PATERNO": "Pérez", "APELLIDO_MATERNO": "Soto", "CAMARA": "S", "PARTIDO": "P.S."},
        {"ID_PARLAMENTARIO": 1002, "NOMBRE": "Daniella ", "APELLIDO_PATERNO": "Rojas", "APELLIDO_MATERNO": "Mora", "CAMARA": "S", "PARTIDO": "R.N."},
        {"ID_PARLAMENTARIO": 1500, "NOMBRE": "Rojo", "APELLIDO_PATERNO": "Gil", "APELLIDO_MATERNO": "  ", "CAMARA": "S", "PARTIDO": "Independiente"}
      ]
    }
  },
  "status": "ok",
  "results": 1
}`

func serve(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
}

func TestSenateImplementsSource(t *testing.T) {
	var s congress.Source = New()
	if s.Chamber() != congress.Senate {
		t.Fatalf("chamber = %s", s.Chamber())
	}
}

func TestCurrentMembers(t *testing.T) {
	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		w.Write([]byte(membersJSON))
	}))
	defer srv.Close()

	got, err := New(WithAPIURL(srv.URL)).CurrentMembers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/hemicycle" || gotQuery != "vigentes=1&camara=S&limit=100" {
		t.Fatalf("request = %s?%s", gotPath, gotQuery)
	}
	want := []congress.Member{
		{Chamber: congress.Senate, ExternalID: 1001, Name: "Ana María Pérez Soto"},
		{Chamber: congress.Senate, ExternalID: 1002, Name: "Daniella Rojas Mora"},
		{Chamber: congress.Senate, ExternalID: 1500, Name: "Rojo Gil"},
	}
	if len(got) != len(want) {
		t.Fatalf("members = %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("member %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestCurrentMembersErrors(t *testing.T) {
	cases := map[string]string{
		"bad json":       `{`,
		"empty list":     `{"data":{"parlamentarios":{"total":0,"data":[]}},"status":"ok"}`,
		"bad status":     `{"data":{"parlamentarios":{"total":1,"data":[{"ID_PARLAMENTARIO":1,"NOMBRE":"A","CAMARA":"S"}]}},"status":"error"}`,
		"truncated list": `{"data":{"parlamentarios":{"total":50,"data":[{"ID_PARLAMENTARIO":1,"NOMBRE":"A","CAMARA":"S"}]}},"status":"ok"}`,
		"no id":          `{"data":{"parlamentarios":{"total":1,"data":[{"NOMBRE":"A","CAMARA":"S"}]}},"status":"ok"}`,
		"wrong chamber":  `{"data":{"parlamentarios":{"total":1,"data":[{"ID_PARLAMENTARIO":1,"NOMBRE":"A","CAMARA":"D"}]}},"status":"ok"}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			srv := serve(body)
			defer srv.Close()
			if _, err := New(WithAPIURL(srv.URL)).CurrentMembers(context.Background()); err == nil {
				t.Fatal("expected error")
			}
		})
	}

	t.Run("http error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer srv.Close()
		if _, err := New(WithAPIURL(srv.URL)).CurrentMembers(context.Background()); err == nil {
			t.Fatal("expected error")
		}
	})
}
