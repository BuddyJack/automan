package middleware

import (
	"fmt"
	"os/exec"
	"strings"
	"unicode"
	"bufio"
	"bytes"
	"io"
	"strconv"
	"sync"
	"../model"
	"../ntos"
)

var (
	mysqlStatMap = make(map[string][2]*MysqlStat)
	mysqlLock    = new(sync.RWMutex)
)
//mysql metrics
const (
	MYSQL_STATUS  = "mysqladmin -h%s -P%d -u%s -p%s status"
	MYSQL_ESTATUS = "mysqladmin -h%s -P%d -u%s -p%s extended-status"
	Slave_Status  = "mysql -h%s -P%d -u%s -p%s -e 'show slave status\\G'"
	Live_Status   = "ps -ef |grep mysqld |grep %d |wc -l"
	Total_Link    = "mysqladmin -h%s -P%d -u%s -p%s processlist| wc -l"
	BaseAdmin_Cmd = "mysqladmin -h%s -P%d -u%s -p%s"
	Inner_Link    = " processlist | awk -F'[|]' '{print $4}' |sed '/^[[:space:]]*$/d;/^\\ Host[[:space:]]*$/d; s/localhost/127.0.0.1/g' | awk -F':' '{print $1}' |sort |uniq |wc -l"
	DB_Link       = " processlist | awk -F'[|]' '{print $5}' |sed '/^[[:space:]]*$/d;/^\\ db[[:space:]]*$/d' | sort | uniq | wc -l"

	Base_Cmd              = "mysql -h%s -P%d -u%s -p%s"
	TPS_CMD               = " -e \"select sum(VARIABLE_VALUE) from information_schema.GLOBAL_STATUS where variable_name in('Com_commit','Com_rollback','Handler_commit','Handler_rollback')\\G\"|grep VARIABLE_VALUE | awk '{print $2}'"
	SLAVE_HOST            = " -e \"show slave hosts\\G\" | wc -l"
	Slave_IO_Running      = " -e \"show slave status\\G\" | grep Slave_IO_Running | awk '{print $2}' | grep Yes | wc -l"
	Slave_SQL_Running     = " -e \"show slave status\\G\" | grep Slave_SQL_Running | awk '{print $2}' | grep Yes | wc -l"
	Seconds_Behind_Master = " -e \"show slave status\\G\" |  grep Seconds_Behind_Master | awk '{print $2}'"
)

type MySqlConfig struct {
	Host   string
	Port   uint64
	User   string
	Passwd string
}
type MysqlStat struct {
	Port                string
	Uptime              string
	Threads             string
	Qps                 string
	InnodbLogWaits      string
	InnodbRowLockWaits  string
	BytesReceived       string
	ByteSent            string
	Alive               string
	TotalConnects       string
	InnerConnects       string
	DBConnects          string
	Tps                 string
	SlaveHosts          string
	SlaveIORunning      string
	SlaveSQLRunning     string
	SecondsBehindMaster string
}

