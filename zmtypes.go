package main

const (
	//ZMMMSGBOOTCODE bootcode
	ZMMMSGBOOTCODE = 0x7e7e
	//ZMMLOGINRQST 登录请求
	ZMMLOGINRQST = 0x0080
	//ZMMLOGINRPLY 登录回复
	ZMMLOGINRPLY = 0x8080
	//ZMMMARKETRQST 市场订阅请求
	ZMMMARKETRQST = 0x00f0
	//ZMMMARKETRPLY 市场订阅回复
	ZMMMARKETRPLY = 0x80f0
	//ZMMIDLERQST 心跳请求
	ZMMIDLERQST = 0x00ff
	//ZMMIDLERPLY 心跳回复
	ZMMIDLERPLY = 0x80ff
	//ZMMHQKZ 快照数据
	ZMMHQKZ = 0x0001
	//ZMMSTATICINFO 静态信息
	ZMMSTATICINFO = 0x0002
	//ZMMZIPDATA 压缩数据
	ZMMZIPDATA = 0x1000
	//ZMMMARKETSTATUS 市场状态
	ZMMMARKETSTATUS = 0x0003
	//ZMMZQDMINFO 证券代码信息
	ZMMZQDMINFO = 1
	//ZMMHQKZSSHQ 实时行情
	ZMMHQKZSSHQ = 1
	//ZMMHQKZSSZS 实时指数
	ZMMHQKZSSZS = 2
	//ZMMHQKZSTEPORDER 逐笔委托
	ZMMHQKZSTEPORDER = 3
	//ZMMHQKZSTEPTRADE 逐笔成交
	ZMMHQKZSTEPTRADE = 4
	//ZMMHQKZORDERQUEUE 一档委托队列
	ZMMHQKZORDERQUEUE = 5
)

/*消息头*/
type ZmMsgHeader struct {
	Bootcode uint16 /* 特征码 */
	Checksum uint16 /* 校验和 */
	Length   uint32 /* 命令长度 */
	Cmd      uint16 /* 命令码 */
	Seqid    uint32 /* 命令的序列号 */
}

/*请求消息*/
type ZmmLoginReq struct {
	version uint32
}

/*应答消息*/
type ZmmLoginResp struct {
	Version uint32
}

/*请求消息*/
type ZmmMarketReq struct {
	compress    uint8 //是否压缩
	marketcount uint8 //市场数量
	/*
		{
		char market_code[3]; //市场代码
		}+
	*/
}

/*实时行情快照*/
type ZmmHqkzHeader struct {
	Flag       uint8
	Marketcode [3]byte //市场代码
	Itemcount  uint32  //
	Itemtype   uint8   //
	Itemsize   uint32  //
}

/*静态信息*/
type ZmmStaticInfoHeader struct {
	Flag       uint8
	Marketcode [3]byte //市场代码
	Itemcount  uint32  //
	Itemtype   uint8   //
	Itemsize   uint32  //
}

//压缩数据
type ZmmZipDataHeader struct {
	Compresstype uint8  //压缩方式,1表示zip,目前支持这一种
	Rawcmd       uint16 //原始命令码
	Rawdatalen   uint32 //
}

//CodeIf 获取代码接口类
type CodeIf interface {
	getCode() string
}

//HQDATABASE  基础结构
type HQDATABASE struct {
	Code [22]byte //
}

func (sel HQDATABASE) getCode() string {
	return string(sel.Code[:8])
}

//ZQDMInfo 证券代码表
type ZQDMInfo struct {
	HQDATABASE
	// Code               [22]byte //
	Name               [60]byte //
	Pinyinname         [16]byte //
	Zqtype             byte
	Volumeunit         uint16
	Preclose           uint32
	Highlimit          uint32
	Lowlimit           uint32
	Pricedigit         uint32
	Pricedivide        byte
	Intrestsettleprice int32
	Crdflag            byte
	Preposition        int32
	Presettleprice     int32
	Exttype            byte
	Subtype            byte
}

// func (sel ZQDMInfo) getCode() string {
// 	return string(sel.Code[:8])
// }

//SSHQ 实时行情快照
type SSHQ struct {
	HQDATABASE
	// Code            [22]byte //
	Last            uint32
	Open            uint32
	High            uint32
	Low             uint32
	Totalvolume     uint64
	Totalamount     uint64
	Totaltradecount uint32
	Position        uint32
	Buyprice        [10]uint32
	Buyvolume       [10]uint32
	Sellprice       [10]uint32
	Sellvolume      [10]uint32
	Date            uint32
	Time            uint32
	TradingPhase    byte
	ReservdWord     [8]byte
}

// func (sel SSHQ) getCode() string {
// 	return string(sel.Code[:8])
// }

//SSZS 实时行情快照
type SSZS struct {
	HQDATABASE
	// Code        [22]byte //
	Last        uint32
	Open        uint32
	High        uint32
	Low         uint32
	Totalvolume uint64
	Totalamount uint64
	Date        uint32
	Time        uint32
}

// func (sel SSZS) getCode() string {
// 	return string(sel.Code[:8])
// }

//StepOrderNative 原生委托
type StepOrderNative struct {
	HQDATABASE
	// Code      [22]byte //
	Date      uint32
	Time      uint32
	Price     uint32
	Orderqty  uint32
	Side      byte
	Ordtype   byte
	Appseqnum uint64
	Channelno uint16
}

// func (sel StepOrderNative) getCode() string {
// 	return string(sel.Code[:8])
// }

//StepTradeNative 原生逐笔
type StepTradeNative struct {
	HQDATABASE
	// Code        [22]byte //
	Date        uint32
	Time        uint32
	Price       uint32
	Tradeqty    uint32
	Bidseqnum   uint64
	Offerseqnum uint64
	Appseqnum   uint64
	Channelno   uint16
	Exectype    byte
}

// func (sel StepTradeNative) getCode() string {
// 	return string(sel.Code[:8])
// }

//委托队列
type OrderQueue struct {
	HQDATABASE
	// Code               [22]byte //
	Date               uint32
	Time               uint32
	Buyprice           uint32
	Sellprice          uint32
	Numbuyorders       uint32
	Numsellorders      uint32
	Buyorderitems      [50]uint32
	Sellorderitems     [50]uint32
	Totalbuyvolume     uint64
	Totalsellvolume    uint64
	Buyavgprice        uint32
	Sellavgprice       uint32
	Numbuyorderstotal  uint32
	Numsellorderstotal uint32
}

// func (sel OrderQueue) getCode() string {
// 	return string(sel.Code[:8])
// }
