package models

import (
	"flag"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/websocket"
)

var (
	yamlFile            = flag.String("file", "./config.yaml", "(Simple) YAML file to read") //配置文件
	SyncMapSumNetworkIn *sync.Map                                                            //统计流量
	SumNetworkInLock    = new(sync.Mutex)                                                    //流量统计锁

	Conf    Config
	WsMutex sync.Mutex
	WsConn  = make(map[websocket.Connection]bool)
)

type Result struct { //全站通用json返回结果
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Config struct { //config.yaml配置文件
	Config struct {
		DriverName     string `yaml:"driverName"`
		DataSourceName string `yaml:"dataSourceName"`
	}
	Mq struct {
		Uri string `yaml:"uri"`
	}
	Push struct {
		Enable    bool     `yaml:"enable"`
		Server    string   `yaml:"server"`
		ApiServer []string `yaml:"apiServer"`
	}
	Aliyun struct {
		KeyId      string `yaml:"keyId"`
		KeySecret  string `yaml:"keySecret"`
		ServiceKey string `yaml:"serviceKey"`
	}
}

type AliyunAccounts struct { //对应数据库表名 阿里云api市场 用户
	Uuid    uuid.UUID `xorm:"pk uuid DEFAULT uuid_generate_v4()"`
	AliUid  string    `xorm:"varchar(16)"`
	Mobile  string    `formam:"bigint"`
	Created time.Time `xorm:"created"`
}

type ApiServices struct { //对应数据库表名 api服务信息表
	Account    string    `xorm:"account"`     //用户的uuid
	OrderBizId string    `xorm:"varchar(16)"` //用户购买后生产的业务实例ID
	OrderId    string    `xorm:"varchar(16)"` //订单ID
	SkuId      string    `xorm:"varchar(16)"` //针对商品的某个版本分配的ID
	HasReq     int64     `xorm:"bigint"`      //剩余次数
	Bill       int32     `xorm:"smallint"`    //计费方式：1，请求次数 2，流量下发
	Created    time.Time `xorm:"created"`     //服务购买时间
	ExpiredOn  time.Time `xorm:"expired_on"`  //失效日期
	IsDelete   bool      `xorm:"is_delete"`   //是否删除
}

type CreateInstance struct { //新购商品
	AliUid          string `formam:"aliUid"` //用户唯一标识
	AccountQuantity int    `formam:"accountQuantity"`
	Mobile          string `formam:"mobile"`
	OrderBizId      string `formam:"orderBizId"` //云市场业务 ID, 用作instanceId
	OrderId         string `formam:"orderId"`    //云市场订单 ID
	SkuId           string `formam:"skuId"`      //商品规格标识，与商品唯一对应，可在商品管理的销售信息中查看
	Token           string `formam:"token"`
	Action          string `formam:"action"`
}

type RenewInstance struct { //续费
	InstanceId string    `formam:"instanceId"` //用户唯一标识
	ExpiredOn  time.Time `formam:"expiredOn"`  //失效日期 ( yyyy - MM - dd HH:mm:ss )
	Token      string    `formam:"token"`
	Action     string    `formam:"action"`
}
