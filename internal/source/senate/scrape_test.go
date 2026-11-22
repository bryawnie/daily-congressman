package senate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

var _ congress.Source = (*Source)(nil)

var (
	balladares = congress.Member{Chamber: congress.Senate, ExternalID: 1501, Name: "Andrea Balladares Letelier"}
	edwards    = congress.Member{Chamber: congress.Senate, ExternalID: 1328, Name: "Rojo Edwards"}
	fixedNow   = func() time.Time { return time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC) }
)

const hemicycleJSON = `{"status":"ok","data":{"parlamentarios":{"total":2,"data":[
 {"ID_PARLAMENTARIO":1501,"SLUG":"andrea-balladares-letelier-sen","NOMBRE":"Andrea","APELLIDO_PATERNO":"Balladares","APELLIDO_MATERNO":"Letelier","CAMARA":"S","PARTIDO":"R.N.","CIRCUNSCRIPCION_ID":9,"REGION":"Región del Maule","SEXO_ETIQUETA":"Mujer","IMAGEN_450":"http://img/1501.jpg","PERIODOS":[{"CAMARA":"S","DESDE":"2026","HASTA":"2030","VIGENTE":1}]},
 {"ID_PARLAMENTARIO":1328,"SLUG":"rojo-edwards-silva-sen","NOMBRE":"Rojo","APELLIDO_PATERNO":"Edwards","APELLIDO_MATERNO":"  ","CAMARA":"S","PARTIDO":"Independiente","CIRCUNSCRIPCION_ID":7,"REGION":"Región Metropolitana de Santiago","SEXO_ETIQUETA":"Hombre","IMAGEN_450":"","PERIODOS":[{"CAMARA":"S","DESDE":"2022","HASTA":"2030","VIGENTE":1}]}
]}}}`

const legislaturesJSON = `{"status":"ok","data":[
 {"ID_LEGISLATURA":507,"NUMERO":374,"INICIO":"11/03/2026","TERMINO":"10/03/2027"},
 {"ID_LEGISLATURA":505,"NUMERO":373,"INICIO":"11/03/2025","TERMINO":"10/03/2026"}]}`

// pageHTML builds a senator's page with the data the website embeds.
func pageHTML(t *testing.T, summary, body, twitter, instagram string) string {
	t.Helper()
	data := map[string]any{"props": map[string]any{"pageProps": map[string]any{"resource": map[string]any{"data": map[string]any{
		"parliamentarianDrupalData": map[string]any{"body": map[string]any{"value": body, "summary": summary}},
		"parliamentarianSenadoData": map[string]any{"ENLACES": []map[string]any{{"TWITTER": twitter, "INSTAGRAM": instagram}}},
	}}}}}
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return `<html><head><script id="__NEXT_DATA__" type="application/json">` + string(b) + `</script></head></html>`
}

// fake serves canned responses by path and records the requests.
type fake struct {
	t        *testing.T
	routes   map[string]func(r *http.Request) string
	requests []string
}

func newFake(t *testing.T, routes map[string]func(r *http.Request) string) (*fake, *Source) {
	t.Helper()
	f := &fake{t: t, routes: routes}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests = append(f.requests, r.URL.RequestURI())
		h, ok := f.routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(h(r)))
	}))
	t.Cleanup(srv.Close)
	return f, New(WithAPIURL(srv.URL), WithSiteURL(srv.URL), WithClock(fixedNow))
}

func static(s string) func(*http.Request) string { return func(*http.Request) string { return s } }

func TestAttendance(t *testing.T) {
	f, s := newFake(t, map[string]func(*http.Request) string{
		"/api/legislatures": static(legislaturesJSON),
		"/api/parliamentarians_attendance_room": static(
			`{"status":"ok","data":{"total":88,"asiste":81,"justificacion":2,"sin_justificacion":5,"data":[]}}`),
	})
	got, err := s.Attendance(context.Background(), balladares)
	if err != nil {
		t.Fatal(err)
	}
	want := congress.Attendance{Total: 88, Present: 81, JustifiedAbsent: 2, UnjustifiedAbsent: 5}
	if got != want {
		t.Fatalf("attendance = %+v, want %+v", got, want)
	}
	last := f.requests[len(f.requests)-1]
	if !strings.Contains(last, "id_parlamentario=1501") || !strings.Contains(last, "id_legislatura=507") {
		t.Fatalf("request = %s", last)
	}
}

