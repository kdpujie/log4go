package main

import log "github.com/kdpujie/log4go"

var (
	ProjectName     = ""
	EndPoint        = ""
	AccessKeyID     = ""
	AccessKeySecret = ""
	LogStoreName    = ""
)

func main() {
	w := log.NewAliLogHubWriter(2048)
	w.SetLog("log4go-test", "")
	w.SetProject(ProjectName, LogStoreName)
	w.SetEndpoint(EndPoint)
	w.SetAccessKey(AccessKeyID, AccessKeySecret)
	log.Register(w)
	log.SetLevel(log.DEBUG)
	defer log.Close()

	log.Info("ali-log-hub Writer for log4go")
	log.Debug("ali-log-hub Writer for log4go")
	log.Error("ali-log-hub Writer for log4go")
}
