package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amansk/parksmarter-pp-cli/internal/auth"
	"github.com/amansk/parksmarter-pp-cli/internal/client"
)

func mockClient(t *testing.T) *client.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case client.PathApplicationValidity:
			_, _ = w.Write([]byte(`{"ForceUpgrade":false,"Config":{"IsInMaintenanceMode":false}}`))
		case client.PathLoginPhonePassword:
			_, _ = w.Write([]byte(`{"AuthToken":"test-token","RefreshToken":"refresh"}`))
		case client.PathCards:
			_, _ = w.Write([]byte(`{"Cards":[{"CardId":"1","Last4":"4242","IsDefault":true}]}`))
		case client.PathVehicles:
			_, _ = w.Write([]byte(`{"Vehicles":[{"VehicleId":"9","LicensePlate":"ABC123","State":"CA","IsDefault":true}]}`))
		case client.PathMetersNearby:
			_, _ = w.Write([]byte(`{"Meters":[{"MeterId":"m1","MeterNumber":"100","ZoneName":"Z1","Latitude":32.7,"Longitude":-117.1}]}`))
		case client.PathMeterByNumber:
			_, _ = w.Write([]byte(`{"Meters":[{"MeterId":"m2","MeterNumber":"`+r.URL.Query().Get("MeterNumber")+`"}]}`))
		case client.PathSessionsActive:
			_, _ = w.Write([]byte(`{"Sessions":[{"SessionId":"s1","Status":"active","MeterNumber":"100"}]}`))
		case client.PathSessionsPast:
			_, _ = w.Write([]byte(`{"Sessions":[{"SessionId":"s2","Status":"ended"}]}`))
		case client.PathStartSession:
			_, _ = w.Write([]byte(`{"SessionId":"s-new","Status":"active"}`))
		case client.PathExtendSession:
			_, _ = w.Write([]byte(`{"SessionId":"`+r.URL.Query().Get("SessionId")+`","Status":"extended"}`))
		case client.PathStopSession:
			_, _ = w.Write([]byte(`{"SessionId":"`+r.URL.Query().Get("SessionId")+`","Status":"stopped"}`))
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

func TestLoginWithPhonePassword(t *testing.T) {
	c := mockClient(t)
	c.AllowUnverifiedMutations = true
	s, err := c.LoginWithPhonePassword(client.LoginInput{PhoneNumber: "+15555550100", Password: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	if s.Token() != "test-token" {
		t.Fatalf("token=%q", s.Token())
	}
}

func TestListCardsRedacts(t *testing.T) {
	c := mockClient(t)
	cards, err := c.ListCards()
	if err != nil {
		t.Fatal(err)
	}
	if len(cards) != 1 || cards[0].Last4 != "4242" || cards[0].ID != "1" {
		t.Fatalf("%+v", cards)
	}
}

func TestNearbyMeters(t *testing.T) {
	c := mockClient(t)
	meters, err := c.NearbyMeters(client.NearbyMetersQuery{Latitude: 32.7, Longitude: -117.1})
	if err != nil {
		t.Fatal(err)
	}
	if len(meters) != 1 || meters[0].MeterNumber != "100" {
		t.Fatalf("%+v", meters)
	}
}

func TestLoginWithPhonePasswordBlocksWithoutAcknowledge(t *testing.T) {
	c := mockClient(t)
	_, err := c.LoginWithPhonePassword(client.LoginInput{PhoneNumber: "+15555550100", Password: "secret"})
	if err == nil {
		t.Fatal("expected error without acknowledge")
	}
}

func TestStartSessionDryRunDoesNotMutate(t *testing.T) {
	c := mockClient(t)
	c.DryRun = true
	out, err := c.StartSession(client.StartParkingInput{MeterNumber: "100", Minutes: 60})
	if err != nil {
		t.Fatal(err)
	}
	if out["dry_run"] != true {
		t.Fatalf("out=%v", out)
	}
}

func TestStartSessionBlocksLiveWithoutAcknowledge(t *testing.T) {
	c := mockClient(t)
	_, err := c.StartSession(client.StartParkingInput{MeterNumber: "100", Minutes: 60})
	if err == nil {
		t.Fatal("expected error without acknowledge")
	}
}

func TestStartSession(t *testing.T) {
	c := mockClient(t)
	c.AllowUnverifiedMutations = true
	out, err := c.StartSession(client.StartParkingInput{MeterNumber: "100", Minutes: 60})
	if err != nil {
		t.Fatal(err)
	}
	if out["SessionId"] != "s-new" {
		b, _ := json.Marshal(out)
		t.Fatalf("out=%s", b)
	}
}
