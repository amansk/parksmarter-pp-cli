// Package auth manages Park Smarter consumer bearer tokens.
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AppName    = "parksmarter-pp-cli"
	TokenEnv   = "PARKSMARTER_TOKEN"
	PhoneEnv   = "PARKSMARTER_PHONE"
	PasswordEnv = "PARKSMARTER_PASSWORD"
)

// Session holds bearer auth material. Values are never logged.
type Session struct {
	AuthToken    string `json:"auth_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Source       string `json:"source,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// HomeDir resolves the config directory.
func HomeDir(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if v := os.Getenv("PARKSMARTER_PP_HOME"); v != "" {
		return v, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppName), nil
}

func tokenPath(home string) string {
	return filepath.Join(home, "token.json")
}

// LoadSession reads stored token (0600 file).
func LoadSession(home string) (*Session, error) {
	path := tokenPath(home)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s Session
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("parse token file: %w", err)
	}
	return &s, nil
}

// SaveSession writes token with mode 0600.
func SaveSession(home string, s *Session) error {
	if err := os.MkdirAll(home, 0o700); err != nil {
		return err
	}
	if s.UpdatedAt == "" {
		s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	path := tokenPath(home)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return err
	}
	return os.Chmod(path, 0o600)
}

// Token returns the bearer token if present.
func (s *Session) Token() string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(s.AuthToken)
}

// Status returns safe metadata (no secret values).
func (s *Session) Status() map[string]any {
	if s == nil || s.Token() == "" {
		return map[string]any{"authenticated": false}
	}
	return map[string]any{
		"authenticated":     true,
		"source":            s.Source,
		"updated_at":        s.UpdatedAt,
		"has_refresh":       strings.TrimSpace(s.RefreshToken) != "",
		"token_present":     true,
		"token_fingerprint": fingerprint(s.Token()),
	}
}

func fingerprint(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "…" + token[len(token)-4:]
}

// ResolveSession loads file token, then PARKSMARTER_TOKEN env.
func ResolveSession(home string) (*Session, error) {
	if v := strings.TrimSpace(os.Getenv(TokenEnv)); v != "" {
		return &Session{
			AuthToken: v,
			Source:    "env:" + TokenEnv,
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		}, nil
	}
	return LoadSession(home)
}

// ParseTokenFile reads a raw token string or JSON {"auth_token":"..."}.
func ParseTokenFile(data []byte, source string) (*Session, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, fmt.Errorf("token file is empty")
	}
	if strings.HasPrefix(trimmed, "{") {
		var s Session
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("parse token json: %w", err)
		}
		if s.Token() == "" {
			return nil, fmt.Errorf("token json missing auth_token")
		}
		s.Source = source
		s.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		return &s, nil
	}
	return &Session{
		AuthToken: trimmed,
		Source:    source,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ResolvePhone returns phone from flag or env.
func ResolvePhone(flag string) (string, error) {
	if v := strings.TrimSpace(flag); v != "" {
		return v, nil
	}
	if v := strings.TrimSpace(os.Getenv(PhoneEnv)); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("phone required: pass --phone or set %s", PhoneEnv)
}

// ResolvePassword returns password from flag or env.
func ResolvePassword(flag string) (string, error) {
	if v := flag; v != "" {
		return v, nil
	}
	if v := os.Getenv(PasswordEnv); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("password required: pass --password or set %s", PasswordEnv)
}
