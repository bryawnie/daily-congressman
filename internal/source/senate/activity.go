package senate

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// motionsPageSize is how many motions are requested per call.
const motionsPageSize = 100

type motionRecord struct {
	Bulletin string `json:"BOLETIN"`
	Summary  string `json:"SUMA"`
	Date     string `json:"FECHA"`
	Status   string `json:"ESTADO"`
}

type motionsData struct {
	Total int            `json:"total"`
	Data  []motionRecord `json:"data"`
}

// periodStartYear is the first year of the senator's current period (the
// "DESDE" of the period flagged as current).
func (r senatorRecord) periodStartYear() (int, error) {
	for _, p := range r.Periods {
		if p.Current == 1 {
			y, err := strconv.Atoi(p.From)
			if err != nil {
				return 0, fmt.Errorf("senate: senator %d has a bad period start %q", r.ID, p.From)
			}
			return y, nil
		}
	}
	return 0, fmt.Errorf("senate: senator %d has no current period", r.ID)
}

// Activity returns the motions the senator has presented since the start of the
// current period, with their status. Motions are listed per legislature, so it
// reads every legislature from that start up to the current one.
func (s *Source) Activity(ctx context.Context, m congress.Member) ([]congress.LawProject, error) {
	if err := checkMember(m); err != nil {
		return nil, err
	}
	rec, err := s.senator(ctx, m.ExternalID)
	if err != nil {
		return nil, err
	}
	fromYear, err := rec.periodStartYear()
	if err != nil {
		return nil, err
	}
	all, err := s.legislatures(ctx)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()

	var projects []congress.LawProject
	for _, leg := range all {
		start, err := leg.startTime()
		if err != nil {
			return nil, fmt.Errorf("senate: legislature %d has a bad start %q", leg.Number, leg.Start)
		}
		if start.Year() < fromYear || start.After(now) {
			continue
		}
		motions, err := s.motions(ctx, m.ExternalID, leg.ID)
		if err != nil {
			return nil, err
		}
		for _, r := range motions {
			date, err := slashDate(r.Date)
			if err != nil {
				return nil, err
			}
			projects = append(projects, congress.LawProject{
				Bulletin: r.Bulletin,
				Name:     cleanText(r.Summary),
				Type:     "Moción",
				Status:   r.Status,
				Date:     date,
			})
		}
	}
	return projects, nil
}

// motions reads every page of the motions of a senator in a legislature.
func (s *Source) motions(ctx context.Context, senatorID, legislatureID int) ([]motionRecord, error) {
	var out []motionRecord
	// The offset advances by the rows received, not by the page size asked for,
	// so a server that caps the page size does not make rows get skipped.
	for offset := 0; ; offset = len(out) {
		var page motionsData
		path := fmt.Sprintf("/api/parliamentarians/motions?id_parlamentario=%d&id_legislatura=%d&limit=%d&offset=%d",
			senatorID, legislatureID, motionsPageSize, offset)
		if err := s.getJSON(ctx, path, &page); err != nil {
			return nil, err
		}
		for _, r := range page.Data {
			if r.Bulletin == "" {
				return nil, fmt.Errorf("senate: motion without bulletin for senator %d", senatorID)
			}
		}
		out = append(out, page.Data...)
		if len(page.Data) == 0 || len(out) >= page.Total {
			return out, nil
		}
	}
}
