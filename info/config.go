package info

import (
	"bufio"
	"fmt"
	"log"
	"network/pdu"
	"network/phy"
	"network/slidewindow"
	mactable "network/switch"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	HOST   = "HOST"
	SWITCH = "SWITCH"
	ROUTER = "ROUTER"
	APP    = "APP"
	LNK    = "LNK"
	PHY    = "PHY"
)

type Config struct {
	ELes []*NetEle
	Msg  chan string
	done chan struct{}
}

func MakeConfig() (*Config, error) {
	cfg := &Config{
		ELes: []*NetEle{},
		Msg:  make(chan string, 40),
		done: make(chan struct{}),
	}

	file, err := os.Open("./info/config.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建一个Scanner来读取文件内容
	scanner := bufio.NewScanner(file)

	var deviceId int
	var deviceType string
	// 逐行读取文件内容
	for scanner.Scan() {
		line := scanner.Text()
		if line == "-----------------" || line[0] == '#' {
			continue
		}
		arr := strings.Split(line, ":")
		if arr[0] == HOST || arr[0] == SWITCH || arr[0] == ROUTER {
			m, _ := strconv.Atoi(arr[1])
			deviceId = m
			deviceType = arr[0]
			continue
		}
		e := &NetEle{
			DeviceId:   deviceId,
			DeviceType: deviceType,
			Layer:      arr[0],
		}

		ports := strings.Split(arr[1], "--")
		if ports[0] != "" {
			port, _ := strconv.Atoi(ports[0])
			e.CurLayer = port
		}
		if ports[1] != "" {
			portsArr := strings.Split(ports[1], "&&")
			var portsNum []int
			for _, port := range portsArr {
				p, _ := strconv.Atoi(port)
				portsNum = append(portsNum, p)
			}
			e.LowerLayer = portsNum
		}
		if ports[2] != "" {
			portsArr := strings.Split(ports[2], "&&")
			var portsNum []int
			for _, port := range portsArr {
				p, _ := strconv.Atoi(port)
				portsNum = append(portsNum, p)
			}
			e.UpperLayer = portsNum
		}

		if len(arr) == 3 {
			e.mac = arr[2]
		}
		if arr[0] == LNK {
			if deviceType == SWITCH {
				e.macTable = mactable.New()
			} else if deviceType == HOST {
				e.windowSend = slidewindow.New()
				go e.lnkSendFromApp()
			}
		}

		e.GetServeConn()
		e.msg = make(chan string, 20)
		cfg.ELes = append(cfg.ELes, e)
	}

	// 检查是否发生了读取错误
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 集中处理日志
	go cfg.handleLog()

	return cfg, nil
}

func (cfg *Config) HandleReceive() {
	for _, ele := range cfg.ELes {
		go func(e *NetEle) {
			for {
				recvdata := make([]byte, 1024)
				n, sendAddr, err := e.conn.ReadFromUDP(recvdata)
				if err != nil {
					e.msg <- fmt.Sprintf("error:[recv data] %v\n", err)
					return
				}
				data := recvdata[:n]
				// e.msg <- fmt.Sprintf("info:[recv data] %v", data)
				switch e.Layer {
				case PHY:
					p := phy.New(phy.LittleHard)
					if sendAddr.Port == e.UpperLayer[0] {
						err := e.SendToLower(nil, p.HandleFrame(data), sendAddr.Port)
						if err != nil {
							e.msg <- fmt.Sprintf("error:[sendtolower] %v", err)
							return
						}
					} else {
						err := e.SendToUpper(p.HandleFrame(data))
						if err != nil {
							e.msg <- fmt.Sprintf("error:[sendtoupper] %v", err)
							return
						}
					}
				case LNK:
					if e.DeviceType == HOST { // 主机设备
						if sendAddr.Port == e.UpperLayer[0] {
							frame := pdu.CreateFrame(string(data), e.mac, IdToMacTable[1])
							e.windowSend.AddFrame(frame)
						} else {
							frameRecv, err := pdu.SplitFrame(data)
							// 等待超时处理
							if err != nil {
								e.msg <- fmt.Sprintf("error: [spliteFrame] %v\n", err)
								return
							}
							if frameRecv.MacTarget != e.mac {
								e.msg <- fmt.Sprintf("warning:not match %s %s", frameRecv.MacTarget, e.mac)
								break
							}
							if frameRecv.FrameType == pdu.AckFrameType { // `ack` 帧
								err := e.windowSend.SlideWindowForAck(frameRecv)
								if err != nil {
									e.msg <- fmt.Sprintf("error: [recv ackFrame] ackSerialNum:%d\t%v", frameRecv.SerialNum, err)
									return
								}
								e.msg <- fmt.Sprintf("info: [recv ackFrame] ackSerialNum:%d", frameRecv.SerialNum)
							} else { // `data` 帧
								// windowRecv = append(windowRecv, frameRecv)
								if err := e.SendToUpper(frameRecv.Data); err != nil {
									e.msg <- fmt.Sprintf("error: [recv ackFrame] ackSerialNum:%d", frameRecv.SerialNum)
									return
								}
								// 原路返回 `ack` 帧
								ackFrame := pdu.CreateAckFrameByte(frameRecv.SerialNum, frameRecv.MacTarget, frameRecv.MacSource)
								if err = e.SendToLower([]int{sendAddr.Port}, ackFrame, 0); err != nil {
									e.msg <- fmt.Sprintf("error: [sendTolower return ack] ackSerialNum:%d", frameRecv.SerialNum)
									return
								}
							}
						}
					} else if e.DeviceType == SWITCH { //
						frameRecv, err := pdu.SplitFrame(data)
						// 等待超时处理
						if err != nil {
							e.msg <- fmt.Sprintf("error: [spliteFrame] %v\n", err)
							continue
						}

						e.macTable.StudyFromSource(frameRecv.MacSource, sendAddr.Port)
						port, ok := e.macTable.GetPort(frameRecv.MacTarget)
						if !ok {
							if err = e.SendToLower(nil, data, sendAddr.Port); err != nil {
								e.msg <- fmt.Sprintf("error:[sendtolower all port] %v", err)
								return
							}
						} else {
							if err = e.SendToLower([]int{port}, data, sendAddr.Port); err != nil {
								e.msg <- fmt.Sprintf("error:[sendtolower all port] %v", err)
								return
							}
						}
					} else {
						e.msg <- fmt.Sprintf("warning:not support %s", e.DeviceType)
					}

				case APP:
					if string(data) == "X" {
						e.msg <- "X"
						cfg.done <- struct{}{}
					}
					log.Printf("data[%s] send and recv successfully\n", string(data))
					e.msg <- fmt.Sprintf("important:[app recv data] `%s`", string(data))
				}
			}
		}(ele)
	}
}

func (cfg *Config) HandleSend() {
	scanner := bufio.NewScanner(os.Stdin)
	e := cfg.ELes[0]
	for scanner.Scan() {
		op := scanner.Text()
		if op == "1" {
			for i := 0; i < 20; i++ {
				var buf = fmt.Sprintf("Test-%d", i)
				if err := e.SendToLower(nil, []byte(buf), 0); err != nil {
					return
				}
				e.msg <- fmt.Sprintf("important:[app send data] `%s`", buf)
				time.Sleep(50 * time.Millisecond)
			}
		} else if op == "2" {
			fmt.Println("os:input your send")
			var str string
			fmt.Scan(&str)
			if err := e.SendToLower(nil, []byte(str), 0); err != nil {
				log.Println(err)
				return
			}
			if str == "X" {
				e.msg <- "important:[app] send end flag `X`"
				return
			}
			e.msg <- fmt.Sprintf("important:[app send data] `%s`", str)
		} else {
			fmt.Println("please input again: ")
		}
	}
}

func (cfg *Config) handleLog() {
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, ele := range cfg.ELes {
		go func(e *NetEle) {
			for buf := range e.msg {
				if buf == "X" {
					cfg.Msg <- buf
					return
				}
				packBuf := fmt.Sprintf("device[%d]-layer[%s]:%s\n", e.DeviceId, e.Layer, buf)
				cfg.Msg <- packBuf
			}
		}(ele)
	}
	for msg := range cfg.Msg {
		if msg == "X" {
			return
		}
		file.WriteString(msg)
	}
}

func (cfg *Config) Waiting() {
	<-cfg.done
	// 给节点缓冲时间
	time.Sleep(2 * time.Second)
}
