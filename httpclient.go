package main

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

type Http11Client struct {
	*http.Client
}

type Http20Client struct {
	*http.Client
}

type HttpClient interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

func newTransport11() http.RoundTripper {
	return &http.Transport{}
}

func newTransport20() http.RoundTripper {
	return &http2.Transport{
		TLSClientConfig: TLS_CONFIG,
	}
}

func newhttpclientxxx() *http.Client {

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