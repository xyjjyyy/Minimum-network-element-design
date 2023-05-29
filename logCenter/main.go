package main

import (
	"log"
	"net"
	"os"
)

func main() {
	showAddr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 5050,
	}
	conn, err := net.ListenUDP("udp", showAddr)
	if err != nil {
		log.Fatal(err.Error())
	}
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	for {
		msg := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(msg)
		if err != nil {
			return
		}
		msg = msg[:n]
		s := string(msg)
		if string(s[len(s)-2]) == "X" {
			break
		}
		file.WriteString(s)
	}
}
