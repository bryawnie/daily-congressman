package senate

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// legislature is a legislative year of the Senate (e.g. number 374, from
// 11/03/2026 to 10/03/2027). The back end identifies it with its own ID.
type legislature struct {
	ID     int    `json:"ID_LEGISLATURA"`
	Number int    `json:"NUMERO"`
	Start  string `json:"INICIO"`
	End    string `json:"TERMINO"`
}

func (l legislature) startTime() (time.Time, error) { return time.Parse("02/01/2006", l.Start) }

// endTime is the end of the last day of the legislature.
func (l legislature) endTime() (time.Time, error) {
	t, err := time.Parse("02/01/2006", l.End)
	if err != nil {
		return time.Time{}, err
	}
	return t.Add(24*time.Hour - time.Nanosecond), nil
}

// legislatures lists every legislature, newest first.
func (s *Source) legislatures(ctx context.Context) ([]legislature, error) {
	var list []legislature
	if err := s.getJSON(ctx, "/api/legislatures?limit=100", &list); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("senate: no legislatures found")
	}
	return list, nil
}

// currentLegislature returns the legislature that contains the current date.
func (s *Source) currentLegislature(ctx context.Context) (legislature, error) {
	list, err := s.legislatures(ctx)
	if err != nil {
		return legislature{}, err
	}
	now := s.now().UTC()
	for _, l := range list {
		start, err1 := l.startTime()
		end, err2 := l.endTime()
		if err1 != nil || err2 != nil {
			return legislature{}, fmt.Errorf("senate: legislature %d has bad dates %q - %q", l.Number, l.Start, l.End)
		}
		if !now.Before(start) && !now.After(end) {
			return l, nil
		}
	}
	return legislature{}, fmt.Errorf("senate: no legislature contains %s", now.Format("2006-01-02"))
}

// slashDate converts 02/01/2006 into 2006-01-02. An empty input stays empty.
func slashDate(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	t, err := time.Parse("02/01/2006", s)
	if err != nil {
		return "", fmt.Errorf("senate: bad date %q: %w", s, err)
	}
	return t.Format("2006-01-02"), nil
}
