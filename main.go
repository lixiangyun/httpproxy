package main

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/http2"
)

type SELECT_ADDR func() string

type HttpProxy struct {
	Fun  SELECT_ADDR
	Addr string
	Svc  *http.Server

	GoCnt int
	Que   chan *HttpRequest
	Wait  sync.WaitGroup
	Stop  chan struct{}
}

type HttpRsponse struct {
	status int
	header http.Header
	body   []byte

	err error
}

var requestNum int32

type HttpRequest struct {
	num    int32
	addr   string
	url    string
	method string
	header http.Header
	body   []byte
	rsp    chan *HttpRsponse
}

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

func (h *HttpProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	var err error

	defer req.Body.Close()

	redirect := h.Fun()

	// step 1
	proxyreq := new(HttpRequest)
	proxyreq.num = atomic.AddInt32(&requestNum, 1)
	proxyreq.addr = redirect
	proxyreq.method = req.Method
	proxyreq.header = req.Header
	proxyreq.rsp = make(chan *HttpRsponse, 1)

	if TLS_TYPE == "out" {
		proxyreq.url = "https://" + redirect + req.URL.RequestURI()
	} else {
		proxyreq.url = "http://" + redirect + req.URL.RequestURI()
	}

	proxyreq.body, err = ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if DEBUG {
		headers := fmt.Sprintf("\r\nHeader:\r\n")
		for key, value := range proxyreq.header {
			headers += fmt.Sprintf("\t%s:%v\r\n", key, value)
		}
		var body string
		if len(proxyreq.body) > 0 {
			body = fmt.Sprintf("Body:%s\r\n", string(proxyreq.body))
		}
		log.Printf("[%d]Request Method:%s\r\nURL:%s%s%s\r\n",
			proxyreq.num, proxyreq.method, proxyreq.url, headers, body)
	}

	h.Que <- proxyreq
	proxyrsp := <-proxyreq.rsp

	if DEBUG {
		headers := fmt.Sprintf("\r\nHeader:\r\n")
		for key, value := range proxyrsp.header {
			headers += fmt.Sprintf("\t%s:%v\r\n", key, value)
		}
		var body string
		if len(proxyrsp.body) > 0 {
			body = fmt.Sprintf("Body:%s\r\n", string(proxyrsp.body))
		}
		log.Printf("[%d]Response Code:%d%s%s\r\n",
			proxyreq.num, proxyrsp.status, headers, body)
	}

	// step 2
	if proxyrsp.err != nil {
		log.Println(proxyrsp.err.Error())
		http.Error(rw, proxyrsp.err.Error(), http.StatusInternalServerError)
		return
	}

	// step 3
	for key, value := range proxyrsp.header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}

	rw.WriteHeader(proxyrsp.status)
	rw.Write(proxyrsp.body)
}

func NewHttpProxy(addr string, fun SELECT_ADDR) *HttpProxy {
	proxy := new(HttpProxy)

	proxy.Addr = addr
	proxy.Fun = fun

	lis, err := net.Listen("tcp", proxy.Addr)
	if err != nil {
		log.Println("http listen failed!", err.Error())
		return nil
	}

	log.Printf("Http Proxy Listen %s\r\n", addr)

	if TLS_TYPE == "in" {
		proxy.Svc = &http.Server{
			Handler:      proxy,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			TLSConfig:    TLS_CONFIG}
	} else {
		proxy.Svc = &http.Server{
			Handler:      proxy,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second}
	}

	proxy.GoCnt = 100
	proxy.Que = make(chan *HttpRequest, 1000)
	proxy.Stop = make(chan struct{}, proxy.GoCnt)

	proxy.Wait.Add(proxy.GoCnt)
	for i := 0; i < proxy.GoCnt; i++ {
		go proxy.Process()
	}

	if TLS_TYPE == "in" {
		if DEBUG {
			http2.VerboseLogs = true
		}
		http2.ConfigureServer(proxy.Svc, &http2.Server{})
		go proxy.Svc.ServeTLS(lis, "", "")
	} else {
		go proxy.Svc.Serve(lis)
	}

	return proxy
}

func (h *HttpProxy) Close() {
	log.Println("Http Proxy Shut Down!", h.Addr)
	h.Svc.Close()
	for i := 0; i < h.GoCnt; i++ {
		h.Stop <- struct{}{}
	}
	h.Wait.Wait()
}

var gServerAddr []string
var gIndex int

func GetServerAddr() string {
	idx := gIndex % len(gServerAddr)
	gIndex++
	return gServerAddr[idx]
}

var (
	TLS_CA_FILE   string
	TLS_CERT_FILE string
	TLS_KEY_FILE  string
	TLS_TYPE      string

	LISTEN_ADDR   string
	REDIRECt_ADDR string

	RUNTIME int
	DEBUG   bool
	help    bool

	TLS_CONFIG *tls.Config
)

func init() {
	flag.StringVar(&TLS_CA_FILE, "ca", "", "CA certificate to verify peer against.")
	flag.StringVar(&TLS_CERT_FILE, "cert", "", "certificate file.if [-tls] option been seted then required.")
	flag.StringVar(&TLS_KEY_FILE, "key", "", "private key file name.if [-tls] option been seted then required.")
	flag.StringVar(&TLS_TYPE, "tls", "", "the proxy channel enable https.[in/out]")

	flag.StringVar(&LISTEN_ADDR, "in", "", "listen addr for http/https proxy.")
	flag.StringVar(&REDIRECt_ADDR, "out", "", "redirect to addr for http/https proxy.")
	flag.IntVar(&RUNTIME, "time", 0, "http proxy run time.")
	flag.BoolVar(&DEBUG, "debug", false, "debug mode.")
	flag.BoolVar(&help, "h", false, "this help.")
}

func main() {

	flag.Parse()

	if help || LISTEN_ADDR == "" || REDIRECt_ADDR == "" {
		flag.Usage()
		return
	}

	if TLS_TYPE != "" && TLS_TYPE != "in" && TLS_TYPE != "out" {
		flag.Usage()
		return
	}

	if TLS_TYPE != "" && (TLS_CERT_FILE == "" || TLS_KEY_FILE == "") {
		flag.Usage()
		return
	}

	var err error

	if TLS_TYPE == "in" {
		TLS_CONFIG, err = ServerTlsConfig(TLS_CA_FILE, TLS_CERT_FILE, TLS_KEY_FILE)
	} else if TLS_TYPE == "out" {
		addlist := strings.Split(REDIRECt_ADDR, ":")
		TLS_CONFIG, err = ClientTlsConfig(TLS_CA_FILE, TLS_CERT_FILE, TLS_KEY_FILE, addlist[0])
	}

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if TLS_CONFIG != nil {
		log.Println("Tls Enable : ", TLS_TYPE)
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
