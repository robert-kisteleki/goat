/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Turn a string of comma separated integers into a slice of ints
func makeIntList(csv string) ([]uint, error) {
	idlist := make([]uint, 0)
	separated := strings.Split(csv, ",")
	for _, ids := range separated {
		id, err := strconv.Atoi(ids)
		if err != nil {
			return nil, err
		}
		idlist = append(idlist, uint(id))
	}

	return idlist, nil
}

// We allow datetimes to be supplied as UNIX epoch or (some short versions of) ISO8601
// or perhaps "today" or "yesterday"
// This function parses those into a time.Time
// With ISO8601: SS, MM:SS and HH:MM:SS are optional and default to 0
func parseTimeAlternatives(data string) (time.Time, error) {
	switch data {
	case "":
		return time.Now(), fmt.Errorf("cannot parse date/time '%s'", data)
	case "today":
		now := time.Now().UTC().Unix()
		return time.Unix(now-now%86400, 0), nil
	case "yesterday":
		now := time.Now().UTC().Unix()
		return time.Unix(now-now%86400-86400, 0), nil
	case "tomorrow":
		now := time.Now().UTC().Unix()
		return time.Unix(now-now%86400+86400, 0), nil
	}

	// try parsing as UNIX epoch first
	epoch, err := strconv.Atoi(data)
	if err == nil {
		return time.Unix(int64(epoch), 0), nil
	}

	// try various shortened versions of ISO8601
	const format = "2006-01-02T15:04:05Z"
	parseformat := format
	if len(data) < len(format) {
		parseformat = format[:len(data)]
	}
	return time.Parse(parseformat, data)
}

type multioption []string

func (o *multioption) String() string {
	return fmt.Sprint(*o)
}

func (o *multioption) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		*o = append(*o, part)
	}
	return nil
}
