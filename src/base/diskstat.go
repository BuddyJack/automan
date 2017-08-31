package base

import (
	"io/ioutil"
	"bufio"
	"bytes"
	"io"
	"strings"
	"unsafe"
	"strconv"
	"sync"
	"../model"
)

var (
	diskStatsMap = make(map[string][2]*DiskStat)
	dsLock       = new(sync.RWMutex)
)

type DiskStat struct {
	Device        string
	ReadRequests  uint64
	ReadSec       uint64
	WriteRequests uint64
	WriteSec      uint64
}

func listDiskStats() (diskstats []*DiskStat, err error) {
	statFile := "/proc/diskstats"
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
		fields := strings.Fields(*(*string)(unsafe.Pointer(&line)))
		size := len(fields)
		if 14 != size {
			continue
		}
		if 0 == strings.Compare("0", fields[3]) {
			continue
		}
		oneStat := &DiskStat{Device: fields[2]}
		if oneStat.ReadRequests, err = strconv.ParseUint(fields[3], 10, 64); nil != err {
			return nil, err
		}
		if oneStat.ReadSec, err = strconv.ParseUint(fields[6], 10, 64); nil != err {
			return nil, err
		}
		if oneStat.WriteRequests, err = strconv.ParseUint(fields[7], 10, 64); nil != err {
			return nil, err
		}
		if oneStat.WriteSec, err = strconv.ParseUint(fields[10], 10, 64); nil != err {
			return nil, err
		}

		diskstats = append(diskstats, oneStat)
	}
	return diskstats, nil
}

func (diskStat *DiskStat) Metrics() (metrics []*model.MetricValue) {
	diskList, err := listDiskStats()
	if nil != err {
		return nil
	}
	dsLock.Lock()
	defer dsLock.Unlock()
	for idx := range diskList {
		device := diskList[idx].Device
		diskStatsMap[device] = [2]*DiskStat{diskList[idx], diskStatsMap[device][0]}
		var readSpeed, writeSpeed float64
		if nil == diskStatsMap[device][1] {
			readSpeed = float64(0)
			writeSpeed = float64(0)
		} else {
			nowStat := diskStatsMap[device][0]
			prevStat := diskStatsMap[device][1]
			readSpeed = float64(nowStat.ReadRequests-prevStat.ReadRequests) / float64(nowStat.ReadSec-prevStat.ReadSec)
			writeSpeed = float64(nowStat.WriteRequests-prevStat.WriteRequests) / float64(nowStat.WriteSec-prevStat.WriteSec)
		}
		metrics = append(metrics, &model.MetricValue{Endpoint: "io", Metric: "io.read", Value: strconv.FormatFloat(readSpeed, 'f', 2, 64), Tags: map[string]string{"mount": device}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "io", Metric: "io.write", Value: strconv.FormatFloat(writeSpeed, 'f', 2, 64), Tags: map[string]string{"mount": device}})
	}
	return
}
