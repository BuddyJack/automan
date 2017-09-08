package base

import (
	"os/exec"
	"bufio"
	"bytes"
	"strings"
	"strconv"
	"../model"
)

type TcpConnStat struct {
	TimeWait    uint64
	ESTABLISHED uint64
	CLOSE_WAIT  uint64
}

func listTcpConnStat() *TcpConnStat {
	cmd := exec.Command("sh", "-c", "netstat -n | awk '/^tcp/ {++S[$NF]} END {for(a in S) print a, S[a]}'")
	bs, err := cmd.Output()
	if nil != err {
		return nil
	}
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	oneStat := &TcpConnStat{}
	for {
		line, err := reader.ReadBytes('\n')
		if nil != err {
			break
		}
		fields := strings.Fields(string(line))
		switch fields[0] {
		case "TIME_WAIT":
			if oneStat.TimeWait, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
				oneStat.TimeWait = 0
				continue
			}
		case "ESTABLISHED":
			if oneStat.ESTABLISHED, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
				oneStat.ESTABLISHED = 0
				continue
			}
		case "CLOSE_WAIT":
			if oneStat.CLOSE_WAIT, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
				oneStat.CLOSE_WAIT = 0
				continue
			}
		default:
			continue
		}
	}
	return oneStat
}

func (*TcpConnStat) Metrics() (metrics []*model.MetricValue) {
	oneStat := listTcpConnStat()
	if nil == oneStat {
		return nil
	}
	metrics = append(metrics, &model.MetricValue{Endpoint: "tcpconns", Metric: "tcpconns.time_wait", Value: strconv.FormatUint(oneStat.TimeWait, 10)})
	metrics = append(metrics, &model.MetricValue{Endpoint: "tcpconns", Metric: "tcpconns.established", Value: strconv.FormatUint(oneStat.ESTABLISHED, 10)})
	metrics = append(metrics, &model.MetricValue{Endpoint: "tcpconns", Metric: "tcpconns.close_wait", Value: strconv.FormatUint(oneStat.CLOSE_WAIT, 10)})
	return
}
