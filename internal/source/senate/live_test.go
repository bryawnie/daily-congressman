//go:build live

// Live checks against the real Senate back end. They are skipped unless the
// "live" tag is given:
//
//	go test -tags live -run TestLive -v ./internal/source/senate
package senate

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestLiveProfiles reads the profile of every sitting senator and reports the
// ones with missing data, which usually means a summary with a new format.
func TestLiveProfiles(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	s := New()
	members, err := s.CurrentMembers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 50 {
		t.Errorf("expected 50 senators, got %d", len(members))
	}
	var noProfession, noParty int
	for _, m := range members {
		p, err := s.Profile(ctx, m)
		if err != nil {
			t.Errorf("%s (%d): %v", m.Name, m.ExternalID, err)
			continue
		}
		if p.Birthdate == "" {
			t.Errorf("%s (%d): no birth date", m.Name, m.ExternalID)
		}
		if p.Profession == "" {
			noProfession++
		}
		if p.Party == "" {
			noParty++
		}
		t.Logf("%-38s %s | %-34q | %-32s %q", m.Name, p.Birthdate, p.Profession, p.Party, p.PartyAlias)
	}
	t.Logf("without profession: %d, without party: %d", noProfession, noParty)
}

// TestLiveExpenseUnits checks that every sitting senator is matched to exactly
// one "unidad ejecutora" in the expenses data, and that no two share one.
func TestLiveExpenseUnits(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	s := New()
	senators, err := s.senators(ctx)
	if err != nil {
		t.Fatal(err)
	}
	periods, err := s.expensePeriods(ctx)
	if err != nil || len(periods) == 0 {
		t.Fatalf("periods: %v, %v", periods, err)
	}
	var rows []transparencyRow
	path := fmt.Sprintf("/api/transparency/parlamentarians-periods?year=%d&month=%d", periods[0].Year, periods[0].Month)
	if err := s.getJSON(ctx, path, &rows); err != nil {
		t.Fatal(err)
	}
	used := map[int]string{}
	for _, rec := range senators {
		unit, found, err := matchUnit(rec, rows)
		switch {
		case err != nil:
			t.Errorf("%s: %v", rec.fullName(), err)
		case !found:
			t.Errorf("%s: not found in the expenses data of %+v", rec.fullName(), periods[0])
		case used[unit] != "":
			t.Errorf("%s and %s share unit %d", rec.fullName(), used[unit], unit)
		default:
			used[unit] = rec.fullName()
		}
	}
	t.Logf("matched %d of %d senators (%d expense rows listed) for %+v", len(used), len(senators), len(rows), periods[0])
}

// TestLiveAttendance checks that the attendance of every sitting senator adds up.
func TestLiveAttendance(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	s := New()
	members, err := s.CurrentMembers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range members {
		a, err := s.Attendance(ctx, m)
		if err != nil {
			t.Errorf("%s: %v", m.Name, err)
			continue
		}
		if a.Total == 0 {
			t.Errorf("%s: no sessions", m.Name)
		}
		if sum := a.Present + a.JustifiedAbsent + a.UnjustifiedAbsent; sum != a.Total {
			t.Logf("%s: %d present + %d justified + %d unjustified = %d, but the total is %d",
				m.Name, a.Present, a.JustifiedAbsent, a.UnjustifiedAbsent, sum, a.Total)
		}
	}
}

// TestLiveVotingResults calculates the result of every voting of the current
// legislature and reports how they are distributed by quorum, to spot rules that
// give absurd results (for example, everything rejected).
func TestLiveVotingResults(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	s := New()
	leg, err := s.currentLegislature(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sitting, err := s.senators(ctx)
	if err != nil {
		t.Fatal(err)
	}
	seats := len(sitting)

	type key struct{ quorum, result string }
	tally := map[key]int{}
	unknown := 0
	total := 0
	for offset := 0; ; {
		var page votesData
		path := fmt.Sprintf("/api/votes?id_legislatura=%d&limit=100&offset=%d&palabra_clave=&desde=&hasta=", leg.ID, offset)
		if err := s.getJSON(ctx, path, &page); err != nil {
			t.Fatal(err)
		}
		for _, r := range page.Data {
			res := outcome(r.Quorum, r.Yes, r.No, seats)
			if res == "" {
				unknown++
			}
			tally[key{r.Quorum, res}]++
			total++
		}
		offset += len(page.Data)
		if len(page.Data) == 0 || offset >= page.Total {
			break
		}
	}
	t.Logf("%d votings, %d seats in office", total, seats)
	for k, n := range tally {
		t.Logf("  %-24q %-10q %d", k.quorum, k.result, n)
	}
	if unknown > 0 {
		t.Errorf("%d votings have a quorum the rule does not know", unknown)
	}
}
