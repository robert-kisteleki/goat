/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"encoding/json"
	"fmt"
	"net/netip"
)

type ConnectionResult struct {
	BaseResult
	Event      string
	Controller string
	Asn        uint
	Prefix     netip.Prefix
}

func (result *ConnectionResult) TypeName() string {
	return "connection"
}

func (conn *ConnectionResult) Parse(from string) (err error) {
	var iconn connectionResult
	err = json.Unmarshal([]byte(from), &iconn)
	if err != nil {
		return err
	}
	if iconn.Type != "connection" {
		return fmt.Errorf("this is not a connection result (type=%s)", iconn.Type)
	}
	conn.BaseResult = iconn.BaseResult
	conn.Event = iconn.Event
	conn.Controller = iconn.Controller
	conn.Asn = iconn.Asn
	conn.Prefix = iconn.Prefix

	return nil
}

//////////////////////////////////////////////////////
// API version of an uptime result

// this is the JSON structure as reported by the API
type connectionResult struct {
	BaseResult
	Event      string       `json:"event"`      //
	Controller string       `json:"controller"` //
	Asn        uint         `json:"asn"`        //
	Prefix     netip.Prefix `json:"prefix"`     //
}
