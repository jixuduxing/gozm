package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type htzmclient struct {
	conn          net.Conn
	cDone, sDone  chan int
	seq           uint32
	totalrecvsize int
	compress      uint8
	logstock      string
}

func (sel *htzmclient) Connect(addr string) error {
	var err error
	sel.conn, err = net.Dial("tcp", addr)
	return err
}
func (sel htzmclient) handleStatic(dReader *bytes.Reader) {
	var StaticInfoHeader ZmmStaticInfoHeader
	binary.Read(dReader, binary.LittleEndian, &StaticInfoHeader)
	fmt.Println("StaticInfoHeader:", StaticInfoHeader, string(StaticInfoHeader.Marketcode[:]))
	infobuf := make([]byte, StaticInfoHeader.Itemsize)
	if StaticInfoHeader.Itemtype == ZMMZQDMINFO {
		var info ZQDMInfo
		for i := 0; i < int(StaticInfoHeader.Itemcount); i++ {
			n, _ := dReader.Read(infobuf)
			if n != int(StaticInfoHeader.Itemsize) {
				fmt.Println("n!= StaticInfoHeader.Itemsize", n, StaticInfoHeader.Itemsize)
				break
			}
			reader2 := bytes.NewReader(infobuf)
			binary.Read(reader2, binary.LittleEndian, &info)
			strcode := info.getCode()
			if strcode == sel.logstock {
				fmt.Println(string(info.Name[:12]), string(info.Pinyinname[:8]), info)
			}
		}
	}

}

func (sel htzmclient) handlehqkz(dReader *bytes.Reader) {
	var HqkzHeader ZmmHqkzHeader
	binary.Read(dReader, binary.LittleEndian, &HqkzHeader)
	fmt.Println("HqkzHeader:", HqkzHeader, string(HqkzHeader.Marketcode[:]))
	infobuf := make([]byte, HqkzHeader.Itemsize)
	var info CodeIf
	switch HqkzHeader.Itemtype {
	case ZMMHQKZSSHQ:
		info = new(SSHQ)
		// fmt.Println("SSHQ size:", binary.Size(info), HqkzHeader.Itemsize)
	case ZMMHQKZSSZS:
		info = new(SSZS)
	case ZMMHQKZSTEPORDER:
		info = new(StepOrderNative)
	case ZMMHQKZSTEPTRADE:
		info = new(StepTradeNative)
	case ZMMHQKZORDERQUEUE:
		info = new(OrderQueue)
	default:
		return
	}
	for i := 0; i < int(HqkzHeader.Itemcount); i++ {
		n, _ := dReader.Read(infobuf)
		if n != int(HqkzHeader.Itemsize) {
			fmt.Println("n!= HqkzHeader.Itemsize", n, HqkzHeader.Itemsize)
			break
		}
		reader2 := bytes.NewReader(infobuf)
		binary.Read(reader2, binary.LittleEndian, info)
		strcode := info.getCode()
		// fmt.Print(strcode)
		if strcode == sel.logstock {
			fmt.Println(strcode, info)
		}
	}
}

func (sel htzmclient) handlemsg(cmd uint16, dReader *bytes.Reader) {
	switch cmd {
	case ZMMSTATICINFO:
		sel.handleStatic(dReader)
	case ZMMHQKZ:
		sel.handlehqkz(dReader)
	case ZMMMARKETSTATUS:
		var status MarketStatusEx
		binary.Read(dReader, binary.LittleEndian, &status)
		fmt.Println("status:", status, string(status.Market_code[:]))
	case ZMMZIPDATA:
		// fmt.Println("ZMMZIPDATA")
		var ZipDataHeader ZmmZipDataHeader
		binary.Read(dReader, binary.LittleEndian, &ZipDataHeader)
		fmt.Println("ZipDataHeader:", ZipDataHeader)
		if ZipDataHeader.Compresstype != 1 || ZipDataHeader.Rawdatalen == 0 {
			break
		}
		var out bytes.Buffer
		r, _ := zlib.NewReader(dReader)
		io.Copy(&out, r)
		if out.Len() == int(ZipDataHeader.Rawdatalen) {
			sel.handlemsg(ZipDataHeader.Rawcmd, bytes.NewReader(out.Bytes()))
		}
	case ZMMLOGINRPLY:
		fmt.Println("ZMMLOGINRPLY")
	case ZMMMARKETRPLY:
		fmt.Println("ZMMMARKETRPLY")
	default:
		fmt.Println("unknown cmd :", cmd)
	}
}

