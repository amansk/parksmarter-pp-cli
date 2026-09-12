package cli

import (
	"github.com/amansk/parksmarter-pp-cli/internal/output"
	"github.com/spf13/cobra"
)

func newAccountCmd(opt *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Read consumer account profile data (no secrets)",
	}
	cmd.AddCommand(newAccountCardsCmd(opt))
	cmd.AddCommand(newAccountVehiclesCmd(opt))
	return cmd
}

func newAccountCardsCmd(opt *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "cards",
		Short: "List saved payment methods (ids + last4 only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			cards, err := c.ListCards()
			if err != nil {
				return err
			}
			if opt.JSON || opt.Agent {
				out := make([]map[string]any, 0, len(cards))
				for _, card := range cards {
					item := map[string]any{"id": card.ID}
					if card.Last4 != "" {
						item["last4"] = card.Last4
					}
					if card.Brand != "" {
						item["brand"] = card.Brand
					}
					if card.Default {
						item["is_default"] = true
					}
					out = append(out, item)
				}
				return writeOut(cmd, opt, map[string]any{"cards": out, "count": len(out)})
			}
			if len(cards) == 0 {
				_, _ = cmd.OutOrStdout().Write([]byte("No saved cards.\n"))
				return nil
			}
			rows := make([][]string, 0, len(cards))
			for _, card := range cards {
				def := ""
				if card.Default {
					def = "yes"
				}
				rows = append(rows, []string{card.ID, card.Last4, card.Brand, def})
			}
			return output.Table(cmd.OutOrStdout(), []string{"ID", "LAST4", "BRAND", "DEFAULT"}, rows)
		},
	}
}

func newAccountVehiclesCmd(opt *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "vehicles",
		Short: "List saved vehicles",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			vehicles, err := c.ListVehicles()
			if err != nil {
				return err
			}
			if opt.JSON || opt.Agent {
				return writeOut(cmd, opt, map[string]any{"vehicles": vehicles, "count": len(vehicles)})
			}
			if len(vehicles) == 0 {
				_, _ = cmd.OutOrStdout().Write([]byte("No saved vehicles.\n"))
				return nil
			}
			rows := make([][]string, 0, len(vehicles))
			for _, v := range vehicles {
				def := ""
				if v.Default {
					def = "yes"
				}
				rows = append(rows, []string{v.ID, v.LicensePlate, v.State, def})
			}
			return output.Table(cmd.OutOrStdout(), []string{"ID", "PLATE", "STATE", "DEFAULT"}, rows)
		},
	}
}
