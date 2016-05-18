package compressedhandler

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEncodings(t *testing.T) {
	examples := map[string]codings{

		// Examples from RFC 2616
		"compress, gzip": {gzip: 1.0},
		"":               {},
		"*":              {},
		"compress;q=0.5, gzip;q=1.0":         {gzip: 1.0},
		"gzip;q=1.0, identity; q=0.5, *;q=0": {gzip: 1.0, identity: 0.5},

		// More random stuff
		"AAA;q=1":             {},
		"BBB ; q = 2":         {},
		"gzip, deflate, sdch": {gzip: 1.0, deflate: 1.0},
	}
	for eg, exp := range examples {
		assert.Equal(t, exp, parseEncodings(eg))
	}
}

func TestGzipHandler(t *testing.T) {
	testBody := "aaabbbccc"

	// This just exists to provide something for Handler to wrap.
	handler := Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, testBody)
	}))

	// requests without accept-encoding are passed along as-is

	req1, _ := http.NewRequest("GET", "/whatever", nil)
	res1 := httptest.NewRecorder()
	handler.ServeHTTP(res1, req1)

	assert.Equal(t, 200, res1.Code)
	assert.Equal(t, "", res1.Header().Get("Content-Encoding"))
	assert.Equal(t, testBody, res1.Body.String())

	// but requests with accept-encoding:gzip are compressed if possible

	req2, _ := http.NewRequest("GET", "/whatever", nil)
	req2.Header.Set("Accept-Encoding", "gzip")
	res2 := httptest.NewRecorder()
	handler.ServeHTTP(res2, req2)

	assert.Equal(t, 200, res2.Code)
	assert.Equal(t, "gzip", res2.Header().Get("Content-Encoding"))
	assert.Equal(t, gzipStr(testBody), res2.Body.Bytes())
}

func gzipStr(s string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	io.WriteString(w, s)
	w.Close()
	return b.Bytes()
}
