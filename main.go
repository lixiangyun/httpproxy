package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lixiangyun/go-mesh/mesher/comm"
)

type SELECT_ADDR func() string

type HttpProxy struct {
	Fun  SELECT_ADDR
	Addr string
	Svc  *http.Server
}

func (h *HttpProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	defer req.Body.Close()

	redirect := h.Fun()

	// step 1
	path := "http://" + redirect + "/" + req.URL.Path

	request, err := http.NewRequest(req.Method, path, req.Body)
	if err != nil {
		log.Println(err.Error())
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	for key, value := range req.Header {
		for _, v := range value {
			request.Header.Add(key, v)
		}
	}

	// step 2
	resp, err := comm.HttpClient.Do(request)
	if err != nil {
		log.Println(err.Error())
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

	log.Printf("Http Proxy Listen : %s\r\n", addr)

	proxy.Svc = &http.Server{Handler: proxy}

	go proxy.Svc.Serve(lis)

	return proxy
}

func (h *HttpProxy) Close() {
	h.Svc.Close()
}

var gServerAddr []string
var gIndex int

func GetServerAddr() string {
	idx := gIndex % len(gServerAddr)
	gIndex++
	return gServerAddr[idx]
}

func main() {

	args := os.Args

	if len(args) != 3 {
		fmt.Println("usage : <Listen Addr> <Redirect Addr>")
		return
	}

	fmt.Printf("Listen   At [%s]\r\n", args[1])
	fmt.Printf("Redirect To [%s]\r\n", args[2])

	gServerAddr = strings.Split(args[2], ";")

	proxy := NewHttpProxy(args[1], GetServerAddr)
	if proxy == nil {
		return
	}

	for {
		time.Sleep(1 * time.Second)
	}

	proxy.Close()
}
