package client

import (
	"encoding/json"

	"github.com/amansk/parksmarter-pp-cli/internal/auth"
	"github.com/amansk/parksmarter-pp-cli/internal/exitcode"
)

// LoginWithPhonePassword exchanges phone/password for bearer tokens.
func (c *Client) LoginWithPhonePassword(in LoginInput) (*auth.Session, error) {
	body := map[string]string{
		"PhoneNumber": in.PhoneNumber,
		"Password":    in.Password,
	}
	raw, code, err := c.do("POST", PathLoginPhonePassword, nil, body)
	if err != nil {
		return nil, err
	}
	sess, perr := parseLoginResponse(raw, "login:phone")
	if perr == nil && sess != nil {
		return sess, nil
	}
	if code == 200 && len(raw) == 0 {
		return nil, exitcode.Authf("login failed (empty response); verify phone/password field names in PLAN.md")
	}
	if perr != nil {
		return nil, perr
	}
	return nil, exitcode.Authf("login failed; no auth token in response")
}

// LoginWithCachedToken refreshes a stored token (APK action name verified).
func (c *Client) LoginWithCachedToken(token string) (*auth.Session, error) {
	body := map[string]string{"AuthToken": token}
	raw, _, err := c.do("POST", PathLoginCachedToken, nil, body)
	if err != nil {
		return nil, err
	}
	sess, perr := parseLoginResponse(raw, "login:cached")
	if perr != nil {
		return nil, perr
	}
	if sess == nil {
		return nil, exitcode.Authf("cached token login failed")
	}
	return sess, nil
}

func parseLoginResponse(raw json.RawMessage, source string) (*auth.Session, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, exitcode.APIf("decode login response: %v", err)
	}
	tok := pickString(generic,
		"AuthToken", "authToken", "Token", "token", "AccessToken", "accessToken", "UserToken", "userToken",
	)
	ref := pickString(generic, "RefreshToken", "refreshToken")
	if tok == "" {
		// Some responses nest under Result/Data.
		for _, k := range []string{"Result", "Data", "LoginResult", "data", "result"} {
			if nested, ok := generic[k].(map[string]any); ok {
				tok = pickString(nested, "AuthToken", "authToken", "Token", "token", "AccessToken")
				ref = pickString(nested, "RefreshToken", "refreshToken")
				if tok != "" {
					break
				}
			}
		}
	}
	if tok == "" {
		return nil, nil
	}
	return &auth.Session{
		AuthToken:    tok,
		RefreshToken: ref,
		Source:       source,
	}, nil
}
