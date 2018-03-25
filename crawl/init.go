package crawl

import (
	"sync"
	"time"
)

func init() {
	var err error
	TimeLocLondon, err = time.LoadLocation("Europe/London")
	if err != nil {
		panic(err)
	}
	GrpcErrCount = new(uint64)
	SymbolUpdateCount = new(uint64)
	yamlInit()
	//GrpcPoolInit()
	SyncMap = new(sync.Map)
	//	panic(Conf.Database.Database)
	go yamlWatcher()
}
