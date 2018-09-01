package main

import (
	"crypto/tls"
	"net/http"
	"sync"
)

type HttpServer struct {
	Name      string
	Address   string
	Protocal  ProtoType
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

func NewHttpServer(Listerner []ListernerConfig) *HttpServer {
	proxy := new(HttpServer)

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
