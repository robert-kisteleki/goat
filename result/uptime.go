/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"encoding/json"
	"fmt"
)

type UptimeResult struct {
	BaseResult
	Uptime uint //
}

func (result *UptimeResult) TypeName() string {
	return "uptime"
}

func (uptime *UptimeResult) Parse(from string) (err error) {
	var iuptime uptimeResult
	err = json.Unmarshal([]byte(from), &iuptime)
	if err != nil {
		return err
	}
	if iuptime.Type != "uptime" {
		return fmt.Errorf("this is not an uptime result (type=%s)", iuptime.Type)
	}
	uptime.BaseResult = iuptime.BaseResult
	uptime.Uptime = iuptime.Uptime

	return nil
}

//////////////////////////////////////////////////////
// API version of an uptime result

// this is the JSON structure as reported by the API
type uptimeResult struct {
	BaseResult
	Uptime uint `json:"uptime"` //
}
