package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"network/host/pdu"
	"network/host/slidewindow"
	"network/info"
	"network/phy"
	"time"
)

// 基本配置
var (
	deviceId  = info.DeviceId1
	macSource = info.IdToMacTable[deviceId]
)

var (
	layer string
	id    int
	w     = slidewindow.New()

	netEle *info.NetEle
	conn   *net.UDPConn
	ch     chan struct{}
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
			log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
		}
		data := recvdata[:n]
		switch layer {
		case info.PHY:
			p := phy.New(phy.LittleHard)
			if sendAddr.Port == netEle.UpperLayer[0] {
				err := netEle.SendToLower(conn, nil, p.HandleFrame(data), sendAddr.Port)
				if err != nil {
					log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
				}
			} else {
				err := netEle.SendToUpper(conn, p.HandleFrame(data))
				if err != nil {
					log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
				}
			}
		case info.LNK:
			if sendAddr.Port == netEle.UpperLayer[0] {
				fmt.Println("recv: ", string(data))

				var frameList []*pdu.Frame
				frameList = []*pdu.Frame{pdu.CreateFrame(string(data), macSource, info.IdToMacTable[info.DeviceId2])}
				w.AddFrame(frameList)

			} else {
				frameRecv, err := pdu.SplitFrame(data)
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

				if frameRecv.FrameType == pdu.AckFrameType {
					err := w.SlideWindowForAck(frameRecv)
					if err != nil {
						fmt.Printf("ackSerialNum:%d\t%s\n", frameRecv.SerialNum, err.Error())
					} else {
						fmt.Printf("ackSerialNum:%d  已经被接受\n", frameRecv.SerialNum)
					}
				} else if frameRecv.FrameType == pdu.ErrFrameType {
					err := w.SlideWindowForErr(frameRecv)
					if err != nil {
						fmt.Printf("errorSerialNum:%d\t%s\n", frameRecv.SerialNum, err.Error())
					} else {
						fmt.Printf("errorSerialNum:%d  将被重发\n", frameRecv.SerialNum)
					}
				} else { // 数据帧
					if err = netEle.SendToUpper(conn, frameRecv.Data); err != nil {
						log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
					}
					// 原路返回ACK帧
					ackFrame := pdu.CreateAckFrameByte(frameRecv.SerialNum, frameRecv.MacTarget, frameRecv.MacSource)
					if err = netEle.SendToLower(conn, []int{sendAddr.Port}, ackFrame, 0); err != nil {
						log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
					}
				}
			}

		case info.APP:
			fmt.Println("RECV: ", string(data))

		}

		if string(data) == "QUIT" {
			ch <- struct{}{}
			return
		}
	}
}

func menu() {
	fmt.Println("input you choice：")
	fmt.Println("\t1- stdin\t2- file")
}

func HandleSend() {

	//keyboard.Open()
	//_, key, err := keyboard.GetKey()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// 按下ESC键退出
	//if key == keyboard.KeyEsc {
	//	break
	//}
	//keyboard.Close()

	//menu()

	//var opt int
	//fmt.Scanf("%d", &opt)

	//if opt == 1 {
	for i := 0; i < 20; i++ {
		var buf = fmt.Sprintf("Test-%d", i)
		if err := netEle.SendToLower(conn, nil, []byte(buf), 0); err != nil {
			log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
		}
	}

}

func main() {
	InitFlag()
	var err error
	netEle, err = info.CreateNetEleById(id, layer)
	if err != nil {
		log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
	}
	conn, err = netEle.GetServeConn()
	if err != nil {
		log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
	}
	ch = make(chan struct{})
	go HandleReceive()

	if layer == info.APP {
		go HandleSend()
	}
	if layer == info.LNK {
		go func() {
			for {
				frame, err := w.SlideWindowForSend()
				if frame == nil && err != nil {
					time.Sleep(time.Millisecond * 200)
					continue
				}
				if frame == nil && err != nil {
					log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
				}
				if frame == nil && err == nil {
					if w.FrameWaitAck.Len() != 0 {
						//fmt.Println("正在等待剩余的内容发送完毕")
					}
					time.Sleep(time.Second * 2)
					continue
				}
				fmt.Println(string(frame.Data), frame.FrameType, frame.MacSource)
				err = netEle.SendToLower(conn, nil, frame.AddLocator(), 0)
				if err != nil {
					log.Fatalf("[%d-%s] err:%v\n", id, layer, err)
				}
				go w.StartClock(frame)
			}
		}()
	}
	<-ch
	fmt.Println("end...")
}
