package main

import (
	"regexp"
	"time"
	"net/url"
	"io/ioutil"
	"github.com/spf13/pflag"
	"github.com/tehmoon/errors"
	"runtime"
)

type Flags struct {
	OutputDir string
	Wordlist string
	Target string
	Insecure bool
	File string
	TestWord string
	Routines int
	Proxy *url.URL
	proxy string
	HideRegexes []*regexp.Regexp
	HideCodes []int
	HideNumWords []int
	HideLines []int
	HideChars []int
	hideNumWords []string
	hideLines []string
	hideChars []string
	hideRegexes []string
	hideCodes []string
	ShowRegexes []*regexp.Regexp
	ShowCodes []int
	ShowLines []int
	ShowChars []int
	ShowNumWords []int
	showLines []string
	showChars []string
	showNumWords []string
	showRegexes []string
	showCodes []string
	Verbose bool
	RandUserAgent bool
	ThrottleDuration time.Duration
	MaxRedirects int
	ExtWords []string
	ExtWordlist string
	TestExtWord string
}

func parseFlags() (flags *Flags, err error) {
	flags = &Flags{}

	pflag.StringVarP(&flags.Target, "target", "t", "", "Set the target URL")
	pflag.StringVarP(&flags.TestWord, "test-word", "", "", "Show the request to stdout using the word as the .Word object for the template")
	pflag.StringVarP(&flags.TestExtWord, "test-ext-word", "", "", "Show the request when --test-word is used as the .Ext object for the template")
	pflag.BoolVarP(&flags.Insecure, "insecure", "k", false, "Bypass TLS checks")
	pflag.StringVarP(&flags.Wordlist, "wordlist", "w", "", "Use worldlist items as root object for the template")
	pflag.StringVarP(&flags.File, "file", "f", "", "Specify raw http request file as input")
	pflag.IntVarP(&flags.Routines, "routines", "r", runtime.NumCPU(), "Max parallel running go-routines")
	pflag.StringArrayVarP(&flags.hideCodes, "hc", "", []string{}, "Hide responses with specified status code")
	pflag.StringArrayVarP(&flags.hideLines, "hl", "", []string{}, "Hide responses with specified number of lines")
	pflag.StringArrayVarP(&flags.hideChars, "hh", "", []string{}, "Hide responses with specified number of characters")
	pflag.StringArrayVarP(&flags.hideNumWords, "hnw", "", []string{}, "Hide if number of words matches")
	pflag.StringArrayVarP(&flags.showCodes, "sc", "", []string{"200","204","301","302","307",}, "Show responses with specified status code")
	pflag.StringArrayVarP(&flags.showLines, "sl", "", []string{}, "Show responses with specified number of lines")
	pflag.StringArrayVarP(&flags.showChars, "sh", "", []string{}, "Show responses with specified number of characters")
	pflag.StringArrayVarP(&flags.showNumWords, "snw", "", []string{}, "Show if number of words matches")
	pflag.BoolVarP(&flags.Verbose, "verbose", "v", false, "Output all request as if they all matched")
	pflag.BoolVarP(&flags.RandUserAgent, "use-random-user-agent", "", false, "Force using a random user agent for each request")
	pflag.StringVarP(&flags.proxy, "proxy", "p", "", "Send the requests through a http proxy")
	pflag.IntVarP(&flags.MaxRedirects, "max-redirects", "", 0, "Maximum number of follow http redirect")
	pflag.StringArrayVarP(&flags.hideRegexes, "hr", "", []string{}, "Hide responses with specifed regex from the body")
	pflag.StringArrayVarP(&flags.showRegexes, "sr", "", []string{}, "Show responses with specifed regex from the body")
	pflag.DurationVarP(&flags.ThrottleDuration, "throttle-duration", "d", time.Duration(0), "Throttle sending requests")
	pflag.StringVarP(&flags.OutputDir, "output-dir", "o", "", "Dump matched requests and responses into a directory")
	pflag.StringArrayVarP(&flags.ExtWords, "ext", "x", []string{}, "Words to be process for each word from the main wordlist.")
	pflag.StringVarP(&flags.ExtWordlist, "ext-list", "", "", "Wordlist to be process for each word from the main wordlist.")

	pflag.Parse()

	if flags.Target == "" {
		return nil, errors.Errorf("Flag %q is mandatory", "--target")
	}

	if flags.File == "" {
		return nil, errors.Errorf("Flag %q is mandatory", "--file")
	}

	if flags.Routines < 1 {
		return nil, errors.Errorf("Flag %q must be a positive integer", "--routines")
	}

	flags.ShowCodes, err = StringArrayToInt(flags.showCodes, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of integers", "--sc")
	}

	flags.HideCodes, err = StringArrayToInt(flags.hideCodes, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of integers", "--hc")
	}

	flags.ShowLines, err = StringArrayToInt(flags.showLines, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of integers", "--sl")
	}

	flags.HideLines, err = StringArrayToInt(flags.hideLines, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of integers", "--hl")
	}

	flags.HideChars, err = StringArrayToInt(flags.hideChars, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of integers", "--hh")
	}

	flags.ShowChars, err = StringArrayToInt(flags.showChars, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of integers", "--sh")
	}

	flags.HideNumWords, err = StringArrayToInt(flags.hideNumWords, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of integers", "--hnw")
	}

	flags.ShowNumWords, err = StringArrayToInt(flags.showNumWords, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of integers", "--snw")
	}

	if flags.proxy != "" {
		flags.Proxy, err = url.Parse(flags.proxy)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing url at flag %q", "--proxy")
		}
	}

	flags.HideRegexes, err = StringArrayToRegex(flags.hideRegexes)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of https://godoc.org/regexp", "--hr")
	}

	flags.ShowRegexes, err = StringArrayToRegex(flags.showRegexes)
	if err != nil {
		return nil, errors.Wrapf(err, "Flag %q must be an array of https://godoc.org/regexp", "--sr")
	}

	flags.OutputDir, err = ioutil.TempDir(flags.OutputDir, time.Now().Format("02-01-2006-03:04:05_"))
	if err != nil {
		return nil, errors.Wrapf(err, "Error creating temporary directory for flag %q", "--output-dir")
	}

	return flags, nil
}
