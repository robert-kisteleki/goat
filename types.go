/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Geolocation type
type Geolocation struct {
	Type        string    `json:"type"`
	Coordinates []float32 `json:"coordinates"`
}

// Tag type
type Tag struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ErrorResponse type
type MultiErrorResponse struct {
	ErrorDetail
	Error  ErrorDetail    `json:"error"`
	Errors []ErrorMessage `json:"errors"`
}

type ErrorMessage struct {
	Source ErrorSource `json:"source"`
	Detail string      `json:"detail"`
}

type ErrorSource struct {
	Pointer string `json:"pointer"`
}

// ErrorDetails type
type ErrorDetail struct {
	Detail string         `json:"detail"`
	Status int            `json:"status"`
	Title  string         `json:"title"`
	Code   int            `json:"code"`
	Errors []ErrorMessage `json:"errors"`
}

// valueOrNA turns various types into a string if they have values
// (i.e. pointer is not nil) or "N/A" otherwise
func valueOrNA[T any](prefix string, quote bool, val *T) string {
	if val != nil {
		if quote {
			return fmt.Sprintf("\t\"%s%v\"", prefix, *val)
		} else {
			return fmt.Sprintf("\t%s%v", prefix, *val)
		}
	} else {
		return "\tN/A"
	}
}

// a datetime type that can be unmarshaled from UNIX epoch *or* ISO times
type uniTime time.Time

// output is ISO8601(Z) down to seconds
func (ut *uniTime) MarshalJSON() (b []byte, e error) {
	layout := "2006-01-02T15:04:05Z"
	return []byte("\"" + time.Time(*ut).UTC().Format(layout) + "\""), nil
}

func (ut *uniTime) UnmarshalJSON(data []byte) error {
	// try parsing as UNIX epoch first
	epoch, err := strconv.Atoi(string(data))
	if err == nil {
		*ut = uniTime(time.Unix(int64(epoch), 0))
		return nil
	}

	// try parsing ISO8601(Z)
	layout := "2006-01-02T15:04:05"
	noquote := strings.ReplaceAll(string(data), "\"", "")
	noz := strings.ReplaceAll(noquote, "Z", "")
	unix, err := time.Parse(layout, noz)
	if err != nil {
		return err
	}
	*ut = uniTime(unix)
	return nil
}

// default output format for uniTime type is ISO8601
func (ut uniTime) String() string {
	return time.Time(ut).UTC().Format(time.RFC3339)
}

// something went wrong; see if the error page can be parsed
// it could be a single error or a bunch of them
func parseAPIError(resp *http.Response) error {
	var err error

	var decoded MultiErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&decoded)
	if err != nil {
		return err
	}

	r := make([]string, 0)
	if decoded.Status != 0 {
		r = append(r, decoded.Title)
		for _, e := range decoded.Errors {
			r = append(r, e.Detail)
		}
	}

	if decoded.Error.Status != 0 {
		r = append(r, decoded.Error.Title)
		r = append(r, decoded.Error.Detail)
		for _, e := range decoded.Error.Errors {
			r = append(r, e.Detail)
		}
	}

	return fmt.Errorf("%d %s %s", decoded.Status, decoded.Title, strings.Join(r, ", "))
}
