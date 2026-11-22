// Package congress defines the domain types shared by all chambers.
package congress

import "time"

// Chamber identifies the legislative chamber a congressman belongs to.
type Chamber string

const (
	Deputies Chamber = "deputies"
	Senate   Chamber = "senate"
)

// Valid reports whether c is a known chamber.
func (c Chamber) Valid() bool {
	return c == Deputies || c == Senate
}

// ConstituencyKind says what kind of area a congressman represents.
type ConstituencyKind string

const (
	ConstituencyDistrict ConstituencyKind = "district"
	ConstituencyRegion   ConstituencyKind = "region"
)

// Member is the minimum needed to identify and choose a congressman.
type Member struct {
	Chamber    Chamber
	ExternalID int
	Name       string
}

// Congressman is the profile of a legislator. ID is the internal identifier
// assigned by the store; it is zero until the congressman is persisted.
type Congressman struct {
	ID         int64
	Chamber    Chamber
	ExternalID int

	FirstName      string
	FirstSurname   string
	SecondSurname  string
	ProfilePicture string
	Gender         string
	Birthdate      string // YYYY-MM-DD
	Profession     string

	Party      string
	PartyAlias string

	TwitterUsername   string
	InstagramUsername string

	ConstituencyKind   ConstituencyKind
	ConstituencyName   string
	ConstituencyDetail string

	LastUpdate time.Time
}

// Period is a parliamentary period served by a congressman.
type Period struct {
	From int
	To   int
}

// Attendance summarizes session attendance.
type Attendance struct {
	Total             int
	Present           int
	JustifiedAbsent   int
	UnjustifiedAbsent int
}

// Voting is one voting and the option chosen by the congressman.
type Voting struct {
	VotingID        int
	Date            string // YYYY-MM-DD
	BulletinNumber  string
	DocumentTitle   string
	ArticleText     string
	VotedOption     string
	TotalApproved   int
	TotalRejected   int
	TotalAbstention int
	// Quorum is the rule the voting needed (e.g. "Mayoría simple").
	Quorum string
	// Result is "Aprobado" or "Rechazado". A source that does not publish it may
	// calculate it from the quorum, or leave it empty when it cannot.
	Result string
}

// LawProject is a bill or motion presented by a congressman.
type LawProject struct {
	Bulletin string // e.g. "18442-07"; identifies the project in both chambers
	Name     string
	Type     string
	Status   string
	Date     string // YYYY-MM-DD
}

// OperationalExpense is the amount spent in one category in a month.
type OperationalExpense struct {
	Year        int
	Month       int
	ExpenseType string
	Amount      int
}

// OfficeExpense summarizes office spending in a month.
type OfficeExpense struct {
	Year          int
	Month         int
	OfficesNumber int
	TotalAmount   int
}

// StaffExpense summarizes support staff spending in a month.
type StaffExpense struct {
	Year        int
	Month       int
	HiredStaff  int
	TotalAmount int
}

// DayChoice is the congressman chosen for a given day.
type DayChoice struct {
	Date          string // YYYY-MM-DD
	CongressmanID int64
	ChainID       string
	PulseID       string
	RandOut       string
}
