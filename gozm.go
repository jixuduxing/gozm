package main

import (
	"os"
	// "fmt"
	"time"
	"encoding/xml"
    "io/ioutil"
	"log"
	// "reflect"
)

func test(compress uint8, logstock string,addr string,logfilename string) {
	logfile,err := os.OpenFile(logfilename,os.O_RDWR|os.O_CREATE|os.O_TRUNC,0);
	if err!=nil {
        log.Printf("%s\r\n",err.Error())
        return
    }
	// defer logfile.Close();
	logger := log.New(logfile,"",log.Ldate|log.Ltime|log.Llongfile);
	var zmclient htzmclient
	zmclient.compress = compress
	zmclient.logstock = logstock
	zmclient.logfile = logger
	zmclient.logfile.Println("test!!")
	err = zmclient.Connect(addr)
	if err != nil {
		logger.Println("err Connect")
		return
	}
	logger.Println("Connect Success")
	err = zmclient.InitLogin()
	if err != nil {
		logger.Println("err InitLogin")
		return
	}
	logger.Println("InitLogin")
	markets := []string{"SZ","SH"}
	err = zmclient.RegMarkets(markets)
	if err != nil {
		logger.Println("err RegMarkets")
		return
	}
	logger.Println("RegMarkets")

	zmclient.cDone = make(chan int)
	zmclient.sDone = make(chan int)

	go zmclient.ReadLoop()
	go zmclient.WriteLoop()
}

type Result struct {
	Addr string		`xml:"addr"`
	Logpath string		`xml:"logpath"`
    Clients []Client `xml:"client"`
}

type Client struct {
    Name      string    `xml:"name,attr"`
    Logcode   string       `xml:"logcode,attr"`
    Compress  int    `xml:"compress,attr"`
}

func main() {
	content, err := ioutil.ReadFile("config.xml")
    if err != nil {
        log.Fatal(err)
	}
	var result Result
    err = xml.Unmarshal(content, &result)
    if err != nil {
        log.Fatal(err)
    }
	// go test(1, "SH600837")
	// go test(1, "SH000001")
	// go test(1, "SZ000001")
	// go test(1, "SZ399001")
	for _,client := range result.Clients {
		logfile := result.Logpath + client.Name +".log"
		go test( uint8(client.Compress), client.Logcode,result.Addr, logfile)
	}
	for i := 0; i < 1000000; i++ {
		time.Sleep(10 * time.Second)
	}
	log.Println("end")
	return
}
