package middleware

import (
	"net"
	"strings"
	"strconv"
	"bufio"
	"bytes"
	"../model"
)

type ZkStat struct {
	ZkIsLeader            uint64
	ZkNumAliveConnections uint64
	ZkAvgLatency          float64
	ZkOutStandingRequests uint64
	ZkServiceHealth       uint64
	Port                  string
}

func listZkStats(zkPorts []uint64) (zkStats []*ZkStat) {
	for idx := range zkPorts {
		onePort := zkPorts[idx]
		tcpAddr, err := net.ResolveTCPAddr("tcp4", strings.Join([]string{"127.0.0.1"}, strconv.Itoa(int(onePort))))
		if nil != err {
			continue
		}
		tcpConn, err := net.DialTCP("tcp4", nil, tcpAddr)
		oneStat := &ZkStat{Port: strconv.FormatUint(onePort, 10)}
		defer tcpConn.Close()
		tcpConn.Write([]byte("ruok"))
		var resp = make([]byte, 1024/8)
		tcpConn.Read(resp)
		if 0 == strings.Compare(string(resp), "imok") {
			oneStat.ZkServiceHealth = 1
		} else {
			oneStat.ZkServiceHealth = 0
		}
		tcpConn.Write([]byte("mntr"))
		resp = make([]byte, 1024*8)
		tcpConn.Read(resp)
		reader := bufio.NewReader(bytes.NewBuffer(resp))
		for {
			line, err := reader.ReadBytes('\n')
			if nil != err {
				break
			}
			fields := strings.Fields(string(line))
			switch fields[0] {
			case "zk_avg_latency":
				if oneStat.ZkAvgLatency, err = strconv.ParseFloat(fields[1], 64); nil != err {
					oneStat.ZkAvgLatency = 0
					continue
				}
			case "zk_num_alive_connections":
				if oneStat.ZkNumAliveConnections, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.ZkNumAliveConnections = 0
					continue
				}
			case "zk_outstanding_requests":
				if oneStat.ZkOutStandingRequests, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
					oneStat.ZkOutStandingRequests = 0
					continue
				}
			case "zk_server_state":
				if 0 != strings.Compare("follower", fields[1]) {
					oneStat.ZkIsLeader = 1
				} else {
					oneStat.ZkIsLeader = 0
				}
			}
		}
		zkStats = append(zkStats, oneStat)
	}
	return
}

func (*ZkStat) Metrics() (metrics []*model.MetricValue) {
	zkstatList := listZkStats([]uint64{})
	for idx := range zkstatList {
		onestat := zkstatList[idx]
		metrics = append(metrics, &model.MetricValue{Endpoint: "zookeeper", Metric: "zookeeper.zk_is_leader", Value: strconv.FormatUint(onestat.ZkIsLeader, 10), Tags: map[string]string{"port": onestat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "zookeeper", Metric: "zookeeper.zk_num_alive_connections", Value: strconv.FormatUint(onestat.ZkNumAliveConnections, 10), Tags: map[string]string{"port": onestat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "zookeeper", Metric: "zookeeper.zk_avg_latency", Value: strconv.FormatFloat(onestat.ZkAvgLatency, 'f', 0, 64), Tags: map[string]string{"port": onestat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "zookeeper", Metric: "zookeeper.zk_outstanding_requests", Value: strconv.FormatUint(onestat.ZkOutStandingRequests, 10), Tags: map[string]string{"port": onestat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "zookeeper", Metric: "zookeeper.zk_service_health", Value: strconv.FormatUint(onestat.ZkServiceHealth, 10), Tags: map[string]string{"port": onestat.Port}})
	}
	return
}
