package common

import "strings"

// TrimAuthURL normalizes configured AAI endpoint values to auth root URL form.
func TrimAuthURL(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimSuffix(endpoint, "/")
	endpoint = strings.TrimSuffix(endpoint, "/oauth2/userinfo")

	return endpoint
}
