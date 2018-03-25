package crawl

import (
	"flag"
	"regexp"
	"sync"
	"time"
)

var (
	TimeLocLondon     *time.Location //用于伦敦夏令时
	GrpcErrCount      *uint64        //grpc出错数
	SymbolUpdateCount *uint64        // 交易对更新次数
	yamlFile          = flag.String("file", "../config.yaml", "(Simple) YAML file to read")
	sp                = []sync.Pool{}
	ChanPageReload    = make(chan bool, 1000) //用于统计 完成采集有效次数（每页）
	Re                *regexp.Regexp          //匹配mql5
	Conf              Config
)

type Config struct { //config.yaml配置文件
	Config struct {
		Datahost      string        `yaml:"datahost"`
		GrpgServerArr []string      `yaml:"grpgServerArr"`
		Database      string        `yaml:"database"`
		PageRefresh   time.Duration `yaml:"pageRefresh"`
	}
}

//const (
//	Grpchost = "192.168.0.74"
//	Database = "mql5"
//	Datahost = "127.0.0.1"
//)
//var GrpcServerArr = []string{"127.0.0.1", "localhost"}

//const (
//	Database = "mql5"
//	Datahost = "127.0.0.1"
//)

type Mql5 struct { //database
	Type     string
	Symbol   string
	Title    string
	Bid      float64
	Ask      float64
	UnixNano int64
	TimeStr  string
}

var SyncMap *sync.Map //缓存最新 mql5结构的数据
