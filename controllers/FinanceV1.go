package controllers

import (
	"strings"

	"bzza.com/models"

	"github.com/kataras/iris"
)

func FinanceV1(ctx iris.Context) {
	Cacloudmarketinstanceid := ctx.GetHeader("Cacloudmarketinstanceid")

	var r models.Result
	if ctx.FormValue("symbol") == "" {
		var arr []interface{}
		models.SyncMapMql5.Range(func(k, v interface{}) bool {
			arr = append(arr, v)
			return true
		})
		r = models.Result{Code: 200, Message: "ok", Data: &arr}
		SumNetworkIn(&r, &Cacloudmarketinstanceid)
		ctx.JSON(&r)
		return
	}

	arr := strings.Split(ctx.FormValue("symbol"), ",")
	if len(arr) == 0 {
		r = models.Result{Code: 400, Message: "请求参数错误symbol", Data: nil}
		SumNetworkIn(&r, &Cacloudmarketinstanceid)
		ctx.JSON(&r)
		return
	}

	var arrInerface []interface{}
	for i := 0; i < len(arr); i++ {
		v, ok := models.SyncMapMql5.Load(arr[i])
		if !ok {
			continue
		}
		arrInerface = append(arrInerface, v)
	}
	if len(arrInerface) > 0 {
		r = models.Result{Code: 200, Message: "ok", Data: arrInerface}
	} else {
		r = models.Result{Code: 404, Message: "数据还未有请稍后或参数错误", Data: &arrInerface}
	}

	SumNetworkIn(&r, &Cacloudmarketinstanceid)
	ctx.JSON(&r)

}
