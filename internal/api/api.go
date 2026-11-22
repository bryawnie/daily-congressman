// Package api serves the congressman of the day over HTTP, read-only, from the
// data stored by the jobs. It never reaches a source: it needs no browser and
// no network access, only the database.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bryawnie/daily-congressman/internal/chooser"
	"github.com/bryawnie/daily-congressman/internal/store"
)

const (
	// DefaultLimit is how many days GET /days returns: today and the 7 before.
	DefaultLimit = 8
	// MaxLimit caps the days a request can ask for.
	MaxLimit = 31
)

// Server answers the HTTP requests.
type Server struct {
	Store *store.Store
	// CORSOrigins are the origins allowed to call the API; "*" allows any. Empty
	// sends no CORS headers.
	CORSOrigins []string
	Log         *slog.Logger
}

// Handler returns the routes, wrapped with CORS and the read-only guard.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /days", s.days)
	mux.HandleFunc("GET /days/{date}", s.day)
	return s.cors(mux)
}

func (s *Server) log() *slog.Logger {
	if s.Log != nil {
		return s.Log
	}
	return slog.Default()
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if _, err := s.Store.ListDayChoices(ctx, 1); err != nil {
		s.log().Error("health check failed", "err", err)
		writeError(w, http.StatusServiceUnavailable, "database unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// days answers GET /days?limit=N: the latest days, most recent first. A day
// that cannot be built is reported, not skipped, so the client never shows a
// silently shorter list.
func (s *Server) days(w http.ResponseWriter, r *http.Request) {
	limit := DefaultLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > MaxLimit {
			writeError(w, http.StatusBadRequest, "limit must be a number from 1 to "+strconv.Itoa(MaxLimit))
			return
		}
		limit = n
	}
	choices, err := s.Store.ListDayChoices(r.Context(), limit)
	if err != nil {
		s.internal(w, err)
		return
	}
	days := make([]Day, 0, len(choices))
	for _, c := range choices {
		d, err := BuildDay(r.Context(), s.Store, c)
		if err != nil {
			s.internal(w, err)
			return
		}
		days = append(days, d)
	}
	setCache(w)
	writeJSON(w, http.StatusOK, map[string]any{"days": days})
}

// day answers GET /days/{date}.
func (s *Server) day(w http.ResponseWriter, r *http.Request) {
	date := r.PathValue("date")
	if _, err := chooser.DayStart(date); err != nil || len(date) != len("2006-01-02") {
		writeError(w, http.StatusBadRequest, "date must be YYYY-MM-DD")
		return
	}
	choice, err := s.Store.GetDayChoice(r.Context(), date)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "no congressman was chosen for "+date)
		return
	}
	if err != nil {
		s.internal(w, err)
		return
	}
	d, err := BuildDay(r.Context(), s.Store, choice)
	if err != nil {
		s.internal(w, err)
		return
	}
	setCache(w)
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) internal(w http.ResponseWriter, err error) {
	s.log().Error("request failed", "err", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

// cors adds the CORS headers and answers preflight requests.
func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			if allow := s.allowOrigin(origin); allow != "" {
				w.Header().Set("Access-Control-Allow-Origin", allow)
				w.Header().Add("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) allowOrigin(origin string) string {
	for _, o := range s.CORSOrigins {
		if o == "*" {
			return "*"
		}
		if strings.EqualFold(o, origin) {
			return origin
		}
	}
	return ""
}

// setCache lets browsers and proxies reuse an answer for a few minutes: the
// data changes at most a few times a day.
func setCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=300")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
