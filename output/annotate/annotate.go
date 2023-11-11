/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package annotate

import (
	"bufio"
	"compress/bzip2"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"time"
)

type probeData struct {
	ID      uint          `json:"id"`
	Country *string       `json:"country_code"`
	Asn4    *uint         `json:"asn_v4"`
	Prefix4 *netip.Prefix `json:"prefix_v4"`
	Asn6    *uint         `json:"asn_v6"`
	Prefix6 *netip.Prefix `json:"prefix_v6"`
}

var probeCache map[uint]probeData
var probeCacheLoaded bool
var cacheDir string
var verbose bool
var probeCacheFile string

type probeArchive struct {
	Objects []probeData `json:"objects"`
}

func GetProbeCountry(probeid uint) string {
	if !probeCacheLoaded {
		initProbeCache()
	}
	p, ok := probeCache[probeid]
	if ok && p.Country != nil {
		return *p.Country
	}
	return "N/A"
}

func GetProbePrefix4(probeid uint) string {
	if !probeCacheLoaded {
		initProbeCache()
	}
	p, ok := probeCache[probeid]
	if ok && p.Prefix4 != nil {
		return p.Prefix4.String()
	} else {
		return "N/A"
	}
}

func GetProbePrefix6(probeid uint) string {
	if !probeCacheLoaded {
		initProbeCache()
	}
	p, ok := probeCache[probeid]
	if ok && p.Prefix6 != nil {
		return p.Prefix6.String()
	} else {
		return "N/A"
	}
}

func GetProbeAsn4(probeid uint) string {
	if !probeCacheLoaded {
		initProbeCache()
	}
	p, ok := probeCache[probeid]
	if ok && p.Asn4 != nil {
		return fmt.Sprintf("%d", *p.Asn4)
	} else {
		return "N/A"
	}
}

func GetProbeAsn6(probeid uint) string {
	if !probeCacheLoaded {
		initProbeCache()
	}
	p, ok := probeCache[probeid]
	if ok && p.Asn6 != nil {
		return fmt.Sprintf("%d", *p.Asn6)
	} else {
		return "N/A"
	}
}

func SetCacheDir(cachedir string, beverbose bool) {
	cacheDir = cachedir
	probeCacheFile = cacheDir + "probes.db"
	verbose = beverbose
}

func InitProbeCache() {
	initProbeCache()
}

func initProbeCache() {

	probeCache = make(map[uint]probeData, 0)
	loadProbeCache()
	probeCacheLoaded = true
}

func loadProbeCache() {
	info, err := os.Stat(probeCacheFile)
	if err != nil || info.ModTime().Before(time.Now().Add(time.Duration(-168)*time.Hour)) {
		if verbose {
			fmt.Printf("# Re-populating probe cache\n")
		}
		populateProbeCache()
		return
	}

	if verbose {
		fmt.Printf("# Loading probe cache from %s\n", probeCacheFile)
	}

	// load from file
	cachefile, err := os.Open(probeCacheFile)
	if err != nil {
		if verbose {
			fmt.Printf("# Could not open probe cache: %v\n", err)
		}
		return
	}
	defer cachefile.Close()

	scanner := bufio.NewScanner(bufio.NewReader(cachefile))
	for scanner.Scan() {
		line := scanner.Text()
		var probe probeData
		err = json.Unmarshal([]byte(line), &probe)
		if err != nil {
			return
		}
		probeCache[probe.ID] = probe
	}

}

func populateProbeCache() {
	cachefile, err := os.Create(probeCacheFile)
	if err != nil {
		if verbose {
			fmt.Printf("# Could not create probe cache: %v\n", err)
		}
		return
	}
	defer cachefile.Close()

	query := "https://ftp.ripe.net/ripe/atlas/probes/archive/meta-latest"
	req, err := http.NewRequest("GET", query, nil)
	if err != nil && verbose {
		fmt.Printf("# WARNING: failed to get probe metadata: %v\n", err)
		return
	}
	//req.Header.Set("User-Agent", goatcli.CLIName)

	if verbose {
		fmt.Printf("# GET %s\n", req.URL)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("# WARNING: failed to get probe metadata: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// grab and store the actual content
	var archive probeArchive
	err = json.NewDecoder(bzip2.NewReader(resp.Body)).Decode(&archive)
	if err != nil {
		fmt.Printf("# WARNING: failed to get probe metadata: %v\n", err)
		return
	}

	for _, probe := range archive.Objects {
		// put this in the cache
		probeCache[probe.ID] = probe

		// store for the next run
		b, _ := json.Marshal(&probe)
		cachefile.WriteString(string(b) + "\n")
	}
}
