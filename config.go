/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"flag"
	"fmt"
	"goatcli/output/annotate"
	"os"

	"github.com/google/uuid"

	"github.com/go-ini/ini"
	"github.com/robert-kisteleki/goatapi"
)

var (
	// global arguments
	flagConfig    string
	flagVerbose   bool
	flagAPIKey    string
	flagAPIEnvKey string

	// subcommand specific arguments
	flagsVersion     *flag.FlagSet
	flagsFindProbe   *flag.FlagSet
	flagsFindAnchor  *flag.FlagSet
	flagsFindMsm     *flag.FlagSet
	flagsGetResult   *flag.FlagSet
	flagsStatusCheck *flag.FlagSet
	flagsMeasure     *flag.FlagSet

	apiKey  *uuid.UUID           // specified on the command line explicitly or via env
	apiKeys map[string]uuid.UUID // collected from config file

	probeSpecCc     string
	probeSpecArea   string
	probeSpecAsn    string
	probeSpecPrefix string
	probeSpecList   string
	probeSpecReuse  string
)

const version = "v0.4.0"
const CLIName = "goatCLI " + version

var defaultConfigDir = os.Getenv("HOME") + "/.config"
var defaultConfigFile = defaultConfigDir + "/goat.ini"
var CacheDir string = os.Getenv("HOME") + "/.cache/goat/"

var Subcommands map[string]*flag.FlagSet

// parse and apply global flags
func setupFlags() {
	flag.Usage = func() {
		printUsage()
	}
	flag.StringVar(&flagConfig, "config", "", "Use this config file")
	flag.BoolVar(&flagVerbose, "v", false, "Be verbose")
	flag.BoolVar(&flagVerbose, "verbose", false, "Be verbose")
	flag.StringVar(&flagAPIKey, "key", "", "Use this API key")
	flag.StringVar(&flagAPIEnvKey, "env", "", "Use this environment variable as API key")

	flag.Parse()

	// is an API key was specified, that is used
	if flagAPIKey != "" {
		key, err := uuid.Parse(flagAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse key as API key: %v\n", err)
			os.Exit(1)
		}
		apiKey = &key
	}

	// is an API key was specified via env, that is used
	if flagAPIEnvKey != "" {
		env := os.Getenv(flagAPIEnvKey)
		key, err := uuid.Parse(env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse ENV %s as API key: %v\n", flagAPIEnvKey, err)
			os.Exit(1)
		}
		apiKey = &key
	}
}

// configure: prepare command line parsing, load config file
func configure() {
	flagsVersion = flag.NewFlagSet("version", flag.ExitOnError)
	flagsFindProbe = flag.NewFlagSet("probe", flag.ExitOnError)
	flagsFindAnchor = flag.NewFlagSet("anchor", flag.ExitOnError)
	flagsFindMsm = flag.NewFlagSet("measurement", flag.ExitOnError)
	flagsGetResult = flag.NewFlagSet("result", flag.ExitOnError)
	flagsStatusCheck = flag.NewFlagSet("status", flag.ExitOnError)
	flagsMeasure = flag.NewFlagSet("measure", flag.ExitOnError)

	Subcommands = map[string]*flag.FlagSet{
		flagsVersion.Name():     flagsVersion,
		flagsFindProbe.Name():   flagsFindProbe,
		flagsFindAnchor.Name():  flagsFindAnchor,
		flagsFindMsm.Name():     flagsFindMsm,
		flagsGetResult.Name():   flagsGetResult,
		flagsStatusCheck.Name(): flagsStatusCheck,
		flagsMeasure.Name():     flagsStatusCheck,
	}
	setupFlags()

	// prepare for, and load, the default config file
	apiKeys = make(map[string]uuid.UUID)
	if !readConfig(defaultConfigFile) {
		createConfig(defaultConfigFile)
	}

	// load configuration file that was explicitly specified
	if flagConfig != "" {
		if !readConfig(flagConfig) {
			fmt.Fprintf(os.Stderr, "Failed to read config file: %v\n", flagConfig)
			os.Exit(1)
		}
	}

	goatapi.ModifyUserAgent(CLIName)
	annotate.SetCacheDir(CacheDir, flagVerbose)
}

