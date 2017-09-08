package middleware

import (
	"strings"
	"unicode"
	"fmt"
	"os/exec"
	"bufio"
	"bytes"
	"io"
	"strconv"
	"sync"
	"../model"
)

const (
	Cobar_Cmd_ThreadPool = "mysql -h%s -P%d -u%s -p%s -e 'show @@threadpool\\G'"
	Cobar_Cmd_Proc       = "mysql -h%s -P%d -u%s -p%s -e 'show @@processor\\G'"
	Cobar_Cmd_Table      = "mysql -h%s -P%d -u%s -p%s -e 'show @@sql.stat.table\\G'"
	Cobar_Cmd_Schema     = "mysql -h%s -P%d -u%s -p%s -e 'show @@sql.stat.schema\\G'"
)

var (
	cobarStatsMap       = make(map[string][2]*CobarStat)
	cobarTableStasMap   = make(map[string][2]map[string]*CobarTable)
	cobarSchemaStatsMap = make(map[string][2]map[string]*CobarSchema)
	cobarStatLock       = new(sync.RWMutex)
	cobarTableLock      = new(sync.RWMutex)
	cobarSchemaLock     = new(sync.RWMutex)
)

//mysql metrics
type CobarConf struct {
	Host   string
	Port   uint64
	User   string
	Passwd string
}

type CobarStat struct {
	Port          string
	NetIn         uint64
	NetOut        uint64
	ActiveCount   uint64
	CompletedTask uint64
	TaskQueueSize uint64
}

type CobarTable struct {
	Port      string
	Table     string
	Read      uint64
	Write     uint64
	Cobar10   uint64
	Cobar50   uint64
	Cobar200  uint64
	Cobar1000 uint64
	Cobar2000 uint64
	Node10    uint64
	Node50    uint64
	Node200   uint64
	Node1000  uint64
	Node2000  uint64
}

type CobarSchema struct {
	Port      string
	Schema    string
	Read      uint64
	Write     uint64
	Cobar10   uint64
	Cobar50   uint64
	Cobar200  uint64
	Cobar1000 uint64
	Cobar2000 uint64
	Node10    uint64
	Node50    uint64
	Node200   uint64
	Node1000  uint64
	Node2000  uint64
}

func getStatThread(conf CobarConf) (activeCount, taskQueueSize, completeTask uint64) {
	cmd := exec.Command("ssh", "root@192.168.102.76", fmt.Sprintf(Cobar_Cmd_ThreadPool, conf.Host, conf.Port, conf.User, conf.Passwd))
	bs, err := cmd.Output()
	if nil != err {
		println(err.Error())
	}
	println(string(bs))
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	process := false
	completeTask, taskQueueSize, activeCount = 0, 0, 0
	for {
		bline, err := reader.ReadBytes('\n')
		if io.EOF == err {
			break
		} else if nil != err {
			panic(err)
			break
		}
		line := strings.TrimFunc(string(bline), unicode.IsSpace)
		fields := strings.Split(line, ":")
		if strings.HasPrefix(line, "NAME") {
			if 2 != len(fields) {
				process = false
				continue
			}
			fields[0] = strings.TrimFunc(fields[0], unicode.IsSpace)
			fields[1] = strings.TrimFunc(fields[1], unicode.IsSpace)
			if strings.HasPrefix(fields[1], "Process") {
				process = true
			} else {
				process = false
			}
		}
		if process {
			switch fields[0] {
			case "ACTIVE_COUNT":
				if count, err := strconv.ParseUint(strings.TrimFunc(fields[1], unicode.IsSpace), 10, 64); nil != err {
					activeCount += 0
				} else {
					activeCount += count
				}
			case "COMPLETED_TASK":
				if count, err := strconv.ParseUint(strings.TrimFunc(fields[1], unicode.IsSpace), 10, 64); nil != err {
					completeTask += 0
				} else {
					completeTask += count
				}
			case "TASK_QUEUE_SIZE":
				if count, err := strconv.ParseUint(strings.TrimFunc(fields[1], unicode.IsSpace), 10, 64); nil != err {
					taskQueueSize += 0
				} else {
					taskQueueSize += count
				}
			}
		}
	}
	println("active:" + strconv.FormatUint(activeCount, 10) + "\t taskqueue:" + strconv.FormatUint(taskQueueSize, 10) + "\t complete:" + strconv.FormatUint(completeTask, 10))
	return
}

