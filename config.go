/*
  (C) Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

const version = "v0.7.3"

var (
	uaString      = "goat " + version
	apiBaseURL    = "https://atlas.ripe.net/api/v2/"
	streamBaseURL = "wss://atlas-stream.ripe.net/stream/"
)

// GetUserAgent returns the user agent as a string
func UserAgent() string {
	return uaString
}

// SetAPIBase allows the caller to modify the API to talk to
// This is really only useful to developers who have access to compatible APIs
func SetAPIBase(newAPIBaseURL string) {
	// TODO: check sanity of new API base URL
	apiBaseURL = newAPIBaseURL
}

// SetStreamBase allows the caller to modify the stream to talk to
// This is really only useful to developers who have access to compatible APIs
func SetStreamBase(newStreamBaseURL string) {
	// TODO: check sanity of new API base URL
	streamBaseURL = newStreamBaseURL
}

// GetStreamBase retrieves the currently configured stream base URL
func GetStreamBase() string {
	return streamBaseURL
}
