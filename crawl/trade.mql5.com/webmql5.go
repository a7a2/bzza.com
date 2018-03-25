package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	//	"time"
	"bzza.com/crawl"
	"github.com/gosuri/uilive"
	"github.com/sclevine/agouti"
)

var page *agouti.Page

func NewDriver() *agouti.WebDriver {
	driver := agouti.ChromeDriver()
	//driver := agouti.PhantomJS()
	//driver := agouti.Selenium()
	//driver := agouti.ChromeDriver(agouti.Browser("chrome"))
	if err := driver.Start(); err != nil {
		panic(err)
	}

	return driver
}

func get(page *agouti.Page, url string) {
	if err := page.Navigate(url); err != nil {
		panic(err)
	}
}

func loopPageReload() {
	var doneUint64 uint64
	var err error

	writer := uilive.New()
	// start listening for updates and render
	writer.Start()
	defer writer.Stop() // flush and stop rendering
	for {
		select {
		case _ = <-crawl.ChanPageReload:
			doneUint64 += 1
			fmt.Fprintf(writer, "完成采集有效次数:%v , Grpc出错次数:%v , 交易对更新次数:%v \n", doneUint64, *(crawl.GrpcErrCount), *(crawl.SymbolUpdateCount))
		case _ = <-time.After(time.Second * crawl.Conf.Config.PageRefresh): //一分钟没有数据进入即刷新page
			switch time.Now().UTC().Weekday().String() {
			case "Saturday": //周六
				break
			case "Sunday": //周日都不处理 因为不开市
				break
			default:
				err = page.Refresh()
				if err != nil {
					fmt.Println("page.Refreash:", err)
				}
			}
		}
	}

}

func enterWebMql5() {
	var str string
	var err error

	driver := NewDriver()
	defer driver.Stop()
	page, err = driver.NewPage()
	if err != nil {
		panic(err)
	}
	defer driver.WebDriver.Stop()

	get(page, "https://trade.mql5.com/trade")
	time.Sleep(time.Second)
	go loopPageReload()

	for i := 0; i < 4; i++ {
		page.FindByID("h").SendKeys("TAB")
		time.Sleep(time.Microsecond * 100)
	}

	time.Sleep(time.Microsecond * 100)
	page.FindByID("server").SendKeys("BACK_SPACE")
	time.Sleep(time.Microsecond * 100)
	page.FindByID("server").Fill("ICMarkets-MT5")

	page.FindByID("mt5-platform").Click()
	time.Sleep(time.Second * 2)
	page.FindByID("login").Fill("0000000")
	time.Sleep(time.Microsecond * 500)
	page.FindByID("password").Fill("12345678")
	time.Sleep(time.Microsecond * 100)

	for {
		str, err = page.HTML()
		if err != nil {
			fmt.Println("page.HTML():", err)
			time.Sleep(time.Second)
			continue
		}

		if len(str) > 100000 && strings.Contains(str, "<span class=\"content\">!</span>") {
			arrStr := crawl.Re.FindAllStringSubmatch(str, 39530)
			go check(&arrStr)
		} else {
			//fmt.Println(len(str), " ? 100000   &&  ", strings.Contains(str, "<span class=\"content\">!</span>"))
			time.Sleep(time.Second)
		}
	}

	str, err = page.HTML()
	if err != nil {
		panic(err)
	}
	err = page.Screenshot("./test.jpg")
	fmt.Println(str)
	if err != nil {
		panic(err)
	}

}

