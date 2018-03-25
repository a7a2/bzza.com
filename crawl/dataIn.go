package crawl

import (
	"fmt"
	//	"io/ioutil"
	"net"
	"net/http"
	//	"strconv"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

func dialTimeout(network, addr string) (net.Conn, error) {
	conn, err := net.DialTimeout(network, addr, time.Second*2)
	if err != nil {
		return conn, err
	}

	tcp_conn := conn.(*net.TCPConn)
	tcp_conn.SetKeepAlive(false)

	return tcp_conn, err
}

func httpPost(m *Mql5) {
	transport := http.Transport{
		Dial:              dialTimeout,
		DisableKeepAlives: true,
	}

	client := http.Client{
		Transport: &transport,
	}

	resp, err := client.Post(fmt.Sprintf("http://%s:8086/write?db=mql5", Conf.Config.Datahost),
		"application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("%s,type=%s bid=%v,ask=%v %v", (*m).Symbol, (*m).Type, (*m).Bid, (*m).Ask, m.UnixNano)))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	//	_, err := ioutil.ReadAll(resp.Body)
	//	if err != nil {
	//		panic(err)
	//	}

	//fmt.Println(string(body))
}

func WriteUDP(m *Mql5, tt *time.Time) {
	// Make client
	c, err := client.NewUDPClient(client.UDPConfig{Addr: fmt.Sprintf("%s:8089", Conf.Config.Datahost)})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  Conf.Config.Database,
		Precision: "ms",
	})

	// Create a point and add to batch
	tags := map[string]string{m.Type: m.Symbol}
	fields := map[string]interface{}{
		"bid": m.Bid,
		"ask": m.Ask,
	}
	pt, err := client.NewPoint(m.Symbol, tags, fields, *tt)
	if err != nil {
		panic(err.Error())
	}
	bp.AddPoint(pt)

	// Write the batch
	err = c.Write(bp)
	if err != nil {
		c.Write(bp)
	}
}
