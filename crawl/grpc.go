package crawl

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"
	//	"time"
)

func newClient(arr *string) (client *rpc.Client, err error) {
	address, err := net.ResolveTCPAddr("tcp", (*arr)+":8888")
	if err != nil {
		fmt.Println("连接失败:", address)
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, address)

	if err != nil {
		fmt.Println("net.DialTCP连接失败:", address)
		return nil, errors.New("rpc连接失败")
	}
	client = rpc.NewClient(conn)

	return client, nil
}

func SpInit() {
	sp = make([]sync.Pool, len(Conf.Config.GrpgServerArr))
	for j := 0; j < len(Conf.Config.GrpgServerArr); j++ {
		client, _ := newClient(&Conf.Config.GrpgServerArr[j])
		sp[j] = sync.Pool{
			New: func() interface{} {
				return client
			},
		}
		sp[j].Put(client)

	}
	fmt.Println("sp=", len(sp))
}

func GrpcClientSend(arrMql5 []*Mql5) {
	var err error

	for j := 0; j < len(sp); j++ {
		var client *rpc.Client
		sin := sp[j].Get()
		client, _ = sin.(*rpc.Client)

		if client == nil {
			*GrpcErrCount = atomic.AddUint64(GrpcErrCount, +1)
			//time.Sleep(time.Second)
			if j <= len(Conf.Config.GrpgServerArr) {
				client, _ = newClient(&Conf.Config.GrpgServerArr[j])
				sp[j] = sync.Pool{
					New: func() interface{} {
						return client
					},
				}
			}
			continue
		}

		reply := new(uint64)
		err = client.Call("Arith.Mql5", &arrMql5, reply)
		if err != nil {
			client.Close()
			*GrpcErrCount = atomic.AddUint64(GrpcErrCount, +1)
			//time.Sleep(time.Second)
			if j <= len(Conf.Config.GrpgServerArr) {
				client, _ = newClient(&Conf.Config.GrpgServerArr[j])
				sp[j] = sync.Pool{
					New: func() interface{} {
						return client
					},
				}
			}
		} else {
			*SymbolUpdateCount = atomic.AddUint64(SymbolUpdateCount, +(*reply))
		}
		sp[j].Put(client)
	}
}
