package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/amansk/parksmarter-pp-cli/internal/auth"
	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
)

// Client calls Park Smarter consumer mobile APIs.
type Client struct {
	BaseURL   string
	HTTP      *http.Client
	Session   *auth.Session
	UserAgent string
	DryRun    bool
}

// New constructs a client with optional bearer session.
func New(session *auth.Session) *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
		Session: session,
		UserAgent: "parksmarter-pp-cli/0.1.0 (+https://github.com/amansk/parksmarter-pp-cli)",
	}
}

func (c *Client) url(path string, query url.Values) string {
	u := strings.TrimRight(c.BaseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}

func (c *Client) newRequest(method, path string, query url.Values, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.url(path, query), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Session != nil {
		if tok := c.Session.Token(); tok != "" {
			req.Header.Set("Authorization", "Bearer "+tok)
			// Unverified fallback header seen in ASP.NET mobile clients; harmless if ignored.
			req.Header.Set("AuthToken", tok)
		}
	}
	return req, nil
}

func (c *Client) do(method, path string, query url.Values, body any) (json.RawMessage, int, error) {
	if c.DryRun && method != http.MethodGet && method != http.MethodHead {
		return nil, 0, exitcode.Usagef("dry-run: would %s %s", method, path)
	}
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := c.newRequest(method, path, query, rdr)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, exitcode.Transientf("request failed: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return raw, resp.StatusCode, exitcode.Transientf("rate limited (HTTP 429)")
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return raw, resp.StatusCode, exitcode.Authf("unauthorized (HTTP 401)")
	}
	if resp.StatusCode >= 400 {
		msg := truncate(string(raw), 240)
		if msg == "" {
			msg = resp.Status
		}
		return raw, resp.StatusCode, exitcode.APIf("HTTP %d: %s", resp.StatusCode, msg)
	}
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, resp.StatusCode, nil
	}
	if !json.Valid(raw) {
		return raw, resp.StatusCode, exitcode.APIf("non-JSON response: %s", truncate(string(raw), 120))
	}
	return raw, resp.StatusCode, nil
}

func (c *Client) getJSON(path string, query url.Values, out any) error {
	raw, _, err := c.do(http.MethodGet, path, query, nil)
	if err != nil {
		return err
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func (c *Client) postJSON(path string, body any, out any) error {
	raw, _, err := c.do(http.MethodPost, path, nil, body)
	if err != nil {
		return err
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// Ping checks the public ApplicationValidity endpoint (no auth).
func (c *Client) Ping() error {
	var out map[string]any
	return c.getJSON(PathApplicationValidity, nil, &out)
}

// ProbeSessionAuth checks an authenticated read endpoint.
func (c *Client) ProbeSessionAuth() error {
	if c.Session == nil || c.Session.Token() == "" {
		return exitcode.Authf("no auth token configured")
	}
	var out map[string]any
	if err := c.getJSON(PathSessionsActive, nil, &out); err != nil {
		return err
	}
	return nil
}
