// Package senate implements the Source interface for the Senate.
//
// All the data comes from the Senate's public back end (web-back.senado.cl), the
// same JSON API its website uses, plus the senator's page on www.senado.cl for a
// few profile fields. No browser is needed.
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

const (
	// DefaultAPIURL is the Senate's back end.
	DefaultAPIURL = "https://web-back.senado.cl"
	// DefaultSiteURL is the Senate's website, used for the senator's page.
	DefaultSiteURL = "https://www.senado.cl"
)

// Source is the Senate data source. It embeds congress.Unsupported, so every
// method not implemented here returns congress.ErrNotSupported (the Senate
// does not publish support staff or offices as separate data, for example).
type Source struct {
	congress.Unsupported
	apiURL  string
	siteURL string
	client  *http.Client
	now     func() time.Time
}

// Option customizes a Source.
type Option func(*Source)

// WithAPIURL overrides the back end base URL (used by tests).
func WithAPIURL(u string) Option { return func(s *Source) { s.apiURL = strings.TrimRight(u, "/") } }

// WithSiteURL overrides the website base URL (used by tests).
func WithSiteURL(u string) Option { return func(s *Source) { s.siteURL = strings.TrimRight(u, "/") } }

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(s *Source) { s.client = c } }

// WithClock overrides the clock (used by tests).
func WithClock(now func() time.Time) Option { return func(s *Source) { s.now = now } }

// New returns the Senate source.
func New(opts ...Option) *Source {
	s := &Source{
		apiURL:  DefaultAPIURL,
		siteURL: DefaultSiteURL,
		client:  &http.Client{Timeout: 30 * time.Second},
		now:     time.Now,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Chamber reports the chamber this source belongs to.
func (*Source) Chamber() congress.Chamber { return congress.Senate }

// envelope is the wrapper of every back end response.
type envelope struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
}

// getJSON calls a back end operation (path with its query) and decodes the
// "data" field of the response into out.
func (s *Source) getJSON(ctx context.Context, path string, out any) error {
	body, err := s.fetch(ctx, s.apiURL+path)
	if err != nil {
		return err
	}
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("senate: %s: parse response: %w", path, err)
	}
	if env.Status != "" && env.Status != "ok" {
		return fmt.Errorf("senate: %s: response status %q", path, env.Status)
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("senate: %s: parse data: %w", path, err)
	}
	return nil
}

func (s *Source) fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("senate: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("senate: %s: unexpected status %s", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("senate: %s: read body: %w", url, err)
	}
	return body, nil
}

// checkMember rejects members from other chambers.
func checkMember(m congress.Member) error {
	if m.Chamber != congress.Senate {
		return fmt.Errorf("senate: member %d belongs to chamber %q", m.ExternalID, m.Chamber)
	}
	if m.ExternalID == 0 {
		return fmt.Errorf("senate: member without id")
	}
	return nil
}
