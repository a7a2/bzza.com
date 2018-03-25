package common

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"

	"encoding/json"
	"net/url"

	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"strings"

	"time"

	"bzza.com/models"
)

const pUrl = "https://market.aliyuncs.com/"

type Metering struct { //推送到阿里的计量信息
	InstanceId string     `json:"InstanceId"` //云市场实例ID
	StartTime  int64      `json:"StartTime"`  //计量开始时间，单位秒（格式为unix时间戳）
	EndTime    int64      `json:"EndTime"`    //计量结束时间，单位秒（格式为unix时间戳）
	Entities   []Entities `json:"Entities"`   //计量实体对象

}

type Entities struct {
	Key   string `json:"Key"` //NetworkIn
	Value string `json:"Value"`
}

type common struct { //公共参数
	Format      string `json:"Format"`      //返回值的类型，支持 JSON 与 XML。默认为 XML
	Version     string `json:"Version"`     //API 版本号，为日期形式：YYYY-MM-DD，本版本对应为 2015-11-01。
	AccessKeyId string `json:"AccessKeyId"` //阿里云颁发给用户的访问服务所用的密钥 ID。
	//	Signature            string `json:"Signature"`            //签名结果串
	SignatureMethod      string `json:"SignatureMethod"`      //签名方式，目前支持 HMAC-SHA1。
	Timestamp            string `json:"Timestamp"`            //请求的时间戳 格式为：使用 UTC 时间 YYYY-MM-DDThh:mm:ssZ 例如，2014-05-26T12:00:00Z（为北京时间 2014 年 5 月 26 日 20 点 0 分 0 秒）。
	SignatureVersion     string `json:"SignatureVersion"`     //签名算法版本，目前版本是 1.0。
	SignatureNonce       string `json:"SignatureNonce"`       //唯一随机数，用于防止网络重放攻击。用户在不同请求间要使用不同的随机数值
	ResourceOwnerAccount string `json:"ResourceOwnerAccount"` //本次 API 请求访问到的资源拥有者账户，即登录用户名。
}

func sig(meterArr []*Metering) {
	bytes, err := json.Marshal(meterArr)
	if err != nil {
		fmt.Println(err)
		return
	}

	tUTC := time.Now().UTC()
	ts := tUTC.Format("2006-01-02T15:04:05Z")
	tUnixStr := fmt.Sprintf("%v", tUTC.Unix())
	//&ResourceOwnerAccount=admin@xunfish.com

	originalString := fmt.Sprintf("AccessKeyId=%s&Action=PushMeteringData&Format=JSON&Metering=%s&SignatureMethod=HMAC-SHA1&SignatureNonce=%s&SignatureVersion=1.0&Timestamp=%s&Version=2015-11-01", models.Conf.Aliyun.ServiceKey, string(bytes), tUnixStr, ts)

	meteringValueEscape := url.QueryEscape(string(bytes))
	tempStringToSign := fmt.Sprintf("AccessKeyId=%s&Action=PushMeteringData&Format=JSON&Metering=%s&SignatureMethod=HMAC-SHA1&SignatureNonce=%s&SignatureVersion=1.0&Timestamp=%s&Version=2015-11-01", models.Conf.Aliyun.KeyId, meteringValueEscape, tUnixStr, url.QueryEscape(ts))
	tempStringToSign = url.QueryEscape(tempStringToSign)

	//	str = strings.Replace(str, "+", "%20", -1)
	//	str = strings.Replace(str, "*", "%2A", -1)
	//	str = strings.Replace(str, "%7E", "~", -1)
	//str = strings.Replace(str, "=", "%3D", -1)
	StringToSign := "POST&%2F&" + tempStringToSign

	//	fmt.Println("StringToSign:", StringToSign)
	//	fmt.Println(" ")
	key := []byte(fmt.Sprintf("%s&", models.Conf.Aliyun.KeySecret))
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(StringToSign))

	//fmt.Println(fmt.Sprintf("%x", mac.Sum(nil)))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	originalString = fmt.Sprintf("%s&Signature=%v", originalString, url.QueryEscape(sig))
	PostAliyun(originalString)
}

func PostAliyun(meteringValue string) {
	//fmt.Println("Post:", meteringValue)
	resp, err := http.Post(pUrl,
		"application/x-www-form-urlencoded",
		strings.NewReader(meteringValue))
	if err != nil {
		fmt.Println("PostAliyu 39:", err)
		return
	}

	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("PostAliyu 46:", err)
		//return
	}
	//fmt.Println("PostAliyu——body:", string(body))
}

func LoopPushMeteringData() {
	for {
		select {
		case times := <-time.After(time.Minute): //每分钟调用一下
			if times.Minute()%5 == 0 { //每5分钟上报一次
				mapSumNetworkIn := CollectMeteringData()
				if mapSumNetworkIn == nil {
					continue
				}
				//fmt.Println(times)
				PushMeteringData(mapSumNetworkIn)
			}
		}
	}
}

func PushMeteringData(mapSumNetworkIn map[string]uint64) {
	var meterArr []*Metering
	for k, v := range mapSumNetworkIn {
		var arrEntities []Entities
		value := fmt.Sprintf("%v", v/300)
		if value == "0" || k == "" {
			continue
		}
		fmt.Println(k, " ", v)
		e := Entities{Key: "NetworkIn", Value: value} // 除以300是5分钟的秒数
		arrEntities = append(arrEntities, e)
		m := &Metering{
			InstanceId: k,
			StartTime:  time.Now().Add(time.Minute * -5).Unix(),
			EndTime:    time.Now().Unix(),
			Entities:   arrEntities,
		}

		meterArr = append(meterArr, m)
	}

	if len(meterArr) == 0 { //不用推送了 没有数据呢
		return
	}
	sig(meterArr)
}

func CollectMeteringData() (mapSumNetworkIn map[string]uint64) {
	mapSumNetworkIn = make(map[string]uint64)
	sMapArr := grpcClientGetSMap()
	if len(sMapArr) == 0 {
		return
	}

	for _, mapAddr := range sMapArr { //分拆sMap数组
		for k, v := range *mapAddr {
			mapSumNetworkIn[k] = mapSumNetworkIn[k] + v //累加所有服务器统计的量
		}
	}

	return
}

func grpcClientGetSMap() (sMap []*map[string]uint64) {
	for _, v := range models.Conf.Push.ApiServer {
		reply := make(map[string]uint64)
		client, err := newClient(&v)
		if err != nil {
			continue
		}
		defer client.Close()
		err = client.Call("Arith.CollectMeteringData", false, &reply) //false为拿统计数据 ,true为拿数据后清空统计数据
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Microsecond * 100) //防止耗尽cpu
			models.YamlInit()
			break
		}

		sMap = append(sMap, &reply)
	}

	return sMap

}

func newClient(arr *string) (clientArr *rpc.Client, err error) {
	address, err := net.ResolveTCPAddr("tcp", (*arr)+":8888")
	if err != nil {
		fmt.Println("连接失败:", address)
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, address)

	if err != nil {
		return nil, errors.New("rpc连接失败")
	}
	clientArr = rpc.NewClient(conn)

	return clientArr, nil
}
