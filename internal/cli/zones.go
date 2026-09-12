package cli

import (
	"strconv"

	"github.com/amansk/parksmarter-pp-cli/internal/client"
	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
	"github.com/amansk/parksmarter-pp-cli/internal/output"
	"github.com/spf13/cobra"
)

func newZonesCmd(opt *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "zones",
		Aliases: []string{"zone", "meters", "meter"},
		Short:   "Look up IPS meters/zones/spaces",
	}
	cmd.AddCommand(newZonesLookupCmd(opt))
	cmd.AddCommand(newZonesNearbyCmd(opt))
	return cmd
}

func newZonesLookupCmd(opt *Options) *cobra.Command {
	var meterNumber, zone, space, address string
	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Find meter(s) by meter number, zone/space, or address",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			meters, err := c.LookupMeter(client.LookupMeterQuery{
				MeterNumber: meterNumber,
				ZoneName:    zone,
				SpaceNumber: space,
				Address:     address,
			})
			if err != nil {
				return err
			}
			return writeMeters(cmd, opt, meters)
		},
	}
	cmd.Flags().StringVar(&meterNumber, "meter-number", "", "Meter or space number")
	cmd.Flags().StringVar(&zone, "zone", "", "Zone name or id")
	cmd.Flags().StringVar(&space, "space", "", "Space number (with --zone)")
	cmd.Flags().StringVar(&address, "address", "", "Street address search")
	return cmd
}

func newZonesNearbyCmd(opt *Options) *cobra.Command {
	var lat, lng float64
	cmd := &cobra.Command{
		Use:   "nearby",
		Short: "Find meters near latitude/longitude",
		RunE: func(cmd *cobra.Command, args []string) error {
			if lat == 0 && lng == 0 {
				return exitcode.Usagef("--lat and --lng are required")
			}
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			meters, err := c.NearbyMeters(client.NearbyMetersQuery{Latitude: lat, Longitude: lng})
			if err != nil {
				return err
			}
			return writeMeters(cmd, opt, meters)
		},
	}
	cmd.Flags().Float64Var(&lat, "lat", 0, "Latitude")
	cmd.Flags().Float64Var(&lng, "lng", 0, "Longitude")
	return cmd
}

func writeMeters(cmd *cobra.Command, opt *Options, meters []client.MeterSummary) error {
	if opt.JSON || opt.Agent {
		return writeOut(cmd, opt, map[string]any{"meters": meters, "count": len(meters)})
	}
	if len(meters) == 0 {
		_, _ = cmd.OutOrStdout().Write([]byte("No meters found.\n"))
		return nil
	}
	rows := make([][]string, 0, len(meters))
	for _, m := range meters {
		rows = append(rows, []string{
			m.ID,
			m.MeterNumber,
			m.ZoneName,
			m.SpaceNumber,
			strconv.FormatFloat(m.Latitude, 'f', 5, 64),
			strconv.FormatFloat(m.Longitude, 'f', 5, 64),
		})
	}
	return output.Table(cmd.OutOrStdout(), []string{"ID", "METER", "ZONE", "SPACE", "LAT", "LNG"}, rows)
}
