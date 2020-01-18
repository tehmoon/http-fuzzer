package main

import (
	"fmt"
	"os"
	"github.com/olekukonko/tablewriter"
	"bytes"
	"io/ioutil"
	"path/filepath"
)

func displayFlags(flags *Flags, wcl, twc int) {
	buffer := bytes.NewBuffer(nil)
	data := [][]string{
		[]string{"Target", flags.Target,},
		[]string{"Wordlist", flags.Wordlist,},
		[]string{"Wordlist length", fmt.Sprintf("%d", wcl),},
		[]string{"Ext wordlist", flags.ExtWordlist,},
		[]string{"Ext words length", fmt.Sprintf("%d", len(flags.ExtWords)),},
		[]string{"Total words to be processed", fmt.Sprintf("%d", twc),},
		[]string{"File", flags.File,},
		[]string{"Proxy", flags.proxy,},
		[]string{"Throttle duration", flags.ThrottleDuration.String(),},
		[]string{"Random user agent", fmt.Sprintf("%t", flags.RandUserAgent),},
		[]string{"Output directory", flags.OutputDir,},
		[]string{"Show codes", formatIntArray(flags.ShowCodes, 10),},
		[]string{"Show lines", formatIntArray(flags.ShowLines, 10),},
		[]string{"Show Chars", formatIntArray(flags.ShowChars, 10),},
		[]string{"Show Regexes", fmt.Sprintf("%v", flags.ShowRegexes),},
		[]string{"Insecure", fmt.Sprintf("%t", flags.Insecure),},
		[]string{"Routines", fmt.Sprintf("%d", flags.Routines),},
		[]string{"Test word", flags.TestWord,},
		[]string{"Test ext word", flags.TestExtWord,},
		[]string{"Verbose", fmt.Sprintf("%t", flags.Verbose),},
		[]string{"Use random User-Agent", fmt.Sprintf("%t", flags.RandUserAgent),},
		[]string{"Max HTTP redirects", fmt.Sprintf("%d", flags.MaxRedirects),},
	}

	table := tablewriter.NewWriter(buffer)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	table.SetHeader([]string{"Flag","Value",})
	for _, v := range data {
		table.Append(v)
	}

	table.Render()

	os.Stderr.Write(buffer.Bytes())
	ioutil.WriteFile(filepath.Join(flags.OutputDir, "flags.txt"), buffer.Bytes(), 0644)
}

func display(res *Result, show bool, flags *Flags) {
	if show || flags.Verbose {
		fmt.Fprintf(os.Stderr, "\r\x1b[2K")
		// TODO: adjust padding to the longuest word
		// TODO: fix this ugly condition
		if res.RequestWord.Ext == "" {
			fmt.Printf(
				"%-20s\tcode: %d\twords: %d\tchars: %d\tlines: %d\tredirects: %d\toffset: %d\n",
				res.RequestWord.Word,
				res.Response.Response.StatusCode,
				res.Response.NumWords,
				len(res.Response.Body),
				res.Response.NumLines,
				res.NumRedirects,
				res.RequestWord.Offset)
		} else {
			fmt.Printf(
				"%-20s\text:%-10s\tcode: %d\twords: %d\tchars: %d\tlines: %d\tredirects: %d\toffset: %d\n",
				res.RequestWord.Word,
				res.RequestWord.Ext,
				res.Response.Response.StatusCode,
				res.Response.NumWords,
				len(res.Response.Body),
				res.Response.NumLines,
				res.NumRedirects,
				res.RequestWord.Offset)
		}

		return
	}
}
