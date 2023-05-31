package main

import (
	"fmt"
	"log"
	"network/info"
)

func main() {
	var err error
	cfg, err := info.MakeConfig()
	if err != nil {
		log.Fatalf("MakeConfig err:%v\n", err)
	}

	cfg.HandleReceive()
	cfg.HandleSend()
	cfg.Waiting()
	fmt.Println("end...")
}
