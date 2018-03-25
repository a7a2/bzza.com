package main

import (
	"net/http"
	"regexp"

	"time"

	"bzza.com/crawl"
)

const comNano24, comNano26 = 3600 * 24 * 100000000, 3600 * 26 * 100000000 //用于unixnano时间比较。23，25小时

func init() {
	crawl.Re = regexp.MustCompile(`(<tr id=")([A-Z0-9_#\.]+?)(" class=")([a-z]{0,6})(" title=")(.+?)(" draggable=)(.+?)(</span>)([A-Z0-9_#\.]+?)(</span></div></td><td id="bid" class=")([a-z]{0,4})(" style="text-align: right;"><div class="container"><span class="content">)([0-9\.]+?)(</span></div></td><td id="ask" class=")([a-z]{0,4})(" style="text-align: right;"><div class="container"><span class="content">)([0-9\.]+?)(</span></div></td><td id="spread" class="" style="text-align: right;"><div class="container"><span class="content">)([0-9]+)(</span></div></td><td id="time")(.+?)(class="content">)([0-9_:\.]+)(</span></div></td><td></td></tr>)`)

	//	CacheMap = cacheMql5{
	//		MapLock: make(map[string]*sync.Mutex),
	//		MapMql5: make(map[string]*mql5),
	//	}

	http.DefaultClient.Timeout = time.Second * 6
}
