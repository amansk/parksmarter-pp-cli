package client

import (
	"net/url"
	"strconv"

	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
)

// NearbyMetersQuery finds meters near coordinates.
type NearbyMetersQuery struct {
	Latitude  float64
	Longitude float64
}

// LookupMeterQuery finds a meter by number or zone/space search terms.
type LookupMeterQuery struct {
	MeterNumber string
	ZoneName    string
	SpaceNumber string
	Address     string
}

// NearbyMeters returns meters near a lat/lng (auth required).
func (c *Client) NearbyMeters(q NearbyMetersQuery) ([]MeterSummary, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	query := url.Values{}
	query.Set("Latitude", formatFloat(q.Latitude))
	query.Set("Longitude", formatFloat(q.Longitude))
	var out map[string]any
	if err := c.getJSON(PathMetersNearby, query, &out); err != nil {
		return nil, err
	}
	return mapMeters(out), nil
}

// LookupMeter resolves a meter by number, zone/space, or address search.
func (c *Client) LookupMeter(q LookupMeterQuery) ([]MeterSummary, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	path := PathMeterByNumber
	query := url.Values{}
	switch {
	case q.MeterNumber != "":
		path = PathMeterByNumber
		query.Set("MeterNumber", q.MeterNumber)
	case q.ZoneName != "" && q.SpaceNumber != "":
		path = PathMetersByZoneSpaceSearch
		query.Set("ZoneName", q.ZoneName)
		query.Set("SpaceNumber", q.SpaceNumber)
	case q.ZoneName != "":
		path = PathMetersByZoneSearch
		query.Set("ZoneName", q.ZoneName)
	case q.Address != "":
		path = PathMetersByLocationAddress
		query.Set("Address", q.Address)
	default:
		return nil, exitcode.Usagef("provide --meter-number, --zone, and/or --address")
	}
	var out map[string]any
	if err := c.getJSON(path, query, &out); err != nil {
		return nil, err
	}
	return mapMeters(out), nil
}

func mapMeters(data map[string]any) []MeterSummary {
	rows := findSlice(data, "Meters", "meters", "MetersResult", "metersResult", "Results", "results")
	out := make([]MeterSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, MeterSummary{
			ID:          pickString(row, "MeterId", "MeterID", "Id", "ID", "id"),
			MeterNumber: pickString(row, "MeterNumber", "meterNumber", "SpaceNumber", "spaceNumber"),
			ZoneName:    pickString(row, "ZoneName", "zoneName", "Zone", "zone"),
			SpaceNumber: pickString(row, "SpaceNumber", "spaceNumber"),
			Latitude:    pickFloat(row, "Latitude", "latitude", "Lat", "lat"),
			Longitude:   pickFloat(row, "Longitude", "longitude", "Lng", "lng", "Lon", "lon"),
			Address:     pickString(row, "Address", "StreetAddress", "address"),
		})
	}
	return out
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
