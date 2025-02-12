/*
  (C) Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"net/netip"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/robert-kisteleki/goat"
	"github.com/robert-kisteleki/goat/cmd/goat/output"
)

// struct to receive/store command line args for measurement filtering
type findMsmFlags struct {
	filterID            uint
	filterIDin          string
	filterIDgt          uint
	filterIDgte         uint
	filterIDlt          uint
	filterIDlte         uint
	filterStartTimeGt   string
	filterStartTimeGte  string
	filterStartTimeLt   string
	filterStartTimeLte  string
	filterStopTimeGt    string
	filterStopTimeGte   string
	filterStopTimeLt    string
	filterStopTimeLte   string
	filterOneoff        bool
	filterPeriodic      bool
	filterInterval      uint
	filterIntervalGt    uint
	filterIntervalGte   uint
	filterIntervalLt    uint
	filterIntervalLte   uint
	filterStatus        string
	filterTags          string
	filterType          string
	filterTarget        string
	filterTargetIs      string
	filterTargetHas     string
	filterTargetStarts  string
	filterTargetEnds    string
	filterDescIs        string
	filterDescHas       string
	filterDescStarts    string
	filterDescEnds      string
	filterProbe         uint
	filterAddressFamily uint
	filterProtocol      string
	filterMy            bool

	output  string
	outopts multioption
	sort    string
	limit   uint
	count   bool
}

// Implementation of the "find measurement" subcommand. Parses command line flags
// and interacts with goatAPI to apply those filters+options to fetch results
func commandFindMsm(args []string) {
	flags := parseFindMsmArgs(args)
	filter, options := parseFindMsmFlags(flags)
	formatter := options["output"].(string)

	if !output.Verify(formatter, "msm") {
		fmt.Fprintf(os.Stderr, "ERROR: unknown or unsupported output format '%s' for measurment list\n", formatter)
		os.Exit(1)
	}

	// counting only
	if _, ok := options["count"]; ok {
		count, err := filter.GetMeasurementCount()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(count)
		return
	}

	// most of the work is done by goatAPI
	measurements := make(chan goat.AsyncMeasurementResult)
	go filter.GetMeasurements(measurements)

	// produce output; exact format depends on the "format" option
	output.Setup(formatter, flagVerbose, []string(flags.outopts))
	output.Start(formatter)
	for measurement := range measurements {
		if measurement.Error != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", measurement.Error)
			os.Exit(1)
		} else {
			output.Process(formatter, measurement)
		}
	}
	output.Finish(formatter)
}

// Process flags (filters & options), pass most of them on to goatAPI
// while doing sanity checks on values
func parseFindMsmFlags(flags *findMsmFlags) (
	filter goat.MeasurementFilter,
	options map[string]any,
) {
	options = make(map[string]any)
	filter = goat.NewMeasurementFilter()
	filter.Verbose(flagVerbose)

	// options

	options["output"] = flags.output

	if flags.sort != "" {
		if goat.ValidMeasurementListSortOrder(flags.sort) {
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
	filter.ApiKey(getApiKey("list_measurements"))

	// filters

	if flags.filterID != 0 {
		filter.FilterID(flags.filterID)
		// with ID filtering other filters are irrelevant - so we can make a shortcut here
		return
	}

	if flags.filterMy {
		filter.FilterMy()
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

	if flags.filterStartTimeGt != "" {
		time, err := parseTimeAlternatives(flags.filterStartTimeGt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse start time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStarttimeGt(time)
	}
	if flags.filterStartTimeGte != "" {
		time, err := parseTimeAlternatives(flags.filterStartTimeGte)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse start time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStarttimeGte(time)
	}
	if flags.filterStartTimeLt != "" {
		time, err := parseTimeAlternatives(flags.filterStartTimeLt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse start time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStarttimeLt(time)
	}
	if flags.filterStartTimeLte != "" {
		time, err := parseTimeAlternatives(flags.filterStartTimeLte)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse start time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStarttimeLte(time)
	}
	if flags.filterStopTimeGt != "" {
		time, err := parseTimeAlternatives(flags.filterStopTimeGt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse stop time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStoptimeGt(time)
	}
	if flags.filterStopTimeGte != "" {
		time, err := parseTimeAlternatives(flags.filterStopTimeGte)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse stop time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStoptimeGte(time)
	}
	if flags.filterStopTimeLt != "" {
		time, err := parseTimeAlternatives(flags.filterStopTimeLt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse start time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStoptimeLt(time)
	}
	if flags.filterStopTimeLte != "" {
		time, err := parseTimeAlternatives(flags.filterStopTimeLte)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse start time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStoptimeLte(time)
	}

	if flags.filterOneoff {
		filter.FilterOneoff(true)
	}
	if flags.filterPeriodic {
		filter.FilterOneoff(false)
	}

	if flags.filterInterval != 0 {
		filter.FilterInterval(flags.filterInterval)
	}
	if flags.filterIntervalGt != 0 {
		filter.FilterIntervalGt(flags.filterIntervalGt)
	}
	if flags.filterIntervalGte != 0 {
		filter.FilterIntervalGte(flags.filterIntervalGte)
	}
	if flags.filterIntervalLt != 0 {
		filter.FilterIntervalLt(flags.filterIntervalLt)
	}
	if flags.filterIntervalLte != 0 {
		filter.FilterIntervalLte(flags.filterIntervalLte)
	}

	if flags.filterStatus != "" {
		var fs []uint
		for _, item := range strings.Split(strings.ToLower(flags.filterStatus), ",") {
			switch item {
			case "any":
			case "spe":
				fs = append(fs, goat.MeasurementStatusSpecified)
			case "sch":
				fs = append(fs, goat.MeasurementStatusScheduled)
			case "ong":
				fs = append(fs, goat.MeasurementStatusOngoing)
			case "stp":
				fs = append(fs, goat.MeasurementStatusStopped)
			case "nos":
				fs = append(fs, goat.MeasurementStatusNoSuitableProbes)
			case "den":
				fs = append(fs, goat.MeasurementStatusDenied)
			case "fld":
				fs = append(fs, goat.MeasurementStatusFailed)
			case "fst":
				fs = append(fs, goat.MeasurementStatusForcedStop)
			default:
				fmt.Fprintf(os.Stderr, "ERROR: unknown status filter\n")
				os.Exit(1)
			}
		}
		if len(fs) > 0 {
			filter.FilterStatusIn(fs)
		}
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

	if flags.filterType != "" {
		if flags.filterType == "trace" {
			// allow "trace" a shorthand for "traceroute"
			flags.filterType = "traceroute"
		}
		if !goat.ValidMeasurementType(flags.filterType) {
			fmt.Fprintf(os.Stderr, "ERROR: invalid type: %s\n", flags.filterType)
			os.Exit(1)
		}
		filter.FilterType(flags.filterType)
	}

	if flags.filterTarget != "" {
		// try to parse as an exact IP first
		addr, err := netip.ParseAddr(flags.filterTarget)
		if err == nil {
			if addr.Is4() {
				filter.FilterTarget(netip.PrefixFrom(addr, 32))
			} else {
				filter.FilterTarget(netip.PrefixFrom(addr, 128))
			}
		} else {
			// not an address, see if it's a prefix
			prefix, err := netip.ParsePrefix(flags.filterTarget)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: invalid IP or prefix: %s (%v)\n", flags.filterTarget, err)
				os.Exit(1)
			}
			filter.FilterTarget(prefix)
		}
	}
	if flags.filterTargetIs != "" {
		filter.FilterTargetIs(flags.filterTargetIs)
	}
	if flags.filterTargetHas != "" {
		filter.FilterTargetHas(flags.filterTargetHas)
	}
	if flags.filterTargetStarts != "" {
		filter.FilterTargetStartsWith(flags.filterTargetStarts)
	}
	if flags.filterTargetEnds != "" {
		filter.FilterTargetEndsWith(flags.filterTargetEnds)
	}
	if flags.filterDescIs != "" {
		filter.FilterDescriptionIs(flags.filterDescIs)
	}
	if flags.filterDescHas != "" {
		filter.FilterDescriptionHas(flags.filterDescHas)
	}
	if flags.filterDescStarts != "" {
		filter.FilterDescriptionStartsWith(flags.filterDescStarts)
	}
	if flags.filterDescEnds != "" {
		filter.FilterDescriptionEndsWith(flags.filterDescEnds)
	}

	if flags.filterProbe != 0 {
		filter.FilterProbe(flags.filterProbe)
	}

	if flags.filterAddressFamily != 0 {
		if flags.filterAddressFamily != 4 && flags.filterAddressFamily != 6 {
			fmt.Fprintf(os.Stderr, "ERROR: invalid address family, it should be 4 or 6: %d\n", flags.filterAddressFamily)
			os.Exit(1)
		}
		filter.FilterAddressFamily(flags.filterAddressFamily)
	}

	if flags.filterProtocol != "" {
		if !slices.Contains([]string{"icmp", "udp", "tcp"}, flags.filterProtocol) {
			fmt.Fprintf(os.Stderr, "ERROR: invalid protocol \"%s\", it must be ICMP or UDP or TCP\n", flags.filterProtocol)
			os.Exit(1)
		}
		filter.FilterProtocol(flags.filterProtocol)
	}

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseFindMsmArgs(args []string) *findMsmFlags {
	var flags findMsmFlags
	// filters
	flagsFindMsm.UintVar(&flags.filterID, "id", 0, "A particular measurement ID to fetch. If present, all other filters are ignored")
	flagsFindMsm.UintVar(&flags.filterIDlt, "idlt", 0, "Filter on ID being less than this value")
	flagsFindMsm.UintVar(&flags.filterIDlte, "idlte", 0, "Filter on ID being less than or equal to this value")
	flagsFindMsm.UintVar(&flags.filterIDgt, "idgt", 0, "Filter on ID being greater than this value")
	flagsFindMsm.UintVar(&flags.filterIDgte, "idgte", 0, "Filter on ID being greater than or equal to this value")
	flagsFindMsm.StringVar(&flags.filterIDin, "idin", "", "Filter on ID being in this comma separated list")
	flagsFindMsm.StringVar(&flags.filterStartTimeGt, "startgt", "", "Filter on start time (epoch or ISO) after this")
	flagsFindMsm.StringVar(&flags.filterStartTimeGte, "startgte", "", "Filter on start time (epoch or ISO) at or after this")
	flagsFindMsm.StringVar(&flags.filterStartTimeLt, "startlt", "", "Filter on start time (epoch or ISO) before this")
	flagsFindMsm.StringVar(&flags.filterStartTimeLte, "startlte", "", "Filter on start time (epoch or ISO) at or before this")
	flagsFindMsm.StringVar(&flags.filterStopTimeGt, "stopgt", "", "Filter on stop time (epoch or ISO) after this")
	flagsFindMsm.StringVar(&flags.filterStopTimeGte, "stopgte", "", "Filter on stop time (epoch or ISO) at or after this")
	flagsFindMsm.StringVar(&flags.filterStopTimeLt, "stoplt", "", "Filter on stop time (epoch or ISO) before this")
	flagsFindMsm.StringVar(&flags.filterStopTimeLte, "stoplte", "", "Filter on stop time (epoch or ISO) at or before this")
	flagsFindMsm.BoolVar(&flags.filterOneoff, "oneoff", false, "Filter for one-off measurements")
	flagsFindMsm.BoolVar(&flags.filterPeriodic, "periodic", false, "Filter for periodic measurement")
	flagsFindMsm.UintVar(&flags.filterInterval, "interval", 0, "Filter on interval")
	flagsFindMsm.UintVar(&flags.filterIntervalGt, "intervalgt", 0, "Filter on interval being greater than this value")
	flagsFindMsm.UintVar(&flags.filterIntervalGte, "intervalgte", 0, "Filter on interval being greater than or equal to this value")
	flagsFindMsm.UintVar(&flags.filterIntervalLt, "intervallt", 0, "Filter on interval being less than this value")
	flagsFindMsm.UintVar(&flags.filterIntervalLte, "intervallte", 0, "Filter on interval being less than or equal to this value")
	flagsFindMsm.StringVar(&flags.filterStatus, "status", "any", "Status to filter for: any/spe/sch/ong/stp/fld/fst")
	flagsFindMsm.StringVar(&flags.filterType, "type", "", "Type to filter for: "+strings.Join(goat.MeasurementTypes, ","))
	flagsFindMsm.StringVar(&flags.filterTags, "tags", "", "Filter tags in this comma separated list")
	flagsFindMsm.StringVar(&flags.filterTarget, "target", "", "Filter for target (exact IP or prefix)")
	flagsFindMsm.StringVar(&flags.filterTargetIs, "targetis", "", "Filter for target name")
	flagsFindMsm.StringVar(&flags.filterTargetHas, "targethas", "", "Filter for target name substring")
	flagsFindMsm.StringVar(&flags.filterTargetStarts, "targetstarts", "", "Filter for target name prefix")
	flagsFindMsm.StringVar(&flags.filterTargetEnds, "targetends", "", "Filter for target name suffix")
	flagsFindMsm.StringVar(&flags.filterDescIs, "descis", "", "Filter for description")
	flagsFindMsm.StringVar(&flags.filterDescHas, "deschas", "", "Filter for description substring")
	flagsFindMsm.StringVar(&flags.filterDescStarts, "descstarts", "", "Filter for description prefix")
	flagsFindMsm.StringVar(&flags.filterDescEnds, "descends", "", "Filter for description suffix")
	flagsFindMsm.UintVar(&flags.filterProbe, "probe", 0, "Filter on ID of a participating probe")
	flagsFindMsm.UintVar(&flags.filterAddressFamily, "af", 0, "Filter on address family (4 or 6)")
	flagsFindMsm.StringVar(&flags.filterProtocol, "proto", "", "Filter for protocol (ICMP, UDP, TCP) where it makes sense (i.e. DNS and traceroute)")
	flagsFindMsm.BoolVar(&flags.filterMy, "my", false, "Filter for 'my' measurements. Requires an API key.")

	// options
	flagsFindMsm.BoolVar(&flags.count, "count", false, "Count only, don't show the actual results")
	flagsFindMsm.StringVar(&flags.sort, "sort", "-id", "Result ordering: "+strings.Join(goat.MeasurementListSortOrders, ","))
	flagsFindMsm.StringVar(&flags.output, "output", "some", "Output format: 'id', 'idcsv', 'some' or 'most'")
	flagsFindMsm.Var(&flags.outopts, "opt", "Options to pass to the output formatter")

	// limit
	flagsFindMsm.UintVar(&flags.limit, "limit", 100, "Maximum amount of measurements to retrieve")

	_ = flagsFindMsm.Parse(args)

	return &flags
}
