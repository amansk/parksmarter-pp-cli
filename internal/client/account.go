package client

import (
	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
)

// ListCards returns saved payment methods (ids + last4 only).
func (c *Client) ListCards() ([]CardSummary, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	var out map[string]any
	if err := c.postJSON(PathCards, map[string]any{}, &out); err != nil {
		return nil, err
	}
	return mapCards(out), nil
}

// ListVehicles returns saved vehicles.
func (c *Client) ListVehicles() ([]VehicleSummary, error) {
	if c.Session == nil || c.Session.Token() == "" {
		return nil, exitcode.Authf("authenticated session required; run auth login")
	}
	var out map[string]any
	if err := c.postJSON(PathVehicles, map[string]any{}, &out); err != nil {
		return nil, err
	}
	return mapVehicles(out), nil
}

func mapCards(data map[string]any) []CardSummary {
	rows := findSlice(data, "Cards", "cards", "CardsResult", "cardsResult", "Results", "results")
	out := make([]CardSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, CardSummary{
			ID:      pickString(row, "CardId", "CardID", "Id", "ID", "cardId", "id"),
			Last4:   pickString(row, "Last4", "CardLast4", "last4", "LastFour", "lastFour"),
			Brand:   pickString(row, "Brand", "CardBrand", "brand"),
			Default: pickBool(row, "IsDefault", "isDefault", "Default", "default"),
		})
	}
	return out
}

func mapVehicles(data map[string]any) []VehicleSummary {
	rows := findSlice(data, "Vehicles", "vehicles", "VehiclesResult", "vehiclesResult", "Results", "results")
	out := make([]VehicleSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, VehicleSummary{
			ID:           pickString(row, "VehicleId", "VehicleID", "Id", "ID", "id"),
			LicensePlate: pickString(row, "LicensePlate", "LicensePlateNumber", "Plate", "licensePlate"),
			State:        pickString(row, "State", "LicensePlateState", "PlateState", "state"),
			Default:      pickBool(row, "IsDefault", "isDefault", "Default", "default"),
		})
	}
	return out
}
