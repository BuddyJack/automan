package main

import (
	"os/exec"
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode"
	"strconv"
	"fmt"
)

func main() {
	listStatThread()
}

func listStatThread() {
	cmd := exec.Command("ssh", "root@192.168.102.76", "mysql -h192.168.102.45 -P9066 -uyh_test -pyh_test -e 'show @@threadpool\\G'")
	bs, err := cmd.Output()
	if nil != err {
		println(err.Error())
	}
	println(string(bs))
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	process := false
	var completeTask, taskQueueSize, activeCount uint64 = 0, 0, 0
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

}

func listStatProc() {
	cmd := exec.Command("ssh", "root@192.168.102.76", "mysql -h192.168.102.45 -P9066 -uyh_test -pyh_test -e 'show @@processor\\G'")
	bs, err := cmd.Output()
	if nil != err {
		println(err.Error())
	}
	println(string(bs))
	reader := bufio.NewReader(bytes.NewReader(bs))
	var process = false
	var netIn, netOut uint64
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
	println("netIn:" + strconv.FormatUint(netIn, 10) + "\t netOut:" + strconv.FormatUint(netOut, 10))

}

func listStatTable() {
	cmd := exec.Command("ssh", "root@192.168.102.76", "mysql -h192.168.102.45 -P9066 -uyh_test -pyh_test -e 'show @@sql.stat.table\\G'")
	bs, err := cmd.Output()
	if nil != err {
		println(err.Error())
	}
	println(string(bs))
	reader := bufio.NewReader(bytes.NewReader(bs))
	var tableMap = make(map[string]map[string]string)
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
			tableMap[table] = make(map[string]string)
			continue
		}
		switch fields[0] {
		case "R":
			tableMap[table]["read"] = fields[1]
		case "W":
			tableMap[table]["write"] = fields[1]
		case "ALL_TTL_COUNT":
			ttls := fields[1][1:len(fields[1])-1]
			ttlsFields := strings.Split(ttls, ",")
			tableMap[table]["cobar_10"] = strings.TrimFunc(ttlsFields[0], unicode.IsSpace)
			tableMap[table]["cobar_50"] = strings.TrimFunc(ttlsFields[1], unicode.IsSpace)
			tableMap[table]["cobar_200"] = strings.TrimFunc(ttlsFields[2], unicode.IsSpace)
			tableMap[table]["cobar_1000"] = strings.TrimFunc(ttlsFields[3], unicode.IsSpace)
			tableMap[table]["cobar_2000"] = strings.TrimFunc(ttlsFields[4], unicode.IsSpace)
		case "NODE_TTL_COUNT":
			ttls := fields[1][2:len(fields[1])-1]
			ttlsFields := strings.Split(ttls, ",")
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[0], unicode.IsSpace)
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[1], unicode.IsSpace)
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[2], unicode.IsSpace)
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[3], unicode.IsSpace)
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[4], unicode.IsSpace)
		default:
			continue
		}
	}

}

func listStatSchema() {
	cmd := exec.Command("ssh", "root@192.168.102.76", "mysql -h192.168.102.45 -P9066 -uyh_test -pyh_test -e 'show @@sql.stat.schema\\G'")
	bs, err := cmd.Output()
	if nil != err {
		println(err.Error())
	}
	println(string(bs))
	reader := bufio.NewReader(bytes.NewReader(bs))
	var tableMap = make(map[string]map[string]string)
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
		if 0 == strings.Compare("SCHEMA", fields[0]) {
			table = fields[1]
			tableMap[table] = make(map[string]string)
			continue
		}
		switch fields[0] {
		case "R":
			tableMap[table]["read"] = fields[1]
		case "W":
			tableMap[table]["write"] = fields[1]
		case "ALL_TTL_COUNT":
			ttls := fields[1][1:len(fields[1])-1]
			ttlsFields := strings.Split(ttls, ",")
			tableMap[table]["cobar_10"] = strings.TrimFunc(ttlsFields[0], unicode.IsSpace)
			tableMap[table]["cobar_50"] = strings.TrimFunc(ttlsFields[1], unicode.IsSpace)
			tableMap[table]["cobar_200"] = strings.TrimFunc(ttlsFields[2], unicode.IsSpace)
			tableMap[table]["cobar_1000"] = strings.TrimFunc(ttlsFields[3], unicode.IsSpace)
			tableMap[table]["cobar_2000"] = strings.TrimFunc(ttlsFields[4], unicode.IsSpace)
		case "NODE_TTL_COUNT":
			ttls := fields[1][2:len(fields[1])-1]
			ttlsFields := strings.Split(ttls, ",")
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[0], unicode.IsSpace)
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[1], unicode.IsSpace)
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[2], unicode.IsSpace)
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[3], unicode.IsSpace)
			tableMap[table]["node_10"] = strings.TrimFunc(ttlsFields[4], unicode.IsSpace)
		default:
			continue
		}
	}
	for k,v := range tableMap{
		println("schema:"+k)
		fmt.Println(v)
	}
}
