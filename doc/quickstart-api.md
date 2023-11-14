# Go (RIPE) Atlas Tools - API Wrapper Quick Start Guide

## Finding Probes

### Count Probes Matching Some Criteria

```go
	filter := goat.NewProbeFilter()
	filter.FilterCountry("NL")
	filter.FilterPublic(true)
	count, err := filter.GetProbeCount(false)
	if err != nil {
		// handle the error
	}
	fmt.Println(count)
```

### Search for Probes

```go
	filter := goat.NewProbeFilter()
	filter.FilterCountry("NL")
	filter.Sort("-id")
	probes := make(chan goat.AsyncProbeResult)
	go filter.GetProbes(false, probes) // false means non-verbose
	if err != nil {
		// handle the error
	}
	for probe := range probes {
		if probe.Error != nil {
			// handle the error
		} else {
			// process the result
		}
	}
```

### Get a Particular Probe

```go
	probe, err := filter.GetProbe(false, 10001) // false means non-verbose
	if err != nil {
		// handle the error
	}
	fmt.Println(probe.ShortString())
```


## Finding Anchors

### Count Anchors Matching Some Criteria

```go
	filter := goat.NewAnchorFilter()
	filter.FilterCountry("NL")
	count, err := filter.GetAnchorCount(false)
	if err != nil {
		// handle the error
	}
	fmt.Println(count)
```

### Search for Anchors

```go
	filter := goat.NewAnchorFilter()
	filter.FilterCountry("NL")
	anchors := make(chan goat.AsyncAnchorResult)
	go filter.GetAnchors(false, anchors) // false means non-verbose
	if err != nil {
		// handle the error
	}
	for anchor := range anchors {
		if anchor.Error != nil {
			// handle the error
		} else {
			// process the result
		}
	}
```

### Get a Particular Anchor

```go
	anchor, err := filter.GetAnchor(false, 1080) // false means non-verbose
	if err != nil {
		// handle the error
	}
	fmt.Println(anchor.ShortString())
```

## Finding Measurements

### Count Measurements Matching Some Criteria

```go
	filter := goat.NewMeasurementFilter()
	filter.FilterTarget(netip.ParsePrefix("193.0.0.0/19"))
	filter.FilterType("ping")
	filter.FilterOneoff(true)
	count, err := filter.GetMeasurementCount(false)
	if err != nil {
		// handle the error
	}
	fmt.Println(count)
```

### Search for Measurements

```go
	filter := goat.NewMeasurementFilter()
	filter.FilterTarget(netip.ParsePrefix("193.0.0.0/19"))
	filter.FilterType("ping")
	filter.FilterOneoff(true)
	filter.Sort("-id")
	msms := make(chan goat.AsyncMeasurementResult)
	go filter.GetMeasurements(false, msms) // false means non-verbose
	if err != nil {
		// handle the error
	}
	for msm := range msms {
		if msm.Error != nil {
			// handle the error
		} else {
			// process the result
		}
	}
```

### Get a Particular Measurement

```go
	msm, err := filter.GetMeasurement(false, 1001)
	if err != nil {
		// handle the error
	}
	fmt.Println(msm.ShortString())
```

## Processing results

All result types are defined as object types (PingResult, TracerouteResult, DnsResult, ...). The Go types try to be more useful than what the API natively provides, i.e. there's a translation from what the API gives to objects that have more meaning and simpler to understand fields and methods.

Results can be fetched via a channel. The filter support measurement ID, start/end time, probe IDs, "latest" results and combinations of these.

**Note: this is beta level code. Some fields may be mis-interpreted, old result versions are may not be processed correctly, a lot of corner cases are likely not handled properly and some objects should be passed as pointers instead. There's possibly a lot of work to be done here.**

An example of retrieving and processing results from the data API:

```go
	filter := goat.NewResultsFilter()
	filter.FilterID(10001)
	filter.FilterLatest()

	results := make(chan result.AsyncResult)
	go filter.GetResults(false, results) // false means verbose mode is off

	for result := range results {
		// do something with a result
	}
```

An example of retrieving and processing results from result streaming:

```go
	filter := goat.NewResultsFilter()
	filter.FilterID(10001)
	filter.Stream(true)

	results := make(chan result.AsyncResult)
	go filter.GetResults(false, results) // false means verbose mode is off

	for result := range results {
		// do something with a result
	}
```

An example of retrieving and processing results from a file:

```go
	filter := goat.NewResultsFilter()
	filter.FilterFile("-") // stdin, one can also use a proper file name
	// note: other filters can be added (namely start, stop and probe)

	results := make(chan result.AsyncResult)
	go filter.GetResults(false, results) // false means verbose mode is off

	for result := range results {
		// do something with a result
	}
```

