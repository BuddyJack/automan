package middleware

import (
	"strings"
	"net"
	"bufio"
	"bytes"
	"unicode"
	"strconv"
	"../model"
	"sync"
)

var (
	memcacheStatMap = make(map[string][2]*MemcacheStat)
	memcacheLock    = new(sync.RWMutex)
	interval        = 10
)

type MemcacheStat struct {
	Port            string
	CurrItems       uint64
	Bytes           uint64
	LimitMaxbytes   uint64
	CurrConnections uint64
	Evictions       uint64
	GetHits         uint64
	GetMisses       uint64
	DeleteHits      uint64
	DeleteMisses    uint64
	IncrHits        uint64
	IncrMisses      uint64
	DecrHits        uint64
	DecrMisses      uint64
	CmdGet          uint64
	CmdSet          uint64
}

type MemcacheConfig struct {
	Host string
	Port uint64
}

func listMemcacheStats(configs []MemcacheConfig) (memcacheStats []*MemcacheStat) {
	for _, oneConfig := range configs {
		onePort := strconv.FormatUint(oneConfig.Port, 10)
		tcpAddr, err := net.ResolveTCPAddr("tcp4", oneConfig.Host+":"+onePort)
		if nil != err {
			continue
		}
		tcpConn, err := net.DialTCP("tcp4", nil, tcpAddr)
		tcpConn.Write([]byte("stats\n"))
		defer tcpConn.Close()
		var res = make([]byte, 1024*2)
		tcpConn.Read(res)
		tcpConn.CloseRead()
		reader := bufio.NewReader(bytes.NewBuffer(res))
		oneStat := &MemcacheStat{Port: onePort}
		for {
			line, err := reader.ReadBytes('\n')
			if nil != err {
				break
			}
			lineStr := strings.TrimFunc(string(line), unicode.IsSpace)
			if 0 == strings.Compare("END", lineStr) {
				break
			}
			fields := strings.Fields(lineStr)
			if 3 != len(fields) {
				continue
			}
			value, err := strconv.ParseUint(fields[2], 10, 64)
			if nil != err {
				continue
			}
			switch fields[1] {
			case "curr_items":
				oneStat.CurrItems = value
			case "bytes":
				oneStat.Bytes = value
			case "limit_maxbytes":
				oneStat.LimitMaxbytes = value
			case "curr_connections":
				oneStat.CurrConnections = value
			case "evictions":
				oneStat.Evictions = value
			case "get_hits":
				oneStat.GetHits = value
			case "get_misses":
				oneStat.GetMisses = value
			case "delete_hits":
				oneStat.DeleteHits = value
			case "delete_misses":
				oneStat.DeleteMisses = value
			case "incr_hits":
				oneStat.IncrHits = value
			case "incr_misses":
				oneStat.IncrMisses = value
			case "decr_hits":
				oneStat.DecrHits = value
			case "decr_misses":
				oneStat.DecrMisses = value
			case "cmd_get":
				oneStat.CmdGet = value
			case "cmd_set":
				oneStat.CmdSet = value
			}
		}
		memcacheStats = append(memcacheStats, oneStat)
	}
	return
}

func (*MemcacheStat) Metrics(configs []MemcacheConfig) (metrics []*model.MetricValue) {
	memcacheStatList := listMemcacheStats(configs)
	if nil == memcacheStatList {
		return nil
	}
	memcacheLock.Lock()
	defer memcacheLock.Unlock()
	for idx, oneStat := range memcacheStatList {
		onePort := oneStat.Port
		memcacheStatMap[onePort] = [2]*MemcacheStat{memcacheStatList[idx], memcacheStatMap[onePort][0]}
		var currItems, currConns, evictions uint64
		var cacheUsage, getHitRatio, delHitRatio, incrHitRatio, decrHitRatio, getPerSec, setPerSec float64
		prevStat := memcacheStatMap[onePort][1]
		if nil == prevStat {
			evictions = 0
			getHitRatio = 0
			delHitRatio = 0
			incrHitRatio = 0
			decrHitRatio = 0
			getPerSec = 0
			setPerSec = 0
		} else {
			evictions = oneStat.Evictions - prevStat.Evictions
			getHitRatio = float64(oneStat.GetHits-prevStat.GetHits) / float64(oneStat.GetHits+oneStat.GetMisses-prevStat.GetHits-prevStat.GetMisses)
			decrHitRatio = float64(oneStat.DecrHits-prevStat.DecrHits) / float64(oneStat.DecrMisses+oneStat.DecrHits-prevStat.DecrMisses-prevStat.DecrHits)
			incrHitRatio = float64(oneStat.IncrHits-prevStat.IncrHits) / float64(oneStat.IncrHits+oneStat.IncrMisses-prevStat.IncrHits-prevStat.IncrMisses)
			delHitRatio = float64(oneStat.DeleteHits-prevStat.DeleteHits) / float64(oneStat.DeleteHits+oneStat.DeleteMisses-prevStat.DeleteHits-prevStat.DeleteMisses)
			getPerSec = float64(oneStat.CmdGet-prevStat.CmdGet) / float64(interval)
			setPerSec = float64(oneStat.CmdSet-prevStat.CmdSet) / float64(interval)
		}
		currItems = oneStat.CurrItems
		currConns = oneStat.CurrConnections
		cacheUsage = float64(oneStat.Bytes) / float64(oneStat.LimitMaxbytes)
		metrics = append(metrics, &model.MetricValue{Endpoint: "memcached", Metric: "memcached.curr_items", Value: strconv.FormatUint(currItems, 10), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.cached_used.percent", Value: strconv.FormatFloat(cacheUsage, 'f', 2, 64), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.curr_connections", Value: strconv.FormatUint(currConns, 10), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.evictions", Value: strconv.FormatUint(evictions, 10), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.get_hitratio", Value: strconv.FormatFloat(getHitRatio, 'f', 2, 64), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.delete_hitratio", Value: strconv.FormatFloat(delHitRatio, 'f', 2, 64), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.incr_hitratio", Value: strconv.FormatFloat(incrHitRatio, 'f', 2, 64), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.decr_hitratio", Value: strconv.FormatFloat(decrHitRatio, 'f', 2, 64), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.get_pers", Value: strconv.FormatFloat(getPerSec, 'f', 2, 64), Tags: map[string]string{"port": onePort}},
			&model.MetricValue{Endpoint: "memcached", Metric: "memcached.set_pers", Value: strconv.FormatFloat(setPerSec, 'f', 2, 64), Tags: map[string]string{"port": onePort}},
		)
	}
	return
}
