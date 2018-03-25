package controllers

import (
	"fmt"
	"time"

	"bzza.com/models"
	"bzza.com/services"

	"github.com/kataras/iris"
	"github.com/monoculum/formam"
)

func AliyunIsv(ctx iris.Context) {
	if AliyunTokenCheck(ctx) == false {
		return
	}

	fmt.Println(time.Now(), ":", ctx.Request().Form)

	var err error
	var dec *formam.Decoder
	var result interface{}

	dec = formam.NewDecoder(&formam.DecoderOptions{TagName: "formam"})

	switch ctx.FormValue("action") {
	case "createInstance": //新购商品
		m := models.CreateInstance{}
		if err = dec.Decode(ctx.Request().Form, &m); err != nil {
			fmt.Println(err)
			return
		}
		result = services.CreateInstance(&m)

	case "renewInstance": //商品续费
		result = map[string]bool{"success": true}
	case "upgradeInstance": //商品升级

	}
	ctx.JSON(result)
}
