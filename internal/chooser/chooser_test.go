package chooser

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bryawnie/daily-congressman/internal/beacon"
	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/source"
)

func members(chamber congress.Chamber, n int) []congress.Member {
	out := make([]congress.Member, n)
	for i := range out {
		out[i] = congress.Member{Chamber: chamber, ExternalID: 1000 + i, Name: "M"}
	}
	return out
}

func TestPickSmallValue(t *testing.T) {
	ms := []congress.Member{
		{Chamber: congress.Deputies, ExternalID: 3},
		{Chamber: congress.Deputies, ExternalID: 1},
		{Chamber: congress.Deputies, ExternalID: 2},
	}
	got, err := Pick("0A", ms) // 10 mod 3 = 1 -> second in sorted order
	if err != nil {
		t.Fatal(err)
	}
	if got.ExternalID != 2 {
		t.Fatalf("got %d, want 2", got.ExternalID)
	}
}

// The expected indexes were computed independently with Python:
// int(value, 16) % N, so the algorithm can be reproduced outside Go.
func TestPickLargeValueMatchesPython(t *testing.T) {
	const randOut = "A3F90123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEFFFFFFFFFFFFFFFFF"

	ms := members(congress.Deputies, 155)
	got, err := Pick(randOut, ms)
	if err != nil {
		t.Fatal(err)
	}
	if got.ExternalID != 1000+127 {
		t.Fatalf("N=155: got %d, want %d", got.ExternalID, 1000+127)
	}

	// 155 deputies + 50 senators; deputies sort before senators.
	all := append(members(congress.Senate, 50), members(congress.Deputies, 155)...)
	got, err = Pick(randOut, all)
	if err != nil {
		t.Fatal(err)
	}
	if got.Chamber != congress.Deputies || got.ExternalID != 1000+17 {
		t.Fatalf("N=205: got %+v, want deputies/%d", got, 1000+17)
	}
}

func TestPickDoesNotDependOnInputOrder(t *testing.T) {
	a := members(congress.Deputies, 30)
	b := append([]congress.Member(nil), a...)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	for _, r := range []string{"1", "FF", "DEADBEEF", "0123456789ABCDEF"} {
		x, _ := Pick(r, a)
		y, _ := Pick(r, b)
		if x != y {
			t.Fatalf("order changed the result for %s: %+v vs %+v", r, x, y)
		}
	}
}

func TestPickDoesNotMutateInput(t *testing.T) {
	ms := []congress.Member{{Chamber: congress.Senate, ExternalID: 9}, {Chamber: congress.Deputies, ExternalID: 1}}
	if _, err := Pick("1", ms); err != nil {
		t.Fatal(err)
	}
	if ms[0].Chamber != congress.Senate {
		t.Fatal("Pick reordered the caller's slice")
	}
}

func TestPickErrors(t *testing.T) {
	if _, err := Pick("AB", nil); err == nil {
		t.Fatal("expected error for no members")
	}
	if _, err := Pick("xyz", members(congress.Deputies, 3)); err == nil {
		t.Fatal("expected error for non-hex value")
	}
	dup := []congress.Member{{Chamber: congress.Deputies, ExternalID: 1}, {Chamber: congress.Deputies, ExternalID: 1}}
	if _, err := Pick("1", dup); err == nil {
		t.Fatal("expected error for duplicated member")
	}
}

func TestDayStartHandlesDaylightSaving(t *testing.T) {
	summer, err := DayStart("2026-01-15") // Chile on UTC-3
	if err != nil {
		t.Fatal(err)
	}
	if got := summer.UTC().Format(time.RFC3339); got != "2026-01-15T03:00:00Z" {
		t.Fatalf("summer = %s", got)
	}
	winter, _ := DayStart("2026-07-15") // Chile on UTC-4
	if got := winter.UTC().Format(time.RFC3339); got != "2026-07-15T04:00:00Z" {
		t.Fatalf("winter = %s", got)
	}
	if _, err := DayStart("15-01-2026"); err == nil {
		t.Fatal("expected error for bad format")
	}
}

func TestTodayUsesChileTime(t *testing.T) {
	// 02:00 UTC on Jan 16 is still Jan 15 at night in Chile (UTC-3).
	now := time.Date(2026, 1, 16, 2, 0, 0, 0, time.UTC)
	if got := Today(now); got != "2026-01-15" {
		t.Fatalf("Today = %s", got)
	}
}

type fakePulses struct {
	pulse beacon.Pulse
	err   error
	asked time.Time
}

func (f *fakePulses) PulseAt(_ context.Context, t time.Time) (beacon.Pulse, error) {
	f.asked = t
	return f.pulse, f.err
}

type fakeSource struct {
	congress.Unsupported
	chamber congress.Chamber
	members []congress.Member
}

func (f fakeSource) Chamber() congress.Chamber { return f.chamber }
func (f fakeSource) CurrentMembers(context.Context) ([]congress.Member, error) {
	if f.members == nil {
		return nil, congress.ErrNotSupported
	}
	return f.members, nil
}

func TestChooseUsesMidnightPulseAndAllActiveChambers(t *testing.T) {
	pulses := &fakePulses{pulse: beacon.Pulse{ChainID: "1", PulseID: "42", RandOut: "05"}}
	reg := source.NewRegistry(
		fakeSource{chamber: congress.Deputies, members: members(congress.Deputies, 3)},
		fakeSource{chamber: congress.Senate, members: members(congress.Senate, 2)},
	)
	res, err := (&Chooser{Pulses: pulses, Sources: reg}).Choose(context.Background(), "2026-01-15")
	if err != nil {
		t.Fatal(err)
	}
	if got := pulses.asked.UTC().Format(time.RFC3339); got != "2026-01-15T03:00:00Z" {
		t.Fatalf("pulse asked for %s", got)
	}
	// 5 members sorted: d1000 d1001 d1002 s1000 s1001; 5 mod 5 = 0 -> deputies/1000
	if res.Member.Chamber != congress.Deputies || res.Member.ExternalID != 1000 {
		t.Fatalf("unexpected member %+v", res.Member)
	}
	if res.Pulse.PulseID != "42" || res.Day != "2026-01-15" {
		t.Fatalf("unexpected result %+v", res)
	}
}

func TestChooseFailsIfAChamberCannotListMembers(t *testing.T) {
	reg := source.NewRegistry(
		fakeSource{chamber: congress.Deputies, members: members(congress.Deputies, 3)},
		fakeSource{chamber: congress.Senate}, // not implemented yet
	)
	_, err := (&Chooser{Pulses: &fakePulses{pulse: beacon.Pulse{RandOut: "1"}}, Sources: reg}).
		Choose(context.Background(), "2026-01-15")
	if !errors.Is(err, congress.ErrNotSupported) || !strings.Contains(err.Error(), "senate") {
		t.Fatalf("err = %v", err)
	}
}

func TestChoosePropagatesBeaconError(t *testing.T) {
	boom := errors.New("beacon down")
	reg := source.NewRegistry(fakeSource{chamber: congress.Deputies, members: members(congress.Deputies, 3)})
	_, err := (&Chooser{Pulses: &fakePulses{err: boom}, Sources: reg}).Choose(context.Background(), "2026-01-15")
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}