func getStatProc(conf CobarConf) (netIn, netOut uint64) {
	cmd := exec.Command("ssh", "root@192.168.102.76", fmt.Sprintf(Cobar_Cmd_Proc, conf.Host, conf.Port, conf.User, conf.Passwd))
	bs, err := cmd.Output()
	if nil != err {
		println(err.Error())
	}
	println(string(bs))
	reader := bufio.NewReader(bytes.NewReader(bs))
	var process = false
	netIn, netOut = 0, 0
	for {
		bline, err := reader.ReadBytes('\n')
		if io.EOF == err {
			break
		} else if nil != err {
			panic(err)
			break
		}
		line := strings.TrimFunc(string(bline), unicode.IsSpace)
		fields := strings.Split(line, ":")
		if 2 != len(fields) {
			process = false
			continue
		}
		fields[0] = strings.TrimFunc(fields[0], unicode.IsSpace)
		fields[1] = strings.TrimFunc(fields[1], unicode.IsSpace)
		if 0 == strings.Compare("NAME", fields[0]) {
			if strings.HasPrefix(fields[1], "Processor") {
				process = true
			} else {
				process = false
			}
		}
		if process {
			switch fields[0] {
			case "NET_IN":
				if count, err := strconv.ParseUint(fields[1], 10, 64); nil != err {
					netIn += 0
				} else {
					netIn += count
				}
			case "NET_OUT":
				if count, err := strconv.ParseUint(fields[1], 10, 64); nil != err {
					netOut += 0
				} else {
					netOut += count
				}
			}
		}
	}
	return

}

func getStatTable(conf CobarConf) (tableMap map[string]*CobarTable) {
	cmd := exec.Command("ssh", "root@192.168.102.76", fmt.Sprintf(Cobar_Cmd_Table, conf.Host, conf.Port, conf.User, conf.Passwd))
	bs, err := cmd.Output()
	if nil != err {
		println(err.Error())
	}
	println(string(bs))
	reader := bufio.NewReader(bytes.NewReader(bs))
	tableMap = make(map[string]*CobarTable)
	var table string
	for {
		bline, err := reader.ReadBytes('\n')
		if io.EOF == err {
			break
		} else if nil != err {
			panic(err)
			break
		}
		line := strings.TrimFunc(string(bline), unicode.IsSpace)
		fields := strings.Split(line, ":")
		if 2 != len(fields) {
			continue
		}
		fields[0] = strings.TrimFunc(fields[0], unicode.IsSpace)
		fields[1] = strings.TrimFunc(fields[1], unicode.IsSpace)
		if 0 == strings.Compare("TABLE", fields[0]) {
			table = fields[1]
			tableMap[table] = &CobarTable{Port: strconv.FormatUint(conf.Port, 10), Table: table}
			continue
		}
		switch fields[0] {
		case "R":
			if tableMap[table].Read, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
				tableMap[table].Read = 0
			}
		case "W":
			if tableMap[table].Write, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
				tableMap[table].Write = 0
			}
		case "ALL_TTL_COUNT":
			ttls := fields[1][1:len(fields[1])-1]
			ttlsFields := strings.Split(ttls, ",")
			if tableMap[table].Cobar10, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[0], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Cobar10 = 0
			}
			if tableMap[table].Cobar50, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[1], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Cobar50 = 0
			}
			if tableMap[table].Cobar200, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[2], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Cobar200 = 0
			}
			if tableMap[table].Cobar1000, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[3], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Cobar1000 = 0
			}
			if tableMap[table].Cobar2000, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[4], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Cobar2000 = 0
			}
		case "NODE_TTL_COUNT":
			ttls := fields[1][2:len(fields[1])-1]
			ttlsFields := strings.Split(ttls, ",")
			if tableMap[table].Node10, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[0], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Node10 = 0
			}
			if tableMap[table].Node50, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[1], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Node50 = 0
			}
			if tableMap[table].Node200, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[2], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Node200 = 0
			}
			if tableMap[table].Node1000, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[3], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Node1000 = 0
			}
			if tableMap[table].Node2000, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[4], unicode.IsSpace), 10, 64); nil != err {
				tableMap[table].Node2000 = 0
			}
		default:
			continue
		}
	}
	return
}

