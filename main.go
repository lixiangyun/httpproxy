package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var HTTP_PROXY string = ":808"

type proxy struct{}

// CopyHeaders copy headers from source to destination.
// Nothing would be returned.
func CopyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// ClearHeaders clear headers.
func ClearHeaders(headers http.Header) {
	for key := range headers {
		headers.Del(key)
	}
}

// RmProxyHeaders remove Hop-by-hop headers.
func RmProxyHeaders(req *http.Request) {
	req.RequestURI = ""
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("TE")
	req.Header.Del("Trailers")
	req.Header.Del("Transfer-Encoding")
	req.Header.Del("Upgrade")
}

func (*proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	fmt.Printf("Received request %s %s %s\n", req.Method, req.Host, req.RemoteAddr)

	fmt.Println(req.URL.Path)

	// step 1
	RmProxyHeaders(req)

	servername := req.Header.Get("ServerName")
	if servername != "" {
		req.Header.Del("ServerName")
		fmt.Println("ServerName : ", servername)
		req.Host = servername

		req.RequestURI = "http://" + servername + "/" + req.URL.Path

		req.URL, _ = url.Parse(req.RequestURI)
	}

	// step 2
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {

		fmt.Println(err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)

		return
	}
	defer resp.Body.Close()

	// step 3
	for key, value := range resp.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}

	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, resp.Body)
	resp.Body.Close()
}

func main() {
	http.ListenAndServe(HTTP_PROXY, &proxy{})
}
