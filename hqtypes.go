package main

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

//OrderQueue 委托队列
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

//MarketStatusEx 市场状态
type MarketStatusEx struct {
	Marketcode [3]byte
	Status     uint8
	Staticdate uint32
	Statictime uint32
	Hqkzdate   uint32
	Hqkztime   uint32
}