func getStatSchema(conf CobarConf) (schemaMap map[string]*CobarSchema) {
	cmd := exec.Command("ssh", "root@192.168.102.76", fmt.Sprintf(Cobar_Cmd_Schema, conf.Host, conf.Port, conf.User, conf.Passwd))
	bs, err := cmd.Output()
	if nil != err {
		println(err.Error())
	}
	println(string(bs))
	reader := bufio.NewReader(bytes.NewReader(bs))
	schemaMap = make(map[string]*CobarSchema)
	var schema string
	for {
		bline, err := reader.ReadBytes('\n')
		if io.EOF == err {
			break
		} else if nil != err {
			panic(err)
			break
		}
		line := strings.TrimFunc(string(bline), unicode.IsSpace)
		fields := strings.Split(line, ":")
		if 2 != len(fields) {
			continue
		}
		fields[0] = strings.TrimFunc(fields[0], unicode.IsSpace)
		fields[1] = strings.TrimFunc(fields[1], unicode.IsSpace)
		if 0 == strings.Compare("SCHEMA", fields[0]) {
			schema = fields[1]
			schemaMap[schema] = &CobarSchema{Port: strconv.FormatUint(conf.Port, 10), Schema: schema}
			continue
		}
		switch fields[0] {
		case "R":
			if schemaMap[schema].Read, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
				schemaMap[schema].Read = 0
			}
		case "W":
			if schemaMap[schema].Write, err = strconv.ParseUint(fields[1], 10, 64); nil != err {
				schemaMap[schema].Write = 0
			}
		case "ALL_TTL_COUNT":
			ttls := fields[1][1:len(fields[1])-1]
			ttlsFields := strings.Split(ttls, ",")
			if schemaMap[schema].Cobar10, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[0], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Cobar10 = 0
			}
			if schemaMap[schema].Cobar50, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[1], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Cobar50 = 0
			}
			if schemaMap[schema].Cobar200, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[2], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Cobar200 = 0
			}
			if schemaMap[schema].Cobar1000, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[3], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Cobar1000 = 0
			}
			if schemaMap[schema].Cobar2000, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[4], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Cobar2000 = 0
			}
		case "NODE_TTL_COUNT":
			ttls := fields[1][2:len(fields[1])-1]
			ttlsFields := strings.Split(ttls, ",")
			if schemaMap[schema].Node10, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[0], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Node10 = 0
			}
			if schemaMap[schema].Node50, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[1], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Node50 = 0
			}
			if schemaMap[schema].Node200, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[2], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Node200 = 0
			}
			if schemaMap[schema].Node1000, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[3], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Node1000 = 0
			}
			if schemaMap[schema].Node2000, err = strconv.ParseUint(strings.TrimFunc(ttlsFields[4], unicode.IsSpace), 10, 64); nil != err {
				schemaMap[schema].Node2000 = 0
			}
		default:
			continue
		}
	}
	return
}

func listCobarStat(cobarConfs []CobarConf) (metrics []*model.MetricValue) {
	cobarStatLock.Lock()
	defer cobarStatLock.Unlock()
	for _, oneConf := range cobarConfs {
		var onePort = strconv.FormatUint(oneConf.Port, 10)
		var oneStat = &CobarStat{Port: onePort}
		oneStat.ActiveCount, oneStat.TaskQueueSize, oneStat.CompletedTask = getStatThread(oneConf)
		oneStat.NetIn, oneStat.NetOut = getStatProc(oneConf)
		cobarStatsMap[onePort] = [2]*CobarStat{oneStat, cobarStatsMap[onePort][0]}
		prevStat := cobarStatsMap[onePort][1]
		var completedTask, netIn, netOut uint64 = 0, 0, 0
		if nil != prevStat {
			completedTask = oneStat.CompletedTask - prevStat.CompletedTask
			netIn = oneStat.NetIn - prevStat.NetIn
			netOut = oneStat.NetOut - prevStat.NetOut
		}
		metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.threadpool.active_count", Value: strconv.FormatUint(oneStat.ActiveCount, 10), Tags: map[string]string{"port": onePort}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.threadpool.task_queue_size", Value: strconv.FormatUint(oneStat.TaskQueueSize, 10), Tags: map[string]string{"port": onePort}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.threadpool.completed_task", Value: strconv.FormatUint(completedTask, 10), Tags: map[string]string{"port": onePort}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.process.net_in", Value: strconv.FormatUint(netIn, 10), Tags: map[string]string{"port": onePort}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.process.net_out", Value: strconv.FormatUint(netOut, 10), Tags: map[string]string{"port": onePort}})
	}
	return
}

