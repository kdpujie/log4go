package main

import (
	"fmt"

	log "github.com/kdpujie/log4go"
)

// SetLog set logger
func SetLog() {
	w1 := log.NewFileWriterWithLevel(log.ERROR)
	if err := w1.SetPathPattern("/tmp/logs/error%Y%M%D%H.log"); err != nil {
		fmt.Printf("file writer SetPathPattern err:%v\n", err)
		return
	}

	w2 := log.NewConsoleWriterWithLevel(log.WARNING)

	log.Register(w1)
	log.Register(w2)
}

func main() {
	SetLog()
	defer log.Close()

	var name = "skoo"
	log.Debug("log4go by %s", name)
	log.Info("log4go by %s", name)
	log.Warn("log4go by %s", name)
	log.Error("log4go by %s", name)
	log.Fatal("log4go by %s", name)
}
