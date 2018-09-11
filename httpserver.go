package main

import (
	//	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/http2"
)

type HttpServer struct {
	Name      string
	Address   string
	Protocal  string
	Router    Router
	TlsConfig *tls.Config

	Svc *http.Server

	GoCnt int
	Que   chan *HttpRequest
	Wait  sync.WaitGroup
	Stop  chan struct{}
}

func (h *HttpServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	var err error

	defer req.Body.Close()

	// step 1
	proxyreq := new(HttpRequest)
	proxyreq.num = atomic.AddInt32(&requestNum, 1)
	proxyreq.method = req.Method
	proxyreq.header = req.Header
	proxyreq.rsp = make(chan *HttpRsponse, 1)

	proxyreq.url = req.URL.RequestURI()

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

func NewHttpServer(addr string, protc string, tls TlsConfig) *HttpServer {
	proxy := new(HttpServer)
	proxy.Address = addr

	lis, err := net.Listen("tcp", proxy.Address)
	if err != nil {
		log.Println("http listen failed!", err.Error())
		return nil
	}

	log.Printf("Http Proxy Listen %s\r\n", addr)

	tlsconfig, err := TlsConfigServer(tls)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	proxy.Svc = &http.Server{
		Handler:      proxy,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    tlsconfig}

	if protc == PROTO_HTTP {
		go proxy.Svc.Serve(lis)
	} else {
		if DEBUG {
			http2.VerboseLogs = true
		}
		http2.ConfigureServer(proxy.Svc, &http2.Server{})
		go proxy.Svc.ServeTLS(lis, "", "")
	}

	return proxy
}

func (h *HttpServer) Close() {
	log.Println("Http Proxy Shut Down!", h.Address)
	h.Svc.Close()
	for i := 0; i < h.GoCnt; i++ {
		h.Stop <- struct{}{}
	}
	h.Wait.Wait()
}
