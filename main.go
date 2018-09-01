package main

import (
	"bytes"
	"flag"

	"io/ioutil"
	"log"

	"net/http"
	"strings"

	"time"

	"golang.org/x/net/http2"
)

var requestNum int32

func newTransport() http.RoundTripper {
	return &http.Transport{}
}

func newTransport2() http.RoundTripper {
	return &http2.Transport{
		TLSClientConfig: TLS_CONFIG,
	}
}

func newhttpclient() *http.Client {

	var Transport http.RoundTripper

	if TLS_TYPE == "out" {
		Transport = newTransport2()
	} else {
		Transport = newTransport()
	}

	return &http.Client{
		Transport: Transport,
		Timeout:   10 * time.Second,
	}
}

func (h *HttpProxy) Process() {
	defer h.Wait.Done()
	httpclient := newhttpclient()

	for {
		select {
		case proxyreq := <-h.Que:
			{
				proxyrsp := new(HttpRsponse)

				request, err := http.NewRequest(proxyreq.method,
					proxyreq.url,
					bytes.NewBuffer(proxyreq.body))
				if err != nil {
					proxyrsp.err = err
					proxyrsp.status = http.StatusInternalServerError

					proxyreq.rsp <- proxyrsp
					continue
				}

				for key, value := range proxyreq.header {
					for _, v := range value {
						request.Header.Add(key, v)
					}
				}

				resp, err := httpclient.Do(request)
				if err != nil {
					proxyrsp.err = err
					proxyrsp.status = http.StatusInternalServerError

					proxyreq.rsp <- proxyrsp
					continue
				} else {
					proxyrsp.status = resp.StatusCode
					proxyrsp.header = resp.Header
				}

				proxyrsp.body, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					proxyrsp.err = err
					proxyrsp.status = http.StatusInternalServerError

					proxyreq.rsp <- proxyrsp
					continue
				}
				resp.Body.Close()

				proxyreq.rsp <- proxyrsp
			}
		case <-h.Stop:
			{
				return
			}
		}
	}
}

var (
	CONFIG_FILE string
	DEBUG       bool
	HELP        bool
)

func init() {
	flag.StringVar(&CONFIG_FILE, "conf", "config.yaml", "http proxy global config file.")
	flag.BoolVar(&DEBUG, "debug", false, "debug mode.")
	flag.BoolVar(&HELP, "h", false, "this help.")
}

func main() {

	flag.Parse()

	if help || LISTEN_ADDR == "" || REDIRECt_ADDR == "" {
		flag.Usage()
		return
	}

	err := LoadConfig("./config.yaml")
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Printf("Listen   At [%s]\r\n", LISTEN_ADDR)
	log.Printf("Redirect To [%s]\r\n", REDIRECt_ADDR)

	if RUNTIME > 0 {
		log.Printf("RunTime     [%s]Sec\r\n", RUNTIME)
	}

	gServerAddr = strings.Split(REDIRECt_ADDR, ";")

	proxy := NewHttpProxy(LISTEN_ADDR, GetServerAddr)
	if proxy == nil {
		return
	}

	if RUNTIME > 0 {
		for i := 0; i < RUNTIME; i++ {
			time.Sleep(1 * time.Second)
		}
	} else {
		for {
			time.Sleep(1000 * time.Second)
		}
	}
	proxy.Close()
}
