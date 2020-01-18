package main

import (
	"fmt"
	"sync"
	"log"
	"bufio"
	"github.com/tehmoon/errors"
	"net/http"
	"os"
	"io/ioutil"
	"github.com/corpix/uarand"
	"regexp"
	"bytes"
	"net/url"
	"time"
)

/* reset terminal:
		- !win: \r\x1b[2K
		- win: \r\r
*/

var (
	ReBody = regexp.MustCompile(`(\r\n|\n){2}((?s).*)`)
	ReHeaders = regexp.MustCompile(`(((.*)(\r\n|\n))*?)(\r\n|\n)`)
	ReRemoveContentLength = regexp.MustCompile(`(\r\n|\n)(Content-Length: .*)`)
	ReAddContentLength = regexp.MustCompile(`(\r\n|\n){2}`)
)

func main() {
	flags, err := parseFlags()
	if err != nil {
		err = errors.Wrap(err, "Error parsing flags")
		panic(err)
	}

	var wcl int = 1
	var wordlistFile *os.File

	if flags.Wordlist != "" {
		wordlistFile, err = os.Open(flags.Wordlist)
		if err != nil {
			err = errors.Wrap(err, "Error opening the wordlist file")
			log.Println(err.Error())
			os.Exit(2)
		}

		wcl, err = countLines(wordlistFile)
		if err != nil {
			err = errors.Wrap(err, "Error counting lines wordlist")
			log.Println(err.Error())
			os.Exit(2)
		}

		// Rewind the file
		_, err = wordlistFile.Seek(0, 0)
		if err != nil {
			err = errors.Wrap(err, "Error rewinding wordlist")
			log.Println(err.Error())
			os.Exit(2)
		}
	}

	if flags.ExtWordlist != "" {
		lines, err := scanLinesFile(flags.ExtWordlist)
		if err != nil {
			err = errors.Wrap(err, "Error processing ext wordlist")
			log.Println(err.Error())
			os.Exit(2)
		}

		flags.ExtWords = append(flags.ExtWords, lines...)
	}

	twc := wcl
	if len(flags.ExtWords) > 0 {
		twc = twc * len(flags.ExtWords)
	} else {
		// Set to empty word because needs to be processed in the main loop
		// "" is not going to trigger anything.
		flags.ExtWords = append(flags.ExtWords, "")
	}

	displayFlags(flags, wcl, twc)

	target, err := url.Parse(flags.Target)
	if err != nil {
		err = errors.Wrapf(err, "Error parsing URL for flag %q", "--target")
		log.Println(err.Error())
		os.Exit(2)
	}

	rawReqFileContent, err := ioutil.ReadFile(flags.File)
	if err != nil {
		err = errors.Wrap(err, "Error reading raw http request file")
		log.Println(err.Error())
		os.Exit(2)
	}

	if flags.TestWord != "" {
		rawReq, err := createRequest(rawReqFileContent, flags.TestWord, flags.TestExtWord)
		if err != nil {
			err = errors.Wrap(err, "Error creating template request")
			log.Println(err.Error())
			os.Exit(2)
		}

		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(rawReq)))
		if err != nil {
			err = errors.Wrap(err, "Error reading the request from the template")
			log.Println(err.Error())
			os.Exit(2)
		}

		if flags.RandUserAgent {
			req.Header.Set("user-agent", uarand.GetRandom())
		}

		buff := bytes.NewBuffer(nil)
		err = req.Write(buff)
		if err != nil {
			err = errors.Wrap(err, "Error writing request to buffer")
			log.Println(err.Error())
			os.Exit(2)
		}

		fmt.Println(string(buff.Bytes()[:]))
		return
	}

	resChan := make(chan *Result)

	doneProcess := &sync.WaitGroup{}
	doneProcess.Add(1)
	go process(flags, resChan, twc, doneProcess)

	if flags.Wordlist == "" {
		config := &SendRequestConfig{
			RequestWord: &RequestWord{
				Word: "",
				Offset: 0,
				Ext: "",
			},
			Target: target,
			Flags: flags,
			ResultChan: resChan,
			FileContent: rawReqFileContent,
		}
		sendRequest(config)
		close(resChan)
		doneProcess.Wait()
		return
	}

	reqChan := make(chan *SendRequestConfig, 0)

	wg := &sync.WaitGroup{}
	for i := 0; i < flags.Routines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			LOOP: for {
				select {
					case config, opened := <- reqChan:
						if ! opened {
							break LOOP
						}
						sendRequest(config)
				}
			}
		}()
	}

	scanner := bufio.NewScanner(wordlistFile)

	var ticker *time.Ticker
	if flags.ThrottleDuration > 0 {
		ticker = time.NewTicker(flags.ThrottleDuration)
		defer ticker.Stop()
	}

	i := 0
	for scanner.Scan() {
		for _, ext := range flags.ExtWords {
			reqChan <- &SendRequestConfig{
				RequestWord: &RequestWord{
					Word: scanner.Text(),
					Offset: i,
					Ext: ext,
				},
				Target: target,
				Flags: flags,
				ResultChan: resChan,
				FileContent: rawReqFileContent,
			}
			if ticker != nil {
				<- ticker.C
			}
		}
		i++
	}
	err = scanner.Err()
	if err != nil {
		err = errors.Wrap(err, "Error reading lines in wordlist file")
		log.Println(err.Error())
	}

	close(reqChan)
	wg.Wait()
	close(resChan)
	doneProcess.Wait()
}
