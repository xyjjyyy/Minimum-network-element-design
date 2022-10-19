package main

import (
	"fmt"
	"log"
	"net"
	"network/host/pdu"
	"network/info"
	"network/switch/macTable"
)

// 基本配置
var (
	deviceId = info.DeviceId2
	layer    = info.LNK
)

// TODO:之后要改成高并发形式
func main() {
	netEle, err := info.CreateNetEleById(deviceId, layer)
	if err != nil {
		log.Fatal(err.Error())
	}

	addrLocal := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: netEle.CurLayer,
	}

	conn, err := net.ListenUDP("udp", addrLocal)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Printf("local server: <%s>\n", conn.LocalAddr().String())

	m := macTable.New()

	for {
		dataWithLocator := make([]byte, 4096)
		n, recvAddr, err := conn.ReadFromUDP(dataWithLocator)
		if err != nil {
			log.Fatal(n)
		}
		fmt.Printf("recv from port: <%d>\n", recvAddr.Port)

		frameRecv, err := pdu.SplitFrame(dataWithLocator)
		if err != nil {
			if frameRecv == nil {
				log.Println(err.Error())
				// 等待超时重传
				continue
			}
			// 返回ack err帧 错误直接丢弃因为无法处理
			// 等待超时处理
			log.Printf("交换机信息： frame %d err : %s,it will be abandoned\n", frameRecv.SerialNum, err.Error())
			continue
		}

		m.StudyFromSource(frameRecv.MacSource, recvAddr.Port)
		port, ok := m.QueryTableByMacAddress(frameRecv.MacTarget)
		if ok {
			netEle.SendToLower(conn, []int{port}, frameRecv.AddLocator(), 0)
		} else {
			netEle.SendToLower(conn, nil, frameRecv.AddLocator(), recvAddr.Port)
		}

	}
}
