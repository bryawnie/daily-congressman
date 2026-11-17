// Package senate implements the Source interface for the Senate (skeleton).
package senate

import "github.com/bryawnie/daily-congressman/internal/congress"

// Source is the Senate data source. It embeds congress.Unsupported, so every
// method returns congress.ErrNotSupported until it is implemented here.
//
// TODO, in the order the data is needed:
//   - CurrentMembers: list of current senators (required to include them in the draw).
//   - Profile.
//   - Attendance.
//   - Votings.
//   - Activity.
//   - Expenses.
type Source struct {
	congress.Unsupported
}

// New returns the Senate source.
func New() *Source { return &Source{} }

// Chamber reports the chamber this source belongs to.
func (*Source) Chamber() congress.Chamber { return congress.Senate }
