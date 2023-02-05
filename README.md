# goatcli - Go (RIPE) Atlas Tools - CLI

goatcli provides a CLI to interact with [RIPE Atlas](https://atlas.ripe.net/) [APIs](https://atlas.ripe.net/api/v2/) using [Golang](https://go.dev/). It uses the [goatapi](https://github.com/robert-kisteleki/goatapi/) under the hood. It is similar to [Magellan](https://github.com/RIPE-NCC/ripe-atlas-tools).

It supports:
* finding probes
* finding anchors
* finding measurements
* downloading and displaying results of measurements and turning them into Go objects
* tuning in to result streaming and turning them into Go objects
* loading a local file containing measurement results and turning them into Go objects
* various kinds of output formatters for displaying and aggregating measurement results
* (more features to come)

The tool needs Go 1.18 to compile.

# Context

[RIPE Atlas](https://atlas.ripe.net) is an open, community based active Internet
measurement network developed by the [RIPE NCC](https://www.ripe.net/) since 2010.
It provides a number of vantage points ("probes") run by volunteers, that allow
various kinds of network measurements (pings, traceroutes, DNS queries, ...) to
be run by any user.


# Quick Start

## Finding Probes

### Count Probes Matching Some Criteria

```sh
$ ./goatcli findprobe --sort -id -cc NL --status C --count
575
```

### Search for Probes

```sh
$ ./goatcli findprobe --sort -id -cc NL --status C --limit 5
1004258	Connected	"NL" N/A	AS44103 N/A	[4.8995 52.3815]
1004216	Connected	"NL" N/A	AS204995 N/A	[4.6675 52.3905]
1004194	Connected	"NL"	"atlas02.netjoe.nl"	AS206238	AS206238	[6.0915 52.5095]
1004164	Connected	"NL"	"Netrebel"	AS38919	AS38919	[5.1805 52.0295]
1004163	Connected	"NL"	"Docker probe"	AS1136	AS1136	[5.6405 52.0385]
```


### Get a Particular Probe

```sh
$ ./goatcli findprobe --id 10001
10001	Connected	"NL" N/A	AS206238	AS206238	[4.9275 52.3475]
```

## Finding Anchors

### Count Anchors Matching Some Criteria

```sh
$ ./goatcli findanchor --asn4 3320 --count
5
```

### Search for Anchors

```sh
$ ./goatcli findanchor --asn4 3320
2899	7040	AT	Vienna	at-vie-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[16.392002 48.193092]
3042	7075	DE	Berlin	de-ber-as3320-client.anchors.atlas.ripe.net	AS3320	AS3320	[13.422857 52.54575]
1963	6724	DE	Frankfurt am Main	de-fra-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[8.682127 50.110924]
2900	7065	US	New York City	us-nyc-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[-74.00833 40.717663]
2045	6748	US	New York	us-nyc-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[-74.00831 40.717728]
```


### Get a Particular Anchor

```sh
$ ./goatcli findanchor --id 2899
2899	7040	AT	Vienna	at-vie-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[16.392002 48.193092]
```

## Finding Measurements

### Count Measurements Matching Some Criteria

```sh
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --count
165
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --type ping --count
25
```

### Search for Measurements

```sh
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --type ping --limit 5
34680813	Ongoing	Periodic	IPv4	2021-12-30T14:23:04Z N/A	900	27	ping	k.root-servers.net
29493618	Ongoing	Periodic	IPv4	2021-04-06T17:13:13Z N/A	240	5	ping	ripe.net
21879480	Ongoing	Periodic	IPv4	2019-06-05T15:18:05Z N/A	240	50	ping	ripe.net
21879479	Ongoing	Periodic	IPv4	2019-06-05T15:17:35Z N/A	240	50	ping	ripe.net
19230504	Ongoing	Periodic	IPv4	2019-01-29T17:01:43Z N/A	60	50	ping	ping.ripe.net
```

### Get a Particular Measurement

```sh
$ ./goatcli findmsm --id 34680813
34680813	Ongoing	Periodic	IPv4	2021-12-30T14:23:04Z N/A	900	27	ping	k.root-servers.net
```

## Processing results

goatCLI can fetch results of exiting measurements (either from the data API, result streaming or from a local file). It's possible to choose what kind of output you want via output formatters:
* `some` and `most` echo some basic properties of the results
* `native` produces native-looking outputs (for now for ping and traceroute)
* `dnsstat` provides basic statistics of DNS results
* `id` and `idcsv` only output the ID of the results (`idcsv` does this in CSV format)

The API call variant supports setting the start time, end time, probe id(s), and a few more filters.

Getting data from the data API:

```sh
$ ./goatcli result --id 1001 --probe 10001 --start today --output most
```

Tuning in to result stream, displaying incoming results in real-time but also saving them to a local file:

```sh
$ ./goatcli result --id 1001 --stream --save output.jsonl
```

Loading data from a local file is simple as well (from a file with one result per line ("format=txt" a.k.a. JSONL a.k.a. NDJSON)):

```sh
$ cat some-results.jsonl | ./goatcli result
```

The output formatters are extensible, feel free to write your own -- and contribute that back to this repo! You only need to make a new package under `output` that implements four functions:
* `setup()` to initialise the output processor
* `start()` to prepare for processing results; may be called once per batch of results
* `process()` to deal with one incoming result
* `finish()` to finish processing, make a summary, etc.; may be called once per batch of results

New output processors need to be registered in `goatcli.go`. See `some.go` or `native.go` for examples.

# Future Additions / TODO

* schedule a new measurement, stop existing measurements
* modify participants of an existing measurement (add/remove probes)
* check credit balance, transfer credits, ...

# Copyright, Contributing

(C) 2022, 2023 [Robert Kisteleki](https://kistel.eu/) & [RIPE NCC](https://www.ripe.net)

Contribution is possible and encouraged via the [Github repo]("https://github.com/robert-kisteleki/goatcli/")

# License

See the LICENSE file