func check(arrStr *[][]string) {
	//	defer func() {
	//		if err := recover(); err != nil {
	//			fmt.Println("出了错：", err)
	//		}
	//	}()

	var arrMql5 []*crawl.Mql5
	var valueInterface interface{}
	var ok bool
	var t time.Time
	var err error
	var unixNano int64
	var symbol, kinds, timeStr, title, bid, ask string //交易对 ，属于什么类型。 如EURUSD，属于forex
	var askF64, bidF64 float64

	for i := 0; i < len(*arrStr) && len((*arrStr)[i]) == 26; i++ {
		symbol, timeStr, title, bid, ask = (*arrStr)[i][10], (*arrStr)[i][24], (*arrStr)[i][6], (*arrStr)[i][14], (*arrStr)[i][18]
		//		switch {
		//		case strings.Contains(symbol, "BTC") || strings.Contains(symbol, "DSH") || strings.Contains(symbol, "LTC") || strings.Contains(symbol, "ETC") || strings.Contains(symbol, "NMC") || strings.Contains(symbol, "PPC") || strings.Contains(symbol, "XMR") || strings.Contains(symbol, "XRP") || symbol == "ZECUSD":
		//			kinds = "Crypto"
		//		case symbol == "XAGUSD" || symbol == "XAUUSD" || symbol == "XPTUSD" || symbol == "XPUSD" || symbol == "XAUEUR" || symbol == "XAGEUR":
		//			kinds = "Metals" //贵金属 icmarket
		//		case symbol == "Coffee_Z7" || symbol == "Coffee_H8" || symbol == "Corn_Z7" || symbol == "Soybean_F8" || symbol == "Sugar_H8" || symbol == "Wheat_Z7" || symbol == "Wheat_H7":
		//			kinds = "Commodities_Softs" //商品 icmarket
		//		case symbol == "BRENT_H8" || symbol == "BRENT_G8" || symbol == "XBRUSD" || symbol == "XTIUSD" || symbol == "WTI_G8" || symbol == "WTI_F8" || symbol == "XNGUSD":
		//			kinds = "Commodities_Energies" //原油 icmarket
		//		case symbol == "AUS200" || symbol == "CHINA50" || symbol == "DE30" || symbol == "ES35" || symbol == "F40" || symbol == "HK50" || symbol == "IT40" || symbol == "JP225" || symbol == "STOXX50" || symbol == "UK100" || symbol == "US2000" || symbol == "US500" || symbol == "US30" || symbol == "USTEC" || symbol == "WIG20":
		//			kinds = "Indices_Spot" //指数 股票 icmarket
		//		case symbol[0:1] == "#":
		//			symbol = strings.Replace(symbol, "#", "CFD_", 1)
		//			kinds = "CFD"
		//		default:
		//			kinds = "Forex"
		//		}

		t, err = time.ParseInLocation("2006-01-02 15:04:05.999", time.Now().Add(time.Hour*2).UTC().Format("2006-01-02 ")+timeStr, time.Local)
		if err != nil {
			panic(err)
		}
		unixNano = t.UnixNano()

		valueInterface, ok = crawl.SyncMap.Load(symbol)
		if ok {
			if valueInterface.(*crawl.Mql5).TimeStr == timeStr { //已经提交过的跳过
				continue
			}

			if (t.Hour() == 23 && t.Minute() > 54) || (t.Hour() == 0 && t.Minute() < 6) { //用于校验时间 在晚上23最后几分钟到0点几分钟之间可能出现的时间差导致跑偏,减少cpu使用率
				comUnixNano := unixNano - (valueInterface.(*crawl.Mql5).UnixNano)
				if comUnixNano > comNano24 && comUnixNano < comNano26 { //大于24小时少于26小时， 休息是2天，第二天开市的少于18小时
					unixNano = t.Add(time.Hour * -24).UnixNano()
				}
			}

		}

		bidF64, err = strconv.ParseFloat(bid, 64)
		if err != nil {
			continue
		}
		askF64, err = strconv.ParseFloat(ask, 64)
		if err != nil {
			continue
		}

		m := crawl.Mql5{
			Type:     kinds,
			Title:    title,
			Symbol:   symbol,
			Bid:      bidF64,
			Ask:      askF64,
			UnixNano: unixNano,
			TimeStr:  timeStr,
		}
		crawl.SyncMap.Store(symbol, &m)

		arrMql5 = append(arrMql5, &m)

		//crawl.WriteUDP(&m, &t)
		//m.httpPost()
	}

	if len(arrMql5) > 0 {
		crawl.ChanPageReload <- true
		crawl.GrpcClientSend(arrMql5)
	}

}
