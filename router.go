package main

type Router struct {
	name string
}

var globalRouterAll map[string]*Router

func (r *Router) Process(req *HttpRequest) string {

	return ""
}

func RouterInit() {
	routerCfg := globalconfig.RouterGetAll()

	for _, v := range routerCfg {

		router := new(Router)

		globalRouterAll[v.Name] = router
	}
}

func RouterProcess(name string, req *HttpRequest) (string, error) {

	return "", nil
}
