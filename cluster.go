package main

import (
	"sync"
)

type ServerType struct {
	Name    string
	Version string
}

type Cluster struct {
	Svc ServerType
	Lb  LoadBalance
	Cli map[string]*HttpClient

	sync.WaitGroup
	sync.RWMutex
}

func NewCluster(cfg ClusterConfig) *Cluster {

	lb := NewLB(cfg.LBType, cfg.Endpoint)
	svc := ServerType{Name: cfg.Name, Version: cfg.Version}
	tlscfg := globalconfig.TlsGet(cfg.TlsName)
	cli := make(map[string]*HttpClient, len(cfg.Endpoint))

	for _, v := range cfg.Endpoint {
		cli[v] = NewHttpClient(cfg.Name, cfg.Protocal, tlscfg)
	}

	return &Cluster{Svc: svc, Lb: lb, Cli: cli}
}

func (c *Cluster) Close() {

}

func (c *Cluster) Do(req *HttpRequest) *HttpRsponse {

	return nil
}
