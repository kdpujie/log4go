/**
@description  把log发送值阿里云的loghub中
@author kdpujie
@data	2018-03-16
**/

package log4go

import (
	"time"

	"github.com/aliyun/aliyun-log-go-sdk"
	"google.golang.org/protobuf/proto"
)

// AliLogHubWriter ali log hub writer
type AliLogHubWriter struct {
	level           int
	logName         string
	logSource       string
	projectName     string
	endpoint        string
	accessKeyId     string
	accessKeySecret string
	storeName       string
	project         *sls.LogProject
	store           *sls.LogStore
	bufLogs         []*sls.Log
	n               int
	err             error
}

// NewAliLogHubWriter create new ali log hub writer
func NewAliLogHubWriter(bufSize int) *AliLogHubWriter {
	return &AliLogHubWriter{
		bufLogs: make([]*sls.Log, bufSize),
	}
}

// Init init ali log hub writer init
func (w *AliLogHubWriter) Init() (err error) {
	w.project, err = sls.NewLogProject(w.projectName, w.endpoint, w.accessKeyId, w.accessKeySecret)
	if err != nil {
		return
	}
	w.project.UsingHTTP = true
	w.store, err = w.project.GetLogStore(w.storeName)
	return
}

// Write ali log hub writer write
func (w *AliLogHubWriter) Write(r *Record) (err error) {
	if r.level < w.level {
		return
	}
	var content []*sls.LogContent
	content = append(content, &sls.LogContent{
		Key:   proto.String("time"),
		Value: proto.String(r.time),
	})
	content = append(content, &sls.LogContent{
		Key:   proto.String("level"),
		Value: proto.String(LevelFlags[r.level]),
	})
	content = append(content, &sls.LogContent{
		Key:   proto.String("code"),
		Value: proto.String(r.code),
	})
	content = append(content, &sls.LogContent{
		Key:   proto.String("info"),
		Value: proto.String(r.info),
	})
	log := &sls.Log{
		Time:     proto.Uint32(uint32(time.Now().Unix())),
		Contents: content,
	}
	if err := w.writeBuf(log); err != nil {
		return err
	}
	return
}

// Flush ali log hub writer flush
func (w *AliLogHubWriter) Flush() error {
	if w.err != nil {
		return w.err
	}
	if w.n == 0 {
		return nil
	}
	logGroup := &sls.LogGroup{
		Topic:  proto.String(w.logName),
		Source: proto.String(w.logSource),
		Logs:   w.bufLogs[0:w.n],
	}
	if w.err = w.store.PutLogs(logGroup); w.err != nil {
		return w.err
	}
	w.n = 0
	return nil
}

// SetLog ...
func (w *AliLogHubWriter) SetLog(logName, logSource string) {
	w.logName = logName
	w.logSource = logSource
}

// SetProject ...
func (w *AliLogHubWriter) SetProject(pName, sName string) {
	w.projectName = pName
	w.storeName = sName
}

// SetEndpoint ...
func (w *AliLogHubWriter) SetEndpoint(endpoint string) {
	w.endpoint = endpoint
}

// SetAccessKey ...
func (w *AliLogHubWriter) SetAccessKey(accessKeyId, accessKeySecret string) {
	w.accessKeyId = accessKeyId
	w.accessKeySecret = accessKeySecret
}

func (w *AliLogHubWriter) writeBuf(log *sls.Log) error {
	if w.available() <= 0 && w.Flush() != nil {
		return w.err
	}
	w.bufLogs[w.n] = log
	w.n++
	return nil
}

func (w *AliLogHubWriter) available() int {
	return len(w.bufLogs) - w.n
}
