/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"testing"
)

// Test if the filter validator does a decent job
func TestAnchorFilterValidator(t *testing.T) {
	var err error
	var filter AnchorFilter

	badcc := "NED"
	goodcc := "NL"
	filter = NewAnchorFilter()
	filter.FilterCountry(badcc)
	err = filter.verifyFilters()
	if err == nil {
		t.Errorf("Bad country code '%s' not filtered properly", badcc)
	}
	filter = NewAnchorFilter()
	filter.FilterCountry(goodcc)
	err = filter.verifyFilters()
	if err != nil {
		t.Errorf("Good country code '%s' is not allowed", goodcc)
	}
}
