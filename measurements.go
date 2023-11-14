/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Measurement object, as it comes from the API
type Measurement struct {
	ID               uint               `json:"id"`
	CreationTime     uniTime            `json:"creation_time"`
	StartTime        uniTime            `json:"start_time"`
	StopTime         *uniTime           `json:"stop_time"`
	Status           MeasurementStatus  `json:"status"`
	GroupID          *uint              `json:"group_id"`
	ResolvedIPs      *[]netip.Addr      `json:"resolved_ips"`
	Description      *string            `json:"description"`
	Type             string             `json:"type"`
	Target           string             `json:"target"`
	TargetASN        *uint              `json:"target_asn"`
	TargetIP         netip.Addr         `json:"target_ip"`
	TargetPrefix     *netip.Prefix      `json:"target_prefix"`
	InWifiGroup      bool               `json:"in_wifi_group"`
	AddressFamily    *uint              `json:"af"`
	AllScheduled     bool               `json:"is_all_scheduled"`
	Interval         *uint              `json:"interval"`
	Spread           *uint              `json:"spread"`
	OneOff           bool               `json:"is_oneoff"`
	Public           bool               `json:"is_public"`
	ResolveOnProbe   bool               `json:"resolve_on_probe"`
	ParticipantCount *uint              `json:"participant_count"`
	ProbesRequested  *int               `json:"probes_requested"`
	ProbesScheduled  *uint              `json:"probes_scheduled"`
	CreditsPerResult uint               `json:"credits_per_result"`
	ResultsPerDay    uint               `json:"estimated_results_per_day"`
	Probes           []ParticipantProbe `json:"probes"`
	Tags             []string           `json:"tags"`
}

// ParticipantProbe - only the ID though
type ParticipantProbe struct {
	ID uint `json:"id"`
}

type AsyncMeasurementResult struct {
	Measurement Measurement
	Error       error
}

// MeasurementListSortOrders lists all the allowed sort orders
var MeasurementListSortOrders = []string{
	"id", "-id",
	"start_time", "-start_time",
	"stop_time", "-stop_time",
	"is_oneoff", "-is_oneoff",
	"interval", "-interval",
	"type", "-type",
	"af", "-af",
	"status.name", "-status.name",
	"status.id", "-status.id",
}

// MeasurementTypes lists all the allowed measurement types
var MeasurementTypes = []string{
	"ping", "traceroute", "dns", "http", "sslcert", "ntp",
}

// various measurement statuses
const (
	MeasurementStatusSpecified        = iota // 0
	MeasurementStatusScheduled               // 1
	MeasurementStatusOngoing                 // 2
	MeasurementStatusUnusedStatus            // 3
	MeasurementStatusStopped                 // 4
	MeasurementStatusForcedStop              // 5
	MeasurementStatusNoSuitableProbes        // 6
	MeasurementStatusFailed                  // 7
	MeasurementStatusDenied                  // 8
	MeasurementStatusCanceled                // 9
)

// MeasurementStatusDict maps the measurement status codes to human readable descriptions
var MeasurementStatusDict = map[uint]string{
	MeasurementStatusSpecified:        "Specified",
	MeasurementStatusScheduled:        "Scheduled",
	MeasurementStatusOngoing:          "Ongoing",
	MeasurementStatusStopped:          "Stopped",
	MeasurementStatusForcedStop:       "ForcedStop",
	MeasurementStatusNoSuitableProbes: "NoSuitableProbes",
	MeasurementStatusFailed:           "Failed",
	MeasurementStatusDenied:           "Denied",
	MeasurementStatusCanceled:         "Canceled",
}

// MeasurementStatusDict maps the measurement variations to human readable descriptions
var MeasurementOneoffDict = map[bool]string{
	true:  "Oneoff",
	false: "Periodic",
}

// MeasurementStatus as defined by the API
type MeasurementStatus struct {
	ID    uint     `json:"id"`
	Name  string   `json:"name"`
	Since *uniTime `json:"when"`
}

