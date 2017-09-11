package transfer

import (
	"net"
	"../model"
	"fmt"
	"encoding/json"
)

//udp client

var (
	connection *net.UDPConn
)

func InitConn(udpAddr string) {
	udp, err := net.ResolveUDPAddr("udp4", udpAddr)
	if nil != err {
		panic("failed to create udp channel to server " + udpAddr + ",error " + err.Error())
	}
	connection, err = net.DialUDP("udp4", nil, udp)
	if nil != err {
		panic("failed to create udp channel to server " + udpAddr + ",error " + err.Error())
	}
	connection.SetWriteBuffer(1024 * 128)
}

func WriteUdpMsg(value model.MetricValue) {
	msgBytes, err := json.Marshal(&value)
	if nil != err {
		panic(err)
	}
	connection.Write(msgBytes)
}
