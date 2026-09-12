package cli

import (
	"github.com/amansk/parksmarter-pp-cli/internal/auth"
	"github.com/amansk/parksmarter-pp-cli/internal/client"
	"github.com/spf13/cobra"
)

func newDoctorCmd(opt *Options) *cobra.Command {
	var live bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check local setup and optional live API connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := opt.ResolveHome()
			if err != nil {
				return err
			}
			sess, err := auth.ResolveSession(home)
			if err != nil {
				return err
			}
			report := map[string]any{
				"ok":          true,
				"cli_version": version,
				"home":        home,
				"api_base":    client.DefaultBaseURL,
				"auth_type":   "bearer",
				"auth":        map[string]any{"authenticated": false},
				"checks":      []map[string]any{},
			}
			if sess != nil {
				report["auth"] = sess.Status()
			}
			checks := report["checks"].([]map[string]any)
			checks = append(checks, map[string]any{"name": "config_dir", "ok": true, "detail": home})
			if sess == nil || sess.Token() == "" {
				checks = append(checks, map[string]any{"name": "session", "ok": false, "detail": "not authenticated; run auth login"})
				report["ok"] = false
			} else {
				checks = append(checks, map[string]any{"name": "session", "ok": true, "detail": "auth token present"})
			}
			if live {
				c, err := opt.newClient()
				if err != nil {
					return err
				}
				if err := c.Ping(); err != nil {
					checks = append(checks, map[string]any{"name": "live_ping", "ok": false, "detail": err.Error()})
					report["ok"] = false
				} else {
					checks = append(checks, map[string]any{"name": "live_ping", "ok": true, "detail": "GET /api/ApplicationValidity succeeded"})
				}
				if sess != nil && sess.Token() != "" {
					if err := c.ProbeSessionAuth(); err != nil {
						checks = append(checks, map[string]any{"name": "live_session", "ok": false, "detail": err.Error()})
						report["ok"] = false
					} else {
						checks = append(checks, map[string]any{"name": "live_session", "ok": true, "detail": "GET /api/ParkingSession/GetActiveParkingSessions succeeded"})
					}
				}
			}
			report["checks"] = checks
			return writeOut(cmd, opt, report)
		},
	}
	cmd.Flags().BoolVar(&live, "live", false, "Probe live apiv3.parksmarter.com (read-only)")
	return cmd
}
