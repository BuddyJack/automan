package middleware

import (
	"github.com/go-redis/redis"
	"strconv"
	"strings"
	"unicode"
	"../model"
	"sync"
)

var (
	clientMap   = make(map[string]*redis.Client)
	rediStatMap = make(map[string][2]*RediStat)
	redisLock   = new(sync.RWMutex)
)

type RediStat struct {
	BlockedClient   uint64
	Maxmemory       uint64
	UsedMemory      uint64
	MemFragRatio    float64
	ConnectedClient uint64
	RejectedConns   uint64
	Keys            uint64
	Expires         uint64
	EvictedKeys     uint64
	KeyspaceHits    uint64
	KeyspaceMisses  uint64
	OpsPerSec       float64
	Port            string
}

func listRediStats(ports []uint64) (rediStats []*RediStat) {
	for idx := range ports {
		onePort := strconv.FormatUint(ports[idx], 10)
		//save client vs reconnect in every batch
		client := clientMap[onePort]
		if nil == client {
			if _, err := client.Ping().Result(); nil != err {
				client = nil
				client = redis.NewClient(&redis.Options{Addr: "", DB: 0})
				clientMap[onePort] = client
			}
		}
		res, err := client.Info().Result()
		if nil != err {
			continue
		}
		items := strings.Split(res, "\n")
		oneStat := &RediStat{Port: onePort}
		for _, oneItem := range items {
			oneItem = strings.TrimFunc(oneItem, unicode.IsSpace)
			if strings.HasPrefix(oneItem, "#") || 0 == len(oneItem) {
				continue
			}
			fields := strings.Split(oneItem, ":")
			switch fields[0] {
			case "blocked_clients":
				if oneStat.BlockedClient, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.BlockedClient = 0
				}
			case "maxmemory":
				if oneStat.Maxmemory, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.Maxmemory = 0
				}
			case "used_memory_rss":
				if oneStat.UsedMemory, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.UsedMemory = 0
				}
			case "mem_fragmentation_ratio":
				if oneStat.MemFragRatio, err = strconv.ParseFloat(fields[1], 64); nil != err {
					oneStat.MemFragRatio = 0
				}
			case "connected_clients":
				if oneStat.ConnectedClient, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.ConnectedClient = 0
				}
			case "rejected_connections":
				if oneStat.RejectedConns, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.RejectedConns = 0
				}
			case "db0":
				chFields := strings.Split(fields[1], ",")
				for _, chField := range chFields {
					value := strings.Split(chField, "=")[1]
					if strings.HasPrefix(chField, "keys") {
						if oneStat.Keys, err = strconv.ParseUint(value, 10, 64); nil != err {
							oneStat.Keys = 0
						}
					} else if strings.HasPrefix(chField, "expires") {
						if oneStat.Expires, err = strconv.ParseUint(value, 10, 64); nil != err {
							oneStat.Expires = 0
						}
					}
				}
			case "evicted_keys":
				if oneStat.EvictedKeys, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.EvictedKeys = 0
				}
			case "keyspace_hits":
				if oneStat.KeyspaceHits, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.KeyspaceHits = 0
				}
			case "keyspace_misses":
				if oneStat.KeyspaceMisses, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.KeyspaceMisses = 0
				}
			case "instantaneous_ops_per_sec":
				if oneStat.OpsPerSec, err = strconv.ParseFloat(fields[1], 64); nil != err {
					oneStat.OpsPerSec = 0
				}
			}
		}
		rediStats = append(rediStats, oneStat)
	}
	return
}

func (*RediStat) Metrics() (metrics []*model.MetricValue) {
	rediStatList := listRediStats([]uint64{})
	redisLock.Lock()
	defer redisLock.Unlock()
	for idx, oneStat := range rediStatList {
		//update
		port := oneStat.Port
		rediStatMap[port] = [2]*RediStat{rediStatList[idx], rediStatMap[port][0]}
		prevStat := rediStatMap[port][1]
		var connectClients, rejectedConns, evictedKeys, blockedClients uint64
		var memFragRatio, usedMemory, expiresRatio, keyspaceHitratio, opsPerSec float64
		if nil == prevStat {
			rejectedConns = 0
			evictedKeys = 0
			keyspaceHitratio = 0
		} else {
			rejectedConns = oneStat.RejectedConns - prevStat.RejectedConns
			evictedKeys = oneStat.EvictedKeys - prevStat.EvictedKeys
			keyspaceHitratio = float64(oneStat.KeyspaceHits-prevStat.KeyspaceHits) / float64(oneStat.KeyspaceMisses+oneStat.KeyspaceHits-prevStat.KeyspaceHits-prevStat.KeyspaceMisses)
		}
		blockedClients = oneStat.BlockedClient
		if 0 == oneStat.Maxmemory {
			usedMemory = 0
		} else {
			usedMemory = float64(oneStat.UsedMemory) / float64(oneStat.Maxmemory)
		}
		memFragRatio = oneStat.MemFragRatio
		connectClients = oneStat.ConnectedClient
		if 0 == oneStat.Keys {
			expiresRatio = 0
		} else {
			expiresRatio = float64(oneStat.Expires) / float64(oneStat.Keys)
		}
		opsPerSec = oneStat.OpsPerSec
		metrics = append(metrics, &model.MetricValue{Endpoint: "redis", Metric: "redis.blocked_clients", Value: strconv.FormatUint(blockedClients, 10), Tags: map[string]string{"port": port}},
			&model.MetricValue{Endpoint: "redis", Metric: "redis.used_memory.percent", Value: strconv.FormatFloat(usedMemory, 'f', 2, 64), Tags: map[string]string{"port": port}},
			&model.MetricValue{Endpoint: "redis", Metric: "redis.mem_fragmentation_ratio", Value: strconv.FormatFloat(memFragRatio, 'f', 2, 64), Tags: map[string]string{"port": port}},
			&model.MetricValue{Endpoint: "redis", Metric: "redis.connected_clients", Value: strconv.FormatUint(connectClients, 10), Tags: map[string]string{"port": port}},
			&model.MetricValue{Endpoint: "redis", Metric: "redis.rejected_connections", Value: strconv.FormatUint(rejectedConns, 10), Tags: map[string]string{"port": port}},
			&model.MetricValue{Endpoint: "redis", Metric: "redis.expires_ratio", Value: strconv.FormatFloat(expiresRatio, 'f', 2, 64), Tags: map[string]string{"port": port}},
			&model.MetricValue{Endpoint: "redis", Metric: "redis.evicted_keys", Value: strconv.FormatUint(evictedKeys, 10), Tags: map[string]string{"port": port}},
			&model.MetricValue{Endpoint: "redis", Metric: "redis.keyspace_hitratio", Value: strconv.FormatFloat(keyspaceHitratio, 'f', 2, 64), Tags: map[string]string{"port": port}},
			&model.MetricValue{Endpoint: "redis", Metric: "redis.instantaneous_ops_per_sec", Value: strconv.FormatFloat(opsPerSec, 'f', 2, 64), Tags: map[string]string{"port": port}},
		)
	}
	return
}
