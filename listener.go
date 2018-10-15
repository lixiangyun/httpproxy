package main

import (
	"log"
)

type Listener struct {
	router  string
	httpsvc *HttpServer
}

var globalListAll map[string]*Listener

func init() {
	globalListAll = make(map[string]*Listener, 0)
}

func (l *Listener) ListenerProcess(rep *HttpRequest) *HttpRsponse {

	cluster, err := RouterProcess(l.router, rep)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return ClusterProcess(cluster, rep)
}

func ListenerInit() {

	listennerCfg := globalconfig.listenerGetAll()
	for _, v := range listennerCfg {

		listener := new(Listener)
		listener.router = v.Router

		tlscfg := globalconfig.TlsGet(v.TlsName)
		if tlscfg != nil {
			listener.httpsvc = NewHttpServer(v.Address, v.Protocal, tlscfg)
		} else {
			listener.httpsvc = NewHttpServer(v.Address, v.Protocal, nil)
		}

		if listener.httpsvc == nil {
			log.Fatal("listenner init failed1")
		}

		listener.httpsvc.FuncHandler(listener.ListenerProcess)
	}
}
