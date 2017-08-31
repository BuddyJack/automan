package base

import (
	"sync"
	"io/ioutil"
	"bufio"
	"bytes"
	"io"
	"strings"
	"unsafe"
	"strconv"
	"../model"
)

var (
	netStatsMap = make(map[string][2]*NetStat)
	netLock     = new(sync.RWMutex)
	interval    = uint64(10)
)

type NetStat struct {
	Device    string
	ReadOct   uint64
	ReadErr   uint64
	ReadDrop  uint64
	WriteOct  uint64
	WriteErr  uint64
	WriteDrop uint64
}

func listNetStats() (netstats []*NetStat, err error) {
	statFile := "/proc/net/dev"
	bs, err := ioutil.ReadFile(statFile)
	if nil != err {
		return nil, err
	}
	reader := bufio.NewReader(bytes.NewBuffer(bs))

	for {
		line, err := reader.ReadBytes('\n')
		if io.EOF == err {
			err = nil
			break
		} else if nil != err {
			return nil, err
		}
		fields := strings.Fields(strings.TrimSpace(*(*string)(unsafe.Pointer(&line))))
		if !strings.HasPrefix(fields[0], "eth") {
			continue
		}
		oneStat := &NetStat{}
		oneStat.Device = strings.Trim(fields[0], ":")
		if 0 == strings.Compare("0", fields[1]) {
			continue
		}
		if oneStat.ReadOct, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
			return nil, err
		}
		if oneStat.ReadErr, err = strconv.ParseUint(fields[3], 10, 64); nil != err {
			return nil, err
		}
		if oneStat.ReadDrop, err = strconv.ParseUint(fields[4], 10, 64); nil != err {
			return nil, err
		}
		if oneStat.WriteOct, err = strconv.ParseUint(fields[9], 10, 64); nil != err {
			return nil, err
		}
		if oneStat.WriteErr, err = strconv.ParseUint(fields[11], 10, 64); nil != err {
			return nil, err
		}
		if oneStat.WriteDrop, err = strconv.ParseUint(fields[12], 10, 64); nil != err {
			return nil, err
		}
		netstats = append(netstats, oneStat)
	}
	return netstats, nil
}

func (netstat *NetStat) Metrics() (metrics []*model.MetricValue) {
	netList, err := listNetStats()
	if nil != err {
		return nil
	}
	netLock.Lock()
	defer netLock.Unlock()
	for idx := range netList {
		device := netList[idx].Device
		netStatsMap[device] = [2]*NetStat{netList[idx], netStatsMap[device][0]}
		prevStat := netStatsMap[device][1]
		var readSpeed, writeSpeed float64
		var readDrop, readErr, writeDrop, writeErr uint64
		nowStat := netList[idx]
		if nil == prevStat {
			readSpeed = 0
			writeSpeed = 0
			readDrop = 0
			readErr = 0
			writeDrop = 0
			writeErr = 0
		} else {
			readSpeed = float64(nowStat.ReadOct-prevStat.ReadOct) / float64(interval*1024*1024)
			writeSpeed = float64(nowStat.WriteOct-prevStat.WriteOct) / float64(interval*1024*1024)
			readDrop = nowStat.ReadDrop - prevStat.ReadDrop
			readErr = nowStat.ReadErr - prevStat.ReadErr
			writeDrop = nowStat.WriteDrop - prevStat.WriteDrop
			writeErr = nowStat.WriteErr - prevStat.WriteErr
		}
		metrics = append(metrics, &model.MetricValue{Endpoint: "net", Metric: "net.read", Value: strconv.FormatFloat(readSpeed, 'f', 2, 64), Tags: map[string]string{"interface": device}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "net", Metric: "net.read.error", Value: strconv.FormatUint(readErr, 10), Tags: map[string]string{"interface": device}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "net", Metric: "net.read.drop", Value: strconv.FormatUint(readDrop, 10), Tags: map[string]string{"interface": device}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "net", Metric: "net.write", Value: strconv.FormatFloat(writeSpeed, 'f', 2, 64), Tags: map[string]string{"interface": device}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "net", Metric: "net.write.error", Value: strconv.FormatUint(writeErr, 10), Tags: map[string]string{"interface": device}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "net", Metric: "net.write.drop", Value: strconv.FormatUint(writeDrop, 10), Tags: map[string]string{"interface": device}})
	}
	return
}