## Result types

The `result` package contains various types to hold corresponding measurement result types:
* `BaseResult` is the basis of all and contains the basic fields such as `MeasurementID`, `ProbeId`, `TimeStamp`, `Type` and such
* `PingResult`, `TracerouteResult`, `DnsResult` etc. contain the type-specific fields

## Measurement Scheduling

You can schedule measuements with virtually all available API options. A quick example:

```go
	spec := goat.NewMeasurementSpec()
	spec.ApiKey(myapikey)

	include := []string{"system-v4"}
	exclude := []string{"badtag", "anotherbad"}
	spec.AddProbesAreaWithTags("ww", 10, &include, &exclude)
	spec.AddProbesList([]uint{1, 99, 999})
	spec.AddProbesCountry("NL", 15)
	spec.AddProbesPrefix(netip.MustParsePrefix("192.0.0.0/8"), 5)

	spec.StartTime(tomorrownoon)
	spec.OneOff(true)

	spec.AddTrace(
		"my traceroute measurement",
		"ping.ripe.net",
		4,
		&goat.BaseOptions{ResolveOnProbe: true},
		&goat.TraceOptions{FirstHop: 4, ParisId: 9},
	)

	msmid, err := spec.Schedule()
	if err != nil {
		// use msmid
	}
```

### Basics

A new measuement object can be created with `NewMeasurementSpec()`. In order to successfully submit this to the API, you need to add an API key using `ApiKey()`. It also needs to contain at least one probe definition and at least one measurement definition.

### Probe Definitions

All variations of probe selection are supported:
* `AddProbesArea()` and `AddProbesAreaWithTags()` to add probes from an area (_WW_, _West_, ...)
* `AddProbesCountry()` and `AddProbesCountryWithTags()` to add probes from an country specified by its country code (ISO 3166-1 alpha-2)
* `AddProbesReuse()` and `AddProbesReuseWithTags()` to reuse probes from a previous measurement
* `AddProbesAsn()` and `AddProbesAsnWithTags()` to add probes from an ASN
* `AddProbesPrefix()` and `AddProbesPrefixWithTags()` to add probes from a prefix (IPv4 or IPv6)
* `AddProbesList()` and `AddProbesListWithTags()` to add probes with explicit probe IDs

Probe tags can be specified to include or exclude ones that have those specific tags.

### Time Definitions

You can specify whether you want a one-off or an ongoing measurement using `Oneoff()`.

Each measurement can have an explicit start time defined with `StartTime()`. Ongoing measurements can also have a predefined end time with `EndTime()`. These have to be sane regarding the current time (they need to be in the future) and to each other (end needs to happen after start). By default start time is as soon as possible with an undefined end time.

### Measurement Definitions

Various measurements can be added with `AddPing()`, `AddTrace()`, `AddDns()`, `AddTls()`, `AddNtp()` and `AddHttp()`. Multiple measurements can be added to one specification; in this case they will use the same probes and timing.

All measurement types support setting common options using a `BaseOptions{}` structure. You can set the measurement interval, spread, the resolve-on-probe flag and so on here. If you are ok with the API defaults then you can leave this parameter to be `nil`.

All measurement types also accept type-specific options via the structures `PingOptions{}`, `TraceOptions{}`, `DnsOptions{}` and so on. If you are ok with the API defaults then you can leave this parameter to be `nil` as well.

### Submitting a Measurement Specification to the API

The `Schedule()` function POSTs the whole specification to the API. It either returns with an `error` or a list of recently created measurement IDs. In case you're only interested in the API-compatible JSON structure without submitting it, then `GetApiJson()` should be called instead.

## Adding and Removing Probes

One can ask for more probes to be added to a measurement, or existing ones to be removed. While the API itself can do both in one call, goatAPI only supports either additions or removals in one query. In order to add or remove probes, the same `AddProbesX()` functions can be used to specify the probe set, then `ParticipationRequest(id, add)` is used with either `add=true` to add or `add=false` to remove probes. Note that for the remove function only an explicit probe list (`AddProbesList()`) can be used in the API.

In order to successfully submit this to the API, you need to add an API key beforehand using `ApiKey()`.

The return value of this function is a list of _participation request IDs_ or an error.

## Stopping a Measurement

You can stop a measuement via:

```go
	spec := goat.NewMeasurementSpec()
	spec.ApiKey(myapikey)

	err := spec.Stop(msmID)
	if err != nil {
		// done
	}
```
