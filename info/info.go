package info

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// * 公共信息
const (
	Model1 = 1 //主机
	Model2 = 2 //交换机
	Model3 = 3 //路由器
)

const (
	DeviceId1 = 0
	DeviceId2 = 1
	DeviceId3 = 2
	DeviceId4 = 3
	DeviceId5 = 4
)

const (
	APP = "APP"
	NET = "NET"
	LNK = "LNK"
	PHY = "PHY"
)

// * 数据链路层验证   每个主机只有一个相对应的
var IdToMacTable = map[int]string{
	DeviceId1: "86ce5924239b",
	DeviceId2: "4091e23db024",
	DeviceId3: "a23bd2c5d3a8",
}

// NetEle * 每一个网元内部的连接关系
/* 也可以说是一个网元的代表
如果初始化为 nil 或 0  代表不存在该层
对应的数据都是端口号 Port
*/
type NetEle struct {
	DeviceId   int
	DeviceType int
	CurLayer   int
	UpperLayer []int
	LowerLayer []int
}

// CreateNetEleById
// 现在是简单的交换机模型
func CreateNetEleById(id int, curLayer string) (*NetEle, error) {
	switch id {
	case DeviceId1:
		switch curLayer {
		case PHY:
			return &NetEle{
				DeviceId:   id,
				CurLayer:   10000,
				UpperLayer: []int{10010},
				LowerLayer: []int{11000},
			}, nil
		case LNK:
			return &NetEle{
				DeviceId:   id,
				DeviceType: Model1,
				CurLayer:   10010,
				UpperLayer: []int{10020},
				LowerLayer: []int{10000},
			}, nil
		case APP:
			return &NetEle{
				DeviceId:   id,
				CurLayer:   10020,
				UpperLayer: nil,
				LowerLayer: []int{10010},
			}, nil
		default:
			return nil, fmt.Errorf("DeviceId %d 不存在该层 %s", id, curLayer)
		}
	case DeviceId2:
		switch curLayer {
		case PHY:
			return &NetEle{
				DeviceId:   id,
				CurLayer:   11000,
				UpperLayer: []int{11010},
				LowerLayer: []int{10000},
			}, nil
		case LNK:
			return &NetEle{
				DeviceId:   id,
				DeviceType: Model1,
				CurLayer:   11010,
				UpperLayer: []int{11020},
				LowerLayer: []int{11000},
			}, nil
		case APP:
			return &NetEle{
				DeviceId:   id,
				CurLayer:   11020,
				UpperLayer: nil,
				LowerLayer: []int{11010},
			}, nil
		default:
			return nil, fmt.Errorf("DeviceId %d 不存在该层 %s", id, curLayer)
		}
	case DeviceId3:
		switch curLayer {
		case LNK:
			return &NetEle{
				DeviceId:   id,
				DeviceType: Model1,
				CurLayer:   12100,
				UpperLayer: nil,
				LowerLayer: []int{12000},
			}, nil
		case PHY:
			return &NetEle{
				DeviceId:   id,
				CurLayer:   12100,
				UpperLayer: nil,
				LowerLayer: []int{12000},
			}, nil
		default:
			return nil, fmt.Errorf("DeviceId %d 不存在该层 %s", id, curLayer)
		}
	default:
		return nil, fmt.Errorf("DeviceId %d 不存在", id)
	}
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
	return conn, err
}

// SendToUpper 对应上层只有一个 不过还是这样写
func (n *NetEle) SendToUpper(conn *net.UDPConn, data []byte) error {
	if n.UpperLayer == nil {
		return errors.New("本网元该层没有上层地址")
	}
	addrList := getAddr(n.UpperLayer)
	for i := 0; i < len(addrList); i++ {
		_, err := conn.WriteToUDP(data, addrList[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// SendToLower 如果index为nil 则发送给全部下层除了recvIndex
// index!=nil 则正常发送给index[] 全部的端口
// 设置recvPort为0可以忽略这个参数
func (n *NetEle) SendToLower(conn *net.UDPConn, port []int, data []byte, recvPort int) error {
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
		_, err := conn.WriteToUDP(data, addrList[i])
		if err != nil {
			return err
		}
	}

	return nil
}
