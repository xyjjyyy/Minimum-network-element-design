package main

import (
	"fmt"
	"log"
	"net"
	"network/host/pdu"
	"network/host/slidewindow"
	"network/info"
	"time"
)

// 基本配置
var (
	deviceId = info.DeviceId1
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

	dataList := []string{"test1", "！", "这是一次尝试嘛2", "嘛！", "*&……%&*（（}3", "test4", "test5", "test6", "test7", "fasdfasdfasdfashfdhasdjfa8", "fasfsadfhasfhas 9", "7583945jkvjk89\"  10"}
	// dataList = []string{"test "}
	var frameList []*pdu.Frame

	// !这是实验数据： 从1发送到3
	for _, data := range dataList {
		frameList = append(frameList, pdu.CreateFrame(data, macSource, info.IdToMacTable[info.DeviceId3]))
	}

	w := slidewindow.New()
	w.AddFrame(frameList)

	// 并发收取

	go recv(conn, w)
	for {
		frame, err := w.SlideWindowForSend()
		if frame == nil && err != nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		if frame == nil && err != nil {
			log.Fatal(err.Error())
		}
		if frame == nil && err == nil {
			if w.FrameWaitAck.Len() != 0 {
				fmt.Println("正在等待剩余的内容发送完毕")
				time.Sleep(time.Millisecond * 100)
				continue
			}
			log.Println("全部发送完毕")
			break
		}
		err = netEle.SendToLower(conn, nil, frame.AddLocator(), 0)
		if err != nil {
			log.Fatal(err.Error())
		}
		go w.StartClock(frame)
	}
}

func recv(conn *net.UDPConn, w *slidewindow.Window) {
	for {
		dataWithLocator := make([]byte, 4096)
		n, _, err := conn.ReadFromUDP(dataWithLocator)
		if err != nil {
			log.Fatal(n)
		}

		frameRecv, err := pdu.SplitFrame(dataWithLocator)
		if err != nil {
			if frameRecv == nil {
				log.Println(err.Error())
				// 等待超时重传
				continue
			}
			// 返回ack err帧 错误直接丢弃因为无法处理
			// 等待超时处理
			log.Printf("frame %d err : %s,it will be abandoned\n", frameRecv.SerialNum, err.Error())
			continue
		}

		// 判断数据是否是这里的
		if frameRecv.MacTarget != macSource {
			log.Println("该数据不属于这里")
			continue
		}

		if frameRecv.FrameType == 1 {
			err := w.SlideWindowForAck(frameRecv)

			// TODO：错误展示出来但是不要中断程序 我们后来设计超时重传
			if err != nil {
				// wg.Add(1)
				fmt.Printf("ackSerialNum:%d\t%s\n", frameRecv.SerialNum, err.Error())
			} else {
				fmt.Printf("ackSerialNum:%d  已经被接受\n", frameRecv.SerialNum)
			}
		} else {
			err := w.SlideWindowForErr(frameRecv)

			// TODO：错误展示出来但是不要中断程序 我们后来设计超时重传
			if err != nil {
				// wg.Add(1)
				fmt.Printf("errorSerialNum:%d\t%s\n", frameRecv.SerialNum, err.Error())
			} else {
				fmt.Printf("errorSerialNum:%d  将被重发\n", frameRecv.SerialNum)
			}
		}
		// wg.Done()
	}
}
