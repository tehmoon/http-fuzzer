## HTTP-FUZZER

This project was born out of one idea that I had: make a fuzzer like wfuzz and gobuster which is burp-compatible, but also being able to fuzz some random parameters in the raw http request.

In order to do this, Go has everything. It is able to parse raw http request, and we can use the `text/template` engine in order to inject our variables. As a side effect using the template engine, it's also using [sprig](https://github.com/Masterminds/sprig) to enrich then engine.

This project features:
- Burp compatible
- Easy templating with Go template package
- Multi-threaded
- Multi architecture support
- Save both templated request AND response to temporary directory every time
- Wfuzz API compatible
- Use random user agent on EACH query to cover your tracks
- Use alternative wordlist to be processed for each word in the wordlist
- Follow specified number of redirects
- Throttling feature to avoid being banned

### Install

You will need to install go first. Once this is done:

`go get -v -u github.com/tehmoon/http-fuzzer`

### How to use it

Simply create a new file and paste the raw request from burp.

Caution: some editor will add a new line at the end of the file when saving. Make sure your file is exactly the same as burp. On `vim` you can do `set noeol` to remove the last end of line when saving.

Then start by testing the template:
```
http-fuzzer -f <your http file> --test-word testing -t <your burp target>
```

If you are satifsfied with the output, you can start the fuzzing with:
```
http-fuzzer -f <your http file> -w <your wordlist file> -t <your burp target>
```

It will automatically save all successful requests to a temporary request

### Usage

```
Usage of ./http-fuzzer:
  -x, --ext stringArray              Words to be process for each word from the main wordlist.
      --ext-list string              Wordlist to be process for each word from the main wordlist.
  -f, --file string                  Specify raw http request file as input
      --hc stringArray               Hide responses with specified status code
      --hh stringArray               Hide responses with specified number of characters
      --hl stringArray               Hide responses with specified number of lines
      --hnw stringArray              Hide if number of words matches
      --hr stringArray               Hide responses with specifed regex from the body
  -k, --insecure                     Bypass TLS checks
      --max-redirects int            Maximum number of follow http redirect
  -o, --output-dir string            Dump matched requests and responses into a directory
  -p, --proxy string                 Send the requests through a http proxy
  -r, --routines int                 Max parallel running go-routines (default 12)
      --sc stringArray               Show responses with specified status code (default [200,204,301,302,307])
      --sh stringArray               Show responses with specified number of characters
      --sl stringArray               Show responses with specified number of lines
      --snw stringArray              Show if number of words matches
      --sr stringArray               Show responses with specifed regex from the body
  -t, --target string                Set the target URL
      --test-ext-word string         Show the request when --test-word is used as the .Ext object for the template
      --test-word string             Show the request to stdout using the word as the .Word object for the template
  -d, --throttle-duration duration   Throttle sending requests
      --use-random-user-agent        Force using a random user agent for each request
  -v, --verbose                      Output all request as if they all matched
  -w, --wordlist string              Use worldlist items as root object for the template
pflag: help requested
```

### Throttling
### Template
### Ext list
### Filtering