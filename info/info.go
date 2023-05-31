package info

import (
	"errors"
	"log"
	"net"
	"network/slidewindow"
	macTable "network/switch"
	"time"
)

const (
	DeviceId1 = iota
	DeviceId2
	DeviceId3
	DeviceId4
	DeviceId5
)

// 数据链路层验证   每个主机只有一个相对应的
var IdToMacTable = map[int]string{
	DeviceId1: "86ce5924239b",
	DeviceId2: "4091e23db024",
	DeviceId3: "a23bd2c5d3a8",
}

// NetEle
type NetEle struct {
	// 类型标识
	DeviceId   int
	DeviceType string
	Layer      string

	// 对应端口
	CurLayer   int
	UpperLayer []int
	LowerLayer []int

	// LNK层独有
	mac        string
	windowSend *slidewindow.Window
	macTable   *macTable.MacTable

	conn *net.UDPConn

	msg chan string //收集必要的log信息
}

func getAddr(port []int) []*net.UDPAddr {
	var addrList []*net.UDPAddr
	for i := 0; i < len(port); i++ {
		addr := &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port[i],
		}
		addrList = append(addrList, addr)
	}
	return addrList
}

func (n *NetEle) GetServeConn() (*net.UDPConn, error) {
	addrLocal := getAddr([]int{n.CurLayer})
	conn, err := net.ListenUDP("udp", addrLocal[0])
	if err != nil {
		log.Fatal(err.Error())
	}
	n.conn = conn
	return conn, err
}

// SendToUpper 对应上层只有一个 不过还是这样写
func (n *NetEle) SendToUpper(data []byte) error {
	if n.UpperLayer == nil {
		return errors.New("本网元该层没有上层地址")
	}
	addrList := getAddr(n.UpperLayer)
	for i := 0; i < len(addrList); i++ {
		_, err := n.conn.WriteToUDP(data, addrList[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// SendToLower 如果port数组为nil 则发送给全部下层除了recvPort
// port!=nil 则正常发送给port全部的端口 类似于发送给指定的端口
// 如果设置recvPort=0 则可以忽略这个参数
func (n *NetEle) SendToLower(port []int, data []byte, recvPort int) error {
	var portSend []int
	if port == nil {
		for _, p := range n.LowerLayer {
			if p != recvPort {
				portSend = append(portSend, p)
			}
		}
	} else {
		portSend = port
	}

	addrList := getAddr(portSend)
	for i := 0; i < len(addrList); i++ {
		_, err := n.conn.WriteToUDP(data, addrList[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// 滑动窗口持续发送
func (n *NetEle) lnkSendFromApp() {
	for {
		frame, err := n.windowSend.SlideWindowForSend()
		if err != nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		err = n.SendToLower(nil, frame.AddLocator(), 0)
		if err != nil {
			return
		}
		go func() {
			n.windowSend.StartClock(frame)
		}()
	}
}
