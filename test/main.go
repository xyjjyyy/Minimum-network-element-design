package main

import (
	"fmt"
	"math/rand"
	"network/utils"
	"strconv"
	"strings"
	"time"
)

func main() {
	// t := utils.ByteToBit([]byte{127, 1})
	// fmt.Println(t)
	// s := "86CE59A1239B"
	// s = strings.ToLower(s)
	// // s="1234567891233"
	// b := utils.Hex12ToBit48(s)
	// fmt.Println(b)

	// h := utils.Bit48ToHex12(b)
	// fmt.Println(h)

	fmt.Println(utils.Code([]byte{127, 27}))
	IP := ByteToIP([]byte{123, 234, 23, 234})
	b := IPToByte(IP)
	fmt.Println(IP, b)
}

func test() string {
	var s string

	rand.Seed(time.Now().UnixMicro())
	for len(s) < 12 {
		randNum := rand.Intn('f' - '0')

		if randNum < 'a'-'0' && randNum > '9'-'0' {
			continue
		}
		t := rune('0' + randNum)
		s += string(t)
	}
	return s
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
