package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type ServerType struct {
	Name    string
	Version string
}

type Cluster struct {
	Svc ServerType
	Lb  LoadBalance
	Que map[string]chan *HttpRequest
	Tls *TlsConfig

	GoNum int
	Stop  chan struct{}
	sync.WaitGroup
	sync.RWMutex
}

func (cluster *Cluster) ClusterProcess(address, proto string) {
	defer cluster.Done()

	httpclient := NewHttpClient(cluster.Svc.Name, proto, cluster.Tls)
	requestque := cluster.Que[address]

	for {
		select {
		case proxyreq := <-requestque:
			{
				proxyrsp := new(HttpRsponse)

				if cluster.Tls != nil {
					proxyreq.url = "https://" + address + proxyreq.url
				} else {
					proxyreq.url = "http://" + address + proxyreq.url
				}

				request, err := http.NewRequest(
					proxyreq.method,
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

				resp, err := httpclient.Do(context.Background(), request)
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
		case <-cluster.Stop:
			{
				log.Println("cluster process exit!")
				return
			}
		}
	}
}

func NewCluster(cfg ClusterConfig) *Cluster {
	lb := NewLB(cfg.LBType, cfg.Endpoint)
	svc := ServerType{Name: cfg.Name, Version: cfg.Version}
	tlscfg := globalconfig.TlsGet(cfg.TlsName)

	cluster := &Cluster{Svc: svc, Lb: lb, Tls: tlscfg}
	cluster.Que = make(map[string]chan *HttpRequest)
	cluster.Stop = make(chan struct{}, 10)

	for _, address := range cfg.Endpoint {
		cluster.Que[address] = make(chan *HttpRequest, 1000)
		for i := 0; i < 10; i++ {
			cluster.Add(1)
			cluster.GoNum++
			go cluster.ClusterProcess(address, cfg.Protocal)
		}
	}
	return cluster
}

func (cluster *Cluster) Close() {
	num := cluster.GoNum
	for i := 0; i < num; i++ {
		cluster.Stop <- struct{}{}
	}
	cluster.Wait()
}

func (cluster *Cluster) Do(req *HttpRequest) *HttpRsponse {
	var proxyrsp *HttpRsponse

	for {
		address := cluster.Lb.Pick()
		requestque := cluster.Que[address]
		requestque <- req

		proxyrsp = <-req.rsp
		if proxyrsp.err != nil {
			log.Println(proxyrsp.err.Error())
			continue
		} else {
			break
		}
	}

	return proxyrsp
}
