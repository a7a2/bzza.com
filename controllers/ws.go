package controllers

import (
	"encoding/json"
	"fmt"

	"strings"

	"bzza.com/models"
	"github.com/kataras/iris/websocket"
)

func WsMain(c websocket.Connection) {
	c.OnDisconnect(func() {
		models.WsMutex.Lock()
		delete(models.WsConn, c)
		models.WsMutex.Unlock()
	})

	Cacloudmarketinstanceid := c.Context().GetHeader("Cacloudmarketinstanceid")

	var r models.Result
	if c.Context().FormValue("symbol") == "" { //全部
		c.SetValue("all", true)

		var arr []interface{}
		models.SyncMapMql5.Range(func(k, v interface{}) bool {
			arr = append(arr, v)
			return true
		})

		r = models.Result{Code: 200, Message: "ok", Data: &arr}
		bytes, _ := json.Marshal(&r)
		SumNetworkIn(&r, &Cacloudmarketinstanceid) //统计流量数据
		c.EmitMessage(bytes)
	}

	arr := strings.Split(c.Context().FormValue("symbol"), ",")
	if len(arr) == 0 {
		r = models.Result{Code: 400, Message: "请求参数错误symbol", Data: nil}
		SumNetworkIn(&r, &Cacloudmarketinstanceid)
		c.Context().JSON(&r)
		return
	}

	//var arrInerface []interface{}
	for i := 0; i < len(arr); i++ {
		v, ok := models.SyncMapMql5.Load(arr[i])
		if !ok {
			continue
		}
		symbol := v.(*models.Mql5).Symbol
		c.SetValue(symbol, true)
		r = models.Result{Code: 200, Message: "ok", Data: v}
		bytes, _ := json.Marshal(&r)
		SumNetworkIn(&r, &Cacloudmarketinstanceid) //统计流量数据
		c.EmitMessage(bytes)
	}
}

func BroadcastSame(Conn *map[websocket.Connection]bool, gate *string, mql5 *models.Mql5) {
	var err error
	r := models.Result{Code: 200, Message: "ok", Data: mql5}
	bytes, err := json.Marshal(r)
	if err != nil {
		return
	}
	for c := range *Conn {
		Cacloudmarketinstanceid := c.Context().GetHeader("Cacloudmarketinstanceid")
		SumNetworkIn(&r, &Cacloudmarketinstanceid) //统计流量数据

		if c.GetValue(*gate) != nil || c.GetValue("all") != nil {
			err = c.EmitMessage(bytes)
			if err != nil {
				fmt.Println("ws BroadcastSame:", err)
				continue
			}
		}

	}
}
