package senate

import (
	"context"
	"fmt"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

type attendanceData struct {
	Total       int `json:"total"`
	Present     int `json:"asiste"`
	Justified   int `json:"justificacion"`
	Unjustified int `json:"sin_justificacion"`
}

// Attendance returns the attendance to the floor sessions of the current
// legislature. The back end already gives the totals, so the detail of each
// session (limit=1) is not requested.
func (s *Source) Attendance(ctx context.Context, m congress.Member) (congress.Attendance, error) {
	if err := checkMember(m); err != nil {
		return congress.Attendance{}, err
	}
	leg, err := s.currentLegislature(ctx)
	if err != nil {
		return congress.Attendance{}, err
	}
	var d attendanceData
	path := fmt.Sprintf("/api/parliamentarians_attendance_room?id_parlamentario=%d&id_legislatura=%d&limit=1&offset=0",
		m.ExternalID, leg.ID)
	if err := s.getJSON(ctx, path, &d); err != nil {
		return congress.Attendance{}, err
	}
	return congress.Attendance{
		Total:             d.Total,
		Present:           d.Present,
		JustifiedAbsent:   d.Justified,
		UnjustifiedAbsent: d.Unjustified,
	}, nil
}
