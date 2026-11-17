// Package deputies implements the Source interface for the Chamber of Deputies.
package deputies

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// DefaultBaseURL is the Chamber's open data web service. It must be HTTPS: the
// plain HTTP endpoint accepts the connection but never answers.
const DefaultBaseURL = "https://opendata.camara.cl/camaradiputados/WServices"

const currentMembersPath = "/WSDiputado.asmx/retornarDiputadosPeriodoActual"

// Source is the Chamber of Deputies data source. Methods not implemented yet
// come from the embedded congress.Unsupported and return ErrNotSupported.
type Source struct {
	congress.Unsupported
	baseURL string
	client  *http.Client
}

// Option customizes a Source.
type Option func(*Source)

// WithBaseURL overrides the web service base URL (used by tests).
func WithBaseURL(u string) Option { return func(s *Source) { s.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(s *Source) { s.client = c } }

// New returns the Chamber of Deputies source.
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
func (*Source) Chamber() congress.Chamber { return congress.Deputies }

// membersDoc mirrors the response of retornarDiputadosPeriodoActual:
// DiputadosPeriodoColeccion > DiputadoPeriodo > Diputado.
type membersDoc struct {
	Deputies []struct {
		ID            int    `xml:"Id"`
		FirstName     string `xml:"Nombre"`
		SecondName    string `xml:"Nombre2"`
		FirstSurname  string `xml:"ApellidoPaterno"`
		SecondSurname string `xml:"ApellidoMaterno"`
	} `xml:"DiputadoPeriodo>Diputado"`
}

// CurrentMembers returns the deputies of the current period, with only the
// data needed to identify them.
func (s *Source) CurrentMembers(ctx context.Context) ([]congress.Member, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+currentMembersPath, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("deputies: fetch members: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deputies: fetch members: unexpected status %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("deputies: read members: %w", err)
	}
	return parseMembers(body)
}

func parseMembers(body []byte) ([]congress.Member, error) {
	var doc membersDoc
	if err := xml.Unmarshal(body, &doc); err != nil {
		return nil, fmt.Errorf("deputies: parse members: %w", err)
	}
	members := make([]congress.Member, 0, len(doc.Deputies))
	for _, d := range doc.Deputies {
		if d.ID == 0 {
			return nil, fmt.Errorf("deputies: member without id: %+v", d)
		}
		name := strings.Join(strings.Fields(
			d.FirstName+" "+d.SecondName+" "+d.FirstSurname+" "+d.SecondSurname), " ")
		members = append(members, congress.Member{Chamber: congress.Deputies, ExternalID: d.ID, Name: name})
	}
	if len(members) == 0 {
		return nil, fmt.Errorf("deputies: no members found in response")
	}
	return members, nil
}
