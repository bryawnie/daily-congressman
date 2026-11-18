// Package senate implements the Source interface for the Senate.
package senate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// DefaultBaseURL is the Senate's back end. The public website embeds the
// responses of this API, so no browser is needed.
const DefaultBaseURL = "https://web-back.senado.cl"

// currentMembersPath lists the sitting senators ("vigentes", chamber "S").
// The limit is above the 50 seats so the list is never cut.
const currentMembersPath = "/api/hemicycle?vigentes=1&camara=S&limit=100"

// Source is the Senate data source. It embeds congress.Unsupported, so every
// method not implemented here returns congress.ErrNotSupported.
//
// TODO, in the order the data is needed:
//   - Profile.
//   - Attendance.
//   - Votings.
//   - Activity.
//   - Expenses.
type Source struct {
	congress.Unsupported
	baseURL string
	client  *http.Client
}

// Option customizes a Source.
type Option func(*Source)

// WithBaseURL overrides the back end base URL (used by tests).
func WithBaseURL(u string) Option { return func(s *Source) { s.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(s *Source) { s.client = c } }

// New returns the Senate source.
func New(opts ...Option) *Source {
	s := &Source{
		baseURL: DefaultBaseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Chamber reports the chamber this source belongs to.
func (*Source) Chamber() congress.Chamber { return congress.Senate }

type hemicycleResponse struct {
	Status string `json:"status"`
	Data   struct {
		Parlamentarios struct {
			Total int `json:"total"`
			Data  []struct {
				ID            int    `json:"ID_PARLAMENTARIO"`
				FirstName     string `json:"NOMBRE"`
				FirstSurname  string `json:"APELLIDO_PATERNO"`
				SecondSurname string `json:"APELLIDO_MATERNO"`
				Chamber       string `json:"CAMARA"`
			} `json:"data"`
		} `json:"parlamentarios"`
	} `json:"data"`
}

// CurrentMembers returns the sitting senators.
//
// The list is rejected if it looks incomplete (fewer entries than the total the
// API announces), because a draw from a partial list would not be uniform.
func (s *Source) CurrentMembers(ctx context.Context) ([]congress.Member, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+currentMembersPath, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("senate: fetch members: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("senate: fetch members: unexpected status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("senate: read members: %w", err)
	}
	return parseMembers(body)
}

func parseMembers(body []byte) ([]congress.Member, error) {
	var doc hemicycleResponse
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("senate: parse members: %w", err)
	}
	if doc.Status != "" && doc.Status != "ok" {
		return nil, fmt.Errorf("senate: members response status %q", doc.Status)
	}
	list := doc.Data.Parlamentarios
	if len(list.Data) == 0 {
		return nil, fmt.Errorf("senate: no members found in response")
	}
	if list.Total != len(list.Data) {
		return nil, fmt.Errorf("senate: incomplete list: announced %d, received %d", list.Total, len(list.Data))
	}

	members := make([]congress.Member, 0, len(list.Data))
	for _, p := range list.Data {
		if p.ID == 0 {
			return nil, fmt.Errorf("senate: member without id: %+v", p)
		}
		if p.Chamber != "S" {
			return nil, fmt.Errorf("senate: member %d is from chamber %q, not the Senate", p.ID, p.Chamber)
		}
		// Some names carry stray spaces, and some have no second surname.
		name := strings.Join(strings.Fields(p.FirstName+" "+p.FirstSurname+" "+p.SecondSurname), " ")
		members = append(members, congress.Member{Chamber: congress.Senate, ExternalID: p.ID, Name: name})
	}
	return members, nil
}
