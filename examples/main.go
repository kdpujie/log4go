/**
@description
@author pujie
@data	2021-04-22
**/
package main

import (
	"fmt"
	"github.com/kdpujie/log4go"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

// SetLog set logger
func InitLog() {
	pwd, _ := os.Getwd()
	viper.AddConfigPath(pwd)
	viper.SetConfigType("yml")
	viper.SetConfigName("conf.yml")
	fmt.Printf("路径：%s \n", pwd)
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
	logConfig := &log4go.LogConfig{}
	if err := viper.UnmarshalKey("log4go", logConfig); err != nil {
		log.Fatalln("[err] Unmarshal log4go config error, ", err)
	}
	if err := log4go.SetupLog(*logConfig); err != nil {
		log.Fatalln("[err] Setup log4go config, ", err)
	}
}

func main() {
	InitLog()
	log4go.Info("console Writer for log4go")
	log4go.Info("console Writer for log4go")
	log4go.Debug("console Writer for log4go")
	log4go.Error("console Writer for log4go")
	time.Sleep(2 * time.Second)
}
