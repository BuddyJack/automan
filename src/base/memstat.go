package base

import (
	"io/ioutil"
	"bufio"
	"bytes"
	"io"
	"strings"
	"unsafe"
	"strconv"
	"../model"
)

var Multi uint64 = 1024

type MemUsage struct {
	Cached   uint64
	MemFree  uint64
	MemTotal uint64
}

var WANT = map[string]struct{}{
	"Cached:":   struct{}{},
	"MemTotal:": struct{}{},
	"MemFree:":  struct{}{},
}

func readMemStat() (*MemUsage, error) {
	statFile := "/proc/meminfo"
	bs, err := ioutil.ReadFile(statFile)
	if nil != err {
		return nil, err
	}
	memInfo := &MemUsage{}
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			err = nil
			break
		} else if nil != err {
			return nil, err
		}
		fields := strings.Fields(*(*string)(unsafe.Pointer(&line)))
		fieldName := fields[0]
		_, ok := WANT[fieldName]
		if ok && 3 == len(fields) {
			val, numerr := strconv.ParseUint(fields[1], 10, 64)
			if nil != numerr {
				continue
			}
			switch fieldName {
			case "Cached:":
				memInfo.Cached = val * Multi
			case "MemFree:":
				memInfo.MemFree = val * Multi
			case "MemTotal":
				memInfo.MemTotal = val * Multi
			default:
				continue
			}
		}

	}

	return memInfo, nil
}

func (memUsage *MemUsage) Metrics() []*model.MetricValue {
	memUsage, err := readMemStat()
	if nil != err {
		return nil
	}
	freePrecent := float64(memUsage.MemFree) / float64(memUsage.MemTotal)
	freeMetric := model.MetricValue{Endpoint: "memory", Metric: "memory.free.percent", Value: strconv.FormatFloat(freePrecent, 'f', 2, 64)}
	freeCachedPrecent := float64(memUsage.MemFree+memUsage.Cached) / float64(memUsage.MemTotal)
	freeCacheMetric := model.MetricValue{Endpoint: "memory", Metric: "memory.freecached.percent", Value: strconv.FormatFloat(freeCachedPrecent, 'f', 2, 64)}
	return []*model.MetricValue{&freeMetric, &freeCacheMetric}
}
