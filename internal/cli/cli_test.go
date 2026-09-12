package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amansk/parksmarter-pp-cli/internal/auth"
	"github.com/amansk/parksmarter-pp-cli/internal/cli"
	"github.com/amansk/parksmarter-pp-cli/internal/client"
)

func mockHTTP(t *testing.T) *client.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case client.PathApplicationValidity:
			_, _ = w.Write([]byte(`{"ForceUpgrade":false}`))
		case client.PathCards:
			_, _ = w.Write([]byte(`{"Cards":[{"CardId":"1","Last4":"4242","IsDefault":true}]}`))
		case client.PathVehicles:
			_, _ = w.Write([]byte(`{"Vehicles":[{"VehicleId":"9","LicensePlate":"ABC123","State":"CA"}]}`))
		case client.PathMetersNearby:
			_, _ = w.Write([]byte(`{"Meters":[{"MeterId":"m1","MeterNumber":"100","ZoneName":"Z1"}]}`))
		case client.PathMeterByNumber:
			_, _ = w.Write([]byte(`{"Meters":[{"MeterId":"m2","MeterNumber":"200"}]}`))
		case client.PathSessionsActive:
			_, _ = w.Write([]byte(`{"Sessions":[{"SessionId":"s1","Status":"active","MeterNumber":"100"}]}`))
		case client.PathSessionsPast:
			_, _ = w.Write([]byte(`{"Sessions":[]}`))
		case client.PathStartSession:
			_, _ = w.Write([]byte(`{"SessionId":"s-new"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	c := client.New(&auth.Session{AuthToken: "tok"})
	c.BaseURL = srv.URL
	c.HTTP = srv.Client()
	return c
}

func runCLI(t *testing.T, args ...string) (int, string, string) {
	return runCLIWithHome(t, t.TempDir(), args...)
}

func runCLIWithHome(t *testing.T, home string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	root := cli.NewRootForTest(&cli.Options{HTTP: mockHTTP(t), Home: home})
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	code := 0
	if err := root.Execute(); err != nil {
		code = cli.ExitCodeForTest(err)
	}
	return code, stdout.String(), stderr.String()
}

func TestVersion(t *testing.T) {
	code, out, _ := runCLI(t, "--version")
	if code != 0 || out == "" {
		t.Fatalf("code=%d out=%q", code, out)
	}
}

func TestDoctorJSON(t *testing.T) {
	code, out, errOut := runCLI(t, "doctor", "--json")
	if code != 0 {
		t.Fatalf("code=%d err=%q out=%q", code, errOut, out)
	}
}

func TestAccountCardsJSON(t *testing.T) {
	code, out, errOut := runCLI(t, "--json", "account", "cards")
	if code != 0 {
		t.Fatalf("code=%d err=%q out=%q", code, errOut, out)
	}
	var payload struct {
		Count int `json:"count"`
		Cards []struct {
			ID    string `json:"id"`
			Last4 string `json:"last4"`
		} `json:"cards"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Count != 1 || payload.Cards[0].Last4 != "4242" {
		t.Fatalf("%+v", payload)
	}
}

func TestZonesNearbyJSON(t *testing.T) {
	code, out, errOut := runCLI(t, "--json", "zones", "nearby", "--lat", "32.7", "--lng", "-117.1")
	if code != 0 {
		t.Fatalf("code=%d err=%q out=%q", code, errOut, out)
	}
}

func TestSessionsListJSON(t *testing.T) {
	code, out, errOut := runCLI(t, "--json", "sessions", "list")
	if code != 0 {
		t.Fatalf("code=%d err=%q out=%q", code, errOut, out)
	}
}

func TestParkingStartRefusesWithoutGates(t *testing.T) {
	code, _, errOut := runCLI(t, "parking", "start", "--meter-number", "100", "--minutes", "60")
	if code != 2 {
		t.Fatalf("code=%d err=%q", code, errOut)
	}
}

func TestParkingStartDryRunWithGates(t *testing.T) {
	code, out, errOut := runCLI(t, "--json", "--dry-run", "parking", "start",
		"--meter-number", "100", "--minutes", "60",
		"--enable-live-parking", "--owner-approved",
		"--confirm", client.StartConfirmPhrase,
	)
	if code != 0 {
		t.Fatalf("code=%d err=%q out=%q", code, errOut, out)
	}
	var payload struct {
		DryRun  bool              `json:"dry_run"`
		WouldGet string           `json:"would_get"`
		Query   map[string]string `json:"query"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.DryRun || payload.WouldGet != client.PathStartSession {
		t.Fatalf("%+v", payload)
	}
}

func TestAuthStatusUnauthenticated(t *testing.T) {
	code, out, _ := runCLI(t, "--json", "auth", "status")
	if code != 0 {
		t.Fatalf("code=%d", code)
	}
	var st struct {
		Authenticated bool `json:"authenticated"`
	}
	if err := json.Unmarshal([]byte(out), &st); err != nil {
		t.Fatal(err)
	}
	if st.Authenticated {
		t.Fatal("expected unauthenticated")
	}
}
