package main

import (
	"fmt"
	"time"
)

func test(compress uint8, logstock string) {
	var zmclient htzmclient
	zmclient.compress = compress
	zmclient.logstock = logstock
	addr := "188.190.12.89:10002"
	err := zmclient.Connect(addr)
	if err != nil {
		fmt.Println("err Connect")
		return
	}
	err = zmclient.InitLogin()
	if err != nil {
		fmt.Println("err InitLogin")
		return
	}
	markets := []string{"SH", "SZ"}
	err = zmclient.RegMarkets(markets)
	if err != nil {
		fmt.Println("err RegMarkets")
		return
	}
	if 0x00f0 == ZMMMARKETRQST {
		fmt.Println("test")
	}
	zmclient.cDone = make(chan int)
	zmclient.sDone = make(chan int)

	go zmclient.ReadLoop()
	go zmclient.WriteLoop()
}

func main() {
	go test(1, "SH600837")
	go test(1, "SH000001")
	go test(1, "SZ000001")
	go test(1, "SZ399001")
	for i := 0; i < 0; i++ {
		go test(uint8(i%2), "SH600837")
	}
	for i := 0; i < 1000000; i++ {
		time.Sleep(10 * time.Second)
	}
	fmt.Println("end")
	return
}
