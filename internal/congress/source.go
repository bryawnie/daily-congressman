package congress

import (
	"context"
	"errors"
)

// ErrNotSupported is returned by a Source for data it cannot provide yet.
var ErrNotSupported = errors.New("congress: not supported by this source")

// Expenses groups the three kinds of monthly expenses of a congressman.
type Expenses struct {
	Operational []OperationalExpense
	Offices     []OfficeExpense
	Staff       []StaffExpense
}

// Source is everything a chamber must be able to provide. A chamber may
// implement only part of it: embed Unsupported and override what is available.
type Source interface {
	Chamber() Chamber
	CurrentMembers(ctx context.Context) ([]Member, error)
	Profile(ctx context.Context, m Member) (Congressman, error)
	Attendance(ctx context.Context, m Member) (Attendance, error)
	Votings(ctx context.Context, m Member, limit int) ([]Voting, error)
	Activity(ctx context.Context, m Member) ([]LawProject, error)
	Expenses(ctx context.Context, m Member) (Expenses, error)
}

// Unsupported implements every Source method except Chamber by returning
// ErrNotSupported. Embed it in a chamber's source so it compiles from day one
// and is completed method by method.
type Unsupported struct{}

func (Unsupported) CurrentMembers(context.Context) ([]Member, error) {
	return nil, ErrNotSupported
}

func (Unsupported) Profile(context.Context, Member) (Congressman, error) {
	return Congressman{}, ErrNotSupported
}

func (Unsupported) Attendance(context.Context, Member) (Attendance, error) {
	return Attendance{}, ErrNotSupported
}

func (Unsupported) Votings(context.Context, Member, int) ([]Voting, error) {
	return nil, ErrNotSupported
}

func (Unsupported) Activity(context.Context, Member) ([]LawProject, error) {
	return nil, ErrNotSupported
}

func (Unsupported) Expenses(context.Context, Member) (Expenses, error) {
	return Expenses{}, ErrNotSupported
}