// ShortString produces a short textual description of the measurement
func (measurement *Measurement) ShortString() string {
	text := fmt.Sprintf("%d\t%s\t%s",
		measurement.ID,
		MeasurementStatusDict[measurement.Status.ID],
		MeasurementOneoffDict[measurement.OneOff],
	)
	text += valueOrNA("IPv", false, measurement.AddressFamily)

	text += valueOrNA("", false, &measurement.StartTime)
	text += valueOrNA("", false, measurement.StopTime)

	text += valueOrNA("", false, measurement.Interval)
	text += valueOrNA("", false, measurement.ParticipantCount)

	text += fmt.Sprintf("\t%s", measurement.Type)
	if measurement.Target != "" {
		text += fmt.Sprintf("\t%s", measurement.Target)
	} else {
		text += "\tN/A"
	}

	return text
}

// LongString produces a longer textual description of the measurement
func (measurement *Measurement) LongString() string {
	text := measurement.ShortString()

	text += valueOrNA("", true, measurement.Description)

	var idlist []uint
	for _, probe := range measurement.Probes {
		idlist = append(idlist, probe.ID)
	}
	text += fmt.Sprintf("\t%v", idlist)

	text += fmt.Sprintf("\t%v", measurement.Tags)

	return text
}

// the API paginates; this describes one such page
type measurementListingPage struct {
	Count        uint          `json:"count"`
	Next         string        `json:"next"`
	Previous     string        `json:"previous"`
	Measurements []Measurement `json:"results"`
}

// MeasurementFilter struct holds specified filters and other options
type MeasurementFilter struct {
	params  url.Values
	id      uint
	limit   uint
	verbose bool
	key     *uuid.UUID
	my      bool
}

// NewMeasurementFilter prepares a new measurement filter object
func NewMeasurementFilter() MeasurementFilter {
	filter := MeasurementFilter{}
	filter.params = url.Values{}
	filter.params.Add("optional_fields", "probes")
	filter.params.Add("format[datetime]", "iso-8601")
	return filter
}

// Verbose sets verbosity
func (filter *MeasurementFilter) Verbose(verbose bool) {
	filter.verbose = verbose
}

// FilterID filters by a particular measurement ID
func (filter *MeasurementFilter) FilterID(id uint) {
	filter.id = id
}

// FilterMy filters to my own measurements - and needs an API key todo so
func (filter *MeasurementFilter) FilterMy() {
	filter.my = true
}

// FilterIDGt filters for msm IDs > some number
func (filter *MeasurementFilter) FilterIDGt(n uint) {
	filter.params.Add("id__gt", fmt.Sprint(n))
}

// FilterIDGte filters for msm IDs >= some number
func (filter *MeasurementFilter) FilterIDGte(n uint) {
	filter.params.Add("id__gte", fmt.Sprint(n))
}

// FilterIDLt filters for msm IDs < some number
func (filter *MeasurementFilter) FilterIDLt(n uint) {
	filter.params.Add("id__lt", fmt.Sprint(n))
}

// FilterIDLt filters for msm IDs <= some number
func (filter *MeasurementFilter) FilterIDLte(n uint) {
	filter.params.Add("id__lte", fmt.Sprint(n))
}

// FilterIDin filters for measurement ID being one of several in the list specified
func (filter *MeasurementFilter) FilterIDin(list []uint) {
	filter.params.Add("id__in", makeCsv(list))
}

// FilterInterval filters for measurement interval being a specific number (seconds)
func (filter *MeasurementFilter) FilterInterval(n uint) {
	filter.params.Add("interval", fmt.Sprint(n))
}

// FilterIntervalGt filters for measurement interval being > a specific number (seconds)
func (filter *MeasurementFilter) FilterIntervalGt(n uint) {
	filter.params.Add("interval__gt", fmt.Sprint(n))
}

// FilterIntervalGte filters for measurement interval being >= a specific number (seconds)
func (filter *MeasurementFilter) FilterIntervalGte(n uint) {
	filter.params.Add("interval__gte", fmt.Sprint(n))
}

