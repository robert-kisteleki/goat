/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"net/url"
)

// Anchor object, as it comes from the API
type Anchor struct {
	ID              uint        `json:"id"`
	Address4        *netip.Addr `json:"ip_v4"`
	ASN4            *uint       `json:"as_v4"`
	IPv4Gateway     *netip.Addr `json:"ip_v4_gateway"`
	IPv4Netmask     *netip.Addr `json:"ip_v4_netmask"`
	Address6        *netip.Addr `json:"ip_v6"`
	ASN6            *uint       `json:"as_v6"`
	IPv6Gateway     *netip.Addr `json:"ip_v6_gateway"`
	IPv6Netmask     *netip.Addr `json:"ip_v6_netmask"`
	FQDN            string      `json:"fqdn"`
	ProbeID         uint        `json:"probe"`
	CountryCode     string      `json:"country"`
	City            string      `json:"city"`
	Company         string      `json:"company"`
	IPv4Only        bool        `json:"is_ipv4_only"`
	Disabled        bool        `json:"is_disabled"`
	NicHandle       string      `json:"nic_handle"`
	Location        Geolocation `json:"geometry"`
	Type            string      `json:"type"`
	TLSARecord      string      `json:"tlsa_record"`
	LiveSince       *uniTime    `json:"date_live"`
	HardwareVersion uint        `json:"hardware_version"`
}

type AsyncAnchorResult struct {
	Anchor Anchor
	Error  error
}

// Translate the anchor version (code) into something more understandable
func (anchor *Anchor) decodeHardwareVersion() string {
	switch anchor.HardwareVersion {
	case 1:
		return "1"
	case 2:
		return "2"
	case 3:
		return "3"
	case 99:
		return "VM"
	default:
		return "?"
	}
}

// ShortString produces a short textual description of the anchor
func (anchor *Anchor) ShortString() string {
	text := fmt.Sprintf("%d\t%d\t%s\t%s\t%s",
		anchor.ID,
		anchor.ProbeID,
		anchor.CountryCode,
		anchor.City,
		anchor.FQDN,
	)

	text += valueOrNA("AS", false, anchor.ASN4)
	text += valueOrNA("AS", false, anchor.ASN6)
	text += fmt.Sprintf("\t%v", anchor.Location.Coordinates)

	return text
}

// LongString produces a longer textual description of the anchor
func (anchor *Anchor) LongString() string {
	text := anchor.ShortString()

	text += valueOrNA("", false, anchor.Address4)
	text += valueOrNA("", false, anchor.Address6)
	if anchor.NicHandle != "" {
		text += "\t" + anchor.NicHandle
	} else {
		text += "\tN/A"
	}

	text += fmt.Sprintf("\t\"%s\" %v %v %s",
		anchor.Company,
		anchor.IPv4Only,
		anchor.Disabled,
		anchor.decodeHardwareVersion(),
	)

	return text
}

// the API paginates; this describes one such page
type anchorListingPage struct {
	Count    uint     `json:"count"`
	Next     string   `json:"next"`
	Previous string   `json:"previous"`
	Anchors  []Anchor `json:"results"`
}

// AnchorFilter struct holds specified filters and other options
type AnchorFilter struct {
	params  url.Values
	id      uint
	limit   uint
	verbose bool
}

// NewAnchorFilter prepares a new anchor filter object
func NewAnchorFilter() AnchorFilter {
	filter := AnchorFilter{}
	filter.params = url.Values{}
	return filter
}

// Verboe sets verbosity
func (filter *AnchorFilter) Verbose(verbose bool) {
	filter.verbose = verbose
}

// FilterID filters by a particular anchor ID
func (filter *AnchorFilter) FilterID(id uint) {
	filter.id = id
}

// FilterCountry filters by a country code (ISO3166-1 alpha-2)
func (filter *AnchorFilter) FilterCountry(cc string) {
	filter.params.Add("country", cc)
}

