/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"encoding/json"
	"fmt"
	"math"
	"net/netip"
	"sort"
)

type PingResult struct {
	BaseResult
	Sent, Received, Duplicates        uint        //
	Minimum, Average, Median, Maximum float64     // -1 if N/A
	PacketSize                        uint        //
	Protocol                          string      //
	Step                              *uint       //
	Ttl                               uint        //
	Replies                           []PingReply //
	Errors                            []string    //
	Timeouts                          uint        //
}

// one successful ping reply
type PingReply struct {
	Rtt       float64
	Source    netip.Addr
	Ttl       uint
	Duplicate bool
}

func (ping *PingResult) Parse(from string) (err error) {
	var iping pingResult
	err = json.Unmarshal([]byte(from), &iping)
	if err != nil {
		return err
	}
	if iping.Type != "ping" {
		return fmt.Errorf("this is not a ping result (type=%s)", iping.Type)
	}
	ping.BaseResult = iping.BaseResult
	ping.Replies = iping.Replies()
	ping.Errors = iping.Errors()
	ping.Timeouts = uint(iping.Timeouts())
	ping.Minimum = iping.Minimum
	ping.Average = iping.Average
	ping.Maximum = iping.Maximum
	ping.Sent = iping.Sent
	ping.Received = iping.Received
	ping.Duplicates = iping.Duplicates
	ping.PacketSize = iping.PacketSize
	ping.Protocol = iping.Protocol
	ping.Step = iping.Step
	if iping.Ttl != nil {
		ping.Ttl = *iping.Ttl
	}
	if len(ping.Replies) == 0 {
		ping.Median = -1
	} else {
		ping.Median = median(ping.ReplyRtts())
	}
	return nil
}

func (result *PingResult) TypeName() string {
	return "ping"
}

func (result *PingResult) ReplyRtts() []float64 {
	r := make([]float64, 0)
	for _, item := range result.Replies {
		r = append(r, item.Rtt)
	}
	return r
}

//////////////////////////////////////////////////////
// API version of a ping result

// this is the JSON structure as reported by the API
type pingResult struct {
	BaseResult
	Minimum    float64 `json:"min"`    //
	Average    float64 `json:"avg"`    //
	Maximum    float64 `json:"max"`    //
	Sent       uint    `json:"sent"`   //
	Received   uint    `json:"rcvd"`   //
	Duplicates uint    `json:"dup"`    //
	PacketSize uint    `json:"size"`   //
	Protocol   string  `json:"proto"`  //
	Step       *uint   `json:"step"`   //
	Ttl        *uint   `json:"ttl"`    //
	RawResult  []any   `json:"result"` //
}

// parse replies in the result
// ignore problems, i.e. IP addresses that don't look like IP addresses
func (result *pingResult) Replies() []PingReply {
	r := make([]PingReply, 0)
	min := 10e10
	max := 0.0
	sum := 0.0
	for _, item := range result.RawResult {
		mapitem := item.(map[string]any)
		if rtt, ok := mapitem["rtt"]; ok {
			// fill in other fields of a reply struct
			pr := PingReply{Rtt: rtt.(float64)}
			min = math.Min(min, pr.Rtt)
			max = math.Max(max, pr.Rtt)
			sum += pr.Rtt
			if src, ok := mapitem["src_addr"]; ok {
				src, err := netip.ParseAddr(src.(string))
				if err == nil {
					pr.Source = src
				}
			} else {
				pr.Source = *result.DestinationAddr // TODO: is this correct?
			}
			if ttl, ok := mapitem["ttl"]; ok {
				pr.Ttl = uint(ttl.(float64))
			} else if result.Ttl != nil {
				pr.Ttl = *result.Ttl
			}
			_, pr.Duplicate = mapitem["dup"]

			r = append(r, pr)
		}
	}

	if len(r) != 0 && (result.Average == -1 || result.Minimum == -1 || result.Maximum == -1) {
		// it may be that we have results but the supplied min/avg/max is not filled in
		// fill them in
		result.Minimum = min
		result.Maximum = max
		result.Average = sum / float64(len(r))
	}
	return r
}

func (result *pingResult) Errors() []string {
	r := make([]string, 0)
	for _, item := range result.RawResult {
		mapitem := item.(map[string]any)
		if err, ok := mapitem["error"]; ok {
			r = append(r, err.(string))
		}
	}
	return r
}

func (result *pingResult) Timeouts() uint {
	var n uint = 0
	for _, item := range result.RawResult {
		mapitem := item.(map[string]any)
		if _, ok := mapitem["x"]; ok {
			n++
		}
	}
	return n
}

func median(vals []float64) float64 {
	n := len(vals)
	slice := vals[:]
	sort.Float64s(slice)

	// follow the definition of median
	if n%2 == 0 {
		return (vals[n/2-1] + vals[n/2]) / 2
	} else {
		return vals[n/2]
	}
}
