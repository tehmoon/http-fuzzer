package main

import (
	"net/url"
	"io"
	"github.com/tehmoon/errors"
	"bufio"
	"bytes"
	"path"
	"net/http"
	"github.com/Masterminds/sprig"
	"crypto/tls"
	"io/ioutil"
	"text/template"
	"log"
	"fmt"
	"github.com/corpix/uarand"
	"os"
	"path/filepath"
	"strconv"
)

type RequestWord struct {
	Word string
	Offset int
	Ext string
}

func sendRequest(config *SendRequestConfig) {
	target, flags := config.Target, config.Flags
	rw, resChan := config.RequestWord, config.ResultChan
	rawReqFileContent := config.FileContent

	rawReq, err := createRequest(rawReqFileContent, rw.Word, rw.Ext)
	if err != nil {
		resChan <- &Result{
			RequestWord: rw,
			Err: errors.Wrap(err, "Error create request template"),
		}
		return
	}

	pr2, pw2 := io.Pipe()
	rawReqReader := bytes.NewReader(rawReq)
	teeReqReader := io.TeeReader(rawReqReader, pw2)

	debug := make(chan []byte)

	go func() {
		reader := bufio.NewReader(pr2)
		req, err := http.ReadRequest(reader)
		if err != nil {
			log.Println(err.Error())
			close(debug)
			return
		}

		if flags.RandUserAgent {
			req.Header.Set("user-agent", uarand.GetRandom())
		}

		req.RequestURI = ""
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		buff := bytes.NewBuffer(nil)
		req.Write(buff)
		debug <- buff.Bytes()
		close(debug)
	}()

	buff := bufio.NewReader(teeReqReader)
	req, err := http.ReadRequest(buff)
	if err != nil {
		resChan <- &Result{
			RequestWord: rw,
			Err: errors.Wrap(err, "Error reading request"),
		}
		return
	}

	var proxy func(*http.Request) (*url.URL, error)
	proxy = http.ProxyFromEnvironment

	if flags.Proxy != nil {
		proxy = http.ProxyURL(flags.Proxy)
	}

	numRedirects := 0
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			numRedirects = len(via)
			if len(via) <= flags.MaxRedirects && flags.MaxRedirects > 0 {
				return nil
			}

			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			Proxy: proxy,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: flags.Insecure,
			},
		},
	}

	if flags.RandUserAgent {
		req.Header.Set("user-agent", uarand.GetRandom())
	}

	req.RequestURI = ""
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host

	if p, err := proxy(req); p != nil && err == nil && req.URL.Scheme == "http" {
		req.URL.Opaque = fmt.Sprintf("%s://%s", req.URL.Scheme, path.Join(req.URL.Host, req.URL.Path))
	}

	resp, err := client.Do(req)
	if err != nil {
		resChan <- &Result{
			RequestWord: rw,
			Err: errors.Wrap(err, "Error sending the request"),
		}
		return
	}

	if numRedirects != 0 && flags.MaxRedirects >= 0 {
		// Remove 1 because via is already set to 1 when the function executes
		numRedirects = numRedirects - 1
	}

	debugReq, open := <- debug
	if ! open {
		debugReq = make([]byte, 0)
	}

	rawResponseBuffer := bytes.NewBuffer(nil)
	resp.Write(rawResponseBuffer)
	rawResponse := rawResponseBuffer.Bytes()

	responseReBody := ReBody.FindStringSubmatch(string(rawResponse[:]))
	responseBody := make([]byte, 0)

	if len(responseReBody) > 1 {
		responseBody = []byte(responseReBody[2])
	}

	cwr := bytes.NewReader(responseBody)
	cw, e := countWords(cwr)
	if e != nil {
		err = errors.Wrap(e, "Error counting words from the body")
	}

	clr := bytes.NewReader(responseBody)
	cl, e := countLines(clr)
	if e != nil {
		err = errors.Wrap(e, "Error counting lines from the body")
	}

	rr := &ResultResponse{
		Response: resp,
		Raw: rawResponse,
		Body: responseBody,
		NumLines: cl,
		NumWords: cw,
	}

	tempdir := filepath.Join(flags.OutputDir, strconv.FormatInt(int64(rw.Offset), 10))
	err = os.Mkdir(tempdir, 0755)
	if err != nil {
		e = errors.Wrap(err, "Error creating the directory inside of the output directory")
	}

	reqOutputFile := filepath.Join(tempdir, "request.http")
	respOutputFile := filepath.Join(tempdir, "response.http")
	wordOutputFile := filepath.Join(tempdir, "word.txt")
	extOutputFile := filepath.Join(tempdir, "ext.txt")
	errOutputFile := filepath.Join(tempdir, "error.txt")

	err = ioutil.WriteFile(reqOutputFile, debugReq, 0644)
	if err != nil {
		e = errors.Wrap(err, "Error writing raw request to output directory")
	}

	err = ioutil.WriteFile(respOutputFile, rawResponse, 0644)
	if err != nil {
		e = errors.Wrap(err, "Error writing raw response to output directory")
	}

	err = ioutil.WriteFile(wordOutputFile, []byte(rw.Word), 0644)
	if err != nil {
		e = errors.Wrap(err, "Error writing word to output directory")
	}

	if rw.Ext != "" {
		err = ioutil.WriteFile(extOutputFile, []byte(rw.Ext), 0644)
		if err != nil {
			e = errors.Wrap(err, "Error writing ext word to output directory")
		}
	}

	if e != nil {
		err = ioutil.WriteFile(errOutputFile, []byte(e.Error()), 0644)
		if err != nil {
			e = errors.Wrap(err, "Error writing word to output directory")
		}
	}

	resChan <- &Result{
		RawRequest: debugReq,
		Err: err,
		RequestWord: rw,
		Response: rr,
		NumRedirects: numRedirects,
	}
}

type TemplateRoot struct {
	Word string
	Ext string
}

func createRequest(rawReqFileContent []byte, word, ext string) (rawReq []byte, err error) {
	pr, pw := io.Pipe()
	tpl := template.Must(template.New("base").Funcs(sprig.TxtFuncMap()).Parse(string(rawReqFileContent[:])))

	go func() {
		pw.CloseWithError(tpl.Templates()[0].Execute(pw, TemplateRoot{Word: word, Ext: ext,}))
	}()

	rawReq, err = ioutil.ReadAll(pr)
	if err != nil {
		err = errors.Wrap(err, "Unable to read template")
		return nil, err
	}

	rawReHeaders := ReHeaders.FindStringSubmatch(string(rawReq[:]))
	if len(rawReHeaders) < 2 {
		return nil, errors.New("No header found in the request!")
	}

	rawHeaders := []byte(rawReHeaders[1])

	rawReBody := ReBody.FindStringSubmatch(string(rawReq[:]))
	var body []byte
	if len(rawReBody) > 0 {
		if len(rawReBody[0]) > 2 {
			body = []byte(rawReBody[2])
		}
	}

	contentLength := len(body)

	rawHeaders = []byte(ReRemoveContentLength.ReplaceAllString(string(rawHeaders[:]), ``))

	buffer := bytes.NewBuffer(nil)
	buffer.Write(rawHeaders)
	fmt.Fprintf(buffer, "Content-Length: %d\r\n\r\n", contentLength)
	buffer.Write(body)

	return buffer.Bytes(), nil
}

type SendRequestConfig struct {
	FileContent []byte
	Flags *Flags
	Target *url.URL
	RequestWord *RequestWord
	ResultChan chan *Result
}
