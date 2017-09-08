package main

import (
	"../config"
	"../base"
	"../middleware"
	"os"
	time2 "time"
	"runtime"
)

var (
	watchConfig  *config.WatchConfig
	localhost    string
	baseInterval uint64
)

func initial() {
	watchConfig = config.ReadConfig()

	if _, exist := watchConfig.Base["host"]; !exist {
		panic("no host in config ,please ensure your agent host...")
		os.Exit(1)
	} else {
		localhost = watchConfig.Base["host"].(string)
	}

	if _, exist := watchConfig.Base["interval"]; !exist {
		baseInterval = 5
	} else {
		baseInterval = uint64(watchConfig.Base["interval"].(float64))
	}
	baseInterval += 1
}
func call() {
	//base call
	go base.Metrics()
	//middleware call
	for _, item := range watchConfig.Enable {
		switch item {
		case "cobar":
			cobars := watchConfig.Cobar
			var cobarConfigs []middleware.CobarConf
			for _, oneCobar := range cobars {
				cobarMap := oneCobar.(map[string]interface{})
				var host string
				if _, exist := cobarMap["host"]; !exist {
					host = localhost
				} else {
					host = cobarMap["host"].(string)
				}

				cobarConfig := middleware.CobarConf{Host: host, Port: uint64(cobarMap["port"].(float64)), User: cobarMap["user"].(string), Passwd: cobarMap["passwd"].(string)}
				cobarConfigs = append(cobarConfigs, cobarConfig)
			}
			cobar := &middleware.CobarStat{}
			go cobar.Metrics(cobarConfigs)
		case "memcache":
			memcacheMap := watchConfig.Memcached
			var memcacheConfigs []middleware.MemcacheConfig
			var host string
			if _, exist := memcacheMap["host"]; !exist {
				host = localhost
			} else {
				host = memcacheMap["host"].(string)
			}
			for _, onePort := range memcacheMap["port"].([]float64) {
				memcacheConfigs = append(memcacheConfigs, middleware.MemcacheConfig{Host: host, Port: uint64(onePort)})
			}
			memcache := &middleware.MemcacheStat{}
			go memcache.Metrics(memcacheConfigs)
		case "mysql":
			mysqls := watchConfig.MySql
			var mysqlConfigs []middleware.MySqlConfig
			for _, oneMysql := range mysqls {
				mysqlMap := oneMysql.(map[string]interface{})
				var host string
				if _, exist := mysqlMap["host"]; !exist {
					host = localhost
				} else {
					host = mysqlMap["host"].(string)
				}
				mysqlConfig := middleware.MySqlConfig{Host: host, Port: uint64(mysqlMap["port"].(float64)), User: mysqlMap["user"].(string), Passwd: mysqlMap["passwd"].(string)}
				mysqlConfigs = append(mysqlConfigs, mysqlConfig)
			}
			mysql := &middleware.MysqlStat{}
			go mysql.Metrics(mysqlConfigs)
		case "nginx":
			nginx := &middleware.NginxStat{}
			go nginx.Metrics()
		case "rabbit":
			rabbits := watchConfig.Rabbit
			var rabbitConfigs []middleware.RabbitConfig
			for _, oneRabbit := range rabbits {
				rabbitMap := oneRabbit.(map[string]interface{})
				var host string
				if _, exist := rabbitMap["host"]; !exist {
					host = localhost
				} else {
					host = rabbitMap["host"].(string)
				}
				rabbitConfig := middleware.RabbitConfig{Host: host, Port: uint64(rabbitMap["port"].(float64)), User: rabbitMap["user"].(string), Passwd: rabbitMap["passwd"].(string)}
				rabbitConfigs = append(rabbitConfigs, rabbitConfig)
			}
			rabbit := &middleware.RabbitNode{}
			go rabbit.Metrics(rabbitConfigs)
		case "redis":
			redisMap := watchConfig.Redis
			var redisConfigs []middleware.RedisConfig
			var host string
			if _, exist := redisMap["host"]; !exist {
				host = localhost
			} else {
				host = redisMap["host"].(string)
			}
			for _, onePort := range redisMap["port"].([]float64) {
				redisConfigs = append(redisConfigs, middleware.RedisConfig{Host: host, Port: uint64(onePort)})
			}
			redis := &middleware.RediStat{}
			go redis.Metrics(redisConfigs)
		case "zookeeper":
			zkMap := watchConfig.Zookeeper
			var zkConfigs []middleware.ZkConfig
			var host string
			if _, exist := zkMap["host"]; !exist {
				host = localhost
			} else {
				host = zkMap["host"].(string)
			}
			for _, onePort := range zkMap["port"].([]float64) {
				zkConfigs = append(zkConfigs, middleware.ZkConfig{Host: host, Port: uint64(onePort)})
			}
			zk := &middleware.ZkStat{}
			go zk.Metrics(zkConfigs)
		}
	}

}

func doAgent(exitChan chan struct{}) {
	timer := time2.NewTicker(3 * time2.Second)
	for {
		select {
		case <-timer.C:
			call()
		}
	}
	timer.Stop()
	close(exitChan)
}


func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	exitChan := make(chan struct{})
	initial()
	//call per 5second
	go doAgent(exitChan)
	<-exitChan
	os.Exit(0)
}
