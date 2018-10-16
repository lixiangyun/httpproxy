package main

import (
	"errors"
	"log"
)

type Router struct {
	name          string
	url_match     Match
	header_match  map[string]Match
	cluser_select *ClusterSelect
	timeout       int
}

var globalRouterAll map[string]*Router

func NewRouter(cfg RouterConfig) *Router {
	router := new(Router)

	router.name = cfg.Name
	router.url_match = NewMatch(cfg.Type, cfg.Value)

	router.header_match = make(map[string]Match, 0)
	for _, v := range cfg.Headers {
		router.header_match[v.Key] = NewMatch(v.Type, v.Value)
	}

	items := make([]ClusterWeigth, len(cfg.Clusters))
	for i, v := range cfg.Clusters {
		items[i].svc = ServerType{Name: v.Name, Version: v.Version}
		items[i].weight = v.Weight
	}

	router.cluser_select = NewClusterWeight(items)
	router.timeout = cfg.Timeout

	return router
}

func (r *Router) Process(req *HttpRequest) *ServerType {
	if !r.url_match.Do(req.url) {
		return nil
	}

	if len(r.header_match) == 0 {
		svc := r.cluser_select.Select()
		return &svc
	}

	for req_key, req_value := range req.header {
		match, b := r.header_match[req_key]
		if b == true {
			for _, value := range req_value {
				match.Do(value)
			}
		}
	}

	return nil
}

func RouterInit() {
	routerCfg := globalconfig.RouterGetAll()
	for _, v := range routerCfg {
		router := NewRouter(v)
		if router == nil {
			log.Fatal("router init failed!")
		}
		globalRouterAll[v.Name] = router
	}
}

func RouterProcess(name string, req *HttpRequest) (*ServerType, error) {
	router, b := globalRouterAll[name]
	if b == false {
		return nil, errors.New("router not found!" + name)
	}
	return router.Process(req), nil
}
