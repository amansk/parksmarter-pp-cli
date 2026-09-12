package cli

import (
	"github.com/amansk/parksmarter-pp-cli/internal/client"
	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
	"github.com/spf13/cobra"
)

func newParkingCmd(opt *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "parking",
		Short: "Preview or mutate parking sessions (mutations are hard-gated)",
	}
	cmd.AddCommand(newParkingPreviewCmd(opt))
	cmd.AddCommand(newParkingStartCmd(opt))
	cmd.AddCommand(newParkingExtendCmd(opt))
	cmd.AddCommand(newParkingStopCmd(opt))
	return cmd
}

func writeMutationResult(cmd *cobra.Command, opt *Options, verb string, out map[string]any) error {
	payload := map[string]any{"result": out}
	if dry, _ := out["dry_run"].(bool); dry {
		payload["dry_run"] = true
		payload["unverified"] = true
	} else {
		payload[verb] = true
		payload["unverified"] = true
	}
	return writeOut(cmd, opt, payload)
}

func newParkingPreviewCmd(opt *Options) *cobra.Command {
	var action, meterNumber, zone, space, sessionID, vehicleID, cardID string
	var minutes int
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview start/extend/stop without charging",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			switch action {
			case "start", "":
				if action == "" {
					action = "start"
				}
				preview, _, err := c.PreviewStart(client.StartParkingInput{
					MeterNumber: meterNumber,
					ZoneName:    zone,
					SpaceNumber: space,
					Minutes:     minutes,
					VehicleID:   vehicleID,
					CardID:      cardID,
				})
				if err != nil {
					return err
				}
				return writeOut(cmd, opt, preview)
			case "extend":
				preview, _, err := c.PreviewExtend(client.ExtendParkingInput{SessionID: sessionID, Minutes: minutes})
				if err != nil {
					return err
				}
				return writeOut(cmd, opt, preview)
			case "stop":
				preview, _, err := c.PreviewStop(client.StopParkingInput{SessionID: sessionID})
				if err != nil {
					return err
				}
				return writeOut(cmd, opt, preview)
			default:
				return exitcode.Usagef("--action must be start, extend, or stop")
			}
		},
	}
	cmd.Flags().StringVar(&action, "action", "start", "start, extend, or stop")
	cmd.Flags().StringVar(&meterNumber, "meter-number", "", "Meter number (start)")
	cmd.Flags().StringVar(&zone, "zone", "", "Zone name (start)")
	cmd.Flags().StringVar(&space, "space", "", "Space number (start)")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "Session id (extend/stop)")
	cmd.Flags().IntVar(&minutes, "minutes", 0, "Minutes to purchase (start/extend)")
	cmd.Flags().StringVar(&vehicleID, "vehicle-id", "", "Saved vehicle id override")
	cmd.Flags().StringVar(&cardID, "card-id", "", "Saved card id override")
	return cmd
}

func newParkingStartCmd(opt *Options) *cobra.Command {
	var meterNumber, zone, space, vehicleID, cardID, confirm string
	var minutes int
	var enableLive, ownerApproved bool
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a live parking session (requires explicit safety gates)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !enableLive || !ownerApproved || confirm != client.StartConfirmPhrase {
				return exitcode.Usagef("refusing live parking: require --enable-live-parking --owner-approved --confirm %q", client.StartConfirmPhrase)
			}
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			out, err := c.StartSession(client.StartParkingInput{
				MeterNumber: meterNumber,
				ZoneName:    zone,
				SpaceNumber: space,
				Minutes:     minutes,
				VehicleID:   vehicleID,
				CardID:      cardID,
			})
			if err != nil {
				return err
			}
			return writeMutationResult(cmd, opt, "started", out)
		},
	}
	addParkingStartFlags(cmd, &meterNumber, &zone, &space, &minutes, &vehicleID, &cardID, &enableLive, &ownerApproved, &confirm, client.StartConfirmPhrase)
	return cmd
}

func newParkingExtendCmd(opt *Options) *cobra.Command {
	var sessionID, confirm string
	var minutes int
	var enableLive, ownerApproved bool
	cmd := &cobra.Command{
		Use:   "extend",
		Short: "Extend a live parking session (requires explicit safety gates)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !enableLive || !ownerApproved || confirm != client.ExtendConfirmPhrase {
				return exitcode.Usagef("refusing live parking: require --enable-live-parking --owner-approved --confirm %q", client.ExtendConfirmPhrase)
			}
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			out, err := c.ExtendSession(client.ExtendParkingInput{SessionID: sessionID, Minutes: minutes})
			if err != nil {
				return err
			}
			return writeMutationResult(cmd, opt, "extended", out)
		},
	}
	cmd.Flags().StringVar(&sessionID, "session-id", "", "Active session id")
	cmd.Flags().IntVar(&minutes, "minutes", 0, "Minutes to add")
	cmd.Flags().BoolVar(&enableLive, "enable-live-parking", false, "Explicit opt-in to charge a payment method")
	cmd.Flags().BoolVar(&ownerApproved, "owner-approved", false, "Explicit owner approval")
	cmd.Flags().StringVar(&confirm, "confirm", "", "Must be exactly: "+client.ExtendConfirmPhrase)
	return cmd
}

func newParkingStopCmd(opt *Options) *cobra.Command {
	var sessionID, confirm string
	var enableLive, ownerApproved bool
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop/end a live parking session (requires explicit safety gates)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !enableLive || !ownerApproved || confirm != client.StopConfirmPhrase {
				return exitcode.Usagef("refusing live parking: require --enable-live-parking --owner-approved --confirm %q", client.StopConfirmPhrase)
			}
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			out, err := c.StopSession(client.StopParkingInput{SessionID: sessionID})
			if err != nil {
				return err
			}
			return writeMutationResult(cmd, opt, "stopped", out)
		},
	}
	cmd.Flags().StringVar(&sessionID, "session-id", "", "Active session id")
	cmd.Flags().BoolVar(&enableLive, "enable-live-parking", false, "Explicit opt-in to end session remotely")
	cmd.Flags().BoolVar(&ownerApproved, "owner-approved", false, "Explicit owner approval")
	cmd.Flags().StringVar(&confirm, "confirm", "", "Must be exactly: "+client.StopConfirmPhrase)
	return cmd
}

func addParkingStartFlags(cmd *cobra.Command, meterNumber, zone, space *string, minutes *int, vehicleID, cardID *string, enableLive, ownerApproved *bool, confirm *string, phrase string) {
	cmd.Flags().StringVar(meterNumber, "meter-number", "", "Meter number")
	cmd.Flags().StringVar(zone, "zone", "", "Zone name")
	cmd.Flags().StringVar(space, "space", "", "Space number")
	cmd.Flags().IntVar(minutes, "minutes", 0, "Minutes to purchase")
	cmd.Flags().StringVar(vehicleID, "vehicle-id", "", "Saved vehicle id override")
	cmd.Flags().StringVar(cardID, "card-id", "", "Saved card id override")
	cmd.Flags().BoolVar(enableLive, "enable-live-parking", false, "Explicit opt-in to charge a payment method")
	cmd.Flags().BoolVar(ownerApproved, "owner-approved", false, "Explicit owner approval")
	cmd.Flags().StringVar(confirm, "confirm", "", "Must be exactly: "+phrase)
}
