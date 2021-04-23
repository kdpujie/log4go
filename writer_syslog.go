// +build !windows,!nacl,!plan9

package log4go

import (
	"errors"
	"log/syslog"
)

// ShortRecord short record
type ShortRecord Record

// String string
func (r *ShortRecord) String() string {
	return "<" + r.code + "> " + r.info
}

// SyslogWriter sys log writer
type SyslogWriter struct {
	level   int
	network string
	addr    string
	tag     string
	writer  *syslog.Writer
}

func NewSyslogWriter() *SyslogWriter {
	return &SyslogWriter{}
}

func (w *SyslogWriter) SetNetwork(network string) {
	w.network = network
}

func (w *SyslogWriter) SetAddr(addr string) {
	w.addr = addr
}

func (w *SyslogWriter) SetTag(tag string) {
	w.tag = tag
}

func (w *SyslogWriter) Init() (err error) {
	w.writer, err = syslog.Dial(w.network, w.addr, syslog.LOG_SYSLOG, w.tag)
	return
}

func (w *SyslogWriter) Write(r *Record) (err error) {
	if r.level < w.level {
		return
	}
	s := ((*ShortRecord)(r)).String()

	switch r.level {
	case DEBUG:
		err = w.writer.Debug(s)
	case INFO:
		err = w.writer.Info(s)
	case WARNING:
		err = w.writer.Warning(s)
	case ERROR:
		err = w.writer.Err(s)
	case FATAL:
		err = w.writer.Crit(s)
	default:
		err = errors.New("invalid level")
	}
	return
}
