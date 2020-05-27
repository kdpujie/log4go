package main

import (
	"fmt"
	"time"

	log "github.com/kdpujie/log4go"
)

func SetLog() {
	w := log.NewFileWriterWithLevel(log.ERROR)
	/*
	   %Y  year    (eg: 2014)
	   %M  month   (eg: 07)
	   %D  day     (eg: 05)
	   %H  hour    (eg: 18)
	   %m  minute  (eg: 29)

	   notice: No second's variable
	*/
	if err := w.SetPathPattern("/tmp/logs/error%Y%M%D%H.log"); err != nil {
		fmt.Printf("file writer SetPathPattern err:%v\n", err)
		return
	}

	log.Register(w)
}

func main() {
	SetLog()
	defer log.Close()

	var name = "skoo"

	for {
		log.Debug("log4go by %s", name)
		log.Info("log4go by %s", name)
		log.Warn("log4go by %s", name)
		log.Error("log4go by %s", name)
		log.Fatal("log4go by %s", name)

		time.Sleep(time.Second * 1)
		break
	}
}