func listMysqlStatus(configs []MySqlConfig) (mysqlStats []*MysqlStat) {
	for _, config := range configs {

		mysqlStat := MysqlStat{Port: strconv.FormatUint(config.Port, 10)}
		cmdArg := fmt.Sprintf(MYSQL_STATUS, config.Host, config.Port, config.User, config.Passwd)
		cmd := exec.Command("sh", "-c", cmdArg)
		bs, err := cmd.Output()
		if nil != err {
		}
		fields := strings.Split(string(bs), "  ")
		for _, field := range fields {
			if kv := strings.Split(field, ":"); 2 == len(kv) {
				switch kv[0] {
				case "Uptime":
					println("uptime:" + strings.TrimFunc(kv[1], unicode.IsSpace))
					mysqlStat.Uptime = strings.TrimFunc(kv[1], unicode.IsSpace)
				case "Threads":
					println("threads:" + strings.TrimFunc(kv[1], unicode.IsSpace))
					mysqlStat.Threads = strings.TrimFunc(kv[1], unicode.IsSpace)
				case "Queries per second avg":
					println("qps:" + strings.TrimFunc(kv[1], unicode.IsSpace))
					mysqlStat.Qps = strings.TrimFunc(kv[1], unicode.IsSpace)
				}
			}
		}
		cmdArg = fmt.Sprintf(MYSQL_ESTATUS, "192.168.102.76", 3306, "root", "123456")
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		reader := bufio.NewReader(bytes.NewBuffer(bs))
		for {
			line, err := reader.ReadBytes('\n')
			if io.EOF == err {
				break
			} else if nil != err {
				continue
			}
			fields := strings.Fields(string(line))
			if 5 == len(fields) {
				switch fields[1] {
				case "Innodb_log_waits":
					println("log_waits:" + fields[3])
					mysqlStat.InnodbLogWaits = fields[3]
				case "Innodb_row_lock_waits":
					println("lock_waits:" + fields[3])
					mysqlStat.InnodbRowLockWaits = fields[3]
				case "Bytes_received":
					println("bytes_recv:" + fields[3])
					mysqlStat.BytesReceived = fields[3]
				case "Bytes_sent":
					println("bytes_send:" + fields[3])
					mysqlStat.ByteSent = fields[3]
				}
			}
		}

		cmdArg = fmt.Sprintf(Live_Status, 3306)
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, err = cmd.Output()
		res := string(bs)
		if count, err := strconv.ParseUint(res, 10, 64); nil != err && 0 < count {
			mysqlStat.Alive = "1"
			println("mysql is alive")
		} else {
			mysqlStat.Alive = "0"
			println("mysql is dead")
		}

		cmdArg = fmt.Sprintf(Total_Link, "192.168.102.76", 3306, "root", "123456")
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		res = strings.TrimFunc(string(bs), unicode.IsSpace)
		mysqlStat.TotalConnects = strings.TrimFunc(string(bs), unicode.IsSpace)
		println(res)

		cmdArg = fmt.Sprintf(BaseAdmin_Cmd, "192.168.102.76", 3306, "root", "123456") + Inner_Link
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		res = strings.TrimFunc(string(bs), unicode.IsSpace)
		mysqlStat.InnerConnects = strings.TrimFunc(string(bs), unicode.IsSpace)
		println(res)

		cmdArg = fmt.Sprintf(BaseAdmin_Cmd, "192.168.102.76", 3306, "root", "123456") + DB_Link
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		res = strings.TrimFunc(string(bs), unicode.IsSpace)
		mysqlStat.DBConnects = strings.TrimFunc(string(bs), unicode.IsSpace)
		println(res)

		cmdArg = fmt.Sprintf(Base_Cmd, "192.168.102.76", 3306, "root", "123456") + TPS_CMD
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		res = strings.TrimFunc(string(bs), unicode.IsSpace)
		mysqlStat.Tps = strings.TrimFunc(string(bs), unicode.IsSpace)
		println("tps:" + res)

		cmdArg = fmt.Sprintf(Base_Cmd, "192.168.102.76", 3306, "root", "123456") + SLAVE_HOST
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		res = strings.TrimFunc(string(bs), unicode.IsSpace)
		mysqlStat.SlaveHosts = strings.TrimFunc(string(bs), unicode.IsSpace)
		println("slave:" + res)

		cmdArg = fmt.Sprintf(Base_Cmd, "192.168.102.76", 3306, "root", "123456") + Slave_IO_Running
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		res = strings.TrimFunc(string(bs), unicode.IsSpace)
		mysqlStat.SlaveIORunning = strings.TrimFunc(string(bs), unicode.IsSpace)
		println("slave_io_running:" + res)

		cmdArg = fmt.Sprintf(Base_Cmd, "192.168.102.76", 3306, "root", "123456") + Slave_SQL_Running
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		res = strings.TrimFunc(string(bs), unicode.IsSpace)
		mysqlStat.SlaveSQLRunning = strings.TrimFunc(string(bs), unicode.IsSpace)
		println("slave_io_running:" + res)

		cmdArg = fmt.Sprintf(Base_Cmd, "192.168.102.76", 3306, "root", "123456") + Seconds_Behind_Master
		cmd = exec.Command("sh", "-c", cmdArg)
		bs, _ = cmd.Output()
		res = strings.TrimFunc(string(bs), unicode.IsSpace)
		mysqlStat.SecondsBehindMaster = strings.TrimFunc(string(bs), unicode.IsSpace)
		println("behind:" + res)

		mysqlStats = append(mysqlStats, &mysqlStat)

	}
	return
}

