# log4go

轻量级log模块，源于google的一项log工程，官方已经停止维护更新，这里fork一份自用。

>with go1.16 and support go mod 

### Install

`go get github.com/kdpujie/log4go`

### example
样例代码在example文件夹下，直接运行时注意conf.yml的位置
```go
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
```

### Features

* 日志输出到文件，支持按日期对文件进行分割
* 日志输出到控制台
* 支持syslog协议.
* 支持写入阿里云loghub
