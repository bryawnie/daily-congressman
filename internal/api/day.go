package api

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/store"
)

// Names of the jobs whose completion tells "fetched" from "not fetched yet".
// They match internal/jobs; they are repeated here so the API does not depend
// on the scheduler.
const (
	jobRefresh  = "refresh"
	jobExpenses = "expenses"
)

// Sections of a day that may be pending.
const (
	SectionAttendance = "attendance"
	SectionVotings    = "votings"
	SectionActivity   = "activity"
	SectionExpenses   = "expenses"
)

// expenseMonths is how many months of expenses a day shows.
const expenseMonths = 6

var monthNames = [...]string{
	"enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

// Day is the congressman of a day with everything the webapp shows.
//
// A section that has not been fetched yet is null (or empty) and listed in
// Pending; a section that was fetched and has no records is an empty value
// that is not in Pending. That is how "no data yet" differs from "zero".
type Day struct {
	Date        string          `json:"date"`
	Beacon      Beacon          `json:"beacon"`
	Congressman Profile         `json:"congressman"`
	Attendance  *Attendance     `json:"attendance"`
	Activity    *Activity       `json:"activity"`
	Votings     []Voting        `json:"votings"`
	Expenses    []MonthExpenses `json:"expenses"`
	Pending     []string        `json:"pending"`
}

// Beacon identifies the pulse the draw came from, so anyone can verify it.
type Beacon struct {
	ChainID string `json:"chainId"`
	PulseID string `json:"pulseId"`
	RandOut string `json:"randOut"`
}

// Profile is the congressman's identity and constituency.
type Profile struct {
	ID                int64        `json:"id"`
	Chamber           string       `json:"chamber"`
	ExternalID        int          `json:"externalId"`
	Name              string       `json:"name"`
	FirstSurname      string       `json:"firstSurname"`
	SecondSurname     string       `json:"secondSurname"`
	Picture           string       `json:"picture"`
	Gender            string       `json:"gender"`
	Birthdate         string       `json:"birthdate"`
	Profession        string       `json:"profession"`
	Party             string       `json:"party"`
	PartyAlias        string       `json:"partyAlias"`
	TwitterUsername   string       `json:"twitterUsername"`
	InstagramUsername string       `json:"instagramUsername"`
	Constituency      Constituency `json:"constituency"`
	Periods           []Period     `json:"periods"`
}

// Constituency is the district or region represented.
type Constituency struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Detail string `json:"detail"`
}

// Period is a parliamentary period.
type Period struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// Attendance summarizes session attendance.
type Attendance struct {
	Attended          int `json:"attended"`
	JustifiedAbsent   int `json:"justifiedAbsent"`
	UnjustifiedAbsent int `json:"unjustifiedAbsent"`
	Total             int `json:"total"`
}

// Activity counts the congressman's bills and motions.
type Activity struct {
	All      int            `json:"all"`
	ByStatus map[string]int `json:"byStatus"`
}

// Voting is one voting and what the congressman voted.
type Voting struct {
	ID              int    `json:"votingId"`
	Date            string `json:"votingDate"`
	BulletinNumber  string `json:"bulletinNumber"`
	DocumentTitle   string `json:"documentTitle"`
	ArticleText     string `json:"articleText"`
	Voted           string `json:"voted"`
	TotalApproved   int    `json:"totalApproved"`
	TotalRejected   int    `json:"totalRejected"`
	TotalAbstention int    `json:"totalAbstention"`
	Quorum          string `json:"quorum"`
	Result          string `json:"result"`
}

// MonthExpenses is the spending of one month, compared with the chamber.
type MonthExpenses struct {
	Year   int             `json:"year"`
	Month  string          `json:"month"`
	Total  int             `json:"total"`
	Detail []ExpenseDetail `json:"detail"`
}

// ExpenseDetail is one expense type of a month. Operational spending lists its
// categories in Expenses.
type ExpenseDetail struct {
	Type         string            `json:"type"`
	Amount       *int              `json:"amount"`
	SupportStaff *int              `json:"supportStaff,omitempty"`
	Stats        Comparison        `json:"stats"`
	Expenses     []ExpenseCategory `json:"expenses,omitempty"`
}

// ExpenseCategory is one operational expense category of a month.
type ExpenseCategory struct {
	Subtype string     `json:"subtype"`
	Amount  int        `json:"amount"`
	Stats   Comparison `json:"stats"`
}

// Comparison puts an amount next to the other congressmen: Global covers the
// whole chamber, Local only the same constituency (same district for deputies,
// same region for senators). Local is null when the congressman's constituency
// is not known.
type Comparison struct {
	Global *Stats `json:"global"`
	Local  *Stats `json:"local"`
}

