package main

import (
	"fmt"
	"log"
	"net"
	"network/host/pdu"
	"network/info"
)

// 基本配置
var (
	deviceId = info.DeviceId3
	// model     = info.Model1 //TODO：等会考虑
	layer     = info.LNK
	macSource = info.IdToMacTable[deviceId]
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
	fmt.Printf("local server: <%s>", conn.LocalAddr().String())

	defer conn.Close()
	recvNum := 0
	for {
		dataWithLocator := make([]byte, 4096)
		n, _, err := conn.ReadFromUDP(dataWithLocator)
		if err != nil {
			log.Fatal(n)
		}

		recvNum++
		frame, err := pdu.SplitFrame(dataWithLocator)
		if err != nil {
			if frame == nil {
				log.Println(err.Error())
				// 等待超时重传
				continue
			}
			log.Printf("frame %d err : %s,it will be retransmitted\n", frame.SerialNum, err.Error())
		}
		// 判断数据是否是这里的
		if frame.MacTarget != macSource {
			log.Println("数据认证失败，直接丢弃")
			continue
		}

		fmt.Printf("recvNum=%d\n", recvNum)

		var frameByteReTransmit = make([]byte, 1024)

		if frame.FrameType == 0 {
			fmt.Printf("SerialNum:%d\trecv data :%s\n", frame.SerialNum, frame.Data)
			frameByteReTransmit = pdu.CreateAckFrameByte(frame.SerialNum, frame.MacSource, frame.MacTarget)
			log.Printf("The ackFrame of `%d` return\n", frame.SerialNum)
		} else {
			frameByteReTransmit = pdu.CreateErrFrameByte(frame.SerialNum, frame.MacSource, frame.MacTarget)
			log.Printf("The errFrame of `%d` return\n", frame.SerialNum)
		}

		err = netEle.SendToLower(conn, nil, frameByteReTransmit, 0)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