func (*MysqlStat) Metrics(configs []MySqlConfig) (metrics []*model.MetricValue) {
	mysqlStats := listMysqlStatus(configs)
	//rw
	mysqlLock.Lock()
	defer mysqlLock.Unlock()
	for _, oneStat := range mysqlStats {

		mysqlStatMap[oneStat.Port] = [2]*MysqlStat{oneStat, mysqlStatMap[oneStat.Port][0]}
		prevStat := mysqlStatMap[oneStat.Port][1]
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.uptime", Value: oneStat.Uptime, Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.threads", Value: oneStat.Threads, Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.qps", Value: oneStat.Qps, Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.Innodb_log_waits", Value: oneStat.InnodbLogWaits, Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.Innodb_row_lock_waits", Value: oneStat.InnodbRowLockWaits, Tags: map[string]string{"port": oneStat.Port}})
		var bsent, brecv, tps float64
		if (nil == prevStat) {
			bsent = 0
			brecv = 0
			tps = 0
		} else {
			var nowByteSent, prevByteSend, nowByteRecv, prevByteRecv, nowTps, PrevTps uint64 = 0, 0, 0, 0, 0, 0
			var err error
			if nowByteSent, err = strconv.ParseUint(oneStat.ByteSent, 10, 64); nil != err {
				nowByteSent = 0
			}
			if prevByteSend, err = strconv.ParseUint(prevStat.ByteSent, 10, 64); nil != err {
				prevByteSend = 0
			}
			bsent = float64(nowByteSent-prevByteSend) / float64(1024)
			if nowByteRecv, err = strconv.ParseUint(oneStat.BytesReceived, 10, 64); nil != err {
				nowByteRecv = 0
			}
			if prevByteRecv, err = strconv.ParseUint(prevStat.BytesReceived, 10, 64); nil != err {
				prevByteRecv = 0
			}
			brecv = float64(nowByteRecv-prevByteRecv) / float64(1024)
			if nowTps, err = strconv.ParseUint(oneStat.Tps, 10, 64); nil != err {
				nowTps = 0
			}
			if PrevTps, err = strconv.ParseUint(prevStat.Tps, 10, 64); nil != err {
				PrevTps = 0
			}
			tps = float64(nowTps-PrevTps) / float64(interval)
		}
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.Bytes_received", Value: ntos.F64toS2(brecv, 2), Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.Bytes_sent", Value: ntos.F64toS2(bsent, 2), Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.status", Value: oneStat.Alive, Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.total_connect", Value: oneStat.TotalConnects, Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.inner_connect", Value: oneStat.InnerConnects, Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.db_connect", Value: oneStat.DBConnects, Tags: map[string]string{"port": oneStat.Port}})

		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.tps", Value: ntos.F64toS2(tps, 2), Tags: map[string]string{"port": oneStat.Port}})
		metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.slave_host.count", Value: oneStat.SlaveHosts, Tags: map[string]string{"port": oneStat.Port}})

		if slaveCount, err := strconv.ParseUint(oneStat.SlaveHosts, 10, 64); nil != err && 0 < slaveCount {
			metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.slave_io_running", Value: oneStat.SlaveIORunning, Tags: map[string]string{"port": oneStat.Port}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.slave_sql_running", Value: oneStat.SlaveSQLRunning, Tags: map[string]string{"port": oneStat.Port}})
			metrics = append(metrics, &model.MetricValue{Endpoint: "mysql", Metric: "mysql.second_behind_master", Value: oneStat.SecondsBehindMaster, Tags: map[string]string{"port": oneStat.Port}})
		}
	}

	return
}
