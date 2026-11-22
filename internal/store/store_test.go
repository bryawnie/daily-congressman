package store

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return s
}

func sampleCongressman(chamber congress.Chamber, externalID int) *congress.Congressman {
	return &congress.Congressman{
		Chamber:          chamber,
		ExternalID:       externalID,
		FirstName:        "Ana",
		FirstSurname:     "Pérez",
		SecondSurname:    "Soto",
		Party:            "Partido X",
		ConstituencyKind: congress.ConstituencyDistrict,
		ConstituencyName: "Distrito 8",
		LastUpdate:       time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC),
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	s := newTestStore(t)
	if err := s.Migrate(context.Background()); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}

func TestForeignKeysAndWALEnabled(t *testing.T) {
	s := newTestStore(t)
	var fk int
	if err := s.db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil || fk != 1 {
		t.Fatalf("foreign_keys = %d, err %v", fk, err)
	}
	var mode string
	if err := s.db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil || mode != "wal" {
		t.Fatalf("journal_mode = %q, err %v", mode, err)
	}
}

func TestUpsertCongressmanKeepsIDAndChildren(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	c := sampleCongressman(congress.Deputies, 1001)
	id, err := s.UpsertCongressman(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertAttendance(ctx, id, congress.Attendance{Total: 10, Present: 9}); err != nil {
		t.Fatal(err)
	}

	c.Party = "Partido Y"
	id2, err := s.UpsertCongressman(ctx, c)
	if err != nil {
		t.Fatal(err)
	}
	if id2 != id {
		t.Fatalf("id changed on upsert: %d -> %d", id, id2)
	}
	got, err := s.GetCongressman(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Party != "Partido Y" || !got.LastUpdate.Equal(c.LastUpdate) {
		t.Fatalf("unexpected congressman after upsert: %+v", got)
	}
	// Child rows must survive the update.
	if _, err := s.GetAttendance(ctx, id); err != nil {
		t.Fatalf("attendance lost after upsert: %v", err)
	}
}

func TestSameExternalIDInDifferentChambers(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	d, err := s.UpsertCongressman(ctx, sampleCongressman(congress.Deputies, 7))
	if err != nil {
		t.Fatal(err)
	}
	sn, err := s.UpsertCongressman(ctx, sampleCongressman(congress.Senate, 7))
	if err != nil {
		t.Fatal(err)
	}
	if d == sn {
		t.Fatal("deputy and senator with same external id share an internal id")
	}
	got, err := s.FindCongressman(ctx, congress.Senate, 7)
	if err != nil || got.ID != sn {
		t.Fatalf("FindCongressman senate: %+v, %v", got, err)
	}
}

func TestInvalidChamberRejected(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.UpsertCongressman(context.Background(), sampleCongressman("house", 1)); err == nil {
		t.Fatal("expected error for invalid chamber")
	}
}

func TestNotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if _, err := s.GetCongressman(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetCongressman: %v", err)
	}
	if _, err := s.GetAttendance(ctx, 999); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetAttendance: %v", err)
	}
	if _, err := s.GetDayChoice(ctx, "2026-01-01"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetDayChoice: %v", err)
	}
}

func TestChildRowsRoundTrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	id, err := s.UpsertCongressman(ctx, sampleCongressman(congress.Deputies, 1))
	if err != nil {
		t.Fatal(err)
	}

	periods := []congress.Period{{From: 2018, To: 2022}, {From: 2022, To: 2026}}
	if err := s.UpsertPeriods(ctx, id, periods); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertPeriods(ctx, id, periods); err != nil { // idempotent
		t.Fatal(err)
	}
	gotPeriods, _ := s.ListPeriods(ctx, id)
	if want := []congress.Period{{From: 2022, To: 2026}, {From: 2018, To: 2022}}; !reflect.DeepEqual(gotPeriods, want) {
		t.Fatalf("periods = %v, want %v", gotPeriods, want)
	}

	att := congress.Attendance{Total: 100, Present: 90, JustifiedAbsent: 7, UnjustifiedAbsent: 3}
	if err := s.UpsertAttendance(ctx, id, att); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.GetAttendance(ctx, id); got != att {
		t.Fatalf("attendance = %+v, want %+v", got, att)
	}

	laws := []congress.LawProject{
		{Bulletin: "100-01", Name: "A", Type: "Moción", Status: "En trámite", Date: "2026-01-01"},
		{Bulletin: "200-02", Name: "B", Type: "Moción", Status: "Archivado", Date: "2026-03-01"},
	}
	if err := s.UpsertLawProjects(ctx, id, laws); err != nil {
		t.Fatal(err)
	}
	gotLaws, _ := s.ListLawProjects(ctx, id)
	if len(gotLaws) != 2 || gotLaws[0].Bulletin != "200-02" {
		t.Fatalf("law projects = %+v", gotLaws)
	}
}

func TestReplaceVotingsDropsPreviousOnes(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	id, _ := s.UpsertCongressman(ctx, sampleCongressman(congress.Deputies, 1))

	first := []congress.Voting{
		{VotingID: 1, Date: "2026-01-01", BulletinNumber: "1234-05", VotedOption: "A favor"},
		{VotingID: 2, Date: "2026-02-01", BulletinNumber: "1235-05", VotedOption: "En contra"},
	}
	if err := s.ReplaceVotings(ctx, id, first); err != nil {
		t.Fatal(err)
	}
	second := []congress.Voting{{VotingID: 3, Date: "2026-03-01", BulletinNumber: "1236-05", VotedOption: "Abstención",
		Quorum: "Mayoría simple", Result: "Aprobado"}}
	if err := s.ReplaceVotings(ctx, id, second); err != nil {
		t.Fatal(err)
	}
	got, err := s.ListVotings(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].VotingID != 3 || got[0].BulletinNumber != "1236-05" ||
		got[0].Quorum != "Mayoría simple" || got[0].Result != "Aprobado" {
		t.Fatalf("votings = %+v", got)
	}
}

