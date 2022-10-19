package main

import (
	"fmt"
	"log"
	"net"
	"network/info"
	"network/router/packet"
	routingtable "network/router/routingTable"
)

// 基本配置
var (
	deviceId = info.DeviceId3
	layer    = info.NET
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

	table := routingtable.New()

	for {
		packetByte := make([]byte, 4096)
		n, recvAddr, err := conn.ReadFromUDP(packetByte)
		if err != nil {
			log.Fatal(n)
		}
		fmt.Printf("recv from port: <%d>\n", recvAddr.Port)

		packet, err := packet.SplitFrameAndCreatePacket(packetByte)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		routingTable, ok := table.QueryTableByDesIP(packet.IPTarget)
		if ok {
			netEle.SendToLower(conn, []int{routingTable.Export}, packetByte, 0)
		} else {
			log.Println("路由表查询不到")
		}
	}
}