// Stats summarizes what a group of congressmen spent. Count is how many have a
// record that month (the congressman included); Ranking is 1 for whoever spent
// the most, out of Count, and is only given for the month totals, never for
// categories. Avg, Min and Max are null when nobody has a record.
type Stats struct {
	// Type says what the group is made of: "all" (every congressman of the
	// chamber), "district" (deputies of the same district) or "region" (senators
	// of the same region).
	Type    string `json:"_type"`
	Count   int    `json:"count"`
	Ranking *int   `json:"ranking,omitempty"`
	Avg     *int   `json:"avg"`
	Min     *int   `json:"min"`
	Max     *int   `json:"max"`
}

// typeAll is the _type of the global stats: the whole chamber.
const typeAll = "all"

// Operational and staff are the expense types of a month.
const (
	typeOperational = "Operacional"
	typeStaff       = "Personal de Apoyo"
)

// BuildDay assembles the day for a stored choice.
func BuildDay(ctx context.Context, st *store.Store, choice congress.DayChoice) (Day, error) {
	c, err := st.GetCongressman(ctx, choice.CongressmanID)
	if err != nil {
		return Day{}, fmt.Errorf("congressman %d: %w", choice.CongressmanID, err)
	}
	d := Day{
		Date:    choice.Date,
		Beacon:  Beacon{ChainID: choice.ChainID, PulseID: choice.PulseID, RandOut: choice.RandOut},
		Pending: []string{},
	}
	d.Congressman = Profile{
		ID: c.ID, Chamber: string(c.Chamber), ExternalID: c.ExternalID,
		Name: c.FirstName, FirstSurname: c.FirstSurname, SecondSurname: c.SecondSurname,
		Picture: c.ProfilePicture, Gender: c.Gender, Birthdate: c.Birthdate,
		Profession: c.Profession, Party: c.Party, PartyAlias: c.PartyAlias,
		TwitterUsername: c.TwitterUsername, InstagramUsername: c.InstagramUsername,
		Constituency: Constituency{Kind: string(c.ConstituencyKind), Name: c.ConstituencyName, Detail: c.ConstituencyDetail},
		Periods:      []Period{},
	}
	periods, err := st.ListPeriods(ctx, c.ID)
	if err != nil {
		return Day{}, err
	}
	for _, p := range periods {
		d.Congressman.Periods = append(d.Congressman.Periods, Period{From: p.From, To: p.To})
	}

	key := fmt.Sprintf("%s:%d", c.Chamber, c.ExternalID)
	refreshed, err := st.EverDone(ctx, jobRefresh, key)
	if err != nil {
		return Day{}, err
	}
	expensesSynced, err := st.EverDone(ctx, jobExpenses, key)
	if err != nil {
		return Day{}, err
	}

	// Attendance: present when stored.
	switch a, err := st.GetAttendance(ctx, c.ID); {
	case err == nil:
		d.Attendance = &Attendance{Attended: a.Present, JustifiedAbsent: a.JustifiedAbsent,
			UnjustifiedAbsent: a.UnjustifiedAbsent, Total: a.Total}
	case errors.Is(err, store.ErrNotFound):
		d.Pending = append(d.Pending, SectionAttendance)
	default:
		return Day{}, err
	}

	votings, err := st.ListVotings(ctx, c.ID)
	if err != nil {
		return Day{}, err
	}
	d.Votings = []Voting{}
	for _, v := range votings {
		d.Votings = append(d.Votings, Voting{ID: v.VotingID, Date: v.Date, BulletinNumber: v.BulletinNumber,
			DocumentTitle: v.DocumentTitle, ArticleText: v.ArticleText, Voted: v.VotedOption,
			TotalApproved: v.TotalApproved, TotalRejected: v.TotalRejected,
			TotalAbstention: v.TotalAbstention, Quorum: v.Quorum, Result: v.Result})
	}
	if len(votings) == 0 && !refreshed {
		d.Pending = append(d.Pending, SectionVotings)
	}

	projects, err := st.ListLawProjects(ctx, c.ID)
	if err != nil {
		return Day{}, err
	}
	if len(projects) > 0 || refreshed {
		act := &Activity{All: len(projects), ByStatus: map[string]int{}}
		for _, p := range projects {
			act.ByStatus[p.Status]++
		}
		d.Activity = act
	} else {
		d.Pending = append(d.Pending, SectionActivity)
	}

	d.Expenses, err = buildExpenses(ctx, st, c)
	if err != nil {
		return Day{}, err
	}
	if len(d.Expenses) == 0 && !expensesSynced {
		d.Pending = append(d.Pending, SectionExpenses)
	}
	return d, nil
}

