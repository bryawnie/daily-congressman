package senate

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// expenseMonths is how many of the latest published months are read.
const expenseMonths = 6

// totalConcept is the sum row of the monthly table. It is not stored, so the
// categories can be added without counting it twice.
const totalConcept = "VALOR TOTAL"

type yearMonth struct{ Year, Month int }

func (a yearMonth) index() int { return a.Year*12 + a.Month }

type availablePeriods struct {
	Data []struct {
		Attributes struct {
			Year  int `json:"ano"`
			Month int `json:"mes"`
		} `json:"attributes"`
	} `json:"data"`
}

// expensePeriods returns the latest published months (not later than today),
// newest first.
func (s *Source) expensePeriods(ctx context.Context) ([]yearMonth, error) {
	q := url.Values{}
	q.Set("pagination[limit]", "500")
	q.Set("filters[endpoint][$eq]", "gastos-operacionales-senadores")
	var data availablePeriods
	if err := s.getJSON(ctx, "/api/transparency/available-periods?"+q.Encode(), &data); err != nil {
		return nil, err
	}
	now := s.now().UTC()
	limit := yearMonth{now.Year(), int(now.Month())}.index()
	seen := map[yearMonth]bool{}
	var all []yearMonth
	for _, d := range data.Data {
		ym := yearMonth{d.Attributes.Year, d.Attributes.Month}
		if ym.Month < 1 || ym.Month > 12 {
			return nil, fmt.Errorf("senate: bad expense period %+v", ym)
		}
		if ym.index() > limit || seen[ym] {
			continue
		}
		seen[ym] = true
		all = append(all, ym)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].index() > all[j].index() })
	if len(all) > expenseMonths {
		all = all[:expenseMonths]
	}
	return all, nil
}

// transparencyRow identifies a senator in the transparency data, where each
// one has an "unidad ejecutora" instead of the senator ID.
type transparencyRow struct {
	Unit          int    `json:"unidad_ejecutora"`
	FirstName     string `json:"nombre"`
	FirstSurname  string `json:"appaterno"`
	SecondSurname string `json:"apmaterno"`
}

var accents = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

// words lowercases, removes accents and keeps only the words of a name.
func words(s string) []string {
	s = accents.Replace(strings.ToLower(s))
	return strings.FieldsFunc(s, func(r rune) bool { return r < 'a' || r > 'z' })
}

func phrase(s string) string { return strings.Join(words(s), " ") }

func firstWord(s string) string {
	w := words(s)
	if len(w) == 0 {
		return ""
	}
	return w[0]
}

// matchUnit finds the senator among the transparency rows. The names are not
// written the same way in both places (a senator can appear with one given
// name in one and two in the other), so it narrows by first surname, then by
// second surname, then by first given name. It returns false if the senator
// is not listed and an error if the match is ambiguous.
func matchUnit(rec senatorRecord, rows []transparencyRow) (unit int, found bool, err error) {
	var cands []transparencyRow
	for _, r := range rows {
		if phrase(r.FirstSurname) == phrase(rec.FirstSurname) {
			cands = append(cands, r)
		}
	}
	narrow := func(keep func(transparencyRow) bool) {
		if len(cands) <= 1 {
			return
		}
		var next []transparencyRow
		for _, r := range cands {
			if keep(r) {
				next = append(next, r)
			}
		}
		if len(next) > 0 {
			cands = next
		}
	}
	narrow(func(r transparencyRow) bool {
		return phrase(r.SecondSurname) == phrase(rec.SecondSurname)
	})
	narrow(func(r transparencyRow) bool {
		return firstWord(r.FirstName) == firstWord(rec.FirstName)
	})
	switch len(cands) {
	case 0:
		return 0, false, nil
	case 1:
		return cands[0].Unit, true, nil
	}
	return 0, false, fmt.Errorf("senate: %s matches %d senators in the expenses data", rec.fullName(), len(cands))
}

type expenseRows struct {
	Data []struct {
		Attributes struct {
			Year    int     `json:"ano"`
			Month   int     `json:"mes"`
			Concept string  `json:"gastos_operacionales"`
			Amount  float64 `json:"monto"`
			Unit    int     `json:"unidad_ejecutora"`
		} `json:"attributes"`
	} `json:"data"`
	Meta struct {
		Pagination struct {
			PageCount int `json:"pageCount"`
		} `json:"pagination"`
	} `json:"meta"`
}

// Expenses returns the operational expenses of the latest published months.
// The Senate does not publish support staff or offices as separate data
// (rent of offices is a concept of the operational expenses), so those lists
// are empty. A month in which the senator is not listed (not yet seated) is
// skipped.
func (s *Source) Expenses(ctx context.Context, m congress.Member) (congress.Expenses, error) {
	if err := checkMember(m); err != nil {
		return congress.Expenses{}, err
	}
	rec, err := s.senator(ctx, m.ExternalID)
	if err != nil {
		return congress.Expenses{}, err
	}
	periods, err := s.expensePeriods(ctx)
	if err != nil {
		return congress.Expenses{}, err
	}

	var out congress.Expenses
	for _, ym := range periods {
		var rows []transparencyRow
		path := fmt.Sprintf("/api/transparency/parlamentarians-periods?year=%d&month=%d", ym.Year, ym.Month)
		if err := s.getJSON(ctx, path, &rows); err != nil {
			return congress.Expenses{}, err
		}
		unit, found, err := matchUnit(rec, rows)
		if err != nil {
			return congress.Expenses{}, err
		}
		if !found {
			continue
		}

		q := url.Values{}
		q.Set("sort", "gastos_operacionales")
		q.Set("pagination[pageSize]", "100")
		q.Set("filters[ano][$eq]", strconv.Itoa(ym.Year))
		q.Set("filters[mes][$eq]", strconv.Itoa(ym.Month))
		q.Set("filters[unidad_ejecutora][$eq]", strconv.Itoa(unit))
		var data expenseRows
		if err := s.getJSON(ctx, "/api/transparency/expenses/senator-Operational-expenses?"+q.Encode(), &data); err != nil {
			return congress.Expenses{}, err
		}
		if data.Meta.Pagination.PageCount > 1 {
			return congress.Expenses{}, fmt.Errorf("senate: expenses of %d-%02d span %d pages", ym.Year, ym.Month, data.Meta.Pagination.PageCount)
		}
		for _, r := range data.Data {
			a := r.Attributes
			// A filter that the back end ignores would return other senators.
			if a.Unit != unit || a.Year != ym.Year || a.Month != ym.Month {
				return congress.Expenses{}, fmt.Errorf("senate: expenses row for unit %d %d-%02d does not match the request (unit %d)", a.Unit, a.Year, a.Month, unit)
			}
			if a.Concept == "" || a.Concept == totalConcept {
				continue
			}
			out.Operational = append(out.Operational, congress.OperationalExpense{
				Year:        a.Year,
				Month:       a.Month,
				ExpenseType: a.Concept,
				Amount:      int(a.Amount),
			})
		}
	}
	return out, nil
}
