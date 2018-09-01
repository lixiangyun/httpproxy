package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ProtoType string

type ListernerConfig struct {
	Name     string    `yaml:"name"`
	Address  string    `yaml:address`
	Protocal ProtoType `yaml:protocol`
	Router   string    `yaml:router`
	TlsName  string    `yaml:tls`
}

type MatchType string

type Header struct {
	Key   string    `yaml:"key"`
	Value string    `yaml:"value"`
	Type  MatchType `yaml:"type"`
}

type DestCluster struct {
	Name    string `yaml:"servername"`
	Version string `yaml:"version"`
	Weight  int    `yaml:"weight"`
}

type RetryPolicy struct {
	Enable  bool `yaml:"retry_on"`
	Times   int  `yaml:"num_retries"`
	Timeout int  `yaml:"try_timeout"`
}

type LoadBalanceType string

type ClusterConfig struct {
	Name     string          `yaml:"servername"`
	Version  string          `yaml:"version"`
	Endpoint []string        `yaml:"endpoints"`
	Protocal ProtoType       `yaml:"protocol"`
	TlsName  string          `yaml:"tls"`
	LBType   LoadBalanceType `yaml:"loadbalance"`
}

type TlsConfig struct {
	Name string `yaml:"name"`
	Cert string `yaml:"cert_chain_file"`
	Key  string `yaml:"private_key_file"`
	CA   string `yaml:"ca_cert_file"`
}

type RouterType string

type RouterConfig struct {
	Name  string     `yaml:"name"`
	Prior int        `yaml:"priority"`
	Type  RouterType `yaml:"rule_type"`
	Value string     `yaml:"rule_value"`

	Headers  []Header      `yaml:"headers"`
	Clusters []DestCluster `yaml:"clusters"`

	Timeout int         `yaml:"timeout"`
	Retry   RetryPolicy `yaml:"retry_policy"`
}

type ConfigV1 struct {
	Listeners []ListernerConfig `yaml:"listeners"`
	Routers   []RouterConfig    `yaml:"router"`
	TlsCfg    []TlsConfig       `yaml:"tls"`
	Clusters  []ClusterConfig   `yaml:"clusters"`
}

var globalConfigV1 ConfigV1

func LoadConfig(filename string) error {

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(body, &globalConfigV1)
	if err != nil {
		return err
	}

	return nil
}

func init() {

	globalConfigV1.Listeners = make([]Listerner, 0)
	globalConfigV1.Routers = make([]Router, 0)
	globalConfigV1.Clusters = make([]ClusterDS, 0)
	globalConfigV1.TlsCfg = make([]TlsConfig, 0)

	//log.Println(globalConfigV1)
}
