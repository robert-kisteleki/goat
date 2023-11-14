/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"fmt"
	"reflect"
	"testing"
)

// Test if the ping parser does a decent job
func TestProbeParser(t *testing.T) {
	var ping PingResult
	err := ping.Parse(`
{
"fw":5040,
"mver":"2.4.1",
"lts":20,
"dst_name":"example.com",
"ttr":1.234,
"af":4,
"dst_addr":"10.1.2.3",
"src_addr":"10.2.3.4",
"proto":"ICMP",
"ttl":54,
"size":64,
"result":[
	{"rtt":10.000},
	{"rtt":15.000},
	{"rtt":4.750},
	{"rtt":5.250},
	{"rtt":25.000}
],
"dup":0,
"rcvd":4,
"sent":4,
"min":4.75,
"max":25.0,
"avg":12.0,
"msm_id":1234567,
"prb_id":2345678,
"timestamp":1655443320,
"msm_name":"Ping",
"from":"192.168.1.1",
"type":"ping",
"group_id":34567890,
"step":10,
"stored_timestamp":1655443322
}
`)
	if err != nil {
		t.Fatalf("Error parsing ping result: %s", err)
	}

	assertEqual(t, *ping.Step, uint(10), "error parsing ping field value for step")

	assertEqual(t, ping.Protocol, "ICMP", "error parsing ping field value for proto")
	assertEqual(t, ping.Ttl, uint(54), "error parsing ping field value for ttl")
	assertEqual(t, ping.PacketSize, uint(64), "error parsing ping field value for size")
	assertEqual(t, ping.Duplicates, uint(0), "error parsing ping field value for dup")
	assertEqual(t, ping.Received, uint(4), "error parsing ping field value for rcvd")
	assertEqual(t, ping.Sent, uint(4), "error parsing ping field value for sent")

	assertEqual(t, ping.Minimum, 4.75, "error parsing ping field value for min")
	assertEqual(t, ping.Average, 12.0, "error parsing ping field value for avg")
	assertEqual(t, ping.Maximum, 25.0, "error parsing ping field value for max")

	rtts := make([]float64, 0)
	for _, reply := range ping.Replies {
		rtts = append(rtts, reply.Rtt)
	}
	assertEqual(t, fmt.Sprint(rtts), "[10 15 4.75 5.25 25]", "error parsing RTTs")
	med := ping.Median
	medexp := 10.0
	assertEqual(t, med, medexp, fmt.Sprintf("median is incorrect, got %f, expected %f", med, medexp))

	err = ping.Parse(`
	{
	"fw":5040,
	"mver":"2.4.1",
	"lts":20,
	"dst_name":"example.com",
	"ttr":1.234,
	"af":4,
	"dst_addr":"10.1.2.3",
	"src_addr":"10.2.3.4",
	"proto":"ICMP",
	"ttl":54,
	"size":64,
	"result":[
		{"rtt":10.000},
		{"rtt":15.000},
		{"rtt":20.000}
	],
	"dup":0,
	"rcvd":3,
	"sent":3,
	"min":-1,
	"max":-1,
	"avg":-1,
	"msm_id":1234567,
	"prb_id":2345678,
	"timestamp":1655443320,
	"msm_name":"Ping",
	"from":"192.168.1.1",
	"type":"ping",
	"group_id":34567890,
	"step":10,
	"stored_timestamp":1655443322
	}
	`)
	if err != nil {
		t.Fatalf("Error parsing ping result: %s", err)
	}

	assertEqual(t, ping.Minimum, 10.0, "error calculating ping field value for min")
	assertEqual(t, ping.Average, 15.0, "error calculating ping field value for avg")
	assertEqual(t, ping.Maximum, 20.0, "error calculating ping field value for max")
	assertEqual(t, ping.Median, 15.0, "error calculating ping field value for median")

	// verify errors
	// verify timeouts
}

func TestMedian(t *testing.T) {
	var list []float64

	list = []float64{10, 40, 30, 20}
	med2 := median(list)
	assertEqual(t, med2, 25.0, "error in median calculation for even list")

	list = []float64{10, 40, 20}
	med1 := median(list)
	assertEqual(t, med1, 20.0, "error in median calculation for odd list")
}

// almost one-liner to reduce boiler plate
func assertEqual(t *testing.T, val1 interface{}, val2 interface{}, msg string) {
	if val1 == val2 {
		return
	}
	t.Errorf("%s: received %v (type %v), expected %v (type %v)", msg, val1, reflect.TypeOf(val1), val2, reflect.TypeOf(val2))
}
