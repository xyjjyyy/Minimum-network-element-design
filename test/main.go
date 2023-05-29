package main

import (
	"fmt"
	"log"
	"network/pdu"
	"network/utils"
	"strconv"
	"strings"
)

func main() {
	// fmt.Println(utils.BitToByte("000000000001110110000110110011100101100100100100001000111001101101000000100100011110001000111101101100000010010001010100011001010111001101110100001"))
	test1()
}

func test1() {
	frame := pdu.CreateFrame("Test8", "86ce5924239b", "4091e23db024")
	frameByte := frame.AddLocator()
	frame, err := pdu.SplitFrame(frameByte)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(frame.Data)
	}

}

func test2() {
	// s := "Hello World"
	// s0 := utils.ByteToBit([]byte(s))
	// fmt.Println("0:", s0)
	b := AddLocator("1111110011111011111")
	_, err := removeLocator(b)
	if err != nil {
		log.Println(err)
	}
}

// 移除定位符
func removeLocator(dataWithLocator []byte) ([]byte, error) {
	var dataByte []byte

	s := utils.ByteToBit(dataWithLocator)
	firstLoactor, secondLocator := -1, -1
	num := 0
	frequency := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
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
		return nil, fmt.Errorf("locator not found")
	}
	s = s[firstLoactor:secondLocator]

	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
			num++
		} else {
			num = 0
		}
		if num == 5 {
			s = s[:i+1] + s[i+2:]
			num = 0
		}
	}
	fmt.Println("s:", s)
	// 转换成byte类型
	dataByte = utils.BitToByte(s)
	return dataByte, nil
}

func AddLocator(s string) []byte {
	// 获得需要CRC32校验的部分 包含 帧类型 Mac地址 帧序号 数据
	num := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
			num++
		} else {
			num = 0
		}

		if num == 5 {
			s = s[:i+1] + "0" + s[i+1:]
			num = 0
			i++
		}
	}
	fmt.Println("l:", s)
	// 添加定位符
	s = "01111110" + s + "01111110"
	return utils.BitToByte(s)
}

func ByteToIP(b []byte) string {
	var IP string
	for i := 0; i < len(b); i++ {
		IP += strconv.Itoa(int(b[i])) + "."
	}
	return IP[:len(IP)-1]
}
func IPToByte(IP string) []byte {
	s := strings.Split(IP, ".")
	b := make([]byte, len(s))
	fmt.Println(s, len(s))

	for i := 0; i < len(s); i++ {
		t, _ := strconv.Atoi(s[i])
		b[i] = byte(int(t))
	}
	fmt.Println(len(b))
	return b
}
