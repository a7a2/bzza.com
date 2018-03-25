package models

import (
	//	"net/rpc"
	"sync"
)

//const GrpcClientSize = 500

//var ChanGrpcClientPool = make(chan *rpc.Client, GrpcClientSize) //grpc连接池

//var MapMql5 = make(map[string]*Mql5) //全部数据

type Mql5 struct { //database
	//Type     string `json:"type"`
	Title    string  `json:"title"`
	Symbol   string  `json:"symbol"`
	Bid      float64 `json:"bid"`
	Ask      float64 `json:"ask"`
	UnixNano int64   `json:"unixnano"`
	TimeStr  string  `json:"timestr"`
}

var SyncMapMql5 *sync.Map
