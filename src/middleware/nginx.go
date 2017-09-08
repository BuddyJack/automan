package middleware

import "github.com/hpcloud/tail"

//jmx  metrics

type NginxStat struct {
}

func (*NginxStat) Metrics() {
	t, err := tail.TailFile("/Users/jack/test.log", tail.Config{ReOpen: true, Follow: true, MaxLineSize: 500, Location: &tail.SeekInfo{Offset: 0, Whence: 2}}, )
	defer t.Cleanup()
	if nil != err {
		panic(err)
	}
	for one := range t.Lines {
		println(one.Text)
	}
}
