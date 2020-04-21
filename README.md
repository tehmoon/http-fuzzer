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

### Go-routines and Throttling

The flags `-r` and `--throttle-duration` are used to control concurrent requests and throtteling.

By default, it uses the number of CPUs available for concurrency and no throtteling, which means that if you have 12 CPUs, you'll have a maxium of 12 processing requests at a time. Since there is not throtteling, requests will be spooled as fast as possible.

When throtteling is set, let's say to 100ms. It means that there is going to be a 100ms delay before spooling the next request. As an example, if you have `-r 3 --throttle-duration 100ms` and each request takes 500ms it means that:

```
0ms 1 request
100ms 2 requests
200ms 3 requests
300ms 3 requests
400ms 3 requests
500ms first request is done, 4th request is spooled
600ms second request is done, 5th request is spooled
700ms third request is done, 6th request is spooled
```

As you can see `-r 3 --throttle-duration` means, not more than 3 concurrent requests and at most 1 request every 100ms.

It is really helpful if you want to be stealthy and/or don't want to DOS the server.

### Template

This project is based on using the [go template](https://godoc.org/text). When you set the template file path with `-f`, http-fuzzer will render the template and parse the http request to make sure it is valid.

The request should look like this:
```
HTTP_VERB URI HTTP/1.1
Header: value
Header: value
...

Body
```

```
GET {{ .Word | urlquery }}

```

Notice how you have to have an empty line between the end of the headers and the body. If you don't have any body, you still have to have that empty line.

As for the template, you have two placeholders:
  - `{{ .Word }}` which is the word in the `-w` flag
  - `{{ .Ext }}` which is the word in the `--ext-word` flag

List of functions:
  - Go [text/template](https://godoc.org/text/template#hdr-Functions) package
  - [Sprig](http://masterminds.github.io/sprig/) functions

Use the `urlquery` function in the URI section so you don't have any parsing problem.

When using `--use-random-user-agent`, the `User-Agent` header will be overwritten automatically from the template.

### Ext list

When specifying ext words, the number of total requests are multiplied by the number of ext words. It is useful if you want to fuzz multiple values for one word:

Wordlist:
  - a
  - b
  - c
  - d

Ext words:
  - 1
  - 2

Output:
  - a 1
  - a 2
  - b 1
  - b 2
  - c 1
  - c 2
  - d 1
  - d 2

For example if you want to fuzz the http verb for each words:
```
{{ .Ext | upper }} {{ .Word | urlquery }} HTTP/1.1


```

`-w raft-medium-directories.txt --ext-word get --ext-word put --ext-word post`

### Filtering

Filtering is similar to the `wfuzz` projects. Most of the flags have been replicated for ease of use.

The engine will analyze the response header and body in each thread (go routine) and send the result to a single thread that will decide to output the result or not.

If you don't specify the `-o` flag, a temporary directory will be used so you don't loose your work.

TODO:
  - finish doc
  - fix offset and ext word
The directory where requests are saved is using this hierarchy:
  - `<output_dir>/<
  - `<offset>` 