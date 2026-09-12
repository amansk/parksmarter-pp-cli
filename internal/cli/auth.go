package cli

import (
	"os"

	"github.com/amansk/parksmarter-pp-cli/internal/auth"
	"github.com/amansk/parksmarter-pp-cli/internal/client"
	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
	"github.com/spf13/cobra"
)

func newAuthCmd(opt *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Park Smarter bearer authentication",
	}
	cmd.AddCommand(newAuthLoginCmd(opt))
	cmd.AddCommand(newAuthStatusCmd(opt))
	return cmd
}

func newAuthLoginCmd(opt *Options) *cobra.Command {
	var phone, password, tokenFile string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Obtain and store a Park Smarter auth token",
		Long:  "Login with phone/password (via flags or PARKSMARTER_PHONE/PARKSMARTER_PASSWORD env), import a token file, or rely on PARKSMARTER_TOKEN. Tokens are stored mode 0600; secrets are never printed.",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := opt.ResolveHome()
			if err != nil {
				return err
			}
			var sess *auth.Session
			switch {
			case tokenFile != "":
				data, err := os.ReadFile(tokenFile)
				if err != nil {
					return exitcode.Usagef("read token file: %v", err)
				}
				sess, err = auth.ParseTokenFile(data, "token-file:"+tokenFile)
				if err != nil {
					return exitcode.Authf("%v", err)
				}
			default:
				phoneVal, err := auth.ResolvePhone(phone)
				if err != nil {
					return exitcode.Usagef("%v", err)
				}
				passVal, err := auth.ResolvePassword(password)
				if err != nil {
					return exitcode.Usagef("%v", err)
				}
				c := client.New(nil)
				if opt.HTTP != nil {
					c = opt.HTTP
				}
				sess, err = c.LoginWithPhonePassword(client.LoginInput{
					PhoneNumber: phoneVal,
					Password:    passVal,
				})
				if err != nil {
					return err
				}
			}
			if sess == nil || sess.Token() == "" {
				return exitcode.Authf("login did not return an auth token")
			}
			if err := auth.SaveSession(home, sess); err != nil {
				return err
			}
			status := sess.Status()
			status["message"] = "session saved"
			status["home"] = home
			return writeOut(cmd, opt, status)
		},
	}
	cmd.Flags().StringVar(&phone, "phone", "", "Account phone (E.164 preferred) or $PARKSMARTER_PHONE")
	cmd.Flags().StringVar(&password, "password", "", "Account password or $PARKSMARTER_PASSWORD")
	cmd.Flags().StringVar(&tokenFile, "token-file", "", "Import bearer token (raw string or {\"auth_token\":\"...\"} JSON)")
	return cmd
}

func newAuthStatusCmd(opt *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status (no secrets)",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := opt.ResolveHome()
			if err != nil {
				return err
			}
			sess, err := auth.ResolveSession(home)
			if err != nil {
				return err
			}
			out := map[string]any{"home": home, "api_base": client.DefaultBaseURL}
			if sess == nil || sess.Token() == "" {
				out["authenticated"] = false
				out["hint"] = "run: parksmarter-pp-cli auth login --phone … (or --token-file)"
				return writeOut(cmd, opt, out)
			}
			for k, v := range sess.Status() {
				out[k] = v
			}
			if env := os.Getenv(auth.TokenEnv); env != "" {
				out["env_override"] = auth.TokenEnv
			}
			return writeOut(cmd, opt, out)
		},
	}
}
