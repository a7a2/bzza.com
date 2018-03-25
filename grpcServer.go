package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"

	"bzza.com/controllers"
	"bzza.com/models"
)

func grpcServer() {
	newServer := rpc.NewServer()
	newServer.Register(new(Arith))
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8888")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("Grpc服务启动")
	newServer.Accept(l)
}

type Arith int

var grpcMql5Lock = new(sync.Mutex)

func (t *Arith) Mql5(arrMql5 *[]*models.Mql5, reply *uint64) error {
	grpcMql5Lock.Lock()
	defer grpcMql5Lock.Unlock()

	*reply = 0
	for _, mql5Addr := range *arrMql5 {
		v, ok := models.SyncMapMql5.Load((*mql5Addr).Symbol)
		if ok && (*mql5Addr).UnixNano <= v.(*models.Mql5).UnixNano {
			continue
		}

		*reply += 1                                                                //有效更新数
		controllers.BroadcastSame(&models.WsConn, &((*mql5Addr).Symbol), mql5Addr) //新数据到达 推送
		models.SyncMapMql5.Store((*mql5Addr).Symbol, mql5Addr)
	}

	return nil
}

func (t *Arith) CollectMeteringData(cleanOk bool, reply *map[string]uint64) error { //收集统计用户流量
	models.SumNetworkInLock.Lock()
	defer models.SumNetworkInLock.Unlock()

	models.SyncMapSumNetworkIn.Range(func(k, v interface{}) bool {
		kStr, ok := k.(string)
		if ok {
			(*reply)[kStr] = v.(uint64)
			models.SyncMapSumNetworkIn.Delete(k)
		}
		return true
	})

	return nil
}
