package client

import "net/url"

// QueryValuesForTest exposes url.Values as a plain map for CLI dry-run output.
func QueryValuesForTest(q url.Values) map[string]any {
	return queryToMap(q)
}
