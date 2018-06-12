package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var HTTP_PROXY string = ":808"

type HttpProxy struct {
	Server string
	Addr   string
}

func (h *HttpProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	fmt.Printf("Received request %s %s %s\n", req.Method, req.Host, req.RemoteAddr)

	fmt.Println(req.URL.Path)

	// step 1
	req.Host = h.Server
	req.RequestURI = "http://" + h.Server + "/" + req.URL.Path

	req.URL, _ = url.Parse(req.RequestURI)

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

	proxy := new(HttpProxy)
	proxy.Addr = HTTP_PROXY
	proxy.Server = "www.baidu.com"

	http.ListenAndServe(proxy.Addr, proxy)
}
