// Package chooser selects the congressman of the day.
//
// The draw is deterministic and easy to reproduce in any language:
//
//  1. Take the beacon pulse generated at 00:00 (America/Santiago) of the day.
//  2. Sort all members by (chamber, external id).
//  3. Read the pulse output value as a big hexadecimal integer R.
//  4. The chosen member is the one at index R mod N, where N is the number of members.
package chooser

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"
	_ "time/tzdata" // embed the time zone database so it works in minimal images

	"github.com/bryawnie/daily-congressman/internal/beacon"
	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/source"
)

// Location is the time zone whose midnight defines a day.
const Location = "America/Santiago"

// Pick chooses one member using the beacon output value (hexadecimal). The
// result does not depend on the order of members.
func Pick(randOut string, members []congress.Member) (congress.Member, error) {
	if len(members) == 0 {
		return congress.Member{}, errors.New("chooser: no members to choose from")
	}
	r, ok := new(big.Int).SetString(randOut, 16)
	if !ok {
		return congress.Member{}, fmt.Errorf("chooser: output value is not hexadecimal: %q", randOut)
	}

	sorted := append([]congress.Member(nil), members...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Chamber != sorted[j].Chamber {
			return sorted[i].Chamber < sorted[j].Chamber
		}
		return sorted[i].ExternalID < sorted[j].ExternalID
	})
	for i := 1; i < len(sorted); i++ {
		if sorted[i].Chamber == sorted[i-1].Chamber && sorted[i].ExternalID == sorted[i-1].ExternalID {
			return congress.Member{}, fmt.Errorf("chooser: duplicated member %s/%d", sorted[i].Chamber, sorted[i].ExternalID)
		}
	}

	idx := new(big.Int).Mod(r, big.NewInt(int64(len(sorted))))
	return sorted[idx.Int64()], nil
}

// DayStart returns midnight of the given day (YYYY-MM-DD) in Chile.
func DayStart(day string) (time.Time, error) {
	loc, err := time.LoadLocation(Location)
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.ParseInLocation("2006-01-02", day, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("chooser: invalid day %q (want YYYY-MM-DD): %w", day, err)
	}
	return t, nil
}

// Today returns the current day (YYYY-MM-DD) in Chile.
func Today(now time.Time) string {
	loc, err := time.LoadLocation(Location)
	if err != nil {
		return now.UTC().Format("2006-01-02")
	}
	return now.In(loc).Format("2006-01-02")
}

// PulseSource gets the beacon pulse for a moment.
type PulseSource interface {
	PulseAt(ctx context.Context, t time.Time) (beacon.Pulse, error)
}

// Chooser draws the congressman of a day from the active chambers.
type Chooser struct {
	Pulses  PulseSource
	Sources *source.Registry
}

// Result is the outcome of a draw.
type Result struct {
	Day    string
	Member congress.Member
	Pulse  beacon.Pulse
}

// Choose draws the congressman for a day (YYYY-MM-DD). It fails if any active
// chamber cannot list its members, so a day is never drawn from a partial list.
func (c *Chooser) Choose(ctx context.Context, day string) (Result, error) {
	start, err := DayStart(day)
	if err != nil {
		return Result{}, err
	}

	var members []congress.Member
	for _, ch := range c.Sources.Chambers() {
		src, err := c.Sources.Get(ch)
		if err != nil {
			return Result{}, err
		}
		ms, err := src.CurrentMembers(ctx)
		if err != nil {
			return Result{}, fmt.Errorf("chooser: members of %s: %w", ch, err)
		}
		members = append(members, ms...)
	}

	pulse, err := c.Pulses.PulseAt(ctx, start)
	if err != nil {
		return Result{}, err
	}
	m, err := Pick(pulse.RandOut, members)
	if err != nil {
		return Result{}, err
	}
	return Result{Day: day, Member: m, Pulse: pulse}, nil
}
