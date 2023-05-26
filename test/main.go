package main

import (
	"fmt"
	"log"
	"math/rand"
	"network/host/pdu"
	"strconv"
	"strings"
	"time"
)

func main() {
	test2()
}

func test2() {
	frame := pdu.CreateAckFrameByte(8, "123456789012", "123456789012")

	if f, err := pdu.SplitFrame(frame); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(string(f.Data))
	}
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
