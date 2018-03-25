package crawl

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

func readFile(path string) *[]byte {
	fi, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	// fmt.Println(string(fd))
	return &fd
}

var yamlLock sync.Mutex

func yamlInit() {
	yamlLock.Lock()
	defer yamlLock.Unlock()

	flag.Parse()
	b := readFile(*yamlFile)
	err := yaml.Unmarshal([]byte(*b), &Conf)
	if err != nil {
		log.Fatalf("readfile(%q): %s", *yamlFile, err)
	}

	SpInit()
}

func yamlWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				//log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					yamlInit()
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("../config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
