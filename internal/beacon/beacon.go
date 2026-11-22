// Package beacon is a client for the Random UChile randomness beacon.
package beacon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL is the Random UChile beacon (version 2.1-beta).
const DefaultBaseURL = "https://random.uchile.cl/beacon/2.1-beta"

// ErrNotAvailable means the beacon has not published the requested pulse yet.
// It is expected right after midnight and is worth retrying a bit later.
var ErrNotAvailable = errors.New("beacon: pulse not available yet")

// Pulse is the beacon record used to draw a day's congressman.
type Pulse struct {
	ChainID string
	PulseID string
	RandOut string // hexadecimal output value
}

// URI returns the public address of the pulse, so anyone can verify it.
func (p Pulse) URI(baseURL string) string {
	return fmt.Sprintf("%s/chain/%s/pulse/%s", strings.TrimRight(baseURL, "/"), p.ChainID, p.PulseID)
}

// Client queries the beacon.
type Client struct {
	baseURL string
	http    *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL overrides the beacon base URL (used by tests).
func WithBaseURL(u string) Option { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

// NewClient returns a beacon client.
func NewClient(opts ...Option) *Client {
	c := &Client{baseURL: DefaultBaseURL, http: &http.Client{Timeout: 30 * time.Second}}
	for _, o := range opts {
		o(c)
	}
	return c
}

// BaseURL returns the beacon base URL.
func (c *Client) BaseURL() string { return c.baseURL }

type pulseResponse struct {
	Pulse struct {
		OutputValue string      `json:"outputValue"`
		ChainIndex  json.Number `json:"chainIndex"`
		PulseIndex  json.Number `json:"pulseIndex"`
	} `json:"pulse"`
}

// PulseAt returns the first pulse generated at or after t.
func (c *Client) PulseAt(ctx context.Context, t time.Time) (Pulse, error) {
	url := fmt.Sprintf("%s/pulse?timeGE=%d", c.baseURL, t.UnixMilli())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Pulse{}, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return Pulse{}, fmt.Errorf("beacon: fetch pulse: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Pulse{}, ErrNotAvailable
	}
	if resp.StatusCode != http.StatusOK {
		return Pulse{}, fmt.Errorf("beacon: fetch pulse: unexpected status %s", resp.Status)
	}
	var r pulseResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Pulse{}, fmt.Errorf("beacon: decode pulse: %w", err)
	}
	if r.Pulse.OutputValue == "" {
		return Pulse{}, fmt.Errorf("beacon: pulse without output value")
	}
	return Pulse{
		ChainID: r.Pulse.ChainIndex.String(),
		PulseID: r.Pulse.PulseIndex.String(),
		RandOut: r.Pulse.OutputValue,
	}, nil
}