func TestRejectsMembersOfOtherChambers(t *testing.T) {
	_, s := newFake(t, map[string]func(*http.Request) string{})
	deputy := congress.Member{Chamber: congress.Deputies, ExternalID: 1}
	if _, err := s.Attendance(context.Background(), deputy); err == nil {
		t.Fatal("expected an error for a deputy")
	}
	if _, err := s.Votings(context.Background(), congress.Member{Chamber: congress.Senate}, 10); err == nil {
		t.Fatal("expected an error for a member without id")
	}
}

func TestCurrentLegislatureErrors(t *testing.T) {
	_, s := newFake(t, map[string]func(*http.Request) string{"/api/legislatures": static(legislaturesJSON)})
	s.now = func() time.Time { return time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC) }
	if _, err := s.currentLegislature(context.Background()); err == nil {
		t.Fatal("expected an error when no legislature contains the date")
	}
	// The last day of a legislature is still part of it.
	s.now = func() time.Time { return time.Date(2027, 3, 10, 23, 30, 0, 0, time.UTC) }
	if l, err := s.currentLegislature(context.Background()); err != nil || l.ID != 507 {
		t.Fatalf("last day: %+v, %v", l, err)
	}
}

func motionsRoute(byLegislature map[string][]string, total map[string]int) func(*http.Request) string {
	return func(r *http.Request) string {
		q := r.URL.Query()
		leg := q.Get("id_legislatura")
		offset, _ := strconv.Atoi(q.Get("offset"))
		rows := byLegislature[leg]
		if offset < len(rows) {
			rows = rows[offset : offset+1] // one row per page, to exercise paging
		} else {
			rows = nil
		}
		return fmt.Sprintf(`{"status":"ok","data":{"total":%d,"data":[%s]}}`, total[leg], strings.Join(rows, ","))
	}
}

func TestActivityReadsTheLegislaturesOfTheCurrentPeriod(t *testing.T) {
	motion := func(b, date string) string {
		return fmt.Sprintf(`{"BOLETIN":%q,"SUMA":"Resumen %s","FECHA":%q,"ESTADO":"En tramitación"}`, b, b, date)
	}
	route := motionsRoute(
		map[string][]string{
			"507": {motion("100-01", "07/07/2026"), motion("101-01", "03/07/2026")},
			"505": {motion("90-01", "02/05/2025")},
		},
		map[string]int{"507": 2, "505": 1},
	)
	_, s := newFake(t, map[string]func(*http.Request) string{
		"/api/hemicycle":                static(hemicycleJSON),
		"/api/legislatures":             static(legislaturesJSON),
		"/api/parliamentarians/motions": route,
	})

	// Balladares' period starts in 2026: only legislature 374, read in two pages.
	got, err := s.Activity(context.Background(), balladares)
	if err != nil {
		t.Fatal(err)
	}
	want := []congress.LawProject{
		{Bulletin: "100-01", Name: "Resumen 100-01", Type: "Moción", Status: "En tramitación", Date: "2026-07-07"},
		{Bulletin: "101-01", Name: "Resumen 101-01", Type: "Moción", Status: "En tramitación", Date: "2026-07-03"},
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("activity = %+v, want %+v", got, want)
	}

	// Edwards' period starts in 2022: both legislatures.
	got, err = s.Activity(context.Background(), edwards)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 motions across 2 legislatures, got %+v", got)
	}
}

func TestActivityRejectsMotionsWithoutBulletin(t *testing.T) {
	_, s := newFake(t, map[string]func(*http.Request) string{
		"/api/hemicycle":    static(hemicycleJSON),
		"/api/legislatures": static(legislaturesJSON),
		"/api/parliamentarians/motions": static(
			`{"status":"ok","data":{"total":1,"data":[{"SUMA":"x","FECHA":"01/01/2026"}]}}`),
	})
	if _, err := s.Activity(context.Background(), balladares); err == nil {
		t.Fatal("expected an error")
	}
}

