package jobs

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bryawnie/daily-congressman/internal/beacon"
	"github.com/bryawnie/daily-congressman/internal/chooser"
	"github.com/bryawnie/daily-congressman/internal/congress"
	"github.com/bryawnie/daily-congressman/internal/retrieval"
	"github.com/bryawnie/daily-congressman/internal/source"
	"github.com/bryawnie/daily-congressman/internal/store"
)

// fakeSource is a senate source with three members. It counts calls and can be
// told to fail.
type fakeSource struct {
	congress.Unsupported
	calls map[string]int
	// failProfile[id] is how many more times Profile(id) fails.
	failProfile map[int]int
	failVotings int
	noExpenses  bool
}

func newFakeSource() *fakeSource {
	return &fakeSource{calls: map[string]int{}, failProfile: map[int]int{}}
}

func (f *fakeSource) Chamber() congress.Chamber { return congress.Senate }

func (f *fakeSource) CurrentMembers(context.Context) ([]congress.Member, error) {
	f.calls["members"]++
	return []congress.Member{
		{Chamber: congress.Senate, ExternalID: 1, Name: "Uno"},
		{Chamber: congress.Senate, ExternalID: 2, Name: "Dos"},
		{Chamber: congress.Senate, ExternalID: 3, Name: "Tres"},
	}, nil
}

func (f *fakeSource) Profile(_ context.Context, m congress.Member) (congress.Congressman, error) {
	f.calls["profile"]++
	if f.failProfile[m.ExternalID] > 0 {
		f.failProfile[m.ExternalID]--
		return congress.Congressman{}, errors.New("boom")
	}
	return congress.Congressman{Chamber: m.Chamber, ExternalID: m.ExternalID, FirstName: m.Name}, nil
}

func (f *fakeSource) Attendance(context.Context, congress.Member) (congress.Attendance, error) {
	f.calls["attendance"]++
	return congress.Attendance{Total: 10, Present: 9}, nil
}

func (f *fakeSource) Votings(context.Context, congress.Member, int) ([]congress.Voting, error) {
	f.calls["votings"]++
	if f.failVotings > 0 {
		f.failVotings--
		return nil, errors.New("votings down")
	}
	return []congress.Voting{{VotingID: 7, BulletinNumber: "1-01"}}, nil
}

func (f *fakeSource) Activity(context.Context, congress.Member) ([]congress.LawProject, error) {
	f.calls["activity"]++
	return []congress.LawProject{{Bulletin: "9-01"}}, nil
}

func (f *fakeSource) Expenses(context.Context, congress.Member) (congress.Expenses, error) {
	f.calls["expenses"]++
	if f.noExpenses {
		return congress.Expenses{}, congress.ErrNotSupported
	}
	return congress.Expenses{Operational: []congress.OperationalExpense{{Year: 2026, Month: 10, ExpenseType: "X", Amount: 5}}}, nil
}

type fakePulses struct {
	ready bool
	calls int
}

func (p *fakePulses) PulseAt(context.Context, time.Time) (beacon.Pulse, error) {
	p.calls++
	if !p.ready {
		return beacon.Pulse{}, beacon.ErrNotAvailable
	}
	return beacon.Pulse{ChainID: "1", PulseID: "2", RandOut: "01"}, nil // 1 mod 3 -> member 2
}

