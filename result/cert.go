/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
)

type CertResult struct {
	BaseResult
	Error           *string            //
	Alert           *CertAlert         //
	Method          string             //
	ConnectTime     float64            //
	ReplyTime       float64            //
	ServerCipher    string             //
	ProtocolVersion string             //
	Certificates    []x509.Certificate //
	DnsError        string             //
}

// CertAlert is an error that could be sent by the server
// see RFC 5246 section 7.2
type CertAlert struct {
	Level       uint //
	Description uint //
}

const (
	AlertLevelWarning = iota
	AlertLevelFatal
)

func (result *CertResult) TypeName() string {
	return "sslcert"
}

func (cert *CertResult) Parse(from string) (err error) {
	var icert certResult
	err = json.Unmarshal([]byte(from), &icert)
	if err != nil {
		return err
	}
	if icert.Type != "sslcert" {
		return fmt.Errorf("this is not a TLS/SSL certificate result (type=%s)", icert.Type)
	}
	cert.BaseResult = icert.BaseResult
	cert.Alert = icert.Alert
	cert.Error = icert.Error
	if icert.Error == nil {
		if icert.DnsError != nil {
			cert.DnsError = *icert.DnsError
		} else {
			// we use dnserror as a hint that there's no real data
			cert.Method = *icert.Method
			cert.ReplyTime = *icert.ReplyTime
			cert.ConnectTime = *icert.ConnectTime
			// if there's an alert, there's no other real data
			if icert.Alert == nil {
				if icert.ServerCipher != nil {
					cert.ServerCipher = *icert.ServerCipher
				}
				cert.ProtocolVersion = *icert.ProtocolVersion
				cert.Certificates, err = icert.Certificates()
				if err != nil {
					return nil
				}
			}
		}
	}

	return nil
}

//////////////////////////////////////////////////////
// API version of an sslcert result

type certResult struct {
	BaseResult
	Alert           *CertAlert `json:"alert"`         //
	Method          *string    `json:"method"`        //
	ReplyTime       *float64   `json:"rt"`            //
	ServerCipher    *string    `json:"server_cipher"` //
	ConnectTime     *float64   `json:"ttc"`           //
	ProtocolVersion *string    `json:"ver"`           //
	Error           *string    `json:"err"`           //
	RawCertificates *[]string  `json:"cert"`          //
	DnsError        *string    `json:"dnserr"`        //
}

func (result *certResult) Certificates() (list []x509.Certificate, err error) {
	list = make([]x509.Certificate, 0)
	if result.RawCertificates == nil {
		return
	}
	for _, item := range *result.RawCertificates {
		block, _ := pem.Decode([]byte(item))
		if block == nil || block.Type != "CERTIFICATE" {
			log.Fatal("failed to decode PEM block containing certificate")
		}
		var cert *x509.Certificate
		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return list, err
		}
		list = append(list, *cert)
	}
	return list, nil
}
