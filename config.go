package main

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	PROTO_HTTP  = "http"
	PROTO_HTTP2 = "http2"
	PROTO_AUTO  = "auto"
)

type ListernerConfig struct {
	Name     string `yaml:"name"`
	Address  string `yaml:address`
	Protocal string `yaml:protocol`
	Router   string `yaml:router`
	TlsName  string `yaml:tls`
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

type ClusterConfig struct {
	Name     string   `yaml:"servername"`
	Version  string   `yaml:"version"`
	Endpoint []string `yaml:"endpoints"`
	Protocal string   `yaml:"protocol"`
	TlsName  string   `yaml:"tls"`
	LBType   string   `yaml:"loadbalance"`
}

type TlsConfig struct {
	Name string `yaml:"name"`
	Cert string `yaml:"cert_chain_file"`
	Key  string `yaml:"private_key_file"`
	CA   string `yaml:"ca_cert_file"`
}

type RouterConfig struct {
	Name  string    `yaml:"name"`
	Prior int       `yaml:"priority"`
	Type  MatchType `yaml:"rule_type"`
	Value string    `yaml:"rule_value"`

	Headers  []Header      `yaml:"headers"`
	Clusters []DestCluster `yaml:"clusters"`

	Timeout int         `yaml:"timeout"`
	Retry   RetryPolicy `yaml:"retry_policy"`
}

type GlobalConfig struct {
	Listeners []ListernerConfig `yaml:"listeners"`
	Routers   []RouterConfig    `yaml:"router"`
	TlsCfg    []TlsConfig       `yaml:"tls"`
	Clusters  []ClusterConfig   `yaml:"clusters"`
}

var globalconfig *GlobalConfig

func LoadConfig(filename string) (*GlobalConfig, error) {

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := new(GlobalConfig)
	config.Listeners = make([]ListernerConfig, 0)
	config.Routers = make([]RouterConfig, 0)
	config.Clusters = make([]ClusterConfig, 0)
	config.TlsCfg = make([]TlsConfig, 0)

	err = yaml.Unmarshal(body, config)
	if err != nil {
		return nil, err
	}

	if !config.Verify() {
		err := errors.New("This config is verity failed!")
		return nil, err
	}

	globalconfig = config

	return config, nil
}

func (c *GlobalConfig) Verify() bool {

	return true
}

func (c *GlobalConfig) listenerGetAll() []ListernerConfig {
	return c.Listeners
}

func (c *GlobalConfig) listenerGet(name string) *ListernerConfig {
	for _, v := range c.Listeners {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func (c *GlobalConfig) RouterGetAll() []RouterConfig {
	return c.Routers
}

func (c *GlobalConfig) RouterGet(name string) *RouterConfig {
	for _, v := range c.Routers {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func (c *GlobalConfig) ClusterGetAll() []ClusterConfig {
	return c.Clusters
}

func (c *GlobalConfig) ClusterGet(name string) *ClusterConfig {
	for _, v := range c.Clusters {
		if v.Name == name {
			return &v
		}
	}
	return nil
}

func (c *GlobalConfig) TlsGetAll() []TlsConfig {
	return c.TlsCfg
}

func (c *GlobalConfig) TlsGet(name string) *TlsConfig {
	for _, v := range c.TlsCfg {
		if v.Name == name {
			return &v
		}
	}
	return nil
}
