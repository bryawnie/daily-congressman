package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/store"
)

func newStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	return st
}

func addCongressman(t *testing.T, st *store.Store, ext int, name string) int64 {
	t.Helper()
	id, err := st.UpsertCongressman(context.Background(), &congress.Congressman{
		Chamber: congress.Senate, ExternalID: ext, FirstName: name, FirstSurname: "Pérez",
		ConstituencyKind: congress.ConstituencyRegion, ConstituencyName: "Región X",
		LastUpdate: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func addIn(t *testing.T, st *store.Store, ext int, region string) int64 {
	t.Helper()
	id, err := st.UpsertCongressman(context.Background(), &congress.Congressman{
		Chamber: congress.Senate, ExternalID: ext, FirstName: "S",
		ConstituencyKind: congress.ConstituencyRegion, ConstituencyName: region, LastUpdate: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func choose(t *testing.T, st *store.Store, date string, id int64) {
	t.Helper()
	if _, err := st.SaveDayChoice(context.Background(), congress.DayChoice{
		Date: date, CongressmanID: id, ChainID: "chain", PulseID: "42", RandOut: "ab"}); err != nil {
		t.Fatal(err)
	}
}

func get(t *testing.T, s *Server, path string, headers ...string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("bad JSON %q: %v", rec.Body.String(), err)
	}
	return v
}

func TestHealth(t *testing.T) {
	s := &Server{Store: newStore(t)}
	if rec := get(t, s, "/health"); rec.Code != 200 {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestDaysPendingAndFull(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	a := addCongressman(t, st, 1, "Ana")
	b := addCongressman(t, st, 2, "Beto")
	choose(t, st, "2026-11-20", a)
	choose(t, st, "2026-11-21", b)

	// Ana: nothing fetched yet. Beto: fetched, with data.
	st.UpsertAttendance(ctx, b, congress.Attendance{Total: 10, Present: 8, JustifiedAbsent: 1, UnjustifiedAbsent: 1})
	st.ReplaceVotings(ctx, b, []congress.Voting{{VotingID: 1, Date: "2026-11-01", VotedOption: "Sí"}})
	st.UpsertLawProjects(ctx, b, []congress.LawProject{
		{Bulletin: "1-07", Status: "En tramitación"}, {Bulletin: "2-07", Status: "Publicado"}, {Bulletin: "3-07", Status: "En tramitación"}})
	st.MarkDone(ctx, "refresh", "senate:2", "2026-11-21", time.Now())
	st.UpsertOperationalExpenses(ctx, a, []congress.OperationalExpense{{Year: 2026, Month: 10, ExpenseType: "Telefonía", Amount: 100}})
	st.UpsertOperationalExpenses(ctx, b, []congress.OperationalExpense{
		{Year: 2026, Month: 10, ExpenseType: "Telefonía", Amount: 300},
		{Year: 2026, Month: 10, ExpenseType: "Difusión", Amount: 200}})

	s := &Server{Store: st}
	rec := get(t, s, "/days")
	if rec.Code != 200 {
		t.Fatalf("status %d: %s", rec.Code, rec.Body)
	}
	days := decode[struct{ Days []Day }](t, rec).Days
	if len(days) != 2 || days[0].Date != "2026-11-21" {
		t.Fatalf("days = %+v", days)
	}

	beto, ana := days[0], days[1]
	if beto.Attendance == nil || beto.Attendance.Attended != 8 || len(beto.Votings) != 1 {
		t.Errorf("beto = %+v", beto)
	}
	if beto.Activity == nil || beto.Activity.All != 3 || beto.Activity.ByStatus["En tramitación"] != 2 {
		t.Errorf("activity = %+v", beto.Activity)
	}
	if got := beto.Pending; len(got) != 0 {
		t.Errorf("beto pending = %v, want none", got)
	}

	// Expenses: Beto spent 500 vs Ana 100, so he ranks 1 of 2.
	if len(beto.Expenses) != 1 {
		t.Fatalf("expenses = %+v", beto.Expenses)
	}
	m := beto.Expenses[0]
	op := m.Detail[len(m.Detail)-1]
	g := op.Stats.Global
	if m.Month != "octubre" || m.Total != 500 || op.Type != "Operacional" || *op.Amount != 500 ||
		*g.Ranking != 1 || g.Count != 2 || *g.Avg != 300 || *g.Min != 100 || *g.Max != 500 ||
		len(op.Expenses) != 2 {
		t.Errorf("month = %+v op = %+v", m, op)
	}

	// Ana: nothing fetched, so sections are pending, not zero.
	if ana.Attendance != nil || ana.Activity != nil {
		t.Errorf("ana should have null attendance and activity: %+v", ana)
	}
	want := map[string]bool{SectionAttendance: true, SectionVotings: true, SectionActivity: true}
	if len(ana.Pending) != len(want) {
		t.Errorf("ana pending = %v", ana.Pending)
	}
	for _, p := range ana.Pending {
		if !want[p] {
			t.Errorf("unexpected pending %q", p)
		}
	}
	if *ana.Expenses[0].Detail[0].Stats.Global.Ranking != 2 {
		t.Errorf("ana ranking = %v", *ana.Expenses[0].Detail[0].Stats.Global.Ranking)
	}
}

func TestFetchedButEmptyIsNotPending(t *testing.T) {
	st := newStore(t)
	id := addCongressman(t, st, 1, "Ana")
	choose(t, st, "2026-11-20", id)
	st.MarkDone(context.Background(), "refresh", "senate:1", "2026-11-20", time.Now())
	st.MarkDone(context.Background(), "expenses", "senate:1", "2026-11", time.Now())

	d := decode[Day](t, get(t, &Server{Store: st}, "/days/2026-11-20"))
	if d.Activity == nil || d.Activity.All != 0 {
		t.Errorf("activity = %+v, want zero not null", d.Activity)
	}
	if d.Expenses == nil || len(d.Expenses) != 0 || d.Votings == nil {
		t.Errorf("empty sections must be [] not null: %+v", d)
	}
	for _, p := range d.Pending {
		if p != SectionAttendance {
			t.Errorf("pending = %v, only attendance should be pending", d.Pending)
		}
	}
}

func TestDayErrors(t *testing.T) {
	st := newStore(t)
	s := &Server{Store: st}
	for path, want := range map[string]int{
		"/days/2026-11-20": 404,
		"/days/hoy":        400,
		"/days/2026-1-2":   400,
		"/days?limit=0":    400,
		"/days?limit=abc":  400,
		"/days?limit=999":  400,
	} {
		if rec := get(t, s, path); rec.Code != want {
			t.Errorf("%s: status %d, want %d", path, rec.Code, want)
		}
	}
	// Empty database is an empty list, not an error.
	if rec := get(t, s, "/days"); rec.Code != 200 || decode[struct{ Days []Day }](t, rec).Days == nil {
		t.Errorf("empty list: %d %s", rec.Code, rec.Body)
	}
}

func TestLimit(t *testing.T) {
	st := newStore(t)
	id := addCongressman(t, st, 1, "Ana")
	for _, d := range []string{"2026-11-18", "2026-11-19", "2026-11-20"} {
		choose(t, st, d, id)
	}
	days := decode[struct{ Days []Day }](t, get(t, &Server{Store: st}, "/days?limit=2")).Days
	if len(days) != 2 || days[0].Date != "2026-11-20" {
		t.Errorf("days = %+v", days)
	}
}

func TestReadOnlyAndCORS(t *testing.T) {
	s := &Server{Store: newStore(t), CORSOrigins: []string{"https://app.example"}}

	req := httptest.NewRequest(http.MethodPost, "/days", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST status %d, want 405", rec.Code)
	}

	if got := get(t, s, "/health", "Origin", "https://app.example").Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Errorf("allowed origin got %q", got)
	}
	if got := get(t, s, "/health", "Origin", "https://evil.example").Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("foreign origin got %q", got)
	}

	pre := httptest.NewRequest(http.MethodOptions, "/days", nil)
	pre.Header.Set("Origin", "https://app.example")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, pre)
	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status %d", rec.Code)
	}

	any := &Server{Store: newStore(t), CORSOrigins: []string{"*"}}
	if got := get(t, any, "/health", "Origin", "https://x.example").Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("wildcard got %q", got)
	}
}

func TestLocalStats(t *testing.T) {
	st := newStore(t)
	ctx := context.Background()
	// Three senators in region A (spending 100, 300, 500) and one in B (900).
	a1 := addIn(t, st, 1, "Región A")
	a2 := addIn(t, st, 2, "Región A")
	a3 := addIn(t, st, 3, "Región A")
	b1 := addIn(t, st, 4, "Región B")
	for id, amount := range map[int64]int{a1: 100, a2: 300, a3: 500, b1: 900} {
		st.UpsertOperationalExpenses(ctx, id, []congress.OperationalExpense{
			{Year: 2026, Month: 10, ExpenseType: "Telefonía", Amount: amount}})
	}
	choose(t, st, "2026-11-20", a2)

	d := decode[Day](t, get(t, &Server{Store: st}, "/days/2026-11-20"))
	op := d.Expenses[0].Detail[0]

	// Chamber: 4 senators, a2 spent 300 -> third of 4, average 450.
	if g := op.Stats.Global; *g.Ranking != 3 || g.Count != 4 || *g.Avg != 450 || *g.Max != 900 {
		t.Errorf("chamber stats = %+v", op)
	}
	// Region A only: 3 senators, a2 is second, average 300.
	l := op.Stats.Local
	if l == nil || l.Count != 3 || *l.Ranking != 2 || *l.Avg != 300 || *l.Min != 100 || *l.Max != 500 {
		t.Fatalf("local stats = %+v", l)
	}
	cat := op.Expenses[0].Stats.Local
	if cat == nil || cat.Count != 3 || cat.Ranking != nil || *cat.Avg != 300 {
		t.Errorf("category local = %+v", cat)
	}
}

func TestLocalStatsUnknownConstituency(t *testing.T) {
	st := newStore(t)
	id := addIn(t, st, 1, "")
	st.UpsertOperationalExpenses(context.Background(), id, []congress.OperationalExpense{
		{Year: 2026, Month: 10, ExpenseType: "Telefonía", Amount: 100}})
	choose(t, st, "2026-11-20", id)
	op := decode[Day](t, get(t, &Server{Store: st}, "/days/2026-11-20")).Expenses[0].Detail[0]
	if op.Stats.Local != nil || op.Expenses[0].Stats.Local != nil || op.Stats.Global == nil {
		t.Errorf("local must be null without a constituency: %+v", op)
	}
}
