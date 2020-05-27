package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/kdpujie/log4go"
)

func main() {
	file := flag.String("c", "log.json", "default log config file")
	flag.Parse()

	fmt.Println(*file)
	if err := log.SetupLogWithConf(*file); err != nil {
		panic(err)
	}
	defer log.Close()

	var name = "skoo"
	log.SetLayout("")
	log.Debug("log4go by %s debug", name)
	log.Info("log4go by %s info", name)
	log.Warn("log4go by %s warn", name)
	log.Error("log4go by %s error", name)
	log.Fatal("log4go by %s fatal", name)

	time.Sleep(1 * time.Second)
}
