/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"testing"
)

// Test if the filter validator does a decent job
func TestMeasurementFilterValidator(t *testing.T) {
	var err error
	var filter MeasurementFilter

	badtag := "*"
	goodtag := "ooo"
	filter = NewMeasurementFilter()
	filter.FilterTags([]string{badtag})
	err = filter.verifyFilters()
	if err == nil {
		t.Errorf("Bad tag '%s' not filtered properly", badtag)
	}
	filter = NewMeasurementFilter()
	filter.FilterTags([]string{"ooo"})
	err = filter.verifyFilters()
	if err != nil {
		t.Errorf("Good tag '%s' is not allowed", goodtag)
	}

	filter = NewMeasurementFilter()
	filter.FilterAddressFamily(5)
	err = filter.verifyFilters()
	if err == nil {
		t.Error("Bad 5 af allowed")
	}
	filter = NewMeasurementFilter()
	filter.FilterAddressFamily(4)
	err = filter.verifyFilters()
	if err != nil {
		t.Error("Good af 4 is not allowed")
	}

	badproto := "sctp"
	goodproto := "udp"
	filter = NewMeasurementFilter()
	filter.FilterProtocol(badproto)
	err = filter.verifyFilters()
	if err == nil {
		t.Errorf("Bad protocol '%s' not filtered properly", badtag)
	}
	filter = NewMeasurementFilter()
	filter.FilterProtocol(goodproto)
	err = filter.verifyFilters()
	if err != nil {
		t.Errorf("Good protocol '%s' is not allowed", goodtag)
	}

	filter = NewMeasurementFilter()
	filter.Sort("abcd")
	err = filter.verifyFilters()
	if err == nil {
		t.Error("Sort order is not filtered properly")
	}
}
