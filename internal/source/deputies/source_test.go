package deputies

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryawnie/daily-congressman/internal/congress"
)

// Fixture with the structure of the real response of
// retornarDiputadosPeriodoActual (names are made up).
const membersXML = `<?xml version="1.0" encoding="utf-8"?>
<DiputadosPeriodoColeccion xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns="http://opendata.camara.cl/camaradiputados/v1">
  <DiputadoPeriodo>
    <FechaInicio>2030-03-10T00:00:00</FechaInicio>
    <FechaTermino xsi:nil="true" />
    <Diputado>
      <Id>1001</Id>
      <Nombre>Ana</Nombre>
      <Nombre2 />
      <ApellidoPaterno>Pérez</ApellidoPaterno>
      <ApellidoMaterno>Soto</ApellidoMaterno>
      <FechaNacimiento>1980-01-02T00:00:00</FechaNacimiento>
      <Sexo Valor="0">Femenino</Sexo>
      <Militancias />
    </Diputado>
  </DiputadoPeriodo>
  <DiputadoPeriodo>
    <FechaInicio>2030-03-10T00:00:00</FechaInicio>
    <FechaTermino xsi:nil="true" />
    <Diputado>
      <Id>1002</Id>
      <Nombre>LUIS</Nombre>
      <Nombre2>ALBERTO</Nombre2>
      <ApellidoPaterno>ROJAS</ApellidoPaterno>
      <ApellidoMaterno />
      <Sexo Valor="1">Masculino</Sexo>
    </Diputado>
  </DiputadoPeriodo>
</DiputadosPeriodoColeccion>`

func TestCurrentMembers(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.Write([]byte(membersXML))
	}))
	defer srv.Close()

	got, err := New(WithBaseURL(srv.URL)).CurrentMembers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if path != "/WSDiputado.asmx/retornarDiputadosPeriodoActual" {
		t.Fatalf("path = %s", path)
	}
	want := []congress.Member{
		{Chamber: congress.Deputies, ExternalID: 1001, Name: "Ana Pérez Soto"},
		{Chamber: congress.Deputies, ExternalID: 1002, Name: "LUIS ALBERTO ROJAS"},
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("members = %+v, want %+v", got, want)
	}
}

func TestCurrentMembersErrors(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"status": func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) },
		"badxml": func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("<x")) },
		"empty":  func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(`<DiputadosPeriodoColeccion/>`)) },
		"noid": func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte(`<C><DiputadoPeriodo><Diputado><Nombre>A</Nombre></Diputado></DiputadoPeriodo></C>`))
		},
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(h)
			defer srv.Close()
			if _, err := New(WithBaseURL(srv.URL)).CurrentMembers(context.Background()); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestUnimplementedMethodsAreNotSupported(t *testing.T) {
	_, err := New().Profile(context.Background(), congress.Member{})
	if !errors.Is(err, congress.ErrNotSupported) {
		t.Fatalf("err = %v", err)
	}
}
