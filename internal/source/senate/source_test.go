package senate

import (
	"context"
	"errors"
	"testing"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

func TestSenateIsASkeleton(t *testing.T) {
	var s congress.Source = New()
	if s.Chamber() != congress.Senate {
		t.Fatalf("chamber = %s", s.Chamber())
	}
	if _, err := s.CurrentMembers(context.Background()); !errors.Is(err, congress.ErrNotSupported) {
		t.Fatalf("err = %v", err)
	}
}
