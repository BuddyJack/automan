package config

import (
	"io/ioutil"
	"io"
	"encoding/json"
)

const (
	ConfigFile = "/Users/jack/Documents/metric_config.json"
)

type WatchConfig struct {
	Base      map[string]interface{} `json:"base"`
	Cobar     map[string]interface{} `json:"cobar"`
	Redis     map[string]interface{} `json:"redis"`
	Zookeeper map[string]interface{} `json:"zk"`
	Rabbit    map[string]interface{} `json:"rabbit"`
	Memcached map[string]interface{} `json:"memcache"`
	MySql     map[string]interface{} `json:"mysql"`
	Nginx     map[string]interface{} `json:"nginx"`
	Enable    []string `json:"enable"`
}

func ReadConfig() (*WatchConfig) {
	bs, err := ioutil.ReadFile(ConfigFile)
	if nil != err && io.EOF != err {
	}
	println(string(bs))
	var configs WatchConfig
	err = json.Unmarshal(bs, &configs)
	if nil != err {
		panic(err)
	}
	return &configs
}
