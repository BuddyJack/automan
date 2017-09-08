package base

import "../model"

type BasicStat interface {
	Metrics() []*model.MetricValue
}

func Metrics() {
	// cpu call
	cpu := &CpuUsage{}
	cpu.Metrics()
	//df call
	df := &DeviceUsage{}
	df.Metrics()
	//disk call
	disk := &DiskStat{}
	disk.Metrics()
	//load call
	load := &LoadAvg{}
	load.Metrics()
	//mem call
	mem := &MemUsage{}
	mem.Metrics()
	//net call
	net := &NetStat{}
	net.Metrics()
	//port call
	port := &PortStat{}
	port.Metrics()
	//proc call
	proc := &ToProcStat{}
	proc.Metrics()
	//tcp call
	tcp := &TcpConnStat{}
	tcp.Metrics()
}