// FilterIntervalLt filters for measurement interval being < a specific number (seconds)
func (filter *MeasurementFilter) FilterIntervalLt(n uint) {
	filter.params.Add("interval__lt", fmt.Sprint(n))
}

// FilterIntervalLte filters for measurement interval being <= a specific number (seconds)
func (filter *MeasurementFilter) FilterIntervalLte(n uint) {
	filter.params.Add("interval__lte", fmt.Sprint(n))
}

// FilterStarttimeGt filters for measurement that start/started after a specific time
func (filter *MeasurementFilter) FilterStarttimeGt(t time.Time) {
	filter.params.Add("start_time__gt", fmt.Sprintf("%d", t.Unix()))
}

// FilterStarttimeGte filters for measurements that start/started after or at a specific time
func (filter *MeasurementFilter) FilterStarttimeGte(t time.Time) {
	filter.params.Add("start_time__gte", fmt.Sprintf("%d", t.Unix()))
}

// FilterStarttimeLt filters for measurements that start/started before a specific time
func (filter *MeasurementFilter) FilterStarttimeLt(t time.Time) {
	filter.params.Add("start_time__lt", fmt.Sprintf("%d", t.Unix()))
}

// FilterStarttimeLte filters for measurements that start/started before or at a specific time
func (filter *MeasurementFilter) FilterStarttimeLte(t time.Time) {
	filter.params.Add("start_time__lte", fmt.Sprintf("%d", t.Unix()))
}

// FilterStoptimeGt filters for measurements that stop/stopped after a specific time
func (filter *MeasurementFilter) FilterStoptimeGt(t time.Time) {
	filter.params.Add("stop_time__gt", fmt.Sprintf("%d", t.Unix()))
}

// FilterStoptimeGte filters for measurements that stop/stopped after or at a specific time
func (filter *MeasurementFilter) FilterStoptimeGte(t time.Time) {
	filter.params.Add("stop_time__gte", fmt.Sprintf("%d", t.Unix()))
}

// FilterStoptimeLt filters for measurements that stop/stopped before a specific time
func (filter *MeasurementFilter) FilterStoptimeLt(t time.Time) {
	filter.params.Add("stop_time__lt", fmt.Sprintf("%d", t.Unix()))
}

// FilterStoptimeLte filters for measurements that stop/stopped before or at a specific time
func (filter *MeasurementFilter) FilterStoptimeLte(t time.Time) {
	filter.params.Add("stop_time__lte", fmt.Sprintf("%d", t.Unix()))
}

// FilterStatus filters for measurements that have a specific status
// See: const MeasurementStatusSpecified*
func (filter *MeasurementFilter) FilterStatus(n uint) {
	filter.params.Add("status", fmt.Sprint(n))
}

// FilterStatusIn filters for measurements that are in a set of statuses
// See: const MeasurementStatusSpecified*
func (filter *MeasurementFilter) FilterStatusIn(list []uint) {
	filter.params.Add("status__in", makeCsv(list))
}

// FilterOneoff filters for one-off or ongoing measurements
// If oneoff is true then only one-offs are returned
// If oneoff is false then only ongoings are returned
func (filter *MeasurementFilter) FilterOneoff(oneoff bool) {
	filter.params.Add("is_oneoff", fmt.Sprint(oneoff))
}

// FilterTags filters for measurements with the specified tags
// Speficying multiple tags results in an AND behaviour
func (filter *MeasurementFilter) FilterTags(list []string) {
	filter.params.Add("tags", strings.Join(list, ","))
}

// FilterTarget filters for measurements that target a specific IP (IPv4 /32 or IPv6 /128)
// or are withing a particular IPV4 or IPv6 prefix ("more specific" search)
func (filter *MeasurementFilter) FilterTarget(target netip.Prefix) {
	filter.params.Add("target_ip", fmt.Sprint(target))
}

// FilterTargetIs filters for measurements that have a specific target as
// a string (DNS name or IP address)
func (filter *MeasurementFilter) FilterTargetIs(what string) {
	filter.params.Add("target", what)
}

