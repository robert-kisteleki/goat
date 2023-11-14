/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"encoding/json"
	"fmt"
)

type NtpResult struct {
	BaseResult
	Protocol           string     //
	Version            uint       //
	LeapIndicator      string     // "no", "59", "61" or "unknown"
	Mode               string     //
	Stratum            uint       //
	PollInterval       uint       //
	Precision          float64    //
	RootDelay          float64    //
	RootDispersion     float64    //
	ReferenceID        string     //
	ReferenceTimestamp float64    //
	Replies            []NtpReply //
	Errors             []string   //
}

// one successful ntp reply
type NtpReply struct {
	OriginTimestamp   float64 //
	TransmitTimestamp float64 //
	ReceiveTimestamp  float64 //
	FinalTimestamp    float64 //
	Offset            float64 //
	Rtt               float64 //
}

func (result *NtpResult) TypeName() string {
	return "ntp"
}

func (ntp *NtpResult) Parse(from string) (err error) {
	var intp ntpResult
	err = json.Unmarshal([]byte(from), &intp)
	if err != nil {
		return err
	}
	if intp.Type != "ntp" {
		return fmt.Errorf("this is not an NTP result (type=%s)", intp.Type)
	}
	ntp.BaseResult = intp.BaseResult
	ntp.Protocol = intp.Protocol
	ntp.Version = intp.Version
	ntp.LeapIndicator = intp.LeapIndicator
	ntp.Mode = intp.Mode
	ntp.Stratum = intp.Stratum
	ntp.PollInterval = intp.PollInterval
	ntp.Precision = intp.Precision
	ntp.RootDelay = intp.RootDelay
	ntp.RootDispersion = intp.RootDispersion
	ntp.ReferenceID = intp.ReferenceID
	ntp.ReferenceTimestamp = intp.ReferenceTimestamp
	ntp.Replies = intp.Replies()
	ntp.Errors = intp.Errors()

	return nil
}

//////////////////////////////////////////////////////
// API version of an NTP result

// this is the JSON structure as reported by the API
type ntpResult struct {
	BaseResult
	Protocol           string  `json:"proto"`           //
	Version            uint    `json:"version"`         //
	LeapIndicator      string  `json:"li"`              // "no", "59", "61" or "unknown"
	Mode               string  `json:"mode"`            //
	Stratum            uint    `json:"stratum"`         //
	PollInterval       uint    `json:"poll"`            //
	Precision          float64 `json:"precision"`       //
	RootDelay          float64 `json:"root-delay"`      //
	RootDispersion     float64 `json:"root-dispersion"` //
	ReferenceID        string  `json:"ref-id"`          //
	ReferenceTimestamp float64 `json:"ref-ts"`          //
	RawResult          []any   `json:"result"`          //
}

func (result *ntpResult) Replies() []NtpReply {
	r := make([]NtpReply, 0)
	for _, item := range result.RawResult {
		mapitem := item.(map[string]any)
		if rtt, ok := mapitem["rtt"]; ok {
			// fill in other fields of a reply struct
			ntpr := NtpReply{Rtt: rtt.(float64)}
			if offset, ok := mapitem["offset"]; ok {
				ntpr.Offset = offset.(float64)
			}
			if origints, ok := mapitem["origin-ts"]; ok {
				ntpr.OriginTimestamp = origints.(float64)
			}
			if transmitts, ok := mapitem["transmit-ts"]; ok {
				ntpr.TransmitTimestamp = transmitts.(float64)
			}
			if receivets, ok := mapitem["receive-ts"]; ok {
				ntpr.ReceiveTimestamp = receivets.(float64)
			}
			if finalts, ok := mapitem["final-ts"]; ok {
				ntpr.FinalTimestamp = finalts.(float64)
			}

			r = append(r, ntpr)
		}
	}
	return r
}

func (result *ntpResult) Errors() []string {
	r := make([]string, 0)
	for _, item := range result.RawResult {
		mapitem := item.(map[string]any)
		if err, ok := mapitem["x"]; ok {
			r = append(r, err.(string))
		}
	}
	return r
}
