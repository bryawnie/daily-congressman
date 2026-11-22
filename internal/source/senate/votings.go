package senate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// votesPageSize is how many votings are requested per call. Each one lists
// every senator by vote, so pages are large.
const votesPageSize = 50

// voterList is the list of senators who cast one kind of vote. The back end
// sends the number 0, and not an empty list, when nobody did.
type voterList []struct {
	ID int `json:"PARLID"`
}

func (l *voterList) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) > 0 && b[0] == '[' {
		var raw []struct {
			ID int `json:"PARLID"`
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			return err
		}
		*l = raw
		return nil
	}
	if string(b) == "null" {
		*l = nil
		return nil
	}
	n, err := strconv.Atoi(string(b))
	if err != nil || n != 0 {
		return fmt.Errorf("senate: unexpected voters value %s", b)
	}
	*l = nil
	return nil
}

func (l voterList) has(id int) bool {
	for _, v := range l {
		if v.ID == id {
			return true
		}
	}
	return false
}

// text accepts a JSON string, a number or null.
type text string

func (t *text) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if string(b) == "null" {
		*t = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*t = text(s)
		return nil
	}
	*t = text(b)
	return nil
}

type voteRecord struct {
	ID       int    `json:"ID_VOTACION"`
	Date     string `json:"FECHA_VOTACION"`
	Topic    string `json:"TEMA"`
	Bulletin text   `json:"BOLETIN"`
	Quorum   string `json:"QUORUM"`
	Yes      int    `json:"SI"`
	No       int    `json:"NO"`
	Abstain  int    `json:"ABS"`
	Ballots  struct {
		Yes     voterList `json:"SI"`
		No      voterList `json:"NO"`
		Abstain voterList `json:"ABSTENCION"`
		Pair    voterList `json:"PAREO"`
	} `json:"VOTACIONES"`
}

type votesData struct {
	Total int          `json:"total"`
	Data  []voteRecord `json:"data"`
}

// option names the vote the senator cast in this voting, using the same
// labels as the Chamber of Deputies. It is empty if the senator is not listed
// (absent).
func (r voteRecord) option(senatorID int) string {
	switch {
	case r.Ballots.Yes.has(senatorID):
		return "Afirmativo"
	case r.Ballots.No.has(senatorID):
		return "En Contra"
	case r.Ballots.Abstain.has(senatorID):
		return "Abstención"
	case r.Ballots.Pair.has(senatorID):
		return "Pareo"
	}
	return ""
}

// Votings returns the latest votings of the current legislature in which the
// senator took part (up to limit), newest first. The Senate lists the
// voters of each voting, so the list is walked page by page until enough of
// them include the senator.
//
// The Senate does not say whether a voting was approved, so Result is
// calculated from the quorum and the totals (see outcome). It uses the number of
// sitting senators as the senators in office, and is empty for a quorum it does
// not know.
func (s *Source) Votings(ctx context.Context, m congress.Member, limit int) ([]congress.Voting, error) {
	if err := checkMember(m); err != nil {
		return nil, err
	}
	if limit <= 0 {
		return nil, fmt.Errorf("senate: votings limit must be positive, got %d", limit)
	}
	leg, err := s.currentLegislature(ctx)
	if err != nil {
		return nil, err
	}
	sitting, err := s.senators(ctx)
	if err != nil {
		return nil, err
	}
	seats := len(sitting)

	var out []congress.Voting
	// The offset advances by the rows received, not by the page size asked for,
	// so a server that caps the page size does not make votings get skipped.
	for offset := 0; len(out) < limit; {
		var page votesData
		path := fmt.Sprintf("/api/votes?id_legislatura=%d&limit=%d&offset=%d&palabra_clave=&desde=&hasta=",
			leg.ID, votesPageSize, offset)
		if err := s.getJSON(ctx, path, &page); err != nil {
			return nil, err
		}
		for _, r := range page.Data {
			if r.ID == 0 {
				return nil, fmt.Errorf("senate: voting without id")
			}
			opt := r.option(m.ExternalID)
			if opt == "" {
				continue
			}
			date, err := time.Parse("02-01-2006 15:04:05", r.Date)
			if err != nil {
				return nil, fmt.Errorf("senate: voting %d has a bad date %q: %w", r.ID, r.Date, err)
			}
			title, article := splitTopic(cleanText(r.Topic))
			out = append(out, congress.Voting{
				VotingID:        r.ID,
				Date:            date.Format("2006-01-02"),
				BulletinNumber:  string(r.Bulletin),
				DocumentTitle:   title,
				ArticleText:     article,
				VotedOption:     opt,
				TotalApproved:   r.Yes,
				TotalRejected:   r.No,
				TotalAbstention: r.Abstain,
				Quorum:          r.Quorum,
				Result:          outcome(r.Quorum, r.Yes, r.No, seats),
			})
			if len(out) == limit {
				break
			}
		}
		offset += len(page.Data)
		if len(page.Data) == 0 || offset >= page.Total {
			break
		}
	}
	return out, nil
}

// splitTopic separates the first line of the topic (its title) from the rest.
func splitTopic(topic string) (title, rest string) {
	lines := strings.Split(strings.TrimSpace(topic), "\n")
	title = strings.TrimSpace(lines[0])
	var parts []string
	for _, l := range lines[1:] {
		if l = strings.TrimSpace(l); l != "" {
			parts = append(parts, l)
		}
	}
	return title, strings.Join(parts, " ")
}
