/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"testing"
)

// Test if the filter validator does a decent job
func TestProbeFilterValidator(t *testing.T) {
	var err error
	var filter *ProbeFilter

	badtag := "*"
	goodtag := "ooo"
	filter = NewProbeFilter()
	filter.FilterTags([]string{badtag})
	err = filter.verifyFilters()
	if err == nil {
		t.Fatalf("Bad tag '%s' not filtered properly", badtag)
	}
	filter = NewProbeFilter()
	filter.FilterTags([]string{"ooo"})
	err = filter.verifyFilters()
	if err != nil {
		t.Fatalf("Good tag '%s' is not allowed", goodtag)
	}

	badcc := "NED"
	goodcc := "NL"
	filter = NewProbeFilter()
	filter.FilterCountry(badcc)
	err = filter.verifyFilters()
	if err == nil {
		t.Fatalf("Bad country code '%s' not filtered properly", badcc)
	}
	filter = NewProbeFilter()
	filter.FilterCountry(goodcc)
	err = filter.verifyFilters()
	if err != nil {
		t.Fatalf("Good country code '%s' is not allowed", goodcc)
	}

	filter = NewProbeFilter()
	filter.Sort("abcd")
	err = filter.verifyFilters()
	if err == nil {
		t.Fatalf("Sort order is not filtered properly")
	}
}
