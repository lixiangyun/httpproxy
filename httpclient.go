package main

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

type HttpRsponse struct {
	status int
	header http.Header
	body   []byte

	err error
}

type HttpRequest struct {
	num    int32
	addr   string
	url    string
	method string
	header http.Header
	body   []byte
	rsp    chan *HttpRsponse
}

type Http11Client struct {
	*http.Client
}

type Http20Client struct {
	*http.Client
}

type HttpClient interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

func NewHttpClient(server string, protc ProtoType, tls TlsConfig) *http.Client {

	var transport http.RoundTripper

	tlsconfig, err := TlsConfigClient(tls, server)
	if err != nil {
		return nil
	}

	if protc == PROTO_HTTP {
		transport = &http.Transport{TLSClientConfig: tlsconfig}
	} else {
		transport = &http2.Transport{TLSClientConfig: tlsconfig}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
}