// FilterTargetHas filters for measurements that have a substring in their target
func (filter *MeasurementFilter) FilterTargetHas(what string) {
	filter.params.Add("target__contains", what)
}

// FilterTargetStartsWith filters for measurements that start with this string in their target
func (filter *MeasurementFilter) FilterTargetStartsWith(what string) {
	filter.params.Add("target__startswith", what)
}

// FilterTargetEndsWith filters for measurements that end with this string in their target
func (filter *MeasurementFilter) FilterTargetEndsWith(what string) {
	filter.params.Add("target__endswith", what)
}

// FilterType filters for measurements that have a specific type
// See also: MeasurementTypes
func (filter *MeasurementFilter) FilterType(typ string) {
	filter.params.Add("type", typ)
}

// FilterProbe filters for measurements that have a particular probe participating in them
func (filter *MeasurementFilter) FilterProbe(id uint) {
	filter.params.Add("current_probes", fmt.Sprint(id))
}

// FilterDescriptionIs filters for measurements that have a specific description
func (filter *MeasurementFilter) FilterDescriptionIs(what string) {
	filter.params.Add("description", what)
}

// FilterDescriptionHas filters for measurements that have a specific string in
// their description
func (filter *MeasurementFilter) FilterDescriptionHas(what string) {
	filter.params.Add("description__contains", what)
}

// FilterDescriptionStartsWith filters for measurements that start with this string in their description
func (filter *MeasurementFilter) FilterDescriptionStartsWith(what string) {
	filter.params.Add("description__startswith", what)
}

// FilterDescriptionEndsWith filters for measurements that end with this string in their description
func (filter *MeasurementFilter) FilterDescriptionEndsWith(what string) {
	filter.params.Add("description__endswith", what)
}

// FilterAddressFamily filters for measurements using a particular address family
// 4 for IPv4, 6 for IPv6
func (filter *MeasurementFilter) FilterAddressFamily(af uint) {
	filter.params.Add("af", fmt.Sprint(af))
}

// FilterProtocol filters for measurements using a particular protocol
// Protocol can be ICMP, UDP or TCP for traceroutes, UDP or TCP for DNS
func (filter *MeasurementFilter) FilterProtocol(protocol string) {
	filter.params.Add("protocol", protocol)
}

// Sort asks the result list to be sorted by some ordering
// See also: MeasurementListSortOrders
func (filter *MeasurementFilter) Sort(by string) {
	filter.params.Add("sort", by)
}

// Limit limits the number of result retrieved
func (filter *MeasurementFilter) Limit(limit uint) {
	filter.limit = limit
}

// ApiKey sets the API key to be used
// This key should have the "list_measurements" permission
func (filter *MeasurementFilter) ApiKey(key *uuid.UUID) {
	filter.key = key
}

// Verify sanity of applied filters
func (filter *MeasurementFilter) verifyFilters() error {

	// there is a finite set of possible orderings
	if filter.params.Has("sort") {
		sort := filter.params.Get("sort")
		if !ValidMeasurementListSortOrder(sort) {
			return fmt.Errorf("invalid sort order")
		}
	}

	// tags need to be slugs
	if filter.params.Has("tags") {
		tags := strings.Split(filter.params.Get("tags"), ",")
		re, _ := regexp.Compile(`^[\w\-]+$`)
		for _, tag := range tags {
			if !re.MatchString(tag) {
				return fmt.Errorf("invalid tag: %s", tag)
			}
		}
	}

	// af needs to be 4 or 6
	if filter.params.Has("af") {
		af := filter.params.Get("af")
		if af != "4" && af != "6" {
			return fmt.Errorf("invalid address family: %s", af)
		}
	}

	// protocol is ICMP, UDP or TCP
	if filter.params.Has("protocol") {
		protocol := strings.ToLower(filter.params.Get("protocol"))
		if !slices.Contains([]string{"icmp", "udp", "tcp"}, protocol) {
			return fmt.Errorf("invalid protocol")
		}
	}

	// 'my' needs and API key
	if filter.my && filter.key == nil {
		return fmt.Errorf("'my' needs an API key")
	}

	return nil
}

