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
	if data == "today" {
		now := time.Now().UTC().Unix()
		return time.Unix(now-now%86400, 0), nil
	}
	if data == "yesterday" {
		now := time.Now().UTC().Unix()
		return time.Unix(now-now%86400-86400, 0), nil
	}

	// try parsing as UNIX epoch first
	epoch, err := strconv.Atoi(data)
	if err == nil {
		return time.Unix(int64(epoch), 0), nil
	}

	var t time.Time
	layout := "2006-01-02T15:04:05Z"
	// try parsing variations of (shortened) ISO8601
	// perhaps there's a nicer way of doing this
	t, err = time.Parse(layout, data)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(layout, data+"Z")
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(layout, data+":00Z")
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(layout, data+":00:00Z")
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(layout, data+"00:00:00Z")
	if err == nil {
		return t, nil
	}
	t, err = time.Parse(layout, data+"T00:00:00Z")
	if err == nil {
		return t, nil
	}

	// we failed. Return an error; the time value can be ignored
	return time.Now().UTC(), fmt.Errorf("could not parse time: %s", data)
}