// readConfig deals with configuration file loading
func readConfig(confFile string) bool {
	if flagVerbose {
		fmt.Println("# Attempting to read config file (" + confFile + ")")
	}

	cfg, err := ini.LoadSources(
		ini.LoadOptions{AllowBooleanKeys: true, AllowShadows: true},
		confFile,
	)
	if err != nil {
		return false
	}

	// record stuff that was in the config file
	loadApiKey(cfg, "list_measurements")
	loadApiKey(cfg, "create_measurements")
	// TODO: add more API key variations here

	// allow config to override where the API is
	apibase := cfg.Section("").Key("apibase").MustString("")
	if apibase != "" {
		goatapi.SetAPIBase(apibase)
	}

	// allow config to override where the stream is
	streambase := cfg.Section("").Key("streambase").MustString("")
	if streambase != "" {
		goatapi.SetStreamBase(streambase)
	}

	// allow config to override the cache directory
	cachedir := cfg.Section("").Key("cachedir").MustString("")
	if cachedir != "" {
		CacheDir = cachedir
	}
	// we deliberately ignore errors on creating this dir as it may exist
	_ = os.MkdirAll(CacheDir, os.FileMode(0755))

	// load probe specification defaults
	probeSpecCc = cfg.Section("probespec").Key("probecc").MustString("")
	probeSpecArea = cfg.Section("probespec").Key("probearea").MustString("")
	probeSpecAsn = cfg.Section("probespec").Key("probeasn").MustString("")
	probeSpecPrefix = cfg.Section("probespec").Key("probeprefix").MustString("")
	probeSpecList = cfg.Section("probespec").Key("probelist").MustString("")
	probeSpecReuse = cfg.Section("probespec").Key("probereuse").MustString("")

	return true
}

// createConfig tries to create a (default) config file
func createConfig(confFile string) {
	// TODO: perhaps try to make the config directory first?

	// we deliberately ignore errors on creating this dir as it may exist
	_ = os.MkdirAll(defaultConfigDir, os.FileMode(0700))

	f, err := os.Create(confFile)
	if err != nil && flagVerbose {
		fmt.Fprintf(os.Stderr, "# Could not create default config file (%s): %v\n", confFile, err)
		return
	}
	defer f.Close()

	// having the default config file contents here allows us to distribute
	// a single binary, without the accompanying default config
	f.WriteString(`#
# this configuration file defines defaults for goatCLI
#

# apibase and streambase let you override the default API locations
# useful only for compatible APIs, i.e. proxies, API development, ...
apibase = ""
streambase = ""

# cachedir defines where to put cache files (ie. probe data)
cachedir = ""

# apikeys is where the various (private) API keys are defined
[apikeys]

# List your measurements
list_measurements = ""

# Create new measurement(s)
create_measurements = ""

# default probe specifications for new measurements
[probespec]
probecc = ""
probearea = ""
probeasn = ""
probeprefix = ""
probelist = ""
probereuse = ""
`)

	if flagVerbose {
		fmt.Fprintf(os.Stderr, "# Created default config file (%s)\n", confFile)
	}
}

// Try to scoop up a particular API key. Check its syntax.
func loadApiKey(cfg *ini.File, keyname string) {
	keyvalue := cfg.Section("apikeys").Key(keyname).String()
	if keyvalue != "" {
		uuid, err := uuid.Parse(keyvalue)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse API key input %s as API key: %v\n", keyname, err)
			os.Exit(1)
		}
		apiKeys[keyname] = uuid
	}
}

// Retrieve an appropriate API key for the function (command) the user really
// wants to execute. This may be from the config file, or explicitly
// sepecified on the command line.
func getApiKey(function string) *uuid.UUID {
	// if there's a key specified on the command line
	// then use that, regardless of the function
	if apiKey != nil {
		return apiKey
	}

	// look up the API key based on the function
	if function != "" {
		key, ok := apiKeys[function]
		if ok {
			return &key
		}
	}

	return nil
}

func getProbeSpecDefault(spec string) string {
	switch spec {
	case "cc":
		return probeSpecCc
	case "area":
		return probeSpecArea
	case "asn":
		return probeSpecAsn
	case "prefix":
		return probeSpecPrefix
	case "list":
		return probeSpecList
	case "reuse":
		return probeSpecReuse
	default:
		return ""
	}
}
