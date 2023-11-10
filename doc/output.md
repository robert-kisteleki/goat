# Output Formatters

Measurement results can come from a variety of sources:
* from downloaded historical results
* from downloaded 'latest' results
* real-time from the [result stream](https://atlas-stream.ripe.net/)

In each case, displaying the results can happen in different ways. These are called _output formatters_.
Each output formatter can work on one or more result type, and can have various options it can accept. They are also capable of providing aggregates, statistics or other summaries in addition to (or instead of) the individual results.

Below a description of the available output formatters and their options.

## none

The `none` formatter produces no output per result. In verbose mode it only produces a summary at the end about how many results it did not show.

## id and idcsv

The `id` and `idcsv` formatters only output the ID of the found probe / anchor / measurement in a one-per-line or a comma separated values format, respectively. They don't work on results.

## some

The `some` formatter displays basic fields for each object that needs to be displayed.

### For Measurement Results

The output has one line per result with the following basic fields for each result type:

* measurement ID
* probe ID
* result timestamp
* target name
* target IP

In addition, on the same line as the basic fields, the following type-specific fields are displayed:

* `ping`: sent/received/duplicate packets, min/avg/med/max RTTs (msec)
* `traceroute`: protocol used, number of hops
* `dns`: number of responses, number of errors
* `tls`: error, if observed OR alert, if observed OR method, protocol version, reply time (msec), number of certificates received
* `ntp`: reference ID, stratum, number of replies, number of errors
* `http`: URI

### For Metadata

The following fields are included for metadata:

* probes:
    * ID
    * connection status
    * country code
    * ASN for IPv4
    * ASN for IPv6
    * latitude and longitude
    * description

* anchors:
    * ID
    * probe ID
    * country code
    * city
    * FQDN
    * ASN for IPv4
    * ASN for IPv6
    * latitude and longitude

* measurements:
    * ID
    * status
    * one-off or ongoing
    * address family (IPv4 or IPv6)
    * start time
    * stop time
    * interval
    * probe participants count
    * type
    * target

* status checks:
    * global alert (true/false)
		* number of probes in alerting status
		* total number of probes involved

## most

The `most` formatter extends the output of the `some` formatter with more fields.

### For Measurement Results

The following fields are appended to the `some` output:

* `ping`:
    * country code of the probe
    * protocol used
    * all reply RTTs

* `traceroute`:
    * was the destination reached (true/false)
    * paris ID

* `dns`: the list of responses; for each response:
    * answers count
    * queries count
    * NS count
    * additional count
    * for each answer: class, type and data

* `tls`
    * server cipher used
    * for each certificate in the chain: serial and subject name 

* `ntp`
    * protocol
    * replies

* `http`
    * error (if any)
    * HTTP method
    * result code
    * header size
    * body size

### For Metadata

The following fields are appended to the `some` output:

* probes:
    * IPv4 address and prefix
    * IPv6 address and prefix
    * first and last connection time
    * total connected time (seconds)
    * IPv4 address and prefix
    * is this probe an anchor?
    * is this probe public?
    * list of tags

* anchors:
    * IPv4 and IPv6 addresses
    * NIC handle of the host
    * company name of the host
    * is the anchor IPv4 only?
    * is the anchor disabled?
    * hardware version

* measurements:
    * description
    * list of involved probes
    * tags

* status checks:
    * list of probes in alerting status

## native

The `native` formatter tries to display results close to how the native UNIX/Linux tools do. This applies to `ping` and `traceroute` results at the moment.

An example `ping` output:

```
PROBE 6856 PING 193.0.14.129 (193.0.14.129): 24 data bytes
32 bytes from 193.0.14.129: icmp_seq=0 ttl=60 time=0.804 ms
32 bytes from 193.0.14.129: icmp_seq=1 ttl=60 time=0.805 ms
32 bytes from 193.0.14.129: icmp_seq=2 ttl=60 time=0.793 ms
--- 193.0.14.129 ping statistics ---
3 packets transmitted, 3 packets received, 0.0% packet loss
round-trip min/avg/med/max = 0.793/0.801/0.804/0.805 ms
```

An example `traceroute` output:

```
PROBE 7066 traceroute to 193.0.14.129 (193.0.14.129): 255 hops max, 40 byte packets
  1  91.213.39.1 (91.213.39.1) 0.966 ms  0.965 ms  0.903 ms
  2  91.228.64.254 (91.228.64.254) 1.107 ms  1.047 ms  0.994 ms
  3  185.0.12.35 (185.0.12.35) 7.481 ms  7.297 ms  7.240 ms
  4  92.53.94.94 (92.53.94.94) 18.017 ms  17.721 ms  18.049 ms
  5  92.53.93.71 (92.53.93.71) 17.584 ms  17.731 ms  17.741 ms
  6  92.53.93.6 (92.53.93.6) 18.335 ms  18.069 ms  17.866 ms
  7  193.0.14.129 (193.0.14.129) 17.722 ms  17.736 ms  17.914 ms
```

## dnsstat

The `dnsstat` formatter consumes multiple DNS results and produces a summary of those results. For example:

```
$ ./goatcli result -stream -id 10001 -output dnsstat -limit 10
9	"[[IN SOA . '.	86400	IN	SOA	a.root-servers.net. nstld.verisign-grs.com. 2023102600 1800 900 604800 86400']]"
1	"REFUSED"
```

The above means that out of 10 results that were retrieved and processes, 9 were the same "correct" answer, and 1 was unsuccessful. In this basic form the aggregate only shows how many probes saw the different answers and it is mostly useful to find out if, from a population of probes, the answers are consistent or not.

This output formatter accepts three options:
* `ccstat` for aggregation per country of the probe
* `asnstat` for aggregation per ASN of the probe
* `type:TYPE1+TYPE2+...` to give type hints / focuses to aggregate on

One example of country code based aggregation:

```
$ ./goatcli result -stream -id 30001  -limit 10 -output dnsstat -opt ccstat
18	"NXDOMAIN"	 ID:3 PT:3 DK:2 NO:2 PL:2 UA:2 SE:1 NL:1 HU:1 IE:1
1	"TIMEOUT"	 IE:1
```

Similarly with ASN based aggregation:

```
$ ./goatcli result -stream -id 30001  -limit 10 -output dnsstat -opt asnstat
20	"NXDOMAIN"	 AS2860:3 AS47583:3 AS132420:3 AS7922:2 AS3209:2 AS57344:2 AS63473:1 AS(N/A):1 AS4771:1 AS3214:1
1	"TIMEOUT"	 AS63473:1
```

In order to make the aggregates, the formatter uses the annotation helper, which maintains a cache of basic probe metadata (in `~/.cache/goat/probes.db`).

The `type` hint can come handy if you want to "zoom in" on a particular answer type; other answers will be disregarded for the purposes of aggregation. For exampe if you're processing results and wan to check NS records only, use `-opt type:NS`.