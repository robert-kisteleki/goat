# goatcli - Go (RIPE) Atlas Tools - CLI

goatcli provides a CLI to interact with [RIPE Atlas](https://atlas.ripe.net/) [APIs](https://atlas.ripe.net/api/v2/) using [Golang](https://go.dev/). It uses the [goatapi](https://github.com/robert-kisteleki/goatapi/) under the hood. It is similar to [Magellan](https://github.com/RIPE-NCC/ripe-atlas-tools).

It supports:
* finding probes
* finding anchors
* finding measurements
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

```
$ ./goatcli findprobe --sort -id -cc NL --status C --count
575
```

### Search for Probes

```
$ ./goatcli findprobe --sort -id -cc NL --status C --limit 5
1004258	Connected	"NL" N/A	AS44103 N/A	[4.8995 52.3815]
1004216	Connected	"NL" N/A	AS204995 N/A	[4.6675 52.3905]
1004194	Connected	"NL"	"atlas02.netjoe.nl"	AS206238	AS206238	[6.0915 52.5095]
1004164	Connected	"NL"	"Netrebel"	AS38919	AS38919	[5.1805 52.0295]
1004163	Connected	"NL"	"Docker probe"	AS1136	AS1136	[5.6405 52.0385]
```


### Get a Particular Probe

```
$ ./goatcli findprobe --id 10001
10001	Connected	"NL" N/A	AS206238	AS206238	[4.9275 52.3475]
```

## Finding Anchors

### Count Anchors Matching Some Criteria

```
$ ./goatcli findanchor --asn4 3320 --count
5
```

### Search for Anchors

```
$ ./goatcli findanchor --asn4 3320
2899	7040	AT	Vienna	at-vie-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[16.392002 48.193092]
3042	7075	DE	Berlin	de-ber-as3320-client.anchors.atlas.ripe.net	AS3320	AS3320	[13.422857 52.54575]
1963	6724	DE	Frankfurt am Main	de-fra-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[8.682127 50.110924]
2900	7065	US	New York City	us-nyc-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[-74.00833 40.717663]
2045	6748	US	New York	us-nyc-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[-74.00831 40.717728]
```


### Get a Particular Anchor

```
$ ./goatcli findanchor --id 2899
2899	7040	AT	Vienna	at-vie-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[16.392002 48.193092]
```

## Finding Measurements

### Count Measurements Matching Some Criteria

```
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --count
165
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --type ping --count
25
```

### Search for Measurements

```
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --type ping --limit 5
34680813	Ongoing	Periodic	IPv4	2021-12-30T14:23:04Z N/A	900	27	ping	k.root-servers.net
29493618	Ongoing	Periodic	IPv4	2021-04-06T17:13:13Z N/A	240	5	ping	ripe.net
21879480	Ongoing	Periodic	IPv4	2019-06-05T15:18:05Z N/A	240	50	ping	ripe.net
21879479	Ongoing	Periodic	IPv4	2019-06-05T15:17:35Z N/A	240	50	ping	ripe.net
19230504	Ongoing	Periodic	IPv4	2019-01-29T17:01:43Z N/A	60	50	ping	ping.ripe.net
```

### Get a Particular Measurement

```
$ ./goatcli findmsm --id 34680813
34680813	Ongoing	Periodic	IPv4	2021-12-30T14:23:04Z N/A	900	27	ping	k.root-servers.net
```

# Future Additions / TODO

* schedule a new measurement, stop existing measurements
* modify participants of an existing measurement (add/remove probes)
* fetch results for, or listen to real-time result stream of, an already scheduled measurement
* check credit balance, transfer credits, ...


# Copyright, Contributing

(C) 2022, [Robert Kisteleki](https://kistel.eu/) & [RIPE NCC](https://www.ripe.net)

Contribution is possible and encouraged via the [Github repo]("https://github.com/robert-kisteleki/goatcli/")

# License

See the LICENSE file
