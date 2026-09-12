package client

import (
	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
)

// ListActiveSessions returns in-progress parking sessions.
func (c *Client) ListActiveSessions() ([]SessionSummary, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	var out map[string]any
	if err := c.getJSON(PathSessionsActive, nil, &out); err != nil {
		return nil, err
	}
	return mapSessions(out), nil
}

// ListPastSessions returns historical parking sessions.
func (c *Client) ListPastSessions() ([]SessionSummary, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	var out map[string]any
	if err := c.getJSON(PathSessionsPast, nil, &out); err != nil {
		return nil, err
	}
	return mapSessions(out), nil
}

// GetSession finds one session by id from active then past lists.
func (c *Client) GetSession(id string) (SessionSummary, error) {
	if id == "" {
		return SessionSummary{}, exitcode.Usagef("session id required")
	}
	active, err := c.ListActiveSessions()
	if err != nil {
		return SessionSummary{}, err
	}
	for _, s := range active {
		if s.ID == id {
			return s, nil
		}
	}
	past, err := c.ListPastSessions()
	if err != nil {
		return SessionSummary{}, err
	}
	for _, s := range past {
		if s.ID == id {
			return s, nil
		}
	}
	return SessionSummary{}, exitcode.NotFoundf("session %s not found", id)
}

func mapSessions(data map[string]any) []SessionSummary {
	rows := findSlice(data, "Sessions", "sessions", "ParkingSessions", "parkingSessions", "Results", "results")
	out := make([]SessionSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, SessionSummary{
			ID:          pickString(row, "SessionId", "SessionID", "ParkingSessionId", "Id", "ID", "id"),
			Status:      pickString(row, "Status", "status", "SessionStatus"),
			MeterNumber: pickString(row, "MeterNumber", "meterNumber"),
			ZoneName:    pickString(row, "ZoneName", "zoneName"),
			Starts:      pickString(row, "StartTime", "Starts", "startTime", "starts"),
			Ends:        pickString(row, "EndTime", "Ends", "endTime", "ends"),
			Plate:       pickString(row, "LicensePlate", "licensePlate", "Plate", "plate"),
		})
	}
	return out
}
