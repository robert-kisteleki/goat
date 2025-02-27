# goat changelog

## next

* CHANGED: fix issues reported by `golangci-lint`
* CHANGED: dependency updates
* CHANGED: minor speedup by setting `page_size` to `0` when counting stuff

## v0.7.3

* CHANGED: if the result download stop time is precisely on a day boundary, then it's adjusted
  to be 23:59:59 of the previous day instead
* NEW: `some` output formatter for connections, including a processed `table` format. Note: this needs sorting of the API output so it uses more memory than just streaming the results.
* FIX: don't try to parse abuf in DNS responses if it's not there
* FIX: only consider IN and CHAOS classes for `dnsstat` output
* ADD: add WrittenOff filter to finding probes
* CHANGED: dependency updates

## v0.7.2

* FIX: measurement spec uses "country" not "cc" as probe country code filter
* FIX: measurement spec probe ASN filter was incorrect when tags were not used

## v0.7.1

* FIX: (TLS) server cipher may be missing even if there is no alert
* CHANGED: added `-timeout` parameter to result streaming

## v0.7.0

* CHANGED: merged *goatAPI* and *goatCLI* into one package *goat* with goatCLI as a command (simply named *goat*)

## v0.6.0

CLI:
* FIX: native trace out didn't show hop errors nicely
* NEW: add shortcuts for measurement types, e.g. `goat ping ping.me` just works
* NEW: pre-flight checks for output format
* NEW: add `type:TYPE` option to the dnsstat output formatter, automatically add this for new DNS measurements
* CHANGED: update to go1.21 and update dependencies
* NEW: add a progress indicator to the dnsstat output formatter
* NEW: add DNS type to the native output formatter (for a `dig`-like output)

API:
* FIX: traceroute hop details can have 'error' instead of actual data
* CHANGED: firmware version 1 results are not supported
* FIX: improved handling of some old results which encode firmware version as string
* CHANGED: update to go 1.21 and module updates to newest
* NEW: added NSID parsing in DNS responses

## v0.5.0

CLI:
* CHANGED: renamed `--stop` to `--end` to indicate preferred end time of periodic measurements
* NEW: add support to stop existing measurements
* NEW: add support to add or remove probes to/from existing measurements (`--add` and `--remove`)

API:
* CHANGED: renamed `Start()` and `Stop()` to `StartTime()` and `EndTime()`
* CHANGED: renamed `Submit()` to `Schedule()`
* NEW: support for stopping measurements via `MeasurementSpec.Stop()`
* NEW: support for adding or removing probes to/from existing measurements via `MeasurementSpec.ParticipationRequest()`

## v0.4.1

CLI:
* NEW: measurement scheduling now immediately tunes in to the result stream to show results for one-offs
* CHANGED: renamed the `--ongoing` flag to `--periodic` for measurement scheduling
* CHANGED: various enhancements to output formats and accommodate goatapi result changes

API:
* FIX: handle dnserrors in TLS results
* FIX: traceroute numerical errors in hops were not handled properly
* CHANGED: fix typos
* CHANGED: improve stream EOF handling

## v0.4.0

CLI:
* NEW: add support for measurement status checks (`status` subcommand)
* NEW: add support for measurement scheduling (`measure` subcommand)
  * support all measurement types with lots of options and sane defaults
  * support all probe selection criteria
  * support default probe selection specification in config file

API:
* FIX: do not blow up if connection to stream fails
* FIX: defaulted to file read if stream was turned on but no measurement ID was set
* FIX: (ping parser) blew up if the generic Rtt field was missing
* FIX: (ping parser) min/avg/max were not reported correctly if they were not present but could otherwise be calculated
* FIX: typo in ErrorDetail JSON "detail"
* NEW: support for measurement status checks
* CHANGED: verbosity is now a setting, not a parameter
* CHANGED: ErrorDetail can have embedded error messages
* NEW: support for measurement scheduling with probe selection, timing, all measurement types and options

## v0.3.0

CLI:
* NEW: add option to stream results from the result stream instead of the data API
* NEW: add option to save all results obtained (from data API, stream API or even
  from a file) to a file
* NEW: added a dummy output formatter ("none")
* CHANGED: change default limit on downloading results to 0 (no limit)

API:
* NEW: add support to stream results from the result stream instead of the data API
* NEW: add support to save all results obtained (from data API, stream API or even
  from a file) to a file
* CHANGED: change default limit on downloading results to 0 (no limit)

## v0.2.2

CLI:
* NEW: output formatters can now accept options (`--opt X`)
* NEW: output annotator: output formatters can get probe ASN, CC and prefix
  information. Probe metadata is cached across runs, for a week.
* CHANGED: output formatters now have a `start()` method to signal processing
  of new (batch of) results; this maybe even multiple times - just like with
  `finish()` from now on. `setup()` is only called once, before any results are
  processed.
* NEW: `--opt ccstat` and `--opt asnstat` options for the `dnsstats` formatter

## v0.2.1

CLI:
* CHANGED: internal changes on how output formatters work
* CHANGED: output filters are now used for probes, anchors and measurements too
* CHANGED: adapt code to goatAPI v0.2.1 async results

API:
* CHANGED: better error handling for non-200 responses
* CHANGED: rdata is exposed in DNS results
* CHANGED: API GET calls are not handles in one function
* CHANGED: GetMeasurement() can also use an API key
* CHANGED: all responses (probes, acnhors, ...) are async/channel based now

## v0.2.0

CLI:
* NEW: support for downloading results of a measurement
  * with start time, stop time and probe filters
  * with option to get "latest" only
* NEW: support for processing results from an already downloaded file
* NEW: preliminary support for output processors
  * some, most: basic properties of the results
  * native: a native-looking output (i.e. similar to ping, traceroute, ...)
  * dnsstat: a simple DNS result summariser
* CHANGED: minor verbose output format changes
* CHANGED: output for "some" and "most" moved here from goatAPI

API:
* NEW: support for result processing (API call or local file)
  * including downloading "latest" results
  * all result types are supported (need a lot of work for testig / older resuls though)

## v0.1.0

CLI:
* support listing probes, anchors, measurements with virtually all filtering options
* support counting items, retrieveing all matching ones or just a specific one by ID
* support for a (primitive) configuration file (~/.config/goat.ini) and command line flags
* support for "list_measurements" API key via the config file

API:
* support listing probes, anchors, measurements with virtually all filtering options
* support counting items, retrieveing all matching ones or just a specific one by ID
* support for "list_measurements" API key