// FilterSearch filters within the fields `city`, `fqdn` and `company`
func (filter *AnchorFilter) FilterSearch(text string) {
	filter.params.Add("search", text)
}

// FilterASN4 filters for an ASN in IPv4 space
func (filter *AnchorFilter) FilterASN4(as uint) {
	filter.params.Add("as_v4", fmt.Sprint(as))
}

// FilterASN6 filters for an ASN in IPv6 space
func (filter *AnchorFilter) FilterASN6(as uint) {
	filter.params.Add("as_v6", fmt.Sprint(as))
}

// Limit limits the number of result retrieved
func (filter *AnchorFilter) Limit(max uint) {
	filter.limit = max
}

// Verify sanity of applied filters
func (filter *AnchorFilter) verifyFilters() error {
	if filter.params.Has("country") {
		cc := filter.params.Get("country")
		// TODO: properly verify country code
		if len(cc) != 2 {
			return fmt.Errorf("invalid country code")
		}
	}

	return nil
}

// GetAnchorCount returns the count of anchors by filtering
func (filter *AnchorFilter) GetAnchorCount() (
	count uint,
	err error,
) {
	// sanity checks - late in the process, but not too late
	err = filter.verifyFilters()
	if err != nil {
		return
	}

	// counting needs application of the specified filters
        filter.params.Add("page_size", "0")
	query := apiBaseURL + "anchors/?" + filter.params.Encode()

	resp, err := apiGetRequest(filter.verbose, query, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// grab and store the actual content
	var page anchorListingPage
	err = json.NewDecoder(resp.Body).Decode(&page)
	if err != nil {
		return 0, err
	}

	// the only really important data point is the count
	return page.Count, nil
}

// GetAnchors returns a bunch of anchors by filtering
// Results (or an error) appear on a channel
func (filter *AnchorFilter) GetAnchors(
	anchors chan AsyncAnchorResult,
) {
	defer close(anchors)

	// special case: a specific ID was "filtered"
	if filter.id != 0 {
		anchor, err := GetAnchor(filter.verbose, filter.id)
		if err != nil {
			anchors <- AsyncAnchorResult{Anchor{}, err}
			return
		}
		anchors <- AsyncAnchorResult{*anchor, nil}
		return
	}

	// sanity checks - late in the process, but not too late
	err := filter.verifyFilters()
	if err != nil {
		anchors <- AsyncAnchorResult{Anchor{}, err}
		return
	}

	query := apiBaseURL + "anchors/?" + filter.params.Encode()

	resp, err := apiGetRequest(filter.verbose, query, nil)

	// results are paginated with next= (and previous=)
	var total uint = 0
	for {
		if err != nil {
			anchors <- AsyncAnchorResult{Anchor{}, err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			anchors <- AsyncAnchorResult{Anchor{}, err}
			return
		}

		// grab and store the actual content
		var page anchorListingPage
		err = json.NewDecoder(resp.Body).Decode(&page)
		if err != nil {
			anchors <- AsyncAnchorResult{Anchor{}, err}
		}

		// return items while observing the limit
		for _, anchor := range page.Anchors {
			anchors <- AsyncAnchorResult{anchor, nil}
			total++
			if total >= filter.limit {
				return
			}
		}

		// no next page => we're done
		if page.Next == "" {
			break
		}

		// just follow the next link
		resp, err = apiGetRequest(filter.verbose, page.Next, nil)
	}
}

// GetAnchor retrieves data for a single anchor, by ID
// returns anchor, _ if an anchor was found
// returns nil, _ if an anchor was not found
// returns _, err on error
func GetAnchor(
	verbose bool,
	id uint,
) (
	anchor *Anchor,
	err error,
) {
	query := fmt.Sprintf("%sanchors/%d/", apiBaseURL, id)

	resp, err := apiGetRequest(verbose, query, nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, parseAPIError(resp)
	}

	// grab and store the actual content
	err = json.NewDecoder(resp.Body).Decode(&anchor)
	if err != nil {
		return
	}

	return
}
