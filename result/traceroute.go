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

type TracerouteResult struct {
	BaseResult
	EndTime       uniTime         //
	ParisID       uint            //
	Protocol      string          //
	PacketSize    uint            //
	TypeOfService uint            //
	Hops          []TracerouteHop //
}

// one hop - error or data
type TracerouteHop struct {
	HopNumber uint                //
	SendError *string             //
	Responses []TraceRouteHopData //
}

type TraceRouteHopData struct {
	Error            *string         //
	Timeout          bool            //
	ErrorCode        string          // N/H/A/P/p/h/(int)
	From             netip.Addr      //
	ITypeOfService   *uint           //
	ITtl             *uint           //
	ErrorDestination *netip.Addr     //
	Late             *uint           //
	Mtu              *uint           //
	Rtt              float64         //
	Size             uint            //
	Ttl              int             //
	Flags            *string         //
	DestOptSize      *uint           //
	HopByHopOptSize  *uint           //
	IcmpExtensions   []IcmpExtension //
}

type IcmpExtension struct {
	Version uint                  //
	Rfc4884 uint                  //
	Objects []IcmpExtensionObject //
}

type IcmpExtensionObject struct {
	Class      uint         //
	Type       uint         //
	MplsObject []MplsObject //
}

type MplsObject struct {
	Experimental  uint //
	Label         uint //
	BottomOfStack uint //
	Ttl           uint //
}

func (result *TracerouteResult) TypeName() string {
	return "traceroute"
}

func (trace *TracerouteResult) Parse(from string) (err error) {
	var itrace tracerouteResult
	err = json.Unmarshal([]byte(from), &itrace)
	if err != nil {
		return err
	}
	if itrace.Type != "traceroute" {
		return fmt.Errorf("this is not a traceroute result (type=%s)", itrace.Type)
	}
	trace.BaseResult = itrace.BaseResult
	trace.Protocol = itrace.Protocol
	trace.EndTime = itrace.EndTime
	trace.ParisID = itrace.ParisID
	trace.PacketSize = itrace.PacketSize
	trace.TypeOfService = itrace.TypeOfService

	trace.Hops = make([]TracerouteHop, 0)
	for _, ihop := range itrace.RawResult {
		hop := TracerouteHop{}
		hop.HopNumber = ihop.HopNumber
		if ihop.SendError != nil {
			hop.SendError = ihop.SendError
			trace.Hops = append(trace.Hops, hop)
			continue
		}
		hop.Responses = make([]TraceRouteHopData, 0)
		for _, ihopdata := range *ihop.HopData {
			hopdata := TraceRouteHopData{}
			if ihopdata.Timeout != nil {
				hopdata.Timeout = true
				hop.Responses = append(hop.Responses, hopdata)
				continue // on timeout: no useful data
			}
			if ihopdata.Error != nil {
				hopdata.Error = ihopdata.Error
				hop.Responses = append(hop.Responses, hopdata)
				continue // on error: no useful data
			}
			if ihopdata.ErrorCode != nil {
				hopdata.ErrorCode = fmt.Sprint(*ihopdata.ErrorCode)
			}
			hopdata.From = ihopdata.From
			if ihopdata.Size != nil {
				hopdata.Size = *ihopdata.Size
			}
			if ihopdata.Ttl != nil {
				hopdata.Ttl = *ihopdata.Ttl
			}
			if ihopdata.Late != nil {
				hopdata.Late = ihopdata.Late
				continue // no other data is it was a LATE packet
			}
			if ihopdata.Rtt != nil {
				hopdata.Rtt = *ihopdata.Rtt
			}
			if ihopdata.ITtl != nil {
				hopdata.ITtl = ihopdata.ITtl
			}
			if ihopdata.ITypeOfService != nil {
				hopdata.ITypeOfService = ihopdata.ITypeOfService
			}
			if ihopdata.ErrorDestination != nil {
				hopdata.ErrorDestination = ihopdata.ErrorDestination
			}
			if ihopdata.Mtu != nil {
				hopdata.Mtu = ihopdata.Mtu
			}
			if ihopdata.Flags != nil {
				hopdata.Flags = ihopdata.Flags
			}
			if ihopdata.DestOptSize != nil {
				hopdata.DestOptSize = ihopdata.DestOptSize
			}
			if ihopdata.HopByHopOptSize != nil {
				hopdata.HopByHopOptSize = ihopdata.HopByHopOptSize
			}
			hopdata.IcmpExtensions = make([]IcmpExtension, 0)
			// process one extension - change if there can be more
			if ihopdata.IcmpExtension != nil {
				iext := ihopdata.IcmpExtension
				ext := IcmpExtension{
					iext.Version,
					iext.Rfc4884,
					make([]IcmpExtensionObject, 0),
				}
				hopdata.IcmpExtensions = append(hopdata.IcmpExtensions, ext)
				for _, iextobj := range iext.Objects {
					extobj := IcmpExtensionObject{
						iextobj.Class,
						iextobj.Type,
						make([]MplsObject, 0),
					}
					if iextobj.MplsObject != nil {
						for _, implsobj := range *iextobj.MplsObject {
							extobj.MplsObject = append(extobj.MplsObject,
								MplsObject{
									implsobj.Experimental,
									implsobj.Label,
									implsobj.BottomOfStack,
									implsobj.Ttl,
								},
							)
						}
					}
					ext.Objects = append(ext.Objects, extobj)
				}
			}
			hop.Responses = append(hop.Responses, hopdata)
		}
		trace.Hops = append(trace.Hops, hop)
	}

	return nil
}