// GetMeasurementCount returns the count of measurements by filtering
func (filter *MeasurementFilter) GetMeasurementCount() (
	count uint,
	err error,
) {
	// sanity checks - late in the process, but not too late
	err = filter.verifyFilters()
	if err != nil {
		return
	}

	// counting needs application of the specified filters
	query := apiBaseURL + "measurements/"
	if filter.my {
		query += "my/"
	}
	query += "?" + filter.params.Encode()

	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", uaString)
	if filter.key != nil {
		req.Header.Set("Authorization", "Key "+filter.key.String())
	}

	// results are paginated with next= (and previous=)
	if filter.verbose {
		msg := fmt.Sprintf("# API call: GET %s", req.URL)
		if filter.key != nil {
			msg += fmt.Sprintf(" (using API key %s...)", filter.key.String()[:8])
		}
		fmt.Println(msg)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// grab and store the actual content
	var page measurementListingPage
	err = json.NewDecoder(resp.Body).Decode(&page)
	if err != nil {
		return 0, err
	}

	// the only really important data point is the count
	return page.Count, nil
}

// GetMeasurements returns a bunch of measurements by filtering
// Results (or an error) appear on a channel
func (filter *MeasurementFilter) GetMeasurements(
	measurements chan AsyncMeasurementResult,
) {
	defer close(measurements)

	// special case: a specific ID was "filtered"
	if filter.id != 0 {
		msm, err := GetMeasurement(filter.verbose, filter.id, filter.key)
		if err != nil {
			measurements <- AsyncMeasurementResult{Measurement{}, err}
			return
		}
		measurements <- AsyncMeasurementResult{*msm, nil}
		return
	}

	// sanity checks - late in the process, but not too late
	err := filter.verifyFilters()
	if err != nil {
		measurements <- AsyncMeasurementResult{Measurement{}, err}
		return
	}

	query := apiBaseURL + "measurements/"
	if filter.my {
		query += "my/"
	}
	query += "?" + filter.params.Encode()

	resp, err := apiGetRequest(filter.verbose, query, filter.key)

	var total uint = 0
	// results are paginated with next= (and previous=)
	for {
		if err != nil {
			measurements <- AsyncMeasurementResult{Measurement{}, err}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			measurements <- AsyncMeasurementResult{Measurement{}, parseAPIError(resp)}
			return
		}

		// grab and store the actual content
		var page measurementListingPage
		err = json.NewDecoder(resp.Body).Decode(&page)
		if err != nil {
			measurements <- AsyncMeasurementResult{Measurement{}, err}
			return
		}

		// return items while observing the limit
		for _, msm := range page.Measurements {
			measurements <- AsyncMeasurementResult{msm, nil}
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
		resp, err = apiGetRequest(filter.verbose, page.Next, filter.key)
	}
}

// GetMeasurement retrieves data for a single measurement, by ID
// returns measurement, nil if a measurement was found
// returns nil, nil if no such measurement was found
// returns nil, err on error
func GetMeasurement(
	verbose bool,
	id uint,
	key *uuid.UUID,
) (
	*Measurement,
	error,
) {
	var measurement *Measurement

	query := fmt.Sprintf("%smeasurements/%d/", apiBaseURL, id)

	resp, err := apiGetRequest(verbose, query, key)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, parseAPIError(resp)
	}

	// grab and store the actual content
	err = json.NewDecoder(resp.Body).Decode(&measurement)
	if err != nil {
		return nil, err
	}

	return measurement, nil
}

// ValidMeasurementListSortOrder checks if a sort order is supported
func ValidMeasurementListSortOrder(sort string) bool {
	return slices.Contains(MeasurementListSortOrders, sort)
}

// ValidMeasurementType checks if a measurment type is recognised
func ValidMeasurementType(typ string) bool {
	return slices.Contains(MeasurementTypes, typ)
}
