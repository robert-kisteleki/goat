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

// We allow datetimes to be supplied as UNIX epoch or ISO8601
// This function parses that into a Time
func parseTimeAlternatives(data string) (time.Time, error) {
	// try parsing as UNIX epoch first
	epoch, err := strconv.Atoi(data)
	if err == nil {
		return time.Unix(int64(epoch), 0), nil
	}

	// try parsing ISO8601
	layout := "2006-01-02T15:04:05Z"
	if !strings.Contains(data, "Z") {
		data += "Z"
	}
	unix, err := time.Parse(layout, data)
	if err == nil {
		return unix, nil
	}

	// we failed. Return an error; the time can be ignored
	return time.Now(), fmt.Errorf("could not parse time: %s", data)
}
