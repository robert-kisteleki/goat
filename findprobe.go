/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"goatcli/output"
	"net/netip"
	"os"
	"regexp"
	"strings"

	"github.com/robert-kisteleki/goatapi"
)

// struct to receive/store command line args for probe filtering
type findProbeFlags struct {
	filterID        uint
	filterIDin      string
	filterIDgt      uint
	filterIDgte     uint
	filterIDlt      uint
	filterIDlte     uint
	filterASN       uint
	filterASN4      uint
	filterASN4in    string
	filterASN6      uint
	filterASN6in    string
	filterCC        string
	filterCCin      string
	filterDist      float64
	filterLat       float64
	filterLatgt     float64
	filterLatgte    float64
	filterLatlt     float64
	filterLatlte    float64
	filterLon       float64
	filterLongt     float64
	filterLongte    float64
	filterLonlt     float64
	filterLonlte    float64
	filterPrefix4   string
	filterPrefix6   string
	filterStatus    string
	filterAnchor    bool
	filterNotAnchor bool
	filterPublic    bool
	filterNotPublic bool
	filterTags      string

	output  string
	outopts multioption
	sort    string
	limit   uint
	count   bool
}

// Implementation of the "find probe" subcommand. Parses command line flags
// and interacts with goatAPI to apply those filters+options to fetch results
func commandFindProbe(args []string) {
	flags := parseFindProbeArgs(args)
	filter, options := parseFindProbeFlags(flags)
	formatter := options["output"].(string)

	if !output.Verify(formatter) {
		fmt.Fprintf(os.Stderr, "ERROR: unknown output format '%s'\n", formatter)
		os.Exit(1)
	}

	// counting only
	if _, ok := options["count"]; ok {
		count, err := filter.GetProbeCount(flagVerbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(count)
		return
	}

	// most of the work is done by goatAPI
	probes := make(chan goatapi.AsyncProbeResult)
	go filter.GetProbes(flagVerbose, probes)

	// produce output; exact format depends on the "format" option
	output.Setup(formatter, flagVerbose, flags.outopts)
	output.Start(formatter)
	for probe := range probes {
		if probe.Error != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", probe.Error)
			os.Exit(1)
		} else {
			output.Process(formatter, probe)
		}
	}
	output.Finish(formatter)
}

