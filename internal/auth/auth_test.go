package auth_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/amansk/parksmarter-pp-cli/internal/auth"
)

func TestParseTokenFileRaw(t *testing.T) {
	s, err := auth.ParseTokenFile([]byte("secret-token-value"), "test")
	if err != nil {
		t.Fatal(err)
	}
	if s.Token() != "secret-token-value" {
		t.Fatalf("token=%q", s.Token())
	}
}

func TestParseTokenFileJSON(t *testing.T) {
	raw := `{"auth_token":"abc","refresh_token":"def"}`
	s, err := auth.ParseTokenFile([]byte(raw), "test")
	if err != nil {
		t.Fatal(err)
	}
	if s.Token() != "abc" || s.RefreshToken != "def" {
		t.Fatalf("%+v", s)
	}
}

func TestSaveSessionPermissions(t *testing.T) {
	dir := t.TempDir()
	s := &auth.Session{AuthToken: "tok", Source: "test"}
	if err := auth.SaveSession(dir, s); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, "token.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perm=%o", info.Mode().Perm())
	}
}

func TestStatusNeverIncludesSecrets(t *testing.T) {
	s := &auth.Session{AuthToken: "super-secret-token-value", Source: "test"}
	st := s.Status()
	for _, v := range st {
		if str, ok := v.(string); ok && str == "super-secret-token-value" {
			t.Fatal("status leaked secret")
		}
	}
	if st["token_fingerprint"] == "super-secret-token-value" {
		t.Fatal("fingerprint leaked full token")
	}
}