type harness struct {
	j      *Jobs
	src    *fakeSource
	pulses *fakePulses
	st     *store.Store
	now    time.Time
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	h := &harness{
		src:    newFakeSource(),
		pulses: &fakePulses{ready: true},
		st:     st,
		now:    time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
	}
	reg := source.NewRegistry(h.src)
	h.j = &Jobs{
		Store:   st,
		Sources: reg,
		Chooser: &chooser.Chooser{Pulses: h.pulses, Sources: reg},
		Syncer:  &retrieval.Syncer{Store: st, Sources: reg, Attempts: 2, Backoff: time.Millisecond},
		Now:     func() time.Time { return h.now },
		Log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	return h
}

func TestProfilesAreSyncedOncePerMonth(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()

	if err := h.j.SyncProfiles(ctx, false); err != nil {
		t.Fatal(err)
	}
	if h.src.calls["profile"] != 3 {
		t.Fatalf("profile calls = %d, want 3", h.src.calls["profile"])
	}
	if _, err := h.st.FindCongressman(ctx, congress.Senate, 2); err != nil {
		t.Fatalf("profile not stored: %v", err)
	}

	// Same month: nothing to do, not even listing the members.
	h.now = h.now.AddDate(0, 0, 5)
	members := h.src.calls["members"]
	if err := h.j.SyncProfiles(ctx, false); err != nil {
		t.Fatal(err)
	}
	if h.src.calls["profile"] != 3 || h.src.calls["members"] != members {
		t.Fatalf("second run in the same month hit the source: %v", h.src.calls)
	}

	// -force redoes it.
	if err := h.j.SyncProfiles(ctx, true); err != nil {
		t.Fatal(err)
	}
	if h.src.calls["profile"] != 6 {
		t.Fatalf("forced run: profile calls = %d, want 6", h.src.calls["profile"])
	}

	// Next month it is due again.
	h.now = time.Date(2026, 12, 2, 12, 0, 0, 0, time.UTC)
	if err := h.j.SyncProfiles(ctx, false); err != nil {
		t.Fatal(err)
	}
	if h.src.calls["profile"] != 9 {
		t.Fatalf("next month: profile calls = %d, want 9", h.src.calls["profile"])
	}
}

func TestMonthlyRunResumesWithOnlyTheFailedOnes(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.src.failProfile[2] = 100 // fails every attempt of the first run

	err := h.j.SyncProfiles(ctx, false)
	if err == nil || !strings.Contains(err.Error(), "senate:2") {
		t.Fatalf("want an error naming senate:2, got %v", err)
	}
	if done, _ := h.st.IsDone(ctx, JobProfiles, allKey, h.j.Month()); done {
		t.Fatal("a month with failures must not be marked complete")
	}
	if errs, _ := h.st.ListJobErrors(ctx, 10); len(errs) != 1 || errs[0].Key != "senate:2" {
		t.Fatalf("logged errors = %+v", errs)
	}

	h.src.failProfile[2] = 0
	before := h.src.calls["profile"]
	if err := h.j.SyncProfiles(ctx, false); err != nil {
		t.Fatal(err)
	}
	if got := h.src.calls["profile"] - before; got != 1 {
		t.Fatalf("resumed run read %d profiles, want only the failed one", got)
	}
	if done, _ := h.st.IsDone(ctx, JobProfiles, allKey, h.j.Month()); !done {
		t.Fatal("the month should be complete now")
	}
}

func TestTransientErrorsAreRetried(t *testing.T) {
	h := newHarness(t)
	h.src.failProfile[1] = 1 // fails once, Attempts is 2
	if err := h.j.SyncProfiles(context.Background(), false); err != nil {
		t.Fatalf("a transient error should be absorbed by the retry: %v", err)
	}
}

func TestExpensesAreSyncedForEveryone(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	// The profiles were never synced: expenses brings the congressman in.
	if err := h.j.SyncExpenses(ctx, false); err != nil {
		t.Fatal(err)
	}
	for id := 1; id <= 3; id++ {
		c, err := h.st.FindCongressman(ctx, congress.Senate, id)
		if err != nil {
			t.Fatal(err)
		}
		exp, _ := h.st.ListOperationalExpenses(ctx, c.ID)
		if len(exp) != 1 {
			t.Fatalf("senator %d: %d operational rows", id, len(exp))
		}
	}
	if err := h.j.SyncExpenses(ctx, false); err != nil || h.src.calls["expenses"] != 3 {
		t.Fatalf("second run: err=%v expenses calls=%d", err, h.src.calls["expenses"])
	}
}

func TestExpensesNotSupportedIsPendingNotFailure(t *testing.T) {
	h := newHarness(t)
	h.src.noExpenses = true
	if err := h.j.SyncExpenses(context.Background(), false); err != nil {
		t.Fatalf("unsupported data must not fail the job: %v", err)
	}
}

func TestChooseWaitsForThePulseAndStoresNothingMeanwhile(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.pulses.ready = false

	_, _, err := h.j.Choose(ctx, "2026-11-20")
	if !errors.Is(err, ErrPulseNotReady) {
		t.Fatalf("err = %v, want ErrPulseNotReady", err)
	}
	if _, err := h.st.GetDayChoice(ctx, "2026-11-20"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("nothing should be stored yet, got %v", err)
	}

	h.pulses.ready = true
	choice, created, err := h.j.Choose(ctx, "2026-11-20")
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	c, _ := h.st.GetCongressman(ctx, choice.CongressmanID)
	if c.ExternalID != 2 || choice.PulseID != "2" || choice.RandOut != "01" {
		t.Fatalf("choice = %+v (%+v)", choice, c)
	}

	// Chosen once, never again: no pulse request, same result.
	calls := h.pulses.calls
	again, created, err := h.j.Choose(ctx, "2026-11-20")
	if err != nil || created || again != choice || h.pulses.calls != calls {
		t.Fatalf("second choose: created=%v err=%v calls=%d", created, err, h.pulses.calls)
	}
}

func TestChooseRejectsFutureDays(t *testing.T) {
	h := newHarness(t)
	if _, _, err := h.j.Choose(context.Background(), "2026-11-21"); err == nil {
		t.Fatal("a future day must not be drawn")
	}
}

func TestRefreshRetriesOnlyWhileIncomplete(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	if _, _, err := h.j.Choose(ctx, "2026-11-20"); err != nil {
		t.Fatal(err)
	}

	h.src.failVotings = 2 // both attempts of the first run
	if err := h.j.Refresh(ctx, "2026-11-20", false); err == nil {
		t.Fatal("expected the votings failure")
	}
	// The other data was still saved.
	choice, _ := h.st.GetDayChoice(ctx, "2026-11-20")
	if att, _ := h.st.GetAttendance(ctx, choice.CongressmanID); att.Total != 10 {
		t.Fatalf("attendance should be saved despite the votings failure: %+v", att)
	}

	if err := h.j.Refresh(ctx, "2026-11-20", false); err != nil {
		t.Fatal(err)
	}
	if v, _ := h.st.ListVotings(ctx, choice.CongressmanID); len(v) != 1 {
		t.Fatalf("votings = %+v", v)
	}
	calls := h.src.calls["votings"]
	if err := h.j.Refresh(ctx, "2026-11-20", false); err != nil || h.src.calls["votings"] != calls {
		t.Fatalf("a refreshed day must not be refreshed again (err=%v)", err)
	}
	if err := h.j.Refresh(ctx, "2026-11-20", true); err != nil || h.src.calls["votings"] != calls+1 {
		t.Fatalf("force must refresh again (err=%v)", err)
	}
}

func TestRunDueFullCycleOnAFreshDatabase(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()

	// 00:00 pulse not out yet: the monthly work still runs.
	h.pulses.ready = false
	if err := h.j.RunDue(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := h.st.GetDayChoice(ctx, "2026-11-20"); !errors.Is(err, store.ErrNotFound) {
		t.Fatal("no choice expected without a pulse")
	}

	h.pulses.ready = true
	if err := h.j.RunDue(ctx); err != nil {
		t.Fatal(err)
	}
	choice, err := h.st.GetDayChoice(ctx, "2026-11-20")
	if err != nil {
		t.Fatal(err)
	}
	if done, _ := h.st.IsDone(ctx, JobRefresh, "senate:2", "2026-11-20"); !done {
		t.Fatal("the chosen congressman should be refreshed")
	}
	if v, _ := h.st.ListVotings(ctx, choice.CongressmanID); len(v) != 1 {
		t.Fatal("votings of the chosen congressman missing")
	}
	for _, job := range []string{JobProfiles, JobExpenses} {
		if done, _ := h.st.IsDone(ctx, job, allKey, "2026-11"); !done {
			t.Fatalf("%s not complete", job)
		}
	}

	// Everything is done: another tick must not touch the sources or the beacon.
	srcCalls, pulseCalls := map[string]int{}, h.pulses.calls
	for k, v := range h.src.calls {
		srcCalls[k] = v
	}
	if err := h.j.RunDue(ctx); err != nil {
		t.Fatal(err)
	}
	for k, v := range h.src.calls {
		if v != srcCalls[k] {
			t.Fatalf("idle tick called source.%s", k)
		}
	}
	if h.pulses.calls != pulseCalls {
		t.Fatal("idle tick asked the beacon")
	}

	// Next day: a new draw and refresh, monthly work stays done.
	h.now = h.now.AddDate(0, 0, 1)
	if err := h.j.RunDue(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := h.st.GetDayChoice(ctx, "2026-11-21"); err != nil {
		t.Fatalf("next day not drawn: %v", err)
	}
	if h.src.calls["profile"] != srcCalls["profile"] {
		t.Fatal("profiles were synced again within the same month")
	}
}

func TestServeStopsWhenCancelled(t *testing.T) {
	h := newHarness(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	go func() { done <- h.j.Serve(ctx, time.Hour) }()
	time.Sleep(200 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not stop")
	}
	if _, err := h.st.GetDayChoice(context.Background(), "2026-11-20"); err != nil {
		t.Fatalf("Serve should have run once immediately: %v", err)
	}
}