func (sel *htzmclient) ReadLoop() {
	fmt.Println("ReadLoop")
	rdata := make([]byte, 1024*1024*16)
	var header ZmMsgHeader
	pos := 0
	for {
		select {
		case <-sel.cDone: // connection closed
			fmt.Println("receiving cancel signal from conn")
			return
		case <-sel.sDone: // server closed
			fmt.Println("receiving cancel signal from server")
			return
		default:
			if pos == len(rdata) {
				fmt.Println("wrong recv buff full")
				break
			}
			n, err := sel.conn.Read(rdata[pos:])
			if err == nil && n > 0 {
				sel.totalrecvsize += n
				pos += n
				beg := 0
				for pos-beg >= binary.Size(header) {
					dReader := bytes.NewReader(rdata[beg:])
					binary.Read(dReader, binary.LittleEndian, &header)

					if header.Length > uint32(len(rdata)) {
						fmt.Println("wrong Length >recv buff size")
						return
					}
					if header.Length > uint32(pos-beg) {
						break
					}
					fmt.Println("header:", header)
					if header.Length == uint32(binary.Size(header)) {
						fmt.Printf("empty msg,type:%x\n", header.Cmd)
					} else {
						sel.handlemsg(header.Cmd, dReader)
					}
					beg += int(header.Length)
				}
				if beg > 0 {
					copy(rdata, rdata[beg:pos])
					pos -= beg
				}
			}
		}
	}
}

func (sel *htzmclient) heartbeat() {
	fmt.Println("heartbeat")
	buf := new(bytes.Buffer)
	sel.seq++
	var header = ZmMsgHeader{ZMMMSGBOOTCODE, 0, 0, ZMMIDLERQST, sel.seq}
	header.Length = uint32(binary.Size(header))
	binary.Write(buf, binary.LittleEndian, header)
	buff := buf.Bytes()
	// header.length = uint32(len(buff))
	header.Checksum = checkSum(buff)
	buf.Reset()
	binary.Write(buf, binary.LittleEndian, header)
	_, err := sel.conn.Write(buf.Bytes())

	if err != nil {
		fmt.Println("err 2")
		sel.sDone <- 1
		sel.sDone <- 1
	}
}
func (sel *htzmclient) WriteLoop() {

	for {
		select {
		case <-sel.cDone: // connection closed
			fmt.Println("receiving cancel signal from conn")
			return
		case <-sel.sDone: // server closed
			fmt.Println("receiving cancel signal from server")
			return
		default:
			fmt.Println("totalrecvsize:", sel.totalrecvsize/1024, "kB,", sel.compress)
			sel.heartbeat()
		}
		time.Sleep(10 * time.Second)
	}
}

func checkSum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	//以每16位为单位进行求和，直到所有的字节全部求完或者只剩下一个8位字节（如果剩余一个8位字节说明字节数为奇数个）
	for length > 1 {
		sum += uint32(data[index]) + uint32(data[index+1])<<8
		index += 2
		length -= 2
	}
	//如果字节数为奇数个，要加上最后剩下的那个8位字节
	if length > 0 {
		sum += uint32(data[index])
	}
	//加上高16位进位的部分
	sum = sum>>16 + sum&0xFFFF
	sum += (sum >> 16)
	//别忘了返回的时候先求反
	return uint16(^sum)
}

func (sel *htzmclient) InitLogin() error {
	buf := new(bytes.Buffer)
	sel.seq++
	sel.seq++
	var header = ZmMsgHeader{ZMMMSGBOOTCODE, 0, 0, ZMMLOGINRQST, sel.seq}
	var login = ZmmLoginReq{0x20000}
	header.Length = uint32(binary.Size(header) + binary.Size(login))
	binary.Write(buf, binary.LittleEndian, header)
	binary.Write(buf, binary.LittleEndian, login)
	buff := buf.Bytes()
	header.Checksum = checkSum(buff)
	buf.Reset()
	binary.Write(buf, binary.LittleEndian, header)
	binary.Write(buf, binary.LittleEndian, login)
	sel.conn.Write(buf.Bytes())

	return nil
}

func (sel *htzmclient) RegMarkets(markets []string) error {
	buf := new(bytes.Buffer)
	sel.seq++
	var header = ZmMsgHeader{ZMMMSGBOOTCODE, 0, 0, ZMMMARKETRQST, sel.seq}
	var marketreq = ZmmMarketReq{sel.compress, uint8(len(markets))}
	header.Length = uint32(binary.Size(header) + len(markets)*3 + binary.Size(marketreq))

	binary.Write(buf, binary.LittleEndian, header)
	binary.Write(buf, binary.LittleEndian, marketreq)
	for _, market := range markets {
		marketbuff := []byte(market)
		buf.Write(marketbuff[:3])
	}
	buff := buf.Bytes()

	header.Checksum = checkSum(buff)
	buf.Reset()
	binary.Write(buf, binary.LittleEndian, header)
	binary.Write(buf, binary.LittleEndian, marketreq)
	for _, market := range markets {
		marketbuff := []byte(market)
		buf.Write(marketbuff[:3])
	}
	sel.conn.Write(buf.Bytes())

	return nil
}
