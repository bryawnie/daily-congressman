package senate

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

const senatorPagePath = "/senadoras-y-senadores/listado-de-senadoras-y-senadores/"

var (
	nextDataRe = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>(.*?)</script>`)
	tagRe      = regexp.MustCompile(`<[^>]*>`)

	// "(Antofagasta, 16 de junio de 1974)" in the summary of the biography.
	summaryBirthRe = regexp.MustCompile(`(?i)\(([^()]*?),?\s*(\d{1,2})\s+de\s+([a-záéíóú]+)\s+de\s+(\d{4})\)`)
	// "nació en Antofagasta el 16 de junio de 1974" in the biography.
	bodyBirthRe = regexp.MustCompile(`(?i)naci[óo][^.]*?(\d{1,2})\s+de\s+([a-záéíóú]+)\s+de\s+(\d{4})`)
	// The summaries do not share one format. After the birth data they say, for
	// example, ") es abogado y militante del ...", "). Es ingeniero civil
	// industrial y político independiente." or "). Ingeniero comercial y
	// militante del ...", with or without "es".
	professionRe = regexp.MustCompile(`(?i)\)[.,]?\s*(?:(?:es|fue)\s+(?:una?\s+)?)?(.+?)(?:\s+y\s+(?:militante|pol[ií]tic[oa]|parlamentari[oa])|\s+e\s+independiente|\s+militante|\s+independiente|[.,;])`)
	// "militante del Partido por la Democracia (PPD)", "político del Partido
	// Republicano de Chile (REP)", "militante y actual presidenta del Partido
	// Socialista de Chile (PS)" or "militante del Partido Renovación Nacional."
	// (without acronym). The name stops at the acronym or at the end of the
	// sentence, so it cannot reach a later parenthesis.
	partyRe = regexp.MustCompile(`(?i)(?:militante|pol[ií]tic[oa])\s+(?:y\s+actual\s+\w+\s+)?d(?:el|e la)\s+([^.()]+?)(?:\s*\(([^)]*)\))?\s*(?:[.,;]|$)`)
)

var months = map[string]string{
	"enero": "01", "febrero": "02", "marzo": "03", "abril": "04", "mayo": "05", "junio": "06",
	"julio": "07", "agosto": "08", "septiembre": "09", "setiembre": "09", "octubre": "10",
	"noviembre": "11", "diciembre": "12",
}

// senatorPage holds what the senator's page adds to the list: biography
// facts and social links.
type senatorPage struct {
	Summary string
	Body    string
	Links   struct {
		Twitter   string `json:"TWITTER"`
		Instagram string `json:"INSTAGRAM"`
	}
}

type nextData struct {
	Props struct {
		PageProps struct {
			Resource struct {
				Data struct {
					Drupal struct {
						Body struct {
							Value   string `json:"value"`
							Summary string `json:"summary"`
						} `json:"body"`
					} `json:"parliamentarianDrupalData"`
					Senate struct {
						Links []struct {
							Twitter   string `json:"TWITTER"`
							Instagram string `json:"INSTAGRAM"`
						} `json:"ENLACES"`
					} `json:"parliamentarianSenadoData"`
				} `json:"data"`
			} `json:"resource"`
		} `json:"pageProps"`
	} `json:"props"`
}

// parseSenatorPage reads the data the website embeds in the senator's page.
// It uses the page, and not the "_next/data/<build id>/..." address, because
// that address changes with every deployment of the website.
func parseSenatorPage(page []byte) (senatorPage, error) {
	m := nextDataRe.FindSubmatch(page)
	if m == nil {
		return senatorPage{}, fmt.Errorf("senate: senator page has no embedded data")
	}
	var nd nextData
	if err := json.Unmarshal(m[1], &nd); err != nil {
		return senatorPage{}, fmt.Errorf("senate: parse senator page data: %w", err)
	}
	d := nd.Props.PageProps.Resource.Data
	out := senatorPage{
		Summary: plainText(d.Drupal.Body.Summary),
		Body:    plainText(d.Drupal.Body.Value),
	}
	if len(d.Senate.Links) > 0 {
		out.Links.Twitter = d.Senate.Links[0].Twitter
		out.Links.Instagram = d.Senate.Links[0].Instagram
	}
	if out.Summary == "" && out.Body == "" {
		return senatorPage{}, fmt.Errorf("senate: senator page has no biography")
	}
	return out, nil
}

// cleanText decodes the HTML entities that the Senate leaves in some texts
// (for example &quot; in the summary of a motion) and trims the spaces. Unlike
// plainText it keeps the line breaks, which separate the title of a voting
// from its text.
func cleanText(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

// plainText removes the HTML tags and entities and collapses the whitespace.
func plainText(s string) string {
	s = html.UnescapeString(tagRe.ReplaceAllString(s, " "))
	return strings.Join(strings.Fields(s), " ")
}

func isoDate(day, month, year string) string {
	mm, ok := months[strings.ToLower(month)]
	if !ok {
		return ""
	}
	if len(day) == 1 {
		day = "0" + day
	}
	return year + "-" + mm + "-" + day
}

// birthdate looks in the summary first, then in the biography.
func (p senatorPage) birthdate() string {
	if m := summaryBirthRe.FindStringSubmatch(p.Summary); m != nil {
		if d := isoDate(m[2], m[3], m[4]); d != "" {
			return d
		}
	}
	if m := bodyBirthRe.FindStringSubmatch(p.Body); m != nil {
		return isoDate(m[1], m[2], m[3])
	}
	return ""
}

// profession is the occupation named in the summary ("es abogado y militante
// ..."). It is empty when the summary does not state one.
func (p senatorPage) profession() string {
	m := professionRe.FindStringSubmatch(p.Summary)
	if m == nil {
		return ""
	}
	prof := strings.TrimSpace(m[1])
	if prof == "" || utf8.RuneCountInString(prof) > 60 || !looksLikeOccupation(prof) {
		return ""
	}
	return prof
}

// looksLikeOccupation rejects phrases that describe the office or the political
// stance instead of an occupation ("Senador por la 13° Circunscripción",
// "independiente"), which the summaries can place where the occupation goes.
func looksLikeOccupation(s string) bool {
	l := strings.ToLower(s)
	if strings.Contains(l, "circunscripci") {
		return false
	}
	for _, prefix := range []string{"senador", "diputad", "militante", "pol", "independiente", "actual", "ex "} {
		if strings.HasPrefix(l, prefix) {
			return false
		}
	}
	return true
}

// party returns the full name and the acronym when the summary gives them.
func (p senatorPage) party() (name, acronym string) {
	m := partyRe.FindStringSubmatch(p.Summary)
	if m == nil {
		return "", ""
	}
	return upperFirst(strings.TrimSpace(m[1])), strings.TrimSpace(m[2])
}

func upperFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(unicode.ToUpper(r)) + s[size:]
}

// username extracts the account name from a "@name" or a profile address such
// as "https://www.instagram.com/name/channel/" (the name is the first part of
// the path). The values can carry trailing spaces.
func username(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	if i := strings.IndexAny(link, "?#"); i >= 0 {
		link = link[:i]
	}
	if i := strings.Index(link, "://"); i >= 0 {
		link = link[i+3:]
		if j := strings.Index(link, "/"); j >= 0 {
			link = link[j+1:] // drop the host
		} else {
			return ""
		}
	}
	link = strings.Trim(link, "/")
	if i := strings.Index(link, "/"); i >= 0 {
		link = link[:i]
	}
	return strings.TrimPrefix(link, "@")
}

func genderOf(label string) string {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "hombre":
		return "MALE"
	case "mujer":
		return "FEMALE"
	default:
		return "OTHER"
	}
}

// Profile returns the senator's profile. The list of senators gives the party,
// constituency and picture; the senator's page gives the birth date,
// profession, full party name and social accounts. It fails if the page cannot
// be read, so a profile is never silently incomplete.
func (s *Source) Profile(ctx context.Context, m congress.Member) (congress.Congressman, error) {
	if err := checkMember(m); err != nil {
		return congress.Congressman{}, err
	}
	rec, err := s.senator(ctx, m.ExternalID)
	if err != nil {
		return congress.Congressman{}, err
	}
	if rec.Slug == "" {
		return congress.Congressman{}, fmt.Errorf("senate: senator %d has no page address", rec.ID)
	}
	raw, err := s.fetch(ctx, s.siteURL+senatorPagePath+rec.Slug)
	if err != nil {
		return congress.Congressman{}, err
	}
	page, err := parseSenatorPage(raw)
	if err != nil {
		return congress.Congressman{}, err
	}

	party, alias := page.party()
	switch {
	case party == "":
		party, alias = rec.Party, ""
	case alias == "":
		alias = rec.Party // the list's short form (e.g. "R.N.") when the text has no acronym
	}
	return congress.Congressman{
		Chamber:            congress.Senate,
		ExternalID:         rec.ID,
		FirstName:          strings.TrimSpace(rec.FirstName),
		FirstSurname:       strings.TrimSpace(rec.FirstSurname),
		SecondSurname:      strings.TrimSpace(rec.SecondSurname),
		ProfilePicture:     rec.Image,
		Gender:             genderOf(rec.Gender),
		Birthdate:          page.birthdate(),
		Profession:         page.profession(),
		Party:              party,
		PartyAlias:         alias,
		TwitterUsername:    username(page.Links.Twitter),
		InstagramUsername:  username(page.Links.Instagram),
		ConstituencyKind:   congress.ConstituencyRegion,
		ConstituencyName:   strings.TrimSpace(rec.Region),
		ConstituencyDetail: fmt.Sprintf("Circunscripción %d", rec.ConstituencyID),
		LastUpdate:         s.now().UTC(),
	}, nil
}
