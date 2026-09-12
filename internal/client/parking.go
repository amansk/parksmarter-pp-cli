package client

import (
	"net/url"
	"strconv"

	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
)

const (
	startConfirmPhrase  = "START PARK SMARTER PARKING"
	extendConfirmPhrase = "EXTEND PARK SMARTER PARKING"
	stopConfirmPhrase   = "STOP PARK SMARTER PARKING"
)

// Confirm phrases exported for CLI validation.
const (
	StartConfirmPhrase  = startConfirmPhrase
	ExtendConfirmPhrase = extendConfirmPhrase
	StopConfirmPhrase   = stopConfirmPhrase
)

// PreviewStart builds a start-session preview (no charge).
func (c *Client) PreviewStart(in StartParkingInput) (ParkingPreview, url.Values, error) {
	q, err := c.buildStartQuery(in)
	if err != nil {
		return ParkingPreview{}, nil, err
	}
	return ParkingPreview{
		Action:       "start",
		MeterNumber:  in.MeterNumber,
		ZoneName:     in.ZoneName,
		Minutes:      in.Minutes,
		DryRun:       true,
		Message:      "Preview only — no charge. Use parking start with all safety gates to commit.",
		RequestQuery: queryToMap(q),
	}, q, nil
}

// StartSession starts a live parking session (mutating GET — APK verified Sep 2026).
func (c *Client) StartSession(in StartParkingInput) (map[string]any, error) {
	q, err := c.buildStartQuery(in)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.getJSON(PathStartSession, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PreviewExtend builds an extend-session preview.
func (c *Client) PreviewExtend(in ExtendParkingInput) (ParkingPreview, url.Values, error) {
	q, err := c.buildExtendQuery(in)
	if err != nil {
		return ParkingPreview{}, nil, err
	}
	return ParkingPreview{
		Action:       "extend",
		SessionID:    in.SessionID,
		Minutes:      in.Minutes,
		DryRun:       true,
		Message:      "Preview only — no charge. Use parking extend with all safety gates to commit.",
		RequestQuery: queryToMap(q),
	}, q, nil
}

// ExtendSession extends a live parking session.
func (c *Client) ExtendSession(in ExtendParkingInput) (map[string]any, error) {
	q, err := c.buildExtendQuery(in)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.getJSON(PathExtendSession, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// PreviewStop builds a stop-session preview.
func (c *Client) PreviewStop(in StopParkingInput) (ParkingPreview, url.Values, error) {
	q, err := c.buildStopQuery(in)
	if err != nil {
		return ParkingPreview{}, nil, err
	}
	return ParkingPreview{
		Action:       "stop",
		SessionID:    in.SessionID,
		DryRun:       true,
		Message:      "Preview only. Use parking stop with all safety gates to commit.",
		RequestQuery: queryToMap(q),
	}, q, nil
}

// StopSession stops a live parking session.
func (c *Client) StopSession(in StopParkingInput) (map[string]any, error) {
	q, err := c.buildStopQuery(in)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := c.getJSON(PathStopSession, q, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) buildStartQuery(in StartParkingInput) (url.Values, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	if in.Minutes <= 0 {
		return nil, exitcode.Usagef("--minutes is required")
	}
	if in.MeterNumber == "" && in.ZoneName == "" {
		return nil, exitcode.Usagef("--meter-number or --zone is required")
	}
	q := url.Values{}
	if in.MeterNumber != "" {
		q.Set("MeterNumber", in.MeterNumber)
	}
	if in.ZoneName != "" {
		q.Set("ZoneName", in.ZoneName)
	}
	if in.SpaceNumber != "" {
		q.Set("SpaceNumber", in.SpaceNumber)
	}
	q.Set("Minutes", formatInt(in.Minutes))
	if in.VehicleID != "" {
		q.Set("VehicleId", in.VehicleID)
	}
	if in.CardID != "" {
		q.Set("CardId", in.CardID)
	}
	return q, nil
}

func (c *Client) buildExtendQuery(in ExtendParkingInput) (url.Values, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	if in.SessionID == "" {
		return nil, exitcode.Usagef("--session-id is required")
	}
	if in.Minutes <= 0 {
		return nil, exitcode.Usagef("--minutes is required")
	}
	q := url.Values{}
	q.Set("SessionId", in.SessionID)
	q.Set("Minutes", formatInt(in.Minutes))
	return q, nil
}

func (c *Client) buildStopQuery(in StopParkingInput) (url.Values, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	if in.SessionID == "" {
		return nil, exitcode.Usagef("--session-id is required")
	}
	q := url.Values{}
	q.Set("SessionId", in.SessionID)
	return q, nil
}

func formatInt(v int) string {
	return strconv.Itoa(v)
}

func queryToMap(q url.Values) map[string]any {
	out := map[string]any{}
	for k, vals := range q {
		if len(vals) == 1 {
			out[k] = vals[0]
		} else {
			out[k] = vals
		}
	}
	return out
}
