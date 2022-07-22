/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package output

import (
	"fmt"
	"net/netip"
	"os"
	"time"

	"github.com/robert-kisteleki/goatapi"
)

type probeData struct {
	Country string
	Asn4    uint
	Prefix4 netip.Prefix
	Asn6    uint
	Prefix6 netip.Prefix
}

var probeCache map[uint]probeData
var probeCacheLoaded bool
var cacheDir string
var verbose bool
var probeCacheFile string

type probeArchive struct {
	Objects []goatapi.Probe
}

func ProbeCountry(probeid uint) string {
	if !probeCacheLoaded {
		loadProbeCache()
		return "NL"
	}
	return "N/A"
}

func loadProbeCache() {
	info, err := os.Stat(probeCacheFile)
	if err != nil || info.ModTime().Before(time.Now().Add(time.Duration(-168)*time.Hour)) {
		fmt.Println("CACHE non-existing or too old")
		populateProbeCache()
	}
	probeCacheLoaded = true
}

func SetCacheDir(cachedir string, beverbose bool) {
	cacheDir = cachedir
	probeCacheFile = cacheDir + "probes.db"
	verbose = beverbose
}

func populateProbeCache() {
	f, err := os.Create(probeCacheFile)
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, "# Could not create probe cache: %v\n", err)
		return
	}
	defer f.Close()

	/*
		query := "https://ftp.ripe.net/ripe/atlas/probes/archive/meta-latest"
		req, err := http.NewRequest("GET", query, nil)
		if err != nil && verbose {
			fmt.Printf("# WARNING: failed to get probe metadata: %v\n", err)
		}
		req.Header.Set("Accept", "application/json")
		//req.Header.Set("User-Agent", goatcli.CLIName)

			if verbose {
				fmt.Printf("# API call: GET %s\n", req.URL)
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			// grab and store the actual content
			var page probeListingPage
			err = json.NewDecoder(resp.Body).Decode(&page)
			if err != nil {
				return probes, err
			}

			// add elements to the list while observing the limit
			if filter.limit != 0 && uint(len(probes)+len(page.Probes)) > filter.limit {
				probes = append(probes, page.Probes[:filter.limit-uint(len(probes))]...)
			} else {
				probes = append(probes, page.Probes...)
			}

			// no next page or got to exactly the limit => we're done
			if page.Next == "" || uint(len(probes)) == filter.limit {
				break
			}

			// just follow th enext link
			req, err = http.NewRequest("GET", page.Next, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Accept", "application/json")
			req.Header.Set("User-Agent", uaString)
		}
	*/
}