func TestExpensesUpsertAndOrder(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	id, _ := s.UpsertCongressman(ctx, sampleCongressman(congress.Deputies, 1))

	op := []congress.OperationalExpense{
		{Year: 2026, Month: 9, ExpenseType: "Transporte", Amount: 100},
		{Year: 2026, Month: 10, ExpenseType: "Transporte", Amount: 200},
	}
	if err := s.UpsertOperationalExpenses(ctx, id, op); err != nil {
		t.Fatal(err)
	}
	op[1].Amount = 250 // update the same key
	if err := s.UpsertOperationalExpenses(ctx, id, op[1:]); err != nil {
		t.Fatal(err)
	}
	got, _ := s.ListOperationalExpenses(ctx, id)
	if len(got) != 2 || got[0].Month != 10 || got[0].Amount != 250 {
		t.Fatalf("operational = %+v", got)
	}

	if err := s.UpsertOfficeExpenses(ctx, id, []congress.OfficeExpense{{Year: 2026, Month: 10, OfficesNumber: 2, TotalAmount: 500}}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.ListOfficeExpenses(ctx, id); len(got) != 1 || got[0].OfficesNumber != 2 {
		t.Fatalf("offices = %+v", got)
	}
	if err := s.UpsertStaffExpenses(ctx, id, []congress.StaffExpense{{Year: 2026, Month: 10, HiredStaff: 3, TotalAmount: 900}}); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.ListStaffExpenses(ctx, id); len(got) != 1 || got[0].HiredStaff != 3 {
		t.Fatalf("staff = %+v", got)
	}

	if err := s.UpsertOfficeExpenses(ctx, id, []congress.OfficeExpense{{Year: 2026, Month: 13}}); err == nil {
		t.Fatal("expected month check constraint to fail")
	}
}

func TestDayChoiceIsStoredOnce(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	a, _ := s.UpsertCongressman(ctx, sampleCongressman(congress.Deputies, 1))
	b, _ := s.UpsertCongressman(ctx, sampleCongressman(congress.Senate, 2))

	created, err := s.SaveDayChoice(ctx, congress.DayChoice{Date: "2026-11-16", CongressmanID: a, PulseID: "p1"})
	if err != nil || !created {
		t.Fatalf("first save: created=%v err=%v", created, err)
	}
	created, err = s.SaveDayChoice(ctx, congress.DayChoice{Date: "2026-11-16", CongressmanID: b})
	if err != nil || created {
		t.Fatalf("second save: created=%v err=%v", created, err)
	}
	got, _ := s.GetDayChoice(ctx, "2026-11-16")
	if got.CongressmanID != a || got.PulseID != "p1" {
		t.Fatalf("choice was overwritten: %+v", got)
	}

	for _, d := range []string{"2026-11-14", "2026-11-15", "2026-11-17"} {
		if _, err := s.SaveDayChoice(ctx, congress.DayChoice{Date: d, CongressmanID: b}); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := s.ListDayChoices(ctx, 3)
	if len(list) != 3 || list[0].Date != "2026-11-17" || list[2].Date != "2026-11-15" {
		t.Fatalf("list = %+v", list)
	}
}

func TestChooseNonexistentCongressmanFails(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.SaveDayChoice(context.Background(), congress.DayChoice{Date: "2026-01-01", CongressmanID: 42}); err == nil {
		t.Fatal("expected foreign key violation")
	}
}

func TestJobStateAndErrors(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	at := time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

	if done, _ := s.IsDone(ctx, "profiles", "senate:1", "2026-11"); done {
		t.Fatal("nothing is done yet")
	}
	for i := 0; i < 2; i++ { // marking twice is harmless
		if err := s.MarkDone(ctx, "profiles", "senate:1", "2026-11", at); err != nil {
			t.Fatal(err)
		}
	}
	if done, _ := s.IsDone(ctx, "profiles", "senate:1", "2026-11"); !done {
		t.Fatal("should be done")
	}
	if done, _ := s.IsDone(ctx, "profiles", "senate:1", "2026-12"); done {
		t.Fatal("another period is independent")
	}
	if err := s.ClearDone(ctx, "profiles", "2026-11"); err != nil {
		t.Fatal(err)
	}
	if done, _ := s.IsDone(ctx, "profiles", "senate:1", "2026-11"); done {
		t.Fatal("ClearDone should forget it")
	}

	for _, msg := range []string{"first", "second"} {
		if err := s.LogJobError(ctx, JobError{At: at, Job: "expenses", Key: "senate:1", Period: "2026-11", Error: msg}); err != nil {
			t.Fatal(err)
		}
	}
	errs, err := s.ListJobErrors(ctx, 10)
	if err != nil || len(errs) != 2 || errs[0].Error != "second" || !errs[0].At.Equal(at) {
		t.Fatalf("errors = %+v, err = %v", errs, err)
	}
}
