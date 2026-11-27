/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"encoding/json"
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

// Test if the country list filter is verified and mapped properly
func TestProbeFilterCountryIn(t *testing.T) {
	for _, good := range [][]string{{"NL"}, {"NL", "de"}} {
		filter := NewProbeFilter()
		filter.FilterCountryIn(good)
		if err := filter.verifyFilters(); err != nil {
			t.Fatalf("Good country list %q is not allowed: %v", good, err)
		}
	}

	for _, bad := range [][]string{{}, {""}, {"N"}, {"NLD"}, {"N1"}, {"NL", ""}, {"NL", " DE"}, {"NL;DE"}} {
		filter := NewProbeFilter()
		filter.FilterCountryIn(bad)
		if err := filter.verifyFilters(); err == nil {
			t.Fatalf("Bad country list %q not filtered properly", bad)
		}
	}

	filter := NewProbeFilter()
	filter.FilterCountryIn([]string{"NL", "DE"})
	if got := filter.params.Get("country_code__in"); got != "NL,DE" {
		t.Fatalf("Country list parameter is '%s', expected 'NL,DE'", got)
	}
}

// Test if the firmware version of a probe is parsed
func TestProbeParseFirmwareVersion(t *testing.T) {
	var probe Probe

	err := json.Unmarshal([]byte(`{"id":6001,"firmware_version":4900}`), &probe)
	if err != nil {
		t.Fatalf("Error parsing probe: %v", err)
	}
	if probe.FirmwareVer != 4900 {
		t.Fatalf("Firmware version is %d, expected 4900", probe.FirmwareVer)
	}
}
