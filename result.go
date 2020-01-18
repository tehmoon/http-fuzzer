package main

type Result struct {
	Err error
	RawRequest []byte
	Response *ResultResponse
	RequestWord *RequestWord
	NumRedirects int
}
