package base

import (
	"io/ioutil"
	"bufio"
	"bytes"
	"strings"
	"strconv"
	"unsafe"
	"../model"
)

type CpuUsage struct {
	System uint64
	Iowait uint64
	Idle   uint64
	Total  uint64
}

func listCpuStat() (*CpuUsage, error) {
	statFile := "/proc/stat"
	bs, err := ioutil.ReadFile(statFile)
	if nil != err {
		return nil, err
	}
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	for {
		line, err := reader.ReadBytes('\n')
		if nil != err {
			return nil, err
		}
		useage := parseLine(line)
		if nil != useage {
			return useage, nil
		}
	}
}

func parseLine(line []byte) (*CpuUsage) {
	fields := strings.Fields(*(*string)(unsafe.Pointer(&line)))
	if 0 != strings.Compare("cpu", fields[0]) {
		return nil
	}
	cu := new(CpuUsage)
	sz := len(fields)
	for i := 1; i < sz; i++ {
		val, err := strconv.ParseUint(fields[i], 10, 64)
		if nil != err {
			continue
		}
		cu.Total += val
		switch i {
		case 3:
			cu.System = val
		case 4:
			cu.Idle = val
		case 5:
			cu.Iowait = val
		}
	}
	return cu
}

func (cpuUsage *CpuUsage) Metrics() []*model.MetricValue {

	cpuUsage, err := listCpuStat()
	if nil != err {
		return nil
	}
	idlePercent := float64(cpuUsage.Idle) / float64(cpuUsage.Total)
	idleMetric := model.MetricValue{Endpoint: "cpu", Metric: "cpu.idle", Value: strconv.FormatFloat(idlePercent, 'f', 2, 64)}
	iowaitPercent := float64(cpuUsage.Iowait) / float64(cpuUsage.Total)
	iowaitMetric := model.MetricValue{Endpoint: "cpu", Metric: "cpu.iowait", Value: strconv.FormatFloat(iowaitPercent, 'f', 2, 64)}
	systemPercent := float64(cpuUsage.System) / float64(cpuUsage.Total)
	systemMetric := model.MetricValue{Endpoint: "cpu", Metric: "cpu.system", Value: strconv.FormatFloat(systemPercent, 'f', 2, 64)}

	return []*model.MetricValue{&idleMetric, &iowaitMetric, &systemMetric}
}
