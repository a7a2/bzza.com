package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"net/url"
	"strings"

	"fmt"

	"bzza.com/models"

	"github.com/kataras/iris"
	"github.com/monoculum/formam"
)

func decode(m *interface{}, ctx *iris.Context) *error {
	dec := formam.NewDecoder(&formam.DecoderOptions{TagName: "formam"})
	if err := dec.Decode((*ctx).Request().Form, &m); err != nil {
		fmt.Println(err)
		return &err
	}
	return nil
}

func AliyunTokenCheck(ctx iris.Context) bool { //阿里云api市场 token检查
	arr := strings.Split(ctx.Request().URL.Query().Encode(), "&token=")
	if len(arr) != 2 {
		return false
	}
	pStr, err := url.PathUnescape(arr[0])
	if err != nil {
		return false
	}
	reqStr := pStr + "&key=" + models.Conf.Aliyun.ServiceKey
	reqStr = strings.Replace(reqStr, "+", " ", -1)

	if getMd5(&reqStr) == ctx.FormValue("token") {
		return true
	}

	return false
}

func getMd5(str *string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(*str))
	cipherStr := md5Ctx.Sum(nil)
	//fmt.Print(hex.EncodeToString(cipherStr))
	return hex.EncodeToString(cipherStr)
}

func SumNetworkIn(r *models.Result, Cacloudmarketinstanceid *string) {
	models.SumNetworkInLock.Lock()
	defer models.SumNetworkInLock.Unlock()

	bytes, _ := json.Marshal(r)
	uI64, _ := strconv.ParseUint(strconv.Itoa(len(bytes)), 10, 64)
	vItf, ok := models.SyncMapSumNetworkIn.Load(*Cacloudmarketinstanceid)
	var tempUint64 uint64
	if ok {
		tempUint64 = vItf.(uint64) + uI64
	} else {
		tempUint64 = uI64
	}
	models.SyncMapSumNetworkIn.Store(*Cacloudmarketinstanceid, tempUint64)
}
