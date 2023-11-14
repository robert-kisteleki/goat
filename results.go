/*
  (C) 2022, 2023 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package goat

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/robert-kisteleki/goat/result"

	"github.com/gorilla/websocket"
)

// ResultsFilter struct holds specified filters and other options
type ResultsFilter struct {
	params   url.Values
	id       uint   // which measurement
	file     string // which file to read from
	stream   bool   // use result streaming?
	limit    uint
	fetched  uint
	start    *time.Time
	stop     *time.Time
	probes   []uint
	latest   bool
	typehint string
	saveFile *os.File // save results to this file (if not nil)
	saveAll  bool
}

// NewResultsFilter prepares a new result filter object
func NewResultsFilter() ResultsFilter {
	filter := ResultsFilter{}
	filter.params = url.Values{}
	filter.params.Add("format", "txt")
	filter.probes = make([]uint, 0)
	return filter
}

// FilterID filters by a particular measurement ID
func (filter *ResultsFilter) FilterID(id uint) {
	filter.id = id
}

// FilterFile "filters" results from a particular file
func (filter *ResultsFilter) FilterFile(filename string) {
	filter.file = filename
}

// FilterStart filters for results after this timestamp
func (filter *ResultsFilter) FilterStart(t time.Time) {
	filter.start = &t
	filter.params.Add("start", fmt.Sprintf("%d", t.Unix()))
}

// FilterStop filters for results before this timestamp
func (filter *ResultsFilter) FilterStop(t time.Time) {
	filter.stop = &t
	filter.params.Add("stop", fmt.Sprintf("%d", t.Unix()))
}

// FilterProbeIDs filters for results where the probe ID is one of several in the list specified
func (filter *ResultsFilter) FilterProbeIDs(list []uint) {
	filter.probes = list
	filter.params.Add("probe_ids", makeCsv(list))
}

// FilterAnchors filters for results reported by anchors
func (filter *ResultsFilter) FilterAnchors() {
	filter.params.Add("anchors-only", "true")
}

// FilterPublicProbes filters for results reported by public probes
func (filter *ResultsFilter) FilterPublicProbes() {
	filter.params.Add("public-only", "true")
}

// FilterLatest "filters" for downloading the latest results only
func (filter *ResultsFilter) FilterLatest() {
	filter.latest = true
}

// Save the results to this particular file
func (filter *ResultsFilter) Save(file *os.File) {
	filter.saveFile = file
}

// SaveAll determines if all results are saved, or only the matched ones
func (filter *ResultsFilter) SaveAll(all bool) {
	filter.saveAll = all
}

// Stream switches between using the streaming or the data API
func (filter *ResultsFilter) Stream(useStream bool) {
	filter.stream = useStream
}

// Limit limits the number of result retrieved
func (filter *ResultsFilter) Limit(max uint) {
	filter.limit = max
}

// Verify sanity of applied filters
func (filter *ResultsFilter) verifyFilters() error {
	if filter.id == 0 && filter.file == "" {
		return fmt.Errorf("ID or filename must be specified")
	}

	return nil
}

// GetResult returns results via various means by filtering
// Results (or an error) appear on a channel
func (filter *ResultsFilter) GetResults(
	verbose bool,
	results chan result.AsyncResult,
) {
	switch {
	case filter.id != 0 && !filter.stream:
		filter.downloadResults(verbose, results)
	case filter.id != 0 && filter.stream:
		filter.streamResults(verbose, results)
	case filter.id == 0 && filter.stream:
		results <- result.AsyncResult{Result: nil, Error: fmt.Errorf("no ID was speficied for stream")}
		close(results)
	case filter.file != "":
		filter.getFileResults(verbose, results)
	default:
		results <- result.AsyncResult{Result: nil, Error: fmt.Errorf("neither ID nor input file were specified")}
		close(results)
	}
}

// DownloadResults returns results from the data API
// via a channel by applying the specified filters
func (filter *ResultsFilter) downloadResults(
	verbose bool,
	results chan result.AsyncResult,
) {
	defer close(results)

	// prepare to read results
	read, err := filter.openNetworkResults(verbose)
	if err != nil {
		results <- result.AsyncResult{Result: nil, Error: err}
		return
	}

	filter.readResults(verbose, read, results)
}

// StreamResults returns results from the streaming API
// via a channel by applying the specified filters
func (filter *ResultsFilter) streamResults(
	verbose bool,
	results chan result.AsyncResult,
) {
	// connect to the streaming API
	conn, _, err := websocket.DefaultDialer.Dial(streamBaseURL, nil)
	if err != nil {
		results <- result.AsyncResult{Result: nil, Error: err}
		close(results)
		return
	}

	// handle the resuts coming form the websocket
	go filter.streamReceiveHandler(verbose, conn, results)

	// using types and marshaling may be overkill - but it's flexible
	subscription := make([]any, 2)
	subscription[0] = "atlas_subscribe"
	type params struct {
		StreamType  string `json:"streamType"`
		Measurement uint   `json:"msm"`
	}
	subscription[1] = params{"result", filter.id}

	conn.WriteJSON(subscription)
}

// getFileResults returns results from a file via a channel
// If the file is "-" then it reads from stdin
func (filter *ResultsFilter) getFileResults(
	verbose bool,
	results chan result.AsyncResult,
) {
	defer close(results)

	var file *os.File
	if filter.file == "-" {
		file = os.Stdin
		if verbose {
			fmt.Printf("# Reading results from stdin\n")
		}
	} else {
		var err error
		file, err = os.Open(filter.file)
		if err != nil {
			results <- result.AsyncResult{Result: nil, Error: err}
			return
		}
		defer file.Close()

		if verbose {
			fmt.Printf("# Reading results from file: %s\n", filter.file)
		}
	}

	read := bufio.NewScanner(bufio.NewReader(file))

	filter.readResults(verbose, read, results)
}

func (filter *ResultsFilter) readResults(
	verbose bool,
	read *bufio.Scanner,
	results chan result.AsyncResult,
) {
	for read.Scan() && (filter.limit == 0 || filter.fetched < filter.limit) {
		line := read.Text()
		filter.processResult(line, verbose, results)
	}
}

func (filter *ResultsFilter) streamReceiveHandler(
	verbose bool,
	connection *websocket.Conn,
	results chan result.AsyncResult,
) {
	defer connection.Close()
	defer close(results)

	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			err := fmt.Errorf("error reading from stream: %v", err)
			results <- result.AsyncResult{Result: nil, Error: err}
			return
		}

		// instead of parsing the full message as JSON, we make a shortcut
		const expectedSubscribePrefix = "[\"atlas_subscribed\","
		const expectedResultPrefix = "[\"atlas_result\","

		switch {
		case expectedResultPrefix == string(msg[:len(expectedResultPrefix)]):
			// cool, a result
		case expectedSubscribePrefix == string(msg[:len(expectedSubscribePrefix)]):
			// cool, subscribe has been confirmed
			continue
		default:
			err := fmt.Errorf("unknown stream message received: %v", string(msg))
			results <- result.AsyncResult{Result: nil, Error: err}
			return
		}

		pduresult := strings.TrimPrefix(string(msg), expectedResultPrefix)
		pduresult = strings.TrimSuffix(pduresult, "]")

		filter.processResult(pduresult, verbose, results)

		if filter.limit > 0 && filter.fetched >= filter.limit {
			return
		}

	}
}

func (filter *ResultsFilter) processResult(
	resultString string,
	verbose bool,
	results chan result.AsyncResult,
) {
	saveResult := func() {
		if filter.saveFile != nil {
			_, err := filter.saveFile.WriteString(resultString + "\n")
			if err != nil {
				results <- result.AsyncResult{Result: nil, Error: err}
			}
			// continue regardless of whether writing was successful
		}
	}

	if filter.saveAll {
		saveResult()
	}

	res, err := result.ParseWithTypeHint(resultString, filter.typehint)
	if err != nil {
		results <- result.AsyncResult{Result: nil, Error: err}
		return
	}

	// check if time interval and probe constraints match (applicable if we're
	// reading from a file), and if so, put the result on the channel
	ts := time.Time(res.GetTimeStamp())
	if (filter.start == nil || filter.start.Before(ts.Add(time.Duration(1)))) &&
		(filter.stop == nil || filter.stop.After(ts.Add(time.Duration(-1)))) &&
		(len(filter.probes) == 0 || slices.Contains(filter.probes, res.GetProbeID())) {
		results <- result.AsyncResult{Result: &res, Error: nil}
		filter.fetched++

		if !filter.saveAll {
			saveResult()
		}
	}

	// a type hint makes parsing much faster
	if filter.typehint == "" {
		filter.typehint = res.TypeName()
	}
}

// prepare fetching results, i.e. verify parameters, connect to the API, etc.
func (filter *ResultsFilter) openNetworkResults(
	verbose bool,
) (
	read *bufio.Scanner,
	err error,
) {
	// sanity checks - late in the process, but not too late
	err = filter.verifyFilters()
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("%smeasurements/%d/", apiBaseURL, filter.id)
	if filter.latest {
		query += "latest/"
	} else {
		query += "results/"
	}
	query += fmt.Sprintf("?%s", filter.params.Encode())

	resp, err := apiGetRequest(verbose, query, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, parseAPIError(resp)
	}

	// we're reading one result per line, a scanner is simple enough
	return bufio.NewScanner(bufio.NewReader(resp.Body)), nil
}
