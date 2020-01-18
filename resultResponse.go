package main

import (
	"net/http"
)

type ResultResponse struct {
	NumWords int
	Response *http.Response
	NumLines int
	Body []byte
	Raw []byte
}
