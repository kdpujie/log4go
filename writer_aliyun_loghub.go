/**
@description  把log发送值阿里云的loghub中
@author kdpujie
@data	2018-03-16
**/

package log4go

import (
	"time"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"google.golang.org/protobuf/proto"
)

var DefaultBufSize = 10

// AliLogHubWriter ali log hub writer
type AliLogHubWriter struct {
	level   int
	config  *ConfAliLogHubWriter
	project *sls.LogProject
	store   *sls.LogStore
	bufLogs []*sls.Log
	n       int
	err     error
}

// NewAliLogHubWriter create new ali log hub writer
func NewAliLogHubWriter(conf *ConfAliLogHubWriter) *AliLogHubWriter {
	if conf.BufSize == 0 {
		conf.BufSize = DefaultBufSize
	}
	return &AliLogHubWriter{
		level:   getLevel(conf.Level),
		config:  conf,
		bufLogs: make([]*sls.Log, conf.BufSize),
	}
}

func NewAliLogHubWriterWithLevel(level int, conf *ConfAliLogHubWriter) *AliLogHubWriter {
	defaultLevel := DEBUG
	maxLevel := len(LevelFlags)
	// maxLevel >= 1 always true
	maxLevel = maxLevel - 1

	if level >= defaultLevel && level <= maxLevel {
		defaultLevel = level
	}
	return &AliLogHubWriter{
		level:  defaultLevel,
		config: conf,
	}
}

// Init init ali log hub writer init
func (w *AliLogHubWriter) Init() (err error) {
	if w.project, err = sls.NewLogProject(w.config.ProjectName, w.config.Endpoint, w.config.AccessKeyId, w.config.AccessKeySecret); err != nil {
		return err
	}
	w.project.UsingHTTP = true
	w.store, err = w.project.GetLogStore(w.config.LogStoreName)
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
		Topic:  proto.String(w.config.Topic),
		Source: proto.String(w.config.Source),
		Logs:   w.bufLogs[0:w.n],
	}
	if w.err = w.store.PutLogs(logGroup); w.err != nil {
		return w.err
	}
	w.n = 0
	return nil
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