func (trace *TracerouteResult) DestinationReached() bool {
	if len(trace.Hops) == 0 {
		return false
	}
	for _, ans := range trace.Hops[len(trace.Hops)-1].Responses {
		if ans.From == *trace.DestinationAddr {
			return true
		}
	}
	return false
}

//////////////////////////////////////////////////////
// API version of a traceroute result

// this is the JSON structure as reported by the API

type tracerouteResult struct {
	BaseResult
	EndTime       uniTime       `json:"endtime"`  //
	ParisID       uint          `json:"paris_id"` //
	Protocol      string        `json:"proto"`    //
	PacketSize    uint          `json:"size"`     //
	TypeOfService uint          `json:"tos"`      //
	RawResult     []rawTraceHop `json:"result"`   //
}

// one hop - error or data
type rawTraceHop struct {
	HopNumber uint               `json:"hop"`    //
	SendError *string            `json:"error"`  //
	HopData   *[]rawTraceHopData `json:"result"` //
}

type errorCode string

func (e *errorCode) UnmarshalJSON(b []byte) error {
	var val any
	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}
	switch v := val.(type) {
	case int:
		*e = errorCode(fmt.Sprintf("%v", v))
	case float64:
		*e = errorCode(fmt.Sprintf("%d", int(v)))
	case string:
		*e = errorCode(v)
	default:
		return fmt.Errorf("unexpected error code with type %T and value %v", v, v)
	}
	return nil
}

// one hop detail
type rawTraceHopData struct {
	Timeout          *string           `json:"x"`          //
	Error            *string           `json:"error"`      //
	ErrorCode        *errorCode        `json:"err"`        // N/H/A/P/p/h/(int)
	From             netip.Addr        `json:"from"`       //
	ITypeOfService   *uint             `json:"itos"`       //
	ITtl             *uint             `json:"ittl"`       //
	ErrorDestination *netip.Addr       `json:"edst"`       //
	Late             *uint             `json:"late"`       //
	Mtu              *uint             `json:"mtu"`        //
	Rtt              *float64          `json:"rtt"`        //
	Size             *uint             `json:"size"`       //
	Ttl              *int              `json:"ttl"`        //
	Flags            *string           `json:"flags"`      //
	DestOptSize      *uint             `json:"dstoptsize"` //
	HopByHopOptSize  *uint             `json:"hbhoptsize"` //
	IcmpExtension    *rawIcmpExtension `json:"icmpext"`    //
}

type rawIcmpExtension struct {
	Version uint                     `json:"version"` //
	Rfc4884 uint                     `json:"rfc4884"` //
	Objects []rawIcmpExtensionObject `json:"obj"`     //
}

type rawIcmpExtensionObject struct {
	Class      uint             `json:"class"` //
	Type       uint             `json:"type"`  //
	MplsObject *[]rawMplsObject `json:"mpls"`  //
}

type rawMplsObject struct {
	Label         uint `json:"label"` //
	BottomOfStack uint `json:"s"`     //
	Ttl           uint `json:"ttl"`   //
	Experimental  uint `json:"exp"`   //
}
