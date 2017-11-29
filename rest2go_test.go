package main

import (
	"testing"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"bytes"
)

func mockServe(string, http.Handler) error {
	return nil
}

func mockServeTls(string, string, string, http.Handler) error {
	return nil
}

var r = 5

func mockPsName(ps httprouter.Params) string {
	if r == 5 {
		r = 4
		return "debug"
	}
	if r == 4 {
		r = 3
		return "info"
	}
	if r == 3 {
		r = 2
		return "warn"
	}
	if r == 2 {
		r = 1
		return "error"
	}
	if r == 1 {
		r = 5
		return "critical"
	}
	return "false"
}

type ClosingBuffer struct {
	*bytes.Buffer
}
func (cb *ClosingBuffer) Close() (err error) {
	//we don't actually have to do anything here, since the buffer is just some data in memory
	//and the error is initialized to no-error
	return
}

func TestMainCode(t *testing.T) {
	httpServe = mockServe
	httpServeTls = mockServeTls
	main()
}

func TestHandler(t *testing.T) {
	w := new(http.ResponseWriter)
	r := new(http.Request)
	ps := new(httprouter.Params)
	psNameCall = mockPsName

	fakeRun=true

	cb := &ClosingBuffer{bytes.NewBufferString("{\"Message\":\"Some\\r\\nLogging Message\\nTesting\\nAlll\\nLogging modes\"}")}
	r.Body = cb
	logLine(*w,r,*ps)

	cb = &ClosingBuffer{bytes.NewBufferString("{\"Message\":\"Some Simple log line\"}")}
	r.Body = cb
	logLine(*w,r,*ps)

	cb = &ClosingBuffer{bytes.NewBufferString("{\"Message\":\"A very long line 123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890\"}")}
	r.Body = cb
	logLine(*w,r,*ps)
}