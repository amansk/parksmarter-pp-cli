package cli

import (
	"github.com/amansk/parksmarter-pp-cli/internal/client"
	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
	"github.com/amansk/parksmarter-pp-cli/internal/output"
	"github.com/spf13/cobra"
)

func newSessionsCmd(opt *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sessions",
		Aliases: []string{"session"},
		Short:   "List or fetch parking sessions",
	}
	cmd.AddCommand(newSessionsListCmd(opt))
	cmd.AddCommand(newSessionsGetCmd(opt))
	return cmd
}

func newSessionsListCmd(opt *Options) *cobra.Command {
	var past bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active or past parking sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			var sessions []client.SessionSummary
			if past {
				sessions, err = c.ListPastSessions()
			} else {
				sessions, err = c.ListActiveSessions()
			}
			if err != nil {
				return err
			}
			if opt.JSON || opt.Agent {
				return writeOut(cmd, opt, map[string]any{"sessions": sessions, "count": len(sessions), "past": past})
			}
			label := "active"
			if past {
				label = "past"
			}
			if len(sessions) == 0 {
				_, _ = cmd.OutOrStdout().Write([]byte("No " + label + " sessions.\n"))
				return nil
			}
			rows := make([][]string, 0, len(sessions))
			for _, s := range sessions {
				rows = append(rows, []string{s.ID, s.Status, s.MeterNumber, s.ZoneName, s.Starts, s.Ends})
			}
			return output.Table(cmd.OutOrStdout(), []string{"ID", "STATUS", "METER", "ZONE", "STARTS", "ENDS"}, rows)
		},
	}
	cmd.Flags().BoolVar(&past, "past", false, "List past sessions instead of active")
	return cmd
}

func newSessionsGetCmd(opt *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "get <session-id>",
		Short: "Get one session by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opt.newClient()
			if err != nil {
				return err
			}
			s, err := c.GetSession(args[0])
			if err != nil {
				return err
			}
			if s.ID == "" {
				return exitcode.NotFoundf("session %s not found", args[0])
			}
			return writeOut(cmd, opt, s)
		},
	}
}
