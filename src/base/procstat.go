package base

import (
	"os/exec"
	"bufio"
	"bytes"
	"strings"
	"../model"
)

type ToProcStat struct {
	User   string
	Id     string
	Proc   string
	UseCpu string
}

func listToProcStat() (toProcStats []*ToProcStat) {
	cmd := exec.Command("sh", "-c", "ps aux|grep -v PID|sort -rn -k3|head -3")
	outBs, err := cmd.Output()
	if nil != err {

	}
	reader := bufio.NewReader(bytes.NewBuffer(outBs))
	for {
		line, err := reader.ReadBytes('\n')
		if nil != err {
			break
		}
		fields := strings.Fields(string(line))
		toProcStats = append(toProcStats, &ToProcStat{User: fields[0], Id: fields[1], Proc: fields[10], UseCpu: fields[2]})
	}

	return
}

func (*ToProcStat) Metrics() (metrics []*model.MetricValue) {
	toProcStatList := listToProcStat()
	if nil == toProcStatList {
		return nil
	}
	for idx := range toProcStatList {
		oneStat := toProcStatList[idx]
		metrics = append(metrics, &model.MetricValue{Endpoint: "topcpu", Metric: "topcpu.useage.percent", Value: oneStat.UseCpu, Tags: map[string]string{"proc": oneStat.Proc}, Attributes: map[string]string{"id": oneStat.Id, "user": oneStat.User}})
	}
	return
}
