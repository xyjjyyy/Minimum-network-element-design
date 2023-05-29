package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"network/info"
	"network/pdu"
	"network/phy"
	"network/slidewindow"
	"os"
	"time"
)

// 基本配置
var (
	deviceId  = info.DeviceId1
	macSource = info.IdToMacTable[deviceId]
)

var (
	layer      string
	id         int
	windowSend = slidewindow.New()
	windowRecv pdu.FrameHeap

	netEle *info.NetEle
	conn   *net.UDPConn
	ch     chan struct{}
	msg    chan string
)

func InitFlag() {
	// 将标志参数与变量绑定
	flag.StringVar(&layer, "layerLogo", "", "logo of this layer")
	flag.IntVar(&id, "deviceId", 0, "deviceID")
	// 解析命令行参数
	flag.Parse()
}

func HandleReceive() {
	for {
		recvdata := make([]byte, 1024)
		n, sendAddr, err := conn.ReadFromUDP(recvdata)
		if err != nil {
			msg <- fmt.Sprintf("error:[recv data] %v\n", err)
			return
		}
		data := recvdata[:n]
		switch layer {
		case info.PHY:
			p := phy.New(phy.LittleHard)
			msg <- fmt.Sprintf("info:[recv data] %v", data)
			if sendAddr.Port == netEle.UpperLayer[0] {
				err := netEle.SendToLower(conn, nil, p.HandleFrame(data), sendAddr.Port)
				if err != nil {
					msg <- fmt.Sprintf("error:[sendtolower] %v", err)
					return
				}
			} else {
				err := netEle.SendToUpper(conn, p.HandleFrame(data))
				if err != nil {
					msg <- fmt.Sprintf("error:[sendtoupper] %v", err)
					return
				}
			}
		case info.LNK:
			if sendAddr.Port == netEle.UpperLayer[0] {
				msg <- fmt.Sprintf("info: [recv data] %v", data)
				frame := pdu.CreateFrame(string(data), macSource, info.IdToMacTable[info.DeviceId2])
				windowSend.AddFrame(frame)
			} else {
				frameRecv, err := pdu.SplitFrame(data)
				// 等待超时处理
				if err != nil {
					msg <- fmt.Sprintf("error: [spliteFrame] %v\n", err)
					continue
				}

				if frameRecv.FrameType == pdu.AckFrameType { // `ack` 帧
					err := windowSend.SlideWindowForAck(frameRecv)
					if err != nil {
						msg <- fmt.Sprintf("error: [recv ackFrame] ackSerialNum:%d\t%v", frameRecv.SerialNum, err)
						return
					}
					msg <- fmt.Sprintf("info: [recv ackFrame] ackSerialNum:%d", frameRecv.SerialNum)
				} else { // `data` 帧
					windowRecv = append(windowRecv, frameRecv)
					if err := netEle.SendToUpper(conn, frameRecv.Data); err != nil {
						msg <- fmt.Sprintf("error: [recv ackFrame] ackSerialNum:%d", frameRecv.SerialNum)
						return
					}
					// 原路返回 `ack` 帧
					ackFrame := pdu.CreateAckFrameByte(frameRecv.SerialNum, frameRecv.MacTarget, frameRecv.MacSource)
					if err = netEle.SendToLower(conn, []int{sendAddr.Port}, ackFrame, 0); err != nil {
						msg <- fmt.Sprintf("error: [sendTolower return ack] ackSerialNum:%d", frameRecv.SerialNum)
						return
					}
				}
			}

		case info.APP:
			log.Printf("data[%s] send and recv successfully\n", string(data))
			msg <- fmt.Sprintf("important:[app recv data] `%s`", string(data))
		}
		if string(data) == "X" {
			msg <- "X"
			break
		}
	}
	<-ch
}

func HandleSend() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		op := scanner.Text()
		if op == "1" {
			for i := 0; i < 20; i++ {
				var buf = fmt.Sprintf("Test-%d", i)
				if err := netEle.SendToLower(conn, nil, []byte(buf), 0); err != nil {
					return
				}
				msg <- fmt.Sprintf("important:[app send data] `%s`", buf)
				time.Sleep(50 * time.Millisecond)
			}
		} else if op == "2" {
			fmt.Println("os:input your send")
			var str string
			fmt.Scan(&str)
			if err := netEle.SendToLower(conn, nil, []byte(str), 0); err != nil {
				log.Println(err)
				ch <- struct{}{}
				return
			}
			if str == "X" {
				msg <- "important:[app] send end flag `X`"
				ch <- struct{}{}
				return
			}
			msg <- fmt.Sprintf("important:[app send data] `%s`", str)
		} else {
			fmt.Println("please input again: ")
		}
	}
}

func LnkSendFromApp() {
	for {
		frame, err := windowSend.SlideWindowForSend()
		if err != nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		err = netEle.SendToLower(conn, nil, frame.AddLocator(), 0)
		if err != nil {
			msg <- fmt.Sprintf("error:[SendToLower] %v", err)
			return
		}
		go func() {
			info := windowSend.StartClock(frame)
			msg <- info
		}()
	}
}

func main() {
	InitFlag()
	msg = make(chan string, 20)
	ch = make(chan struct{})

	var err error
	netEle, err = info.CreateNetEleById(id, layer)
	if err != nil {
		log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
	}
	conn, err = netEle.GetServeConn()
	if err != nil {
		log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
	}
	go HandleLog()
	go HandleReceive()

	if layer == info.APP {
		go HandleSend()
	}
	if layer == info.LNK {
		go LnkSendFromApp()
	}
	<-ch
	close(ch)
	time.Sleep(time.Second)
	close(msg)
	fmt.Println("end...")
}

func HandleLog() {
	showAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 5050,
	}
	for buf := range msg {
		packBuf := fmt.Sprintf("device[%d]-layer[%s]:%s\n", deviceId, layer, buf)
		_, err := conn.WriteToUDP([]byte(packBuf), showAddr)
		if err != nil {
			log.Println("send to show center failed")
		}
	}
}