func TestProfile(t *testing.T) {
	page := pageHTML(t,
		"<p>Andrea Balladares Letelier (Chanco, 30 de agosto de 1987) es socióloga y militante del partido Renovación Nacional (RN).</p>",
		"<p>Andrea nació en Chanco el 30 de agosto de 1987.</p>",
		"@AndreaBal  ", "https://www.instagram.com/andreabal/channel/ ")
	f, s := newFake(t, map[string]func(*http.Request) string{
		"/api/hemicycle": static(hemicycleJSON),
		"/senadoras-y-senadores/listado-de-senadoras-y-senadores/andrea-balladares-letelier-sen": static(page),
	})
	got, err := s.Profile(context.Background(), balladares)
	if err != nil {
		t.Fatal(err)
	}
	want := congress.Congressman{
		Chamber: congress.Senate, ExternalID: 1501,
		FirstName: "Andrea", FirstSurname: "Balladares", SecondSurname: "Letelier",
		ProfilePicture: "http://img/1501.jpg", Gender: "FEMALE", Birthdate: "1987-08-30", Profession: "socióloga",
		Party: "Partido Renovación Nacional", PartyAlias: "RN",
		TwitterUsername: "AndreaBal", InstagramUsername: "andreabal",
		ConstituencyKind: congress.ConstituencyRegion, ConstituencyName: "Región del Maule",
		ConstituencyDetail: "Circunscripción 9", LastUpdate: fixedNow(),
	}
	if got != want {
		t.Fatalf("profile = %+v\nwant      %+v", got, want)
	}
	if len(f.requests) < 2 || !strings.HasSuffix(f.requests[len(f.requests)-1], "andrea-balladares-letelier-sen") {
		t.Fatalf("requests = %v", f.requests)
	}
}

func TestProfileOfAnIndependentWithTheOtherSummaryFormat(t *testing.T) {
	page := pageHTML(t,
		"Rojo Edwards (Chicago, Estados Unidos, 15 de julio de 1977). Es ingeniero Civil Industrial y político independiente. Es senador por la 7a Circunscripción.",
		"", "", "")
	_, s := newFake(t, map[string]func(*http.Request) string{
		"/api/hemicycle": static(hemicycleJSON),
		"/senadoras-y-senadores/listado-de-senadoras-y-senadores/rojo-edwards-silva-sen": static(page),
	})
	got, err := s.Profile(context.Background(), edwards)
	if err != nil {
		t.Fatal(err)
	}
	if got.Birthdate != "1977-07-15" || got.Profession != "ingeniero Civil Industrial" ||
		got.Party != "Independiente" || got.PartyAlias != "" || got.Gender != "MALE" || got.SecondSurname != "" {
		t.Fatalf("profile = %+v", got)
	}
}

