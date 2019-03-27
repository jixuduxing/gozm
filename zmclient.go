package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
	"log"
	// "reflect"
)

type htzmclient struct {
	conn          net.Conn
	cDone, sDone  chan int
	seq           uint32
	totalrecvsize int
	compress      uint8
	logstock      string
	logfile		  *log.Logger	
	instancename	string
	HqkzHeader		ZmmHqkzHeader
}

func (sel *htzmclient) Connect(addr string) error {
	var err error
	sel.conn, err = net.Dial("tcp", addr)
	return err
}
func (sel htzmclient) handleStatic(dReader *bytes.Reader) {
	var StaticInfoHeader ZmmStaticInfoHeader
	binary.Read(dReader, binary.LittleEndian, &StaticInfoHeader)
	
	infobuf := make([]byte, StaticInfoHeader.Itemsize)
	if StaticInfoHeader.Itemtype == ZMMZQDMINFO {
		var info ZQDMInfo
		for i := 0; i < int(StaticInfoHeader.Itemcount); i++ {
			n, _ := dReader.Read(infobuf)
			if n != int(StaticInfoHeader.Itemsize) {
				sel.logfile.Println("n!= StaticInfoHeader.Itemsize", n, StaticInfoHeader.Itemsize)
				break
			}
			reader2 := bytes.NewReader(infobuf)
			binary.Read(reader2, binary.LittleEndian, &info)
			strcode := info.getCode()
			if strcode == sel.logstock {
				sel.logfile.Println("StaticInfoHeader:", StaticInfoHeader, string(StaticInfoHeader.Marketcode[:2]))
				sel.logfile.Println(string(info.Name[:12]), string(info.Pinyinname[:4]), info)
			}
		}
	}

}

func (sel htzmclient) handlehqkz(dReader *bytes.Reader) {
	// var HqkzHeader ZmmHqkzHeader
	binary.Read(dReader, binary.LittleEndian, &sel.HqkzHeader)
	// fmt.Println(sel.HqkzHeader)
	infobuf := make([]byte, sel.HqkzHeader.Itemsize)
	var info CodeIf
	switch sel.HqkzHeader.Itemtype {
	case ZMMHQKZSSHQ:
		info = new(SSHQ)
		// sel.logfile.Println("SSHQ size:", binary.Size(info), HqkzHeader.Itemsize)
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
	for i := 0; i < int(sel.HqkzHeader.Itemcount); i++ {
		n, _ := dReader.Read(infobuf)
		if n != int(sel.HqkzHeader.Itemsize) {
			// sel.logfile.Println("n!= HqkzHeader.Itemsize", n, HqkzHeader.Itemsize)
			break
		}
		reader2 := bytes.NewReader(infobuf)
		binary.Read(reader2, binary.LittleEndian, info)
		strcode := info.getCode()
		// sel.logfile.Print(strcode)
		if strcode == sel.logstock {
			// sel.logfile.Println("HqkzHeader:", HqkzHeader, string(HqkzHeader.Marketcode[:2]))
			// sel.logfile.Println(strcode, info)
			if ZMMHQKZSSHQ == sel.HqkzHeader.Itemtype{
				hhkz,_:= info.(*SSHQ)
				// fmt.Println(info,reflect.TypeOf(info))
				// fmt.Println(hhkz,bok)
				playtm:= int64(hhkz.Date)*1000000000+int64(hhkz.Time)
				ntime := time.Now().UnixNano()
				fmt.Printf("%d,%d,%s,%d,%d,%d,%s\r\n",ntime ,ntime-playtm,strcode,hhkz.Totalvolume,hhkz.Date,hhkz.Time,sel.instancename)
			}
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
		// sel.logfile.Println("status:", status, string(status.Marketcode[:2]))
	case ZMMZIPDATA:
		// sel.logfile.Println("ZMMZIPDATA")
		var ZipDataHeader ZmmZipDataHeader
		binary.Read(dReader, binary.LittleEndian, &ZipDataHeader)
		// sel.logfile.Println("ZipDataHeader:", ZipDataHeader)
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
		// sel.logfile.Println("ZMMLOGINRPLY")
	case ZMMMARKETRPLY:
		// sel.logfile.Println("ZMMMARKETRPLY")
	default:
		sel.logfile.Println("unknown cmd :", cmd)
	}
}

func (sel *htzmclient) ReadLoop() {
	sel.logfile.Println("ReadLoop")
	rdata := make([]byte, 1024*1024*16)
	var header ZmMsgHeader
	pos := 0
	for {
		select {
		case <-sel.cDone: // connection closed
			sel.logfile.Println("receiving cancel signal from conn")
			return
		case <-sel.sDone: // server closed
			sel.logfile.Println("receiving cancel signal from server")
			return
		default:
			if pos == len(rdata) {
				// sel.logfile.Println("wrong recv buff full")
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
						sel.logfile.Println("wrong Length >recv buff size")
						return
					}
					if header.Length > uint32(pos-beg) {
						break
					}
					// sel.logfile.Println("header:", header)
					if header.Length == uint32(binary.Size(header)) {
						// sel.logfile.Printf("empty msg,type:%x\n", header.Cmd)
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
		sel.logfile.Println("err 2")
		sel.sDone <- 1
		sel.sDone <- 1
	}
}
func (sel *htzmclient) WriteLoop() {
	sel.logfile.Println("WriteLoop")
	for {
		select {
		case <-sel.cDone: // connection closed
			sel.logfile.Println("receiving cancel signal from conn")
			return
		case <-sel.sDone: // server closed
			sel.logfile.Println("receiving cancel signal from server")
			return
		default:
			sel.logfile.Println("totalrecvsize:", sel.totalrecvsize/1024, "kB,", sel.compress)
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
	sel.logfile.Println("InitLogin  2")
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
	sel.logfile.Println("RegMarkets  2")
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
