// http.go
package main

/* Manage project's HTTP connection including
- timeout management
- log
*/

import (
	//	"github.com/simulot/scantopc/sm"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}
func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}

func HttpGetWithTimeout(url string, headers *map[string]string, connTimeout, respTimeout time.Duration) (resp *http.Response, err error) {
	defer Un(Trace("HttpGetWithTimeout", url))

	client := NewTimeoutClient(connTimeout, respTimeout)
	req, err := http.NewRequest("GET", url, nil)
	if headers != nil && len(*headers) > 0 {
		for key, val := range *headers {
			req.Header.Add(key, val)
		}
	}
	resp, err = client.Do(req)

	if flagTraceHTTP > 0 {
		if err != nil {
			ERROR.Println("http.Get(", url, ") ->", err)
			return
		} else {
			TRACE.Println("http.Get(", url, ") ->", resp.Status, err)
			if flagTraceHTTP > 1 {
				for h, v := range resp.Header {
					TRACE.Println("\t", h, " --> ", v)
				}
			}
		}
	}
	return
}

func HttpGet(url string) (resp *http.Response, err error) {
	defer Un(Trace("HttpGet"))
	resp, err = HttpGetWithTimeout(url, nil, 500*time.Millisecond, 1*time.Second)
	return
}

func HttpPost(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	defer Un(Trace("HttpPost", url, bodyType))

	resp, err = http.Post(url, bodyType, body)
	if flagTraceHTTP > 0 {
		TRACE.Println("http.Post(", url, ",", bodyType, ") ->", resp.Status, err)
		if flagTraceHTTP > 1 {
			for h, v := range resp.Header {
				TRACE.Println("\t", h, " --> ", v)
			}
		}
	}
	return
}

func ioutilReadAll(r io.Reader) (ByteArray []byte, err error) {
	ByteArray, err = ioutil.ReadAll(r)
	if flagTraceHTTP > 1 {
		TRACE.Println("\tResponse:\n", string(ByteArray))
	}
	return
}
