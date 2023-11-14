/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"
	"time"
)

type BaseResult struct {
	FirmwareVersion firmwareVersion `json:"fw"`               //
	CodeVersion     string          `json:"mver"`             //
	MeasurementID   uint            `json:"msm_id"`           //
	GroupID         uint            `json:"group_id"`         //
	ProbeID         uint            `json:"prb_id"`           //
	MeasurementName string          `json:"msm_name"`         // measurement name (better use type)
	Type            string          `json:"type"`             // measurement type
	TimeStamp       uniTime         `json:"timestamp"`        // when was this result collected
	StoreTimeStamp  uniTime         `json:"stored_timestamp"` // when was this result stored
	Bundle          uint            `json:"bundle"`           // ID for a collection of related measurement results
	LastTimeSync    int             `json:"lts"`              // how long ago was the probe's clock synced
	DestinationName string          `json:"dst_name"`         //
	DestinationAddr *netip.Addr     `json:"dst_addr"`         //
	SourceAddr      netip.Addr      `json:"src_addr"`         // source address used by probe
	FromAddr        netip.Addr      `json:"from"`             // IP address of the probe as known by the infra
	AddressFamily   uint            `json:"af"`               // 4 or 6
	ResolveTime     *float64        `json:"ttr"`              // only if resolve-on-probe was used
}

func (result *BaseResult) Parse(from string) (err error) {
	err = json.Unmarshal([]byte(from), &result)
	if err != nil {
		return err
	}
	return nil
}

func (result *BaseResult) BaseString() string {
	d := "N/A"
	if result.DestinationName != "" {
		d = result.DestinationName
	}
	ret := fmt.Sprintf("%d\t%d\t%v\t%s",
		result.MeasurementID,
		result.ProbeID,
		result.TimeStamp,
		d,
	)
	ret += valueOrNA("", false, result.DestinationAddr)
	return ret
}

func (result *BaseResult) BaseDetailString() string {
	return result.BaseString() +
		fmt.Sprintf("\t%d", result.AddressFamily)
}

func (result *BaseResult) TypeName() string {
	return result.Type
}

// Destination yields the destination name
// Or if that's not defined then the destination address
func (result *BaseResult) Destination() string {
	if result.DestinationName != "" {
		return result.DestinationName
	} else {
		return result.DestinationAddr.String()
	}
}

func (result *BaseResult) GetTimeStamp() time.Time {
	return time.Time(result.TimeStamp)
}

func (result *BaseResult) GetProbeID() uint {
	return result.ProbeID
}

func (result *BaseResult) GetFirmwareVersion() uint {
	return uint(result.FirmwareVersion)
}

type firmwareVersion uint

func (fw *firmwareVersion) UnmarshalJSON(b []byte) error {
	var val any
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case int:
		*fw = firmwareVersion(uint(v))
	case float64:
		*fw = firmwareVersion(uint(v))
	case string:
		val, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("unable to parse firmware version with type %T and value %v", v, v)
		}
		*fw = firmwareVersion(uint(val))
	default:
		return fmt.Errorf("unexpected firmware version with type %T and value %v", v, v)
	}
	return nil
}
