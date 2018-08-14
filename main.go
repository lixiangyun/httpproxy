package main

import (
	"bytes"
	"flag"
	"fmt"

	"io/ioutil"
	"log"
	"net"
	"net/http"

	"strings"
	"sync"
	"time"
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

type HttpRequest struct {
	addr   string
	url    string
	method string
	header http.Header
	body   []byte
	rsp    chan *HttpRsponse
}

func newTransport() http.RoundTripper {
	return &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConnsPerHost:   100,
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func newhttpclient() *http.Client {
	return &http.Client{
		Transport: newTransport(),
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

	if DEBUG {
		log.Printf("Request : %v \r\n", req)
	}

	// step 1
	proxyreq := new(HttpRequest)
	proxyreq.addr = redirect
	proxyreq.url = "http://" + redirect + req.URL.RequestURI()
	proxyreq.method = req.Method
	proxyreq.header = req.Header
	proxyreq.rsp = make(chan *HttpRsponse, 1)

	proxyreq.body, err = ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println(err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if DEBUG {
		log.Printf("RequestProxy : %v \r\n", proxyreq)
	}

	h.Que <- proxyreq
	proxyrsp := <-proxyreq.rsp

	if DEBUG {
		log.Printf("nResponse : %v \r\n", proxyrsp)
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

	proxy.Svc = &http.Server{Handler: proxy}

	proxy.GoCnt = 100
	proxy.Que = make(chan *HttpRequest, 1000)
	proxy.Stop = make(chan struct{}, proxy.GoCnt)

	proxy.Wait.Add(proxy.GoCnt)
	for i := 0; i < proxy.GoCnt; i++ {
		go proxy.Process()
	}

	go proxy.Svc.Serve(lis)

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
	LISTEN_ADDR   string
	REDIRECt_ADDR string
	RUNTIME       int
	DEBUG         bool
	help          bool
)

func init() {
	flag.StringVar(&LISTEN_ADDR, "in", "", "listen addr by http proxy.")
	flag.StringVar(&REDIRECt_ADDR, "out", "", "http proxy redirect to addr.")
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

	fmt.Printf("Listen   At [%s]\r\n", LISTEN_ADDR)
	fmt.Printf("Redirect To [%s]\r\n", REDIRECt_ADDR)

	if RUNTIME > 0 {
		fmt.Printf("RunTime     [%s]Sec\r\n", RUNTIME)
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
