package senate

import (
	"context"
	"fmt"
	"strings"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// currentMembersPath lists the sitting senators ("vigentes", chamber "S").
// The limit is above the 50 seats so the list is never cut.
const currentMembersPath = "/api/hemicycle?vigentes=1&camara=S&limit=100"

// senatorRecord is one sitting senator as listed by the back end.
type senatorRecord struct {
	ID             int    `json:"ID_PARLAMENTARIO"`
	Slug           string `json:"SLUG"`
	FirstName      string `json:"NOMBRE"`
	FirstSurname   string `json:"APELLIDO_PATERNO"`
	SecondSurname  string `json:"APELLIDO_MATERNO"`
	Chamber        string `json:"CAMARA"`
	Party          string `json:"PARTIDO"`
	ConstituencyID int    `json:"CIRCUNSCRIPCION_ID"`
	Region         string `json:"REGION"`
	Gender         string `json:"SEXO_ETIQUETA"`
	Image          string `json:"IMAGEN_450"`
	Periods        []struct {
		Chamber string `json:"CAMARA"`
		From    string `json:"DESDE"`
		To      string `json:"HASTA"`
		Current int    `json:"VIGENTE"`
	} `json:"PERIODOS"`
}

type hemicycleData struct {
	Parlamentarios struct {
		Total int             `json:"total"`
		Data  []senatorRecord `json:"data"`
	} `json:"parlamentarios"`
}

// fullName joins the name parts, dropping the stray spaces some records carry
// and the missing second surname of some senators.
func (r senatorRecord) fullName() string {
	return strings.Join(strings.Fields(r.FirstName+" "+r.FirstSurname+" "+r.SecondSurname), " ")
}

// senators returns the sitting senators. The list is rejected if it looks
// incomplete (fewer entries than the total the API announces), because a draw
// from a partial list would not be uniform.
func (s *Source) senators(ctx context.Context) ([]senatorRecord, error) {
	var data hemicycleData
	if err := s.getJSON(ctx, currentMembersPath, &data); err != nil {
		return nil, err
	}
	list := data.Parlamentarios
	if len(list.Data) == 0 {
		return nil, fmt.Errorf("senate: no members found in response")
	}
	if list.Total != len(list.Data) {
		return nil, fmt.Errorf("senate: incomplete list: announced %d, received %d", list.Total, len(list.Data))
	}
	for _, p := range list.Data {
		if p.ID == 0 {
			return nil, fmt.Errorf("senate: member without id: %+v", p)
		}
		if p.Chamber != "S" {
			return nil, fmt.Errorf("senate: member %d is from chamber %q, not the Senate", p.ID, p.Chamber)
		}
	}
	return list.Data, nil
}

// senator returns the record of one sitting senator.
func (s *Source) senator(ctx context.Context, id int) (senatorRecord, error) {
	all, err := s.senators(ctx)
	if err != nil {
		return senatorRecord{}, err
	}
	for _, p := range all {
		if p.ID == id {
			return p, nil
		}
	}
	return senatorRecord{}, fmt.Errorf("senate: senator %d is not among the sitting senators", id)
}

// CurrentMembers returns the sitting senators.
func (s *Source) CurrentMembers(ctx context.Context) ([]congress.Member, error) {
	all, err := s.senators(ctx)
	if err != nil {
		return nil, err
	}
	members := make([]congress.Member, 0, len(all))
	for _, p := range all {
		members = append(members, congress.Member{Chamber: congress.Senate, ExternalID: p.ID, Name: p.fullName()})
	}
	return members, nil
}
