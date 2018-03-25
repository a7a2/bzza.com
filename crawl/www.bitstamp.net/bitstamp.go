package main

import (
	"encoding/json"
	"fmt"

	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bzza.com/crawl"
	"github.com/aglyzov/ws-machine"
)

func main() {
	URL := "wss://ws.pusherapp.com/app/de504dc5763aeef9ff52?protocol=7&client=js&version=2.1.6&flash=true"
	wsm := machine.New(URL, http.Header{})
	var numInt, lastNumInt int
	var i64, ts64, lastTS64 int64 //
	var price float64
	var str, timeStr string
	var err error

	for {
		select {
		case st := <-wsm.Status:
			//	fmt.Println("STATE:", st.State)
			if st.Error != nil {
				fmt.Println("ERROR:", st.Error)
			}
			if st.State == machine.CONNECTED {
				//fmt.Println("SUBSCRIBE: live_trades")
				wsm.Output <- []byte(`{"event":"pusher:subscribe","data":{"channel":"live_trades"}}`)
				//wsm.Output <- []byte(`{"event":"pusher:subscribe","data":{"channel":"live_trades"}}`)
				//wsm.Output <- []byte(`{"event":"pusher:subscribe","data":{"channel":"live_orders"}}`)
				//fmt.Println("SUBSCRIBE: order_book")
				//wsm.Output <- []byte(`{"event":"pusher:subscribe","data":{"channel":"order_book"}}`)
			}
		case msg := <-wsm.Input:
			msgStr := strings.Replace(string(msg), "\\", "", -1)
			msgStr = strings.Replace(msgStr, "\"{", "{", 1)
			msgStr = strings.Replace(msgStr, "}\"", "}", 1)

			btcData := BtcData{}
			err = json.Unmarshal([]byte(msgStr), &btcData)
			if err != nil {
				fmt.Println(err)
				continue
			}

			ts64, err = strconv.ParseInt((*btcData.Data).Timestamp, 10, 64)
			if err != nil {
				continue
			}
			price = (*btcData.Data).Price

			i64 = time.Now().UnixNano()
			str = fmt.Sprintf("%v", i64)
			numInt, _ = strconv.Atoi(str[10:13])
			if numInt > 225 {
				numInt -= 225
			}

			if lastTS64 == ts64 { //上一个时间戳跟当前时间戳一致
				for lastNumInt > numInt { //900 800
					if 999-lastNumInt < 10 {
						numInt = numInt + 1
						if numInt == 0 {
							numInt = 999
						}
					} else {
						rand.Seed(time.Now().Unix())
						numInt = rand.Intn(999-lastNumInt) + numInt
					}
					lastNumInt = numInt
				}
			}

			lastTS64 = ts64

			timeStr = time.Unix(ts64, int64(numInt)*1000000).Format("15:04:05.999")

			ts64 = ts64*1000 + int64(numInt)
			//	priceStr = strconv.FormatFloat(price, 'f', 2, 64)
			m := crawl.Mql5{
				Type:     "digital cash",
				Title:    "bitcoin usd index",
				Symbol:   "BTCUSD",
				Bid:      price,
				Ask:      price,
				UnixNano: ts64 * 1000000,
				TimeStr:  timeStr,
			}

			var arrMql5 []*crawl.Mql5
			arrMql5 = append(arrMql5, &m)
			crawl.GrpcClientSend(arrMql5)
			//fmt.Println(timeStr, "	", ts64*1000000, " __ ", price)
		}
	}
}
