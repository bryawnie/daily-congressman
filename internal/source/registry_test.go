package source

import (
	"testing"

	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/source/deputies"
	"github.com/bryawnie/daily-congressman/internal/source/senate"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry(senate.New(), deputies.New())
	chambers := r.Chambers()
	if len(chambers) != 2 || chambers[0] != congress.Deputies || chambers[1] != congress.Senate {
		t.Fatalf("chambers = %v", chambers)
	}

	only, err := r.Only(congress.Deputies)
	if err != nil {
		t.Fatal(err)
	}
	if got := only.Chambers(); len(got) != 1 || got[0] != congress.Deputies {
		t.Fatalf("only = %v", got)
	}
	if _, err := only.Get(congress.Senate); err == nil {
		t.Fatal("senate should not be in the restricted registry")
	}
	if _, err := r.Only("house"); err == nil {
		t.Fatal("expected error for unknown chamber")
	}
}
