/*
  (C) 2022, 2023 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

// status check result object, as it comes from the API
type StatusCheckResult struct {
	GloblaAlert bool                      `json:"global_alert"`
	TotalAlerts uint                      `json:"total_alerts"`
	Probes      map[uint]StatusCheckProbe `json:"probes"`
}

// status check result for one probe
type StatusCheckProbe struct {
	Alert          bool       `json:"alert"`
	Last           float64    `json:"last"`
	LastPacketLoss float64    `json:"last_packet_loss"`
	Source         string     `json:"source"`
	AllRTTs        *[]float64 `json:"all"`
}

// a status check result: error or value
type AsyncStatusCheckResult struct {
	Status *StatusCheckResult
	Error  error
}

// ShortString produces a short textual description of the status
func (sc *StatusCheckResult) ShortString() string {
	text := fmt.Sprintf("%v\t%d\t%d",
		sc.GloblaAlert,
		sc.TotalAlerts,
		len(sc.Probes),
	)
	return text
}

// LongString produces a longer textual description of the status
func (sc *StatusCheckResult) LongString() string {
	text := sc.ShortString()

	// add the list of alerting probes to the output
	alerted := make([]uint, 0)
	for probe, status := range sc.Probes {
		if status.Alert {
			alerted = append(alerted, probe)
		}
	}
	text += fmt.Sprintf("\t%v", alerted)
	return text
}

// StatusCheckFilter struct holds specified filters and other options
type StatusCheckFilter struct {
	params  url.Values
	id      uint
	showall bool
}

// NewStatusCheckFilter prepares a new status check filter object
func NewStatusCheckFilter() StatusCheckFilter {
	sc := StatusCheckFilter{}
	sc.params = url.Values{}
	return sc
}

// MsmID sets the measurement ID for which we ask the status check
func (filter *StatusCheckFilter) MsmID(id uint) {
	filter.id = id
}

// GetAllRTTs asks for all RTTs to be returned
func (filter *StatusCheckFilter) GetAllRTTs(showall bool) {
	filter.showall = showall
	if showall {
		filter.params.Add("show_all", fmt.Sprintf("%v", showall))
	}
}

// Verify sanity of applied filters
func (filter *StatusCheckFilter) verifyFilters() error {
	if filter.id == 0 {
		return fmt.Errorf("ID must be specified")
	}

	return nil
}

// StatusCheck returns a status check result
func (filter *StatusCheckFilter) StatusCheck(
	verbose bool,
	statuses chan AsyncStatusCheckResult,
) {
	defer close(statuses)

	var status StatusCheckResult

	// sanity checks - late in the process, but not too late
	err := filter.verifyFilters()
	if err != nil {
		statuses <- AsyncStatusCheckResult{&status, err}
		return
	}

	// make the request
	query := fmt.Sprintf("%smeasurements/%d/status-check?%s", apiBaseURL, filter.id, filter.params.Encode())
	resp, err := apiGetRequest(verbose, query, nil)
	if err != nil {
		statuses <- AsyncStatusCheckResult{&status, err}
		return
	}

	// read the response - it is a single JSON
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		statuses <- AsyncStatusCheckResult{&status, err}
		return
	}

	// check for error(s)
	if resp.StatusCode != 200 {
		var errors MultiErrorResponse
		err = json.Unmarshal(data, &errors)
		if err != nil {
			statuses <- AsyncStatusCheckResult{&status, err}
			return
		}
		statuses <- AsyncStatusCheckResult{&status, fmt.Errorf(errors.Error.Detail)}
		return
	}

	// parse the response into a status object
	err = json.Unmarshal(data, &status)
	if err != nil {
		statuses <- AsyncStatusCheckResult{&status, err}
		return
	}

	statuses <- AsyncStatusCheckResult{&status, nil}
}
