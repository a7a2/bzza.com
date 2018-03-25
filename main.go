package main

import (
	"sync"

	"bzza.com/common"
	"bzza.com/controllers"
	"bzza.com/models"
	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
)

//http://aliyun.api.bzza.com/v1/finance

func main() {
	app := iris.New()

	//-----------------我是sb的分界线----websocket start----------------------//
	ws := websocket.New(websocket.Config{
		// to enable binary messages (useful for protobuf):
		// BinaryMessages: true,
	})
	var WsMutex = new(sync.Mutex)
	app.Get("/w1/finance", ws.Handler())
	ws.OnConnection(func(c websocket.Connection) {
		WsMutex.Lock()
		models.WsConn[c] = true
		WsMutex.Unlock()
		controllers.WsMain(c)
	})
	//-----------------我是sb的分界线----websocket end------------------------//

	v1 := app.Party("/v1")                    //版本v1
	v1.Get("/finance", controllers.FinanceV1) //阿里api市场 金融数据接口

	app.Get("/isv", controllers.AliyunIsv) //阿里isv接入

	go grpcServer()

	go common.LoopPushMeteringData()
	app.Run(iris.Addr(":80"))

}
