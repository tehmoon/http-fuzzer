package main

import (
	pb "github.com/cheggaaa/pb/v3"
	"sync"
	"os"
	"log"
)

func processShow(res *Result, flags *Flags) bool {
	if findIntArray(flags.ShowCodes, res.Response.Response.StatusCode) {
		return true
	}

	if findIntArray(flags.ShowNumWords, res.Response.NumWords) {
		return true
	}

	if findIntArray(flags.ShowChars, len(res.Response.Body)) {
		return true
	}

	if findIntArray(flags.ShowLines, res.Response.NumLines) {
		return true
	}

	if matchReInBytes(flags.ShowRegexes, res.Response.Body) {
		return true
	}

	return false
}

func processHide(res *Result, flags *Flags) bool {
	if findIntArray(flags.HideCodes, res.Response.Response.StatusCode) {
		return true
	}

	if findIntArray(flags.HideNumWords, res.Response.NumWords) {
		return true
	}

	if findIntArray(flags.HideChars, len(res.Response.Body)) {
		return true
	}

	if findIntArray(flags.HideLines, res.Response.NumLines) {
		return true
	}

	if matchReInBytes(flags.HideRegexes, res.Response.Body) {
		return true
	}

	return false
}

func process(flags *Flags, c chan *Result, totalLines int, done *sync.WaitGroup) {
	defer done.Done()

	bar := pb.New(totalLines)
	if ! flags.Verbose {
		bar.SetWriter(os.Stderr)
		bar = bar.Start()
	}

	show := false
	for res := range c {
		if res.Err != nil {
			log.Printf("Word %q had error: %q\n", res.RequestWord.Word, res.Err.Error())
			bar.Increment()
			continue
		}

		show = processShow(res, flags)
		show = ! processHide(res, flags) && show

		display(res, show, flags)

		show = false
		if ! flags.Verbose {
			bar.Increment()
		}
	}

	if ! flags.Verbose {
		bar.Finish()
	}
}
