package main

import (
	"context"
	"crypto/tls"
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
	url    string
	method string
	header http.Header
	body   []byte
	rsp    chan *HttpRsponse
}

type HttpClient struct {
	cli *http.Client
}

func (h *HttpClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return h.cli.Do(req)
}

func NewHttpClient(server string, protc string, cfg *TlsConfig) *HttpClient {

	var transport http.RoundTripper
	var tlsconfig *tls.Config
	var err error

	if cfg != nil {
		tlsconfig, err = TlsConfigClient(cfg, server)
		if err != nil {
			return nil
		}
	}

	if protc == PROTO_HTTP {
		transport = &http.Transport{TLSClientConfig: tlsconfig}
	} else {
		transport = &http2.Transport{TLSClientConfig: tlsconfig}
	}

	cli := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	return &HttpClient{cli: cli}
}
