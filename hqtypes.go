package main

//市场状态
type MarketStatusEx struct {
	Market_code [3]byte
	Status      uint8
	Staticdate  uint32
	Statictime  uint32
	Hqkzdate    uint32
	Hqkztime    uint32
}
