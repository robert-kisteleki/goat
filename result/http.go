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

type HttpResult struct {
	BaseResult
	Uri             string   //
	HeaderSize      uint     //
	Headers         []string //
	BodySize        uint     //
	Method          string   //
	Version         string   //
	ResultCode      uint     //
	ReplyTime       float64  //
	TimeToConnect   float64  //
	TimeToFirstByte float64  //
	DnsError        string   //
	Error           string   //

	//ReadTiming *HttpReadTiming `json:"readtiming"` //
	//SubID         *uint               //
	//SubMax        *uint           //
	//Time *uniTime `json:"time"` //
}

func (result *HttpResult) TypeName() string {
	return "http"
}

func (http *HttpResult) Parse(from string) (err error) {
	var ihttp httpResult
	err = json.Unmarshal([]byte(from), &ihttp)
	if err != nil {
		return err
	}
	if ihttp.Type != "http" {
		return fmt.Errorf("this is not a HTTP result (type=%s)", ihttp.Type)
	}

	http.BaseResult = ihttp.BaseResult
	http.Uri = ihttp.Uri
	if len(ihttp.RawHttpReply) >= 1 {
		// only deal with response 1 for now

		resp := ihttp.RawHttpReply[0]
		if resp.Headers != nil {
			http.Headers = *resp.Headers
		}
		http.HeaderSize = resp.HeaderSize
		http.BodySize = resp.BodySize
		http.Method = resp.Method
		http.Version = resp.Version
		http.ResultCode = resp.ResultCode
		http.ReplyTime = resp.ReplyTime
		http.TimeToConnect = resp.TimeToConnect
		http.TimeToFirstByte = resp.TimeToFirstByte
		if resp.DnsError != nil {
			http.DnsError = *resp.DnsError
		}
		if resp.Error != nil {
			http.Error = *resp.Error
		}
	}

	return nil
}

//////////////////////////////////////////////////////
// API version of a http result

type httpResult struct {
	BaseResult
	Uri          string         `json:"uri"`    //
	RawHttpReply []rawHttpReply `json:"result"` //
}

type rawHttpReply struct {
	BodySize        uint            `json:"bsize"`      //
	DnsError        *string         `json:"dnserr"`     //
	DestinationAddr netip.Addr      `json:"dst_addr"`   //
	Error           *string         `json:"err"`        //
	Headers         *[]string       `json:"header"`     //
	HeaderSize      uint            `json:"hsize"`      //
	Method          string          `json:"method"`     //
	ReadTiming      *httpReadTiming `json:"readtiming"` //
	ResultCode      uint            `json:"res"`        //
	ReplyTime       float64         `json:"rt"`         //
	SubID           *uint           `json:"subid"`      //
	SubMax          *uint           `json:"submax"`     //
	Time            *uniTime        `json:"time"`       //
	TimeToConnect   float64         `json:"ttc"`        //
	TimeToFirstByte float64         `json:"ttfb"`       //
	Version         string          `json:"ver"`        //
}

type httpReadTiming struct {
	Offset    uint `json:"o"` //
	TimeSince uint `json:"t"` //
}