func listCobarTableStat(cobarConfs []CobarConf) (metrics []*model.MetricValue) {
	cobarTableLock.Lock()
	defer cobarTableLock.Unlock()
	for _, oneConf := range cobarConfs {
		onePort := strconv.FormatUint(oneConf.Port, 10)
		oneStat := getStatTable(oneConf)
		cobarTableStasMap[onePort] = [2]map[string]*CobarTable{oneStat, cobarTableStasMap[onePort][0]}
		prevStat := cobarTableStasMap[onePort][1]
		for tableName, oneTable := range oneStat {
			var read, write, cobar10, cobar50, cobar200, cobar1000, cobar2000, node10, node50, node200, node1000, node2000 uint64 = 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
			if nil != prevStat && nil != prevStat[tableName] {
				prevTable := prevStat[tableName]
				read = oneTable.Read - prevTable.Read
				write = oneTable.Write - prevTable.Write
				cobar10 = oneTable.Cobar10 - prevTable.Cobar10
				cobar50 = oneTable.Cobar50 - prevTable.Cobar50
				cobar200 = oneTable.Cobar200 - prevTable.Cobar200
				cobar1000 = oneTable.Cobar1000 - prevTable.Cobar1000
				cobar2000 = oneTable.Cobar2000 - prevTable.Cobar2000
				node10 = oneTable.Node10 - prevTable.Node10
				node50 = oneTable.Node50 - prevTable.Node50
				node200 = oneTable.Node200 - prevTable.Node200
				node1000 = oneTable.Node1000 - prevTable.Node1000
				node2000 = oneTable.Node2000 - prevTable.Node2000
			}
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.read", Value: strconv.FormatUint(read, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.write", Value: strconv.FormatUint(write, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar10", Value: strconv.FormatUint(cobar10, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar50", Value: strconv.FormatUint(cobar50, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar200", Value: strconv.FormatUint(cobar200, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar1000", Value: strconv.FormatUint(cobar1000, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar2000", Value: strconv.FormatUint(cobar2000, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node10", Value: strconv.FormatUint(node10, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node50", Value: strconv.FormatUint(node50, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node200", Value: strconv.FormatUint(node200, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node1000", Value: strconv.FormatUint(node1000, 10), Tags: map[string]string{"port": onePort, "table": tableName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node2000", Value: strconv.FormatUint(node2000, 10), Tags: map[string]string{"port": onePort, "table": tableName}})

		}
	}
	return
}

func listCobarSchemaStat(cobarConfs []CobarConf) (metrics []*model.MetricValue) {
	cobarSchemaLock.Lock()
	defer cobarSchemaLock.Unlock()
	for _, oneConf := range cobarConfs {
		onePort := strconv.FormatUint(oneConf.Port, 10)
		oneStat := getStatSchema(oneConf)
		cobarSchemaStatsMap[onePort] = [2]map[string]*CobarSchema{oneStat, cobarSchemaStatsMap[onePort][0]}
		prevStat := cobarSchemaStatsMap[onePort][1]
		for shcemaName, oneSchema := range oneStat {
			var read, write, cobar10, cobar50, cobar200, cobar1000, cobar2000, node10, node50, node200, node1000, node2000 uint64 = 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0
			if nil != prevStat && nil != prevStat[shcemaName] {
				prevSchema := prevStat[shcemaName]
				read = oneSchema.Read - prevSchema.Read
				write = oneSchema.Write - prevSchema.Write
				cobar10 = oneSchema.Cobar10 - prevSchema.Cobar10
				cobar50 = oneSchema.Cobar50 - prevSchema.Cobar50
				cobar200 = oneSchema.Cobar200 - prevSchema.Cobar200
				cobar1000 = oneSchema.Cobar1000 - prevSchema.Cobar1000
				cobar2000 = oneSchema.Cobar2000 - prevSchema.Cobar2000
				node10 = oneSchema.Node10 - prevSchema.Node10
				node50 = oneSchema.Node50 - prevSchema.Node50
				node200 = oneSchema.Node200 - prevSchema.Node200
				node1000 = oneSchema.Node1000 - prevSchema.Node1000
				node2000 = oneSchema.Node2000 - prevSchema.Node2000
			}
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.read", Value: strconv.FormatUint(read, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.write", Value: strconv.FormatUint(write, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar10", Value: strconv.FormatUint(cobar10, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar50", Value: strconv.FormatUint(cobar50, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar200", Value: strconv.FormatUint(cobar200, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar1000", Value: strconv.FormatUint(cobar1000, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.cobar2000", Value: strconv.FormatUint(cobar2000, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node10", Value: strconv.FormatUint(node10, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node50", Value: strconv.FormatUint(node50, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node200", Value: strconv.FormatUint(node200, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node1000", Value: strconv.FormatUint(node1000, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "cobar", Metric: "cobar.table.node2000", Value: strconv.FormatUint(node2000, 10), Tags: map[string]string{"port": onePort, "schema": shcemaName}})

		}
	}
	return
}

func (*CobarStat) Metrics(cobarConfs []CobarConf) (metrics []*model.MetricValue) {
	go listCobarStat(cobarConfs)
	go listCobarTableStat(cobarConfs)
	go listCobarSchemaStat(cobarConfs)

	return
}
