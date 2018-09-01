package main

type ServerType struct {
	Name    string
	Version string
}

type ClusterCtl struct {
	Svc ServerType
	Add []string
	Lb  LoadBalance
	Tls TlsConfig
}

func ClusterCfgUpdate(list []ClusterConfig) {

}
