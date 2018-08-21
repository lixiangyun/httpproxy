package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
)

var globalTls map[string]TlsConfig

func init() {
	globalTls = make(map[string]TlsConfig, 0)
}

func TlsConfigUpdate(list []TlsConfig) {
	for _, v := range list {
		_, b := globalTls[v.Name]
		if b == false {
			globalTls[v.Name] = v
		}
	}
}

func TlsConfigClientGet(tls string, servername string) (*tls.Config, error) {

	v, b := globalTls[tls]
	if b == false {
		return nil, errors.New("can't not found tls cfg " + tls)
	}

	return ClientTlsConfig(v.CA, v.Cert, v.Key, servername)
}

func TlsConfigServerGet(tls string) (*tls.Config, error) {

	v, b := globalTls[tls]
	if b == false {
		return nil, errors.New("can't not found tls cfg " + tls)
	}

	return ServerTlsConfig(v.CA, v.Cert, v.Key)
}

func ClientTlsConfig(ca, cert, key string, addr string) (*tls.Config, error) {

	//服务端证书池
	var pool *x509.CertPool

	if ca != "" {
		//这里读取的是根证书
		buf, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		pool = x509.NewCertPool()
		pool.AppendCertsFromPEM(buf)
	}

	//加载客户端证书
	//这里加载的是服务端签发的
	crt, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	var bSkipVerify bool

	// 如果没有配置服务端根证书，则忽略校验服务端证书有效性。
	if pool == nil {
		bSkipVerify = true
	}

	return &tls.Config{
		ServerName:         addr,
		InsecureSkipVerify: bSkipVerify,
		RootCAs:            pool,
		Certificates:       []tls.Certificate{crt},
	}, nil
}

func ServerTlsConfig(ca, cert, key string) (*tls.Config, error) {

	var pool *x509.CertPool

	if ca != "" {
		//这里读取的是根证书
		buf, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, err
		}
		pool = x509.NewCertPool()
		pool.AppendCertsFromPEM(buf)
	}

	//加载服务端证书
	crt, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	var authtype tls.ClientAuthType

	if pool != nil {
		authtype = tls.RequireAndVerifyClientCert
	} else {
		authtype = tls.RequireAnyClientCert
	}

	return &tls.Config{
		Certificates: []tls.Certificate{crt},
		ClientAuth:   authtype,
		ClientCAs:    pool,
	}, nil
}
