package client

import "encoding/json"

// CardSummary is a redacted payment method row.
type CardSummary struct {
	ID     string `json:"id,omitempty"`
	Last4  string `json:"last4,omitempty"`
	Brand  string `json:"brand,omitempty"`
	Default bool  `json:"is_default,omitempty"`
}

// VehicleSummary is a vehicle row without PII beyond plate metadata.
type VehicleSummary struct {
	ID           string `json:"id,omitempty"`
	LicensePlate string `json:"license_plate,omitempty"`
	State        string `json:"state,omitempty"`
	Default      bool   `json:"is_default,omitempty"`
}

// MeterSummary is a zone/meter lookup row.
type MeterSummary struct {
	ID          string  `json:"id,omitempty"`
	MeterNumber string  `json:"meter_number,omitempty"`
	ZoneName    string  `json:"zone_name,omitempty"`
	SpaceNumber string  `json:"space_number,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Address     string  `json:"address,omitempty"`
}

// SessionSummary is an active or past parking session row.
type SessionSummary struct {
	ID          string `json:"id,omitempty"`
	Status      string `json:"status,omitempty"`
	MeterNumber string `json:"meter_number,omitempty"`
	ZoneName    string `json:"zone_name,omitempty"`
	Starts      string `json:"starts,omitempty"`
	Ends        string `json:"ends,omitempty"`
	Plate       string `json:"license_plate,omitempty"`
}

// ParkingPreview is a dry-run quote for start/extend.
type ParkingPreview struct {
	Action       string         `json:"action"`
	MeterNumber  string         `json:"meter_number,omitempty"`
	ZoneName     string         `json:"zone_name,omitempty"`
	SessionID    string         `json:"session_id,omitempty"`
	Minutes      int            `json:"minutes,omitempty"`
	DryRun         bool           `json:"dry_run"`
	Message        string         `json:"message,omitempty"`
	UnverifiedNote string         `json:"unverified_note,omitempty"`
	RequestQuery   map[string]any `json:"request_query,omitempty"`
}

// LoginInput for phone/password auth.
type LoginInput struct {
	PhoneNumber string
	Password    string
}

// StartParkingInput for live session start.
type StartParkingInput struct {
	MeterNumber string
	ZoneName    string
	SpaceNumber string
	Minutes     int
	VehicleID   string
	CardID      string
}

// ExtendParkingInput for live session extension.
type ExtendParkingInput struct {
	SessionID string
	Minutes   int
}

// StopParkingInput for live session stop.
type StopParkingInput struct {
	SessionID string
}

func pickString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				if t != "" {
					return t
				}
			case float64:
				if t != 0 {
					b, _ := json.Marshal(t)
					return string(b)
				}
			case json.Number:
				return t.String()
			}
		}
	}
	return ""
}

func pickBool(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
	}
	return false
}

func pickFloat(m map[string]any, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case float64:
				return t
			}
		}
	}
	return 0
}

func asObjectSlice(v any) []map[string]any {
	switch t := v.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, item := range t {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case map[string]any:
		return []map[string]any{t}
	default:
		return nil
	}
}

func findSlice(data map[string]any, keys ...string) []map[string]any {
	for _, k := range keys {
		if v, ok := data[k]; ok {
			if rows := asObjectSlice(v); len(rows) > 0 {
				return rows
			}
		}
	}
	for _, v := range data {
		if m, ok := v.(map[string]any); ok {
			if rows := findSlice(m); len(rows) > 0 {
				return rows
			}
		}
	}
	return nil
}
