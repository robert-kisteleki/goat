/*
  (C) 2022, 2023 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Turn a slice of ints to a comma CSV string
func makeCsv(list []uint) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(list)), ","), "[]")
}

func apiGetRequest(
	verbose bool,
	url string,
	key *uuid.UUID,
) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", uaString)
	if key != nil {
		req.Header.Set("Authorization", "Key "+(*key).String())
	}

	if verbose {
		msg := fmt.Sprintf("# API call: GET %s", url)
		if key != nil {
			msg += fmt.Sprintf(" (using API key %s...)", (*key).String()[:8])
		}
		fmt.Println(msg)
	}
	client := &http.Client{}

	return client.Do(req)
}