// Process flags (filters & options), pass most of them on to goatAPI
// while doing sanity checks on values
func parseFindProbeFlags(flags *findProbeFlags) (
	filter *goatapi.ProbeFilter,
	options map[string]any,
) {
	options = make(map[string]any)
	filter = goatapi.NewProbeFilter()

	// options

	options["output"] = flags.output

	if flags.sort != "" {
		if goatapi.ValidProbeListSortOrder(flags.sort) {
			filter.Sort(flags.sort)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: invalid sort order\n")
			os.Exit(1)
		}
	}
	if flags.limit > 0 {
		filter.Limit(flags.limit)
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: limit should be positive\n")
		os.Exit(1)
	}
	if flags.count {
		options["count"] = true
	}

	// filters

	if flags.filterID != 0 {
		filter.FilterID(flags.filterID)
		// with ID filtering other filters are irrelevant - so we can make a shortcut here
		return
	}

	if flags.filterCC != "" {
		// TODO: properly verify country code
		if len(flags.filterCC) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: invalid country code: %s\n", flags.filterCC)
			os.Exit(1)
		}
		filter.FilterCountry(strings.ToUpper(flags.filterCC))
	}

	if flags.filterIDgt != 0 {
		filter.FilterIDGt(flags.filterIDgt)
	}
	if flags.filterIDgte != 0 {
		filter.FilterIDGte(flags.filterIDgte)
	}
	if flags.filterIDlt != 0 {
		filter.FilterIDLt(flags.filterIDlt)
	}
	if flags.filterIDlte != 0 {
		filter.FilterIDLte(flags.filterIDlte)
	}
	if flags.filterIDin != "" {
		list, err := makeIntList(flags.filterIDin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid ID in list: %s\n", flags.filterIDin)
			os.Exit(1)
		}
		filter.FilterIDin(list)
	}

	if flags.filterASN != 0 {
		filter.FilterASN(flags.filterASN)
	}
	if flags.filterASN4 != 0 {
		filter.FilterASN4(flags.filterASN4)
	}
	if flags.filterASN6 != 0 {
		filter.FilterASN6(flags.filterASN6)
	}

	if flags.filterASN4in != "" {
		list, err := makeIntList(flags.filterASN4in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid ID in list: %s\n", flags.filterASN4in)
			os.Exit(1)
		}
		filter.FilterASN4in(list)
	}
	if flags.filterASN6in != "" {
		list, err := makeIntList(flags.filterASN6in)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid ID in list: %s\n", flags.filterASN6in)
			os.Exit(1)
		}
		filter.FilterASN6in(list)
	}

	// TODO: these could be done nicer to allow 0.0 as value -- e.g. with Visit()
	if flags.filterLatgt != 0.0 {
		filter.FilterLatitudeGt(flags.filterLatgt)
	}
	if flags.filterLatgte != 0.0 {
		filter.FilterLatitudeGte(flags.filterLatgte)
	}
	if flags.filterLatlt != 0.0 {
		filter.FilterLatitudeLt(flags.filterLatlt)
	}
	if flags.filterLatlte != 0.0 {
		filter.FilterLatitudeLte(flags.filterLatlte)
	}
	if flags.filterLongt != 0.0 {
		filter.FilterLongitudeGt(flags.filterLongt)
	}
	if flags.filterLongte != 0.0 {
		filter.FilterLongitudeGte(flags.filterLongte)
	}
	if flags.filterLonlt != 0.0 {
		filter.FilterLongitudeLt(flags.filterLonlt)
	}
	if flags.filterLonlte != 0.0 {
		filter.FilterLongitudeLte(flags.filterLonlte)
	}
	if flags.filterPrefix4 != "" {
		prefix, err := netip.ParsePrefix(flags.filterPrefix4)
		if err != nil || !prefix.Addr().Is4() {
			fmt.Fprintf(os.Stderr, "ERROR: invalid IPv4 prefix: %s (%v)\n", flags.filterPrefix4, err)
			os.Exit(1)
		}
		filter.FilterPrefixV4(prefix)
	}
	if flags.filterPrefix6 != "" {
		prefix, err := netip.ParsePrefix(flags.filterPrefix4)
		if err != nil || !prefix.Addr().Is6() {
			fmt.Fprintf(os.Stderr, "ERROR: invalid IPv6 prefix: %s\n", flags.filterPrefix6)
			os.Exit(1)
		}
		filter.FilterPrefixV6(prefix)
	}

	switch strings.ToUpper(flags.filterStatus) {
	case "N":
		filter.FilterStatus(goatapi.ProbeStatusNeverConnected)
	case "C":
		filter.FilterStatus(goatapi.ProbeStatusConnected)
	case "D":
		filter.FilterStatus(goatapi.ProbeStatusDisconnected)
	case "A":
		filter.FilterStatus(goatapi.ProbeStatusAbandoned)
	default:
		fmt.Fprintf(os.Stderr, "ERROR: unknown status filter\n")
		os.Exit(1)
	}

	if flags.filterAnchor {
		filter.FilterAnchor(true)
	}
	if flags.filterNotAnchor {
		filter.FilterAnchor(false)
	}
	if flags.filterPublic {
		filter.FilterPublic(true)
	}
	if flags.filterNotPublic {
		filter.FilterPublic(false)
	}

	if flags.filterDist != 0.0 {
		if flags.filterLat == 0.0 && flags.filterLon == 0.0 {
			fmt.Fprintf(os.Stderr, "ERROR: when using a distance, lat+lon also have to be specified\n")
			os.Exit(1)
		}
		if flags.filterDist < 0.0 {
			fmt.Fprintf(os.Stderr, "ERROR: distance has to be positive\n")
			os.Exit(1)
		}
		filter.FilterRadius(flags.filterLat, flags.filterLon, flags.filterDist)
	}

	if flags.filterTags != "" {
		tags := strings.Split(flags.filterTags, ",")
		re, _ := regexp.Compile(`^[\w\-]+$`)
		for _, tag := range tags {
			if !re.MatchString(tag) {
				fmt.Fprintf(os.Stderr, "ERROR: invalid tag: %s (it should be a slug)\n", tag)
				os.Exit(1)
			}
		}
		filter.FilterTags(tags)
	}

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseFindProbeArgs(args []string) *findProbeFlags {
	var flags findProbeFlags

	// filters
	flagsFindProbe.UintVar(&flags.filterID, "id", 0, "A particular probe ID to fetch. If present, all other filters are ignored")
	flagsFindProbe.UintVar(&flags.filterIDlt, "idlt", 0, "Filter on ID being less than this value")
	flagsFindProbe.UintVar(&flags.filterIDlte, "idlte", 0, "Filter on ID being less than or equal to this value")
	flagsFindProbe.UintVar(&flags.filterIDgt, "idgt", 0, "Filter on ID being greater than this value")
	flagsFindProbe.UintVar(&flags.filterIDgte, "idgte", 0, "Filter on ID being greater than or equal to this value")
	flagsFindProbe.StringVar(&flags.filterIDin, "idin", "", "Filter on ID being in this comma separated list")
	flagsFindProbe.UintVar(&flags.filterASN, "asn", 0, "Filter for probes with an IPv4 or IPv6 address announced by ths AS")
	flagsFindProbe.UintVar(&flags.filterASN4, "asn4", 0, "Filter for probes with an IPv4 address announced by ths AS")
	flagsFindProbe.StringVar(&flags.filterASN4in, "asn4in", "", "Filter on ASN4 being in this comma separated list")
	flagsFindProbe.UintVar(&flags.filterASN6, "asn6", 0, "Filter for probes with an IPv6 address announced by ths AS")
	flagsFindProbe.StringVar(&flags.filterASN6in, "asn6in", "", "Filter on ASN6 being in this comma separated list")
	flagsFindProbe.StringVar(&flags.filterCC, "cc", "", "Filter for country code (2 letter ISO-3166 alpha-2)")
	flagsFindProbe.StringVar(&flags.filterCCin, "ccin", "", "Filter for country code (2 letter ISO-3166 alpha-2) in this comma separated list")
	flagsFindProbe.Float64Var(&flags.filterLat, "lat", 0.0, "Latitude for distance filtering")
	flagsFindProbe.Float64Var(&flags.filterLatgt, "latgt", 0.0, "Filter latitude being greater than ('north of') this value (in degrees)")
	flagsFindProbe.Float64Var(&flags.filterLatgte, "latgte", 0.0, "Filter latitude being greter than or equal to ('north of') this value (in degrees)")
	flagsFindProbe.Float64Var(&flags.filterLatlt, "latlt", 0.0, "Filter latitude being less than ('south of') this value (in degrees)")
	flagsFindProbe.Float64Var(&flags.filterLatlte, "latlte", 0.0, "Filter latitude being less than or equal to ('south of') this value (in degrees)")
	flagsFindProbe.Float64Var(&flags.filterLon, "lon", 0.0, "Longitude for distance filtering (in degrees)")
	flagsFindProbe.Float64Var(&flags.filterLongt, "longt", 0.0, "Filter longitude being greater than ('east of') this value (in degrees)")
	flagsFindProbe.Float64Var(&flags.filterLongte, "longte", 0.0, "Filter longitude being greter than or equal to ('east of') this value (in degrees)")
	flagsFindProbe.Float64Var(&flags.filterLonlt, "lonlt", 0.0, "Filter longitude being less than ('west of') this value (in degrees)")
	flagsFindProbe.Float64Var(&flags.filterLonlte, "lonlte", 0.0, "Filter longitude being less than or equal to ('west of') this value (in degrees)")
	flagsFindProbe.StringVar(&flags.filterPrefix4, "prefix4", "", "Filter for IPv4 prefix")
	flagsFindProbe.StringVar(&flags.filterPrefix6, "prefix6", "", "Filter for IPv6 prefix")
	flagsFindProbe.StringVar(&flags.filterStatus, "status", "C", "Status to filter for: '' any, 'N' never-connected, 'C' connected, 'D' disconnected, 'A' abandoned")
	flagsFindProbe.BoolVar(&flags.filterAnchor, "anchor", false, "Filter for anchors only")
	flagsFindProbe.BoolVar(&flags.filterNotAnchor, "notanchor", false, "Filter for non-anchors only")
	flagsFindProbe.BoolVar(&flags.filterPublic, "public", false, "Filter for public probes only")
	flagsFindProbe.BoolVar(&flags.filterNotPublic, "notpublic", false, "Filter for non-public probes only")
	flagsFindProbe.Float64Var(&flags.filterDist, "dist", 0.0, "Filter for a distance (in km) around lat+lon")
	flagsFindProbe.StringVar(&flags.filterTags, "tags", "", "Filter tags in this comma separated list")

	// options
	flagsFindProbe.BoolVar(&flags.count, "count", false, "Count only, don't show the actual results")
	flagsFindProbe.StringVar(&flags.sort, "sort", "-id", "Result ordering: "+strings.Join(goatapi.ProbeListSortOrders, ","))
	flagsFindProbe.StringVar(&flags.output, "output", "some", "Output format: 'id', 'idcsv', 'some' or 'most'")
	flagsFindProbe.Var(&flags.outopts, "opt", "Options to pass to the output formatter")

	// limit
	flagsFindProbe.UintVar(&flags.limit, "limit", 100, "Maximum amount of probes to retrieve")

	flagsFindProbe.Parse(args)

	return &flags
}