func TestProfileErrors(t *testing.T) {
	t.Run("page without embedded data", func(t *testing.T) {
		_, s := newFake(t, map[string]func(*http.Request) string{
			"/api/hemicycle": static(hemicycleJSON),
			"/senadoras-y-senadores/listado-de-senadoras-y-senadores/andrea-balladares-letelier-sen": static("<html></html>"),
		})
		if _, err := s.Profile(context.Background(), balladares); err == nil {
			t.Fatal("expected an error")
		}
	})
	t.Run("page not found", func(t *testing.T) {
		_, s := newFake(t, map[string]func(*http.Request) string{"/api/hemicycle": static(hemicycleJSON)})
		if _, err := s.Profile(context.Background(), balladares); err == nil {
			t.Fatal("expected an error")
		}
	})
	t.Run("not a sitting senator", func(t *testing.T) {
		_, s := newFake(t, map[string]func(*http.Request) string{"/api/hemicycle": static(hemicycleJSON)})
		other := congress.Member{Chamber: congress.Senate, ExternalID: 9999}
		if _, err := s.Profile(context.Background(), other); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestPageParsingHelpers(t *testing.T) {
	cases := []struct {
		name, summary, body string
		birth, profession   string
		party, acronym      string
	}{
		{"parenthesis format", "X (Antofagasta, 16 de junio de 1974) es abogado y militante del Partido por la Democracia (PPD).", "",
			"1974-06-16", "abogado", "Partido por la Democracia", "PPD"},
		{"political of the party", "X (Santiago, 5 de septiembre de 1978). Es abogado y político del Partido Republicano de Chile (REP).", "",
			"1978-09-05", "abogado", "Partido Republicano de Chile", "REP"},
		{"birth only in the body", "", "Nació en Chanco el 30 de agosto de 1987.", "1987-08-30", "", "", ""},
		{"independent without profession", "X (Temuco, 1 de enero de 1980) es independiente.", "", "1980-01-01", "", "", ""},
		{"nothing known", "Texto sin datos", "Otro texto", "", "", "", ""},
		// Real summaries, whose formats differ from each other.
		{"no 'es' and no acronym",
			"Carlos Kuschel Silva (Frutillar, 16 de marzo de 1953). Ingeniero comercial y militante del Partido Renovación Nacional. Es senador por 13° Circunscripción, región de Los Lagos, por el período 2022-2030. Fue diputado entre 2018-2022 y 1990-2006 y senador (2006- 2014).",
			"", "1953-03-16", "Ingeniero comercial", "Partido Renovación Nacional", ""},
		{"independent without 'es'",
			"Alejandro Kusanovic Glusevic (Punta Arenas, 16 de diciembre de 1960). Ingeniero Civil Mecánico y político independiente. Es Senador por la 15a Circunscripción Electoral, región de Magallanes, por el periodo 2022-2030. Fue consejero regional (2013-2017 y 2017-2021) y presidente del mismo (2019- 2020).",
			"", "1960-12-16", "Ingeniero Civil Mecánico", "", ""},
		{"several occupations and a 'political of'",
			"Vanessa Kaiser Barents-Von Hohenhagen (Santiago de Chile, 25 de octubre de 1977). Es doctora en Ciencia Política y Filosofía, periodista y política del Partido Nacional Libertario. Senadora por la 11a Circunscripción.",
			"", "1977-10-25", "doctora en Ciencia Política y Filosofía", "Partido Nacional Libertario", ""},
		{"'and current president of the party'",
			"Paulina Vodanovic Rojas (Santiago, 20 de septiembre de 1971). Es abogada, militante y actual presidenta del Partido Socialista de Chile (PS). Es senadora por la 9a Circunscripción Electoral, región del Maule, en el período 2026-2034.",
			"", "1971-09-20", "abogada", "Partido Socialista de Chile", "PS"},
		{"'and parliamentarian' is not part of the occupation",
			"Alejandra Sepúlveda Orbenes (Viña del Mar, 13 de noviembre de 1965). Es médica veterinaria y parlamentaria independiente. Es senadora por la 8ª Circunscripción Electoral.",
			"", "1965-11-13", "médica veterinaria", "", ""},
		{"only a party, no occupation",
			"Iván Moreira Barros (Santiago, 8 de diciembre de 1956). Es militante del Partido Unión Demócrata Independiente (UDI). Es senador por la 13a Circunscripción Electoral.",
			"", "1956-12-08", "", "Partido Unión Demócrata Independiente", "UDI"},
		{"the office is not an occupation",
			"X (Santiago, 5 de mayo de 1970). Senador por la 3ª Circunscripción Electoral, región de Antofagasta, período 2026-2034.",
			"", "1970-05-05", "", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := senatorPage{Summary: c.summary, Body: c.body}
			if got := p.birthdate(); got != c.birth {
				t.Errorf("birthdate = %q, want %q", got, c.birth)
			}
			if got := p.profession(); got != c.profession {
				t.Errorf("profession = %q, want %q", got, c.profession)
			}
			if n, a := p.party(); n != c.party || a != c.acronym {
				t.Errorf("party = %q/%q, want %q/%q", n, a, c.party, c.acronym)
			}
		})
	}

	for in, want := range map[string]string{
		"@ArayaPedro":                              "ArayaPedro",
		"@RojoEdwards  ":                           "RojoEdwards",
		"https://twitter.com/user":                 "user",
		"https://x.com/user/":                      "user",
		"https://www.instagram.com/name/channel/ ": "name",
		"https://www.instagram.com/name?igsh=abc":  "name",
		"https://www.instagram.com/":               "",
		"":                                         "",
	} {
		if got := username(in); got != want {
			t.Errorf("username(%q) = %q, want %q", in, got, want)
		}
	}
}

const votesJSON = `{"status":"ok","data":{"total":3,"data":[
 {"ID_VOTACION":3,"FECHA_VOTACION":"20-11-2026 02:18:49","TEMA":"Indicación N° 73 (Mayoría Simple)\nAutores: Ejecutivo\nTexto largo","QUORUM":"Mayoría simple","BOLETIN":"18702-05","SI":2,"NO":1,"ABS":0,
  "VOTACIONES":{"SI":[{"PARLID":1501},{"PARLID":1328}],"NO":[{"PARLID":9}],"ABSTENCION":0,"PAREO":0}},
 {"ID_VOTACION":2,"FECHA_VOTACION":"19-11-2026 10:00:00","TEMA":"Otra votación","QUORUM":"Mayoría simple","BOLETIN":"18702-05","SI":1,"NO":0,"ABS":0,
  "VOTACIONES":{"SI":[{"PARLID":1328}],"NO":0,"ABSTENCION":0,"PAREO":0}},
 {"ID_VOTACION":1,"FECHA_VOTACION":"18-11-2026 09:00:00","TEMA":"Votación con pareo","QUORUM":"Dos tercios Q.C.","BOLETIN":null,"SI":0,"NO":0,"ABS":1,
  "VOTACIONES":{"SI":0,"NO":0,"ABSTENCION":[{"PARLID":1328}],"PAREO":[{"PARLID":1501}]}}
]}}`

func TestVotingsReturnsTheVotesWhereTheSenatorAppears(t *testing.T) {
	f, s := newFake(t, map[string]func(*http.Request) string{
		"/api/hemicycle":    static(hemicycleJSON), // 2 sitting senators
		"/api/legislatures": static(legislaturesJSON),
		"/api/votes":        static(votesJSON),
	})
	got, err := s.Votings(context.Background(), balladares, 10)
	if err != nil {
		t.Fatal(err)
	}
	// Voting 2 does not list her (absent), so it is skipped.
	if len(got) != 2 || got[0].VotingID != 3 || got[1].VotingID != 1 {
		t.Fatalf("votings = %+v", got)
	}
	v := got[0]
	if v.Date != "2026-11-20" || v.BulletinNumber != "18702-05" || v.VotedOption != "Afirmativo" ||
		v.DocumentTitle != "Indicación N° 73 (Mayoría Simple)" || v.ArticleText != "Autores: Ejecutivo Texto largo" ||
		v.TotalApproved != 2 || v.TotalRejected != 1 || v.TotalAbstention != 0 ||
		v.Quorum != "Mayoría simple" || v.Result != "Aprobado" {
		t.Fatalf("voting = %+v", v)
	}
	// Two thirds of the 2 sitting senators needs both votes in favor.
	if got[1].VotedOption != "Pareo" || got[1].BulletinNumber != "" || got[1].Quorum != "Dos tercios Q.C." || got[1].Result != "Rechazado" {
		t.Fatalf("paired voting = %+v", got[1])
	}
	if !strings.Contains(f.requests[len(f.requests)-1], "id_legislatura=507") {
		t.Fatalf("requests = %v", f.requests)
	}

	// The other senator abstained in voting 1 and voted yes in 3 and 2.
	got, _ = s.Votings(context.Background(), edwards, 10)
	if len(got) != 3 || got[2].VotedOption != "Abstención" {
		t.Fatalf("edwards votings = %+v", got)
	}
}

func TestVotingsStopsAtTheLimitAndWalksPages(t *testing.T) {
	// 120 votings in the legislature; the senator is listed in all of them.
	route := func(r *http.Request) string {
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		var rows []string
		for i := offset; i < offset+votesPageSize && i < 120; i++ {
			id := 1000 - i
			who := `[{"PARLID":1501}]`
			if i < 60 { // she is absent from the first 60
				who = `0`
			}
			rows = append(rows, fmt.Sprintf(
				`{"ID_VOTACION":%d,"FECHA_VOTACION":"20-11-2026 02:18:49","TEMA":"t","BOLETIN":"1-1","SI":1,"NO":0,"ABS":0,"VOTACIONES":{"SI":%s,"NO":0,"ABSTENCION":0,"PAREO":0}}`, id, who))
		}
		return fmt.Sprintf(`{"status":"ok","data":{"total":120,"data":[%s]}}`, strings.Join(rows, ","))
	}
	f, s := newFake(t, map[string]func(*http.Request) string{
		"/api/hemicycle":    static(hemicycleJSON),
		"/api/legislatures": static(legislaturesJSON),
		"/api/votes":        route,
	})
	got, err := s.Votings(context.Background(), balladares, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 10 || got[0].VotingID != 1000-60 {
		t.Fatalf("got %d votings, first %+v", len(got), got[0])
	}
	voteCalls := 0
	for _, r := range f.requests {
		if strings.HasPrefix(r, "/api/votes") {
			voteCalls++
		}
	}
	if voteCalls != 2 {
		t.Fatalf("expected 2 pages, got %d (%v)", voteCalls, f.requests)
	}
	if _, err := s.Votings(context.Background(), balladares, 0); err == nil {
		t.Fatal("expected an error for a non positive limit")
	}
}

func TestVoterListAcceptsNumbersAndLists(t *testing.T) {
	var l voterList
	for _, in := range []string{`0`, `null`, `[]`, ` 0 `} {
		if err := json.Unmarshal([]byte(in), &l); err != nil || len(l) != 0 {
			t.Errorf("%s -> %v, %v", in, l, err)
		}
	}
	if err := json.Unmarshal([]byte(`[{"PARLID":5}]`), &l); err != nil || !l.has(5) || l.has(6) {
		t.Errorf("list -> %v, %v", l, err)
	}
	if err := json.Unmarshal([]byte(`3`), &l); err == nil {
		t.Error("a non zero number would hide voters and must fail")
	}
}

func TestMatchUnit(t *testing.T) {
	rows := []transparencyRow{
		{Unit: 1, FirstName: "ANDREA PAZ", FirstSurname: "BALLADARES", SecondSurname: "LETELIER"},
		{Unit: 2, FirstName: "JOSE ROJO", FirstSurname: "EDWARDS", SecondSurname: "SILVA"},
		{Unit: 3, FirstName: "MARIA LORETO", FirstSurname: "CARVAJAL", SecondSurname: "AMBIADO"},
		{Unit: 4, FirstName: "DANIEL", FirstSurname: "NÚÑEZ", SecondSurname: "ARANCIBIA"},
		{Unit: 5, FirstName: "PAULINA", FirstSurname: "NÚÑEZ", SecondSurname: "URRUTIA"},
		{Unit: 6, FirstName: "CAMILA", FirstSurname: "CRUZ-COKE", SecondSurname: "X"},
		{Unit: 7, FirstName: "LUIS ARTURO", FirstSurname: "SQUELLA", SecondSurname: "OVALLE"},
	}
	rec := func(first, p, m string) senatorRecord {
		return senatorRecord{FirstName: first, FirstSurname: p, SecondSurname: m}
	}
	cases := []struct {
		name string
		rec  senatorRecord
		unit int
		ok   bool
	}{
		{"exact", rec("Andrea", "Balladares", "Letelier"), 1, true},
		{"blank second surname, other given names", rec("Rojo", "Edwards", "  "), 2, true},
		{"given name differs, unique surname", rec("Loreto", "Carvajal", "Ambiado"), 3, true},
		{"same first surname, told apart by the second", rec("Paulina", "Núñez", "Urrutia"), 5, true},
		{"accents and hyphens", rec("Camila", "Cruz Coke", "X"), 6, true},
		{"second given name only", rec("Arturo", "Squella", "Ovalle"), 7, true},
		{"not listed", rec("Zoe", "Nadie", "Nada"), 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			unit, ok, err := matchUnit(c.rec, rows)
			if err != nil || ok != c.ok || unit != c.unit {
				t.Fatalf("matchUnit = %d, %v, %v; want %d, %v", unit, ok, err, c.unit, c.ok)
			}
		})
	}
	// Two senators the data cannot tell apart must fail instead of guessing.
	twins := []transparencyRow{
		{Unit: 8, FirstName: "JUAN", FirstSurname: "PEREZ", SecondSurname: "SOTO"},
		{Unit: 9, FirstName: "JUAN", FirstSurname: "PEREZ", SecondSurname: "SOTO"},
	}
	if _, _, err := matchUnit(rec("Juan", "Pérez", "Soto"), twins); err == nil {
		t.Fatal("expected an ambiguity error")
	}
}

func expensesRoutes(t *testing.T, wrongUnit bool) map[string]func(*http.Request) string {
	return map[string]func(*http.Request) string{
		"/api/hemicycle": static(hemicycleJSON),
		"/api/transparency/available-periods": static(`{"status":"ok","data":{"data":[
			{"attributes":{"ano":2026,"mes":1}},{"attributes":{"ano":2026,"mes":2}},{"attributes":{"ano":2026,"mes":3}},
			{"attributes":{"ano":2026,"mes":4}},{"attributes":{"ano":2026,"mes":5}},{"attributes":{"ano":2026,"mes":6}},
			{"attributes":{"ano":2026,"mes":7}},{"attributes":{"ano":2026,"mes":8}},{"attributes":{"ano":2026,"mes":8}},
			{"attributes":{"ano":2026,"mes":12}}]}}`),
		"/api/transparency/parlamentarians-periods": func(r *http.Request) string {
			if r.URL.Query().Get("month") == "4" { // she is not listed that month
				return `{"status":"ok","data":[{"unidad_ejecutora":103,"nombre":"JOSE ROJO","appaterno":"EDWARDS","apmaterno":"SILVA"}]}`
			}
			return `{"status":"ok","data":[
				{"unidad_ejecutora":102,"nombre":"ANDREA PAZ","appaterno":"BALLADARES","apmaterno":"LETELIER"},
				{"unidad_ejecutora":103,"nombre":"JOSE ROJO","appaterno":"EDWARDS","apmaterno":"SILVA"}]}`
		},
		"/api/transparency/expenses/senator-Operational-expenses": func(r *http.Request) string {
			q := r.URL.Query()
			unit := q.Get("filters[unidad_ejecutora][$eq]")
			year, month := q.Get("filters[ano][$eq]"), q.Get("filters[mes][$eq]")
			returned := unit
			if wrongUnit {
				returned = "999"
			}
			row := func(concept string, amount string) string {
				return fmt.Sprintf(`{"id":1,"attributes":{"ano":%s,"mes":%s,"gastos_operacionales":%q,"monto":%s,"unidad_ejecutora":%s}}`,
					year, month, concept, amount, returned)
			}
			return `{"status":"ok","data":{"data":[` + strings.Join([]string{
				row("ARRIENDO OFICINAS", "2362672"), row("CORRESPONDENCIA", "0"), row("VALOR TOTAL", "2362672"),
			}, ",") + `],"meta":{"pagination":{"pageCount":1}}}}`
		},
	}
}

func TestExpenses(t *testing.T) {
	f, s := newFake(t, expensesRoutes(t, false))
	got, err := s.Expenses(context.Background(), balladares)
	if err != nil {
		t.Fatal(err)
	}
	// Latest six published months not after November 2026: 8, 7, 6, 5, 4, 3.
	// Month 4 is skipped because she is not listed; the future month 12 is ignored.
	months := map[int]int{}
	for _, e := range got.Operational {
		months[e.Month]++
		if e.ExpenseType == "VALOR TOTAL" {
			t.Fatal("the total row must not be stored")
		}
	}
	if len(months) != 5 || months[4] != 0 || months[12] != 0 || months[3] != 2 || months[8] != 2 {
		t.Fatalf("months = %v", months)
	}
	if got.Operational[0] != (congress.OperationalExpense{Year: 2026, Month: 8, ExpenseType: "ARRIENDO OFICINAS", Amount: 2362672}) {
		t.Fatalf("first row = %+v", got.Operational[0])
	}
	if got.Offices != nil || got.Staff != nil {
		t.Fatal("the Senate does not publish offices or staff separately")
	}
	var sawUnit bool
	for _, r := range f.requests {
		if strings.Contains(r, "senator-Operational-expenses") && strings.Contains(r, "unidad_ejecutora%5D%5B%24eq%5D=102") {
			sawUnit = true
		}
	}
	if !sawUnit {
		t.Fatalf("the expenses were not requested with unit 102: %v", f.requests)
	}
}

func TestExpensesRejectRowsOfAnotherSenator(t *testing.T) {
	_, s := newFake(t, expensesRoutes(t, true))
	if _, err := s.Expenses(context.Background(), balladares); err == nil {
		t.Fatal("rows of a different unit mean the filter was ignored and must fail")
	}
}

func TestExpensesRejectPaginatedResults(t *testing.T) {
	routes := expensesRoutes(t, false)
	routes["/api/transparency/expenses/senator-Operational-expenses"] = static(
		`{"status":"ok","data":{"data":[],"meta":{"pagination":{"pageCount":2}}}}`)
	_, s := newFake(t, routes)
	if _, err := s.Expenses(context.Background(), balladares); err == nil {
		t.Fatal("expected an error when the rows span several pages")
	}
}

func TestSplitTopicAndSlashDate(t *testing.T) {
	title, rest := splitTopic("  Título\n\n  línea 1 \nlínea 2\n")
	if title != "Título" || rest != "línea 1 línea 2" {
		t.Fatalf("splitTopic = %q, %q", title, rest)
	}
	if title, rest := splitTopic("solo"); title != "solo" || rest != "" {
		t.Fatalf("splitTopic = %q, %q", title, rest)
	}
	if d, err := slashDate("07/07/2026"); err != nil || d != "2026-07-07" {
		t.Fatalf("slashDate = %q, %v", d, err)
	}
	if d, err := slashDate(""); err != nil || d != "" {
		t.Fatalf("slashDate empty = %q, %v", d, err)
	}
	if _, err := slashDate("2026-07-07"); err == nil {
		t.Fatal("expected an error for another format")
	}
}

func TestOutcome(t *testing.T) {
	cases := []struct {
		quorum  string
		yes, no int
		seats   int
		want    string
	}{
		{"Mayoría simple", 13, 11, 50, "Aprobado"},
		{"Mayoría simple", 11, 13, 50, "Rechazado"},
		{"Mayoría simple", 12, 12, 50, "Empate"},
		{"Mayoría simple", 1, 0, 50, "Aprobado"},
		{"Mayoría simple", 0, 0, 50, "Empate"},

		// Qualified quorum: absolute majority of the senators in office (26 of 50).
		{"Q.C.", 26, 0, 50, "Aprobado"},
		{"Q.C.", 25, 0, 50, "Rechazado"},
		{"Q.C.", 25, 0, 49, "Aprobado"}, // 25 of 49 is already a majority
		{"Q.C.", 49, 1, 50, "Aprobado"},

		// Four sevenths of 50 is 28.57, so 29 votes.
		{"Cuatro séptimos Q.C.", 29, 10, 50, "Aprobado"},
		{"Cuatro séptimos Q.C.", 28, 10, 50, "Rechazado"},
		// Two thirds of 50 is 33.3, so 34 votes.
		{"Dos tercios Q.C.", 34, 5, 50, "Aprobado"},
		{"Dos tercios Q.C.", 33, 5, 50, "Rechazado"},
		{"Dos tercios", 34, 5, 50, "Aprobado"},
		{"Dos tercios", 25, 20, 50, "Rechazado"},
		// Three fifths of 50 is 30.
		{"Tres quintos Q.C.", 30, 10, 50, "Aprobado"},
		{"Tres quintos Q.C.", 29, 10, 50, "Rechazado"},

		// The majority of the votes cast does not meet a special quorum.
		{"Dos tercios Q.C.", 10, 0, 50, "Rechazado"},

		// Unknown or empty quorums are not guessed.
		{"Quórum desconocido", 40, 0, 50, ""},
		{"", 40, 0, 50, ""},
	}
	for _, c := range cases {
		if got := outcome(c.quorum, c.yes, c.no, c.seats); got != c.want {
			t.Errorf("outcome(%q, %d yes, %d no, %d seats) = %q, want %q", c.quorum, c.yes, c.no, c.seats, got, c.want)
		}
	}
}