func buildExpenses(ctx context.Context, st *store.Store, c *congress.Congressman) ([]MonthExpenses, error) {
	groups := scopes{
		global: store.Scope{Chamber: c.Chamber},
		local:  store.Scope{Chamber: c.Chamber, ConstituencyKind: c.ConstituencyKind, ConstituencyName: c.ConstituencyName},
	}
	months, err := st.ExpenseMonths(ctx, c.ID, expenseMonths)
	if err != nil {
		return nil, err
	}
	out := []MonthExpenses{}
	if len(months) == 0 {
		return out, nil
	}
	operational, err := st.ListOperationalExpenses(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	staff, err := st.ListStaffExpenses(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	for _, m := range months {
		rec := MonthExpenses{Year: m.Year, Month: monthNames[m.Month-1], Detail: []ExpenseDetail{}}

		// Support staff first, only for chambers that report it.
		staffCmp, err := groups.compare(
			func(sc store.Scope) (store.ExpenseStats, error) { return st.StaffMonthStats(ctx, sc, m.Year, m.Month) },
			func(sc store.Scope) (int, bool, error) { return st.StaffRank(ctx, sc, c.ID, m.Year, m.Month) })
		if err != nil {
			return nil, err
		}
		if staffCmp.Global.Count > 0 {
			det := ExpenseDetail{Type: typeStaff, Stats: staffCmp}
			for _, s := range staff {
				if s.Year == m.Year && s.Month == m.Month {
					amount, n := s.TotalAmount, s.HiredStaff
					det.Amount, det.SupportStaff = &amount, &n
					rec.Total += amount
				}
			}
			rec.Detail = append(rec.Detail, det)
		}

		// Operational spending with its categories.
		opCmp, err := groups.compare(
			func(sc store.Scope) (store.ExpenseStats, error) {
				return st.OperationalMonthStats(ctx, sc, m.Year, m.Month)
			},
			func(sc store.Scope) (int, bool, error) { return st.OperationalRank(ctx, sc, c.ID, m.Year, m.Month) })
		if err != nil {
			return nil, err
		}
		op := ExpenseDetail{Type: typeOperational, Stats: opCmp, Expenses: []ExpenseCategory{}}
		sum := 0
		for _, e := range operational {
			if e.Year != m.Year || e.Month != m.Month {
				continue
			}
			sum += e.Amount
			cmp, err := groups.compare(func(sc store.Scope) (store.ExpenseStats, error) {
				return st.OperationalCategoryStats(ctx, sc, m.Year, m.Month, e.ExpenseType)
			}, nil)
			if err != nil {
				return nil, err
			}
			op.Expenses = append(op.Expenses, ExpenseCategory{Subtype: e.ExpenseType, Amount: e.Amount, Stats: cmp})
		}
		op.Amount = &sum
		rec.Total += sum
		rec.Detail = append(rec.Detail, op)
		out = append(out, rec)
	}
	return out, nil
}

// scopes are the two groups a congressman is compared with.
type scopes struct{ global, local store.Scope }

// compare computes the statistics for both groups. rank may be nil for figures
// that have no ranking. The local group is skipped, without querying, when the
// constituency is not known.
func (g scopes) compare(
	stats func(store.Scope) (store.ExpenseStats, error),
	rank func(store.Scope) (int, bool, error),
) (Comparison, error) {
	global, err := summarize(typeAll, g.global, stats, rank)
	if err != nil {
		return Comparison{}, err
	}
	cmp := Comparison{Global: global}
	if g.local.Local() {
		if cmp.Local, err = summarize(string(g.local.ConstituencyKind), g.local, stats, rank); err != nil {
			return Comparison{}, err
		}
	}
	return cmp, nil
}

func summarize(kind string, sc store.Scope, stats func(store.Scope) (store.ExpenseStats, error), rank func(store.Scope) (int, bool, error)) (*Stats, error) {
	s, err := stats(sc)
	if err != nil {
		return nil, err
	}
	out := &Stats{Type: kind, Count: s.Count}
	out.Avg, out.Min, out.Max = statsValues(s)
	if rank != nil {
		if r, ok, err := rank(sc); err != nil {
			return nil, err
		} else if ok {
			out.Ranking = &r
		}
	}
	return out, nil
}

// statsValues returns nil for all three when nobody has a record.
func statsValues(s store.ExpenseStats) (avg, min, max *int) {
	if s.Count == 0 {
		return nil, nil, nil
	}
	a := int(math.Round(s.Avg))
	return &a, &s.Min, &s.Max
}
