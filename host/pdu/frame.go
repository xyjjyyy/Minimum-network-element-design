package pdu

import (
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"network/utils"
	"strconv"
)

func New() *Frame {
	return &Frame{
		Locator: "01111110",
		// TODO：有机会可以改为动态调整
		Duration: 500, //设置超时时间500ms
		Done:     make(chan struct{}),
	}
}

// GetCRC32  返回CRC32校验码
func (f *Frame) GetCRC32() {
	s := f.getDataNeedCRC()
	var dataNeedCRC []byte
	l := len(s)
	//dataNeedCRC := []byte(s)
	// 转化为我们需要的 []byte 数组类型
	for i := 0; i < l; i += 8 {
		var b int
		for j := 0; j < 8 && i+j < l; j++ {
			if string(s[i+j]) == "1" {
				b += 1 << (7 - j)
			}
		}
		dataNeedCRC = append(dataNeedCRC, byte(b))
	}
	log.Println("dataNeedCRC", dataNeedCRC)
	crc := crc32.ChecksumIEEE(dataNeedCRC)
	log.Println("f:", crc)
	f.CRC32 = fmt.Sprintf("%032s", strconv.FormatUint(uint64(crc), 2))
	fmt.Println("CRC#32:", f.CRC32)
}

// * 获得 需要CRC32校验的部分（包含 帧类型 帧序号 数据 ）的对应二进制字符串
func (f *Frame) getDataNeedCRC() string {
	// 添加帧类型(1)
	s := fmt.Sprintf("%08s", strconv.FormatInt(int64(f.FrameType), 2))
	//log.Println(len(s))
	// 添加帧序号(1)
	s += fmt.Sprintf("%08s", strconv.FormatUint(f.SerialNum, 2))
	//log.Println(len(s))
	// 添加macAddress(12=6+6)
	s += utils.Hex12ToBit48(f.MacSource)
	//log.Println(len(s))
	s += utils.Hex12ToBit48(f.MacTarget)
	//log.Println(len(s))
	// 添加data数据部分(>=0)
	s += utils.ByteToBit(f.Data)
	fmt.Println(string(f.Data))
	//log.Println(len(s))
	return s
}

// AddLocator 加上定位符 并返回最终的生成的帧(byte类型)
func (f *Frame) AddLocator() []byte {
	// 获得需要CRC32校验的部分 包含 帧类型 帧序号 数据
	s := f.getDataNeedCRC()
	// 添加 CRC32 校验码(4)
	s += f.CRC32
	num := 0
	l := len(s)
	for i := 0; i < l; i++ {
		if num == 5 {
			s = s[:i] + "0" + s[i:]
			num = 0
		}
		if string(s[i]) == "1" {
			num++
		} else {
			num = 0
		}
	}

	// 添加定位符(2)
	s = f.Locator + s + f.Locator

	log.Println("S2:", s, len(s))
	log.Println(utils.BitToByte(s))
	return utils.BitToByte(s)
}

// SplitFrame *将 frame 进行分割，返回我们的帧结构(*frame)和error
func SplitFrame(dataWithLocator []byte) (*Frame, error) {
	fullFrame, err := removeLocator(dataWithLocator)
	if err != nil {
		return nil, err
	}
	// fmt.Println("fullFrame:", fullFrame)
	l := len(fullFrame)
	if l < fullFrameMinLength {
		return nil, errors.New("待解析帧长度不符合最低长度")
	}

	//TODO:可以优化代码 更少量的代码
	frameSerialNum := utils.Code(fullFrame[1:2])
	dataWithoutCrc := fullFrame[:l-4]
	crc32Uint := utils.Code(fullFrame[l-4 : l])
	frameType := utils.Code(fullFrame[:1])
	macSource := utils.Bit48ToHex12(utils.ByteToBit(fullFrame[2:8]))
	macTarget := utils.Bit48ToHex12(utils.ByteToBit(fullFrame[8:14]))
	data := fullFrame[14 : l-4]
	log.Println("crcbyte:", fullFrame[l-4:l])
	log.Println("data:", data)
	log.Println("datawithoutCrc:", dataWithoutCrc)

	crc := crc32.ChecksumIEEE(dataWithoutCrc)
	log.Println(crc32Uint, crc)
	if uint64(crc) != crc32Uint {
		// 若CRC32校验出现错误 立马构建ErrFrame并返回
		frame := CreateErrFrame(frameSerialNum, macSource, macTarget)
		return frame, fmt.Errorf("serialNum:%d-->CRC校验未通过", frameSerialNum)
	}

	frame := New()
	frame.Data = data
	frame.FrameType = int(frameType)
	frame.SerialNum = frameSerialNum
	frame.MacSource = macSource
	frame.MacTarget = macTarget
	frame.GetCRC32()

	return frame, nil
}

// * 移除定位符
func removeLocator(dataWithLocator []byte) ([]byte, error) {
	var dataByte []byte
	var s string

	s += utils.ByteToBit(dataWithLocator)
	//fmt.Println("R1: ", s)
	//TODO:要考虑帧定位符找不到或者超过两个情况（已处理）
	firstLoactor, secondLocator := 0, 0
	num := 0
	frequency := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == "1" {
			num++
		} else {
			num = 0
		}
		if num == 6 {
			frequency++
			i += 2
			if frequency == 1 {
				firstLoactor = i
			} else if frequency == 2 {
				secondLocator = i - 8
			}
			num = 0
		}
	}
	if frequency != 2 {
		fmt.Println(s)
		return nil, fmt.Errorf("定位符出错")
	}
	s = s[firstLoactor:secondLocator]
	//fmt.Println("R2:", s)

	s += "0"
	for i := 0; i < len(s); i++ {
		if string(s[i]) == "1" {
			num++
		} else {
			num = 0
		}
		if num == 5 {
			s = s[:i+1] + s[i+2:]
		}
	}
	l := len(s)
	//fmt.Println("R3:", s)

	// 转换成byte类型
	for i := 0; i < l; i += 8 {
		var b int
		for j := 0; j < 8 && i+j < l; j++ {
			if string(s[i+j]) == "1" {
				b += 1 << (7 - j)
			}
		}

		// 后面不足八位的全是当初加上定位符的时候补零多的
		// 应当删除 不加入
		if l-i >= 8 {
			dataByte = append(dataByte, byte(b))
		}
	}

	return dataByte, nil
}
