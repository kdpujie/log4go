package log4go

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	LevelFlags = [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	recordPool *sync.Pool
)

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	FATAL
)

const tunnelSizeDefault = 1024

// Record record struct
type Record struct {
	time  string
	code  string
	info  string
	level int
}

// String record string
func (r *Record) String() string {
	return fmt.Sprintf("%s [%s] <%s> %s\n", r.time, LevelFlags[r.level], r.code, r.info)
}

// Writer writer interface
type Writer interface {
	Init() error
	Write(*Record) error
}

// Rotater rotate interface
type Rotater interface {
	Rotate() error
	SetPathPattern(string) error
}

// Flusher flush interface
type Flusher interface {
	Flush() error
}

// Logger log struct
type Logger struct {
	writers []Writer
	tunnel  chan *Record
	// level       int
	lastTime    int64
	lastTimeStr string
	c           chan bool
	layout      string

	fullPath bool // show full path, default only show file:line_number
	lock     sync.RWMutex
}

// NewLogger create the logger instance
func NewLogger() *Logger {
	if loggerDefault != nil && !takeUP {
		takeUP = true
		return loggerDefault
	}

	l := new(Logger)
	l.writers = make([]Writer, 0, 2)
	l.tunnel = make(chan *Record, tunnelSizeDefault)
	l.c = make(chan bool, 1)
	// l.level = DEBUG
	l.layout = "2006/01/02 15:04:05"

	go bootstrapLogWriter(l)

	return l
}

// Register register logger writer
func (l *Logger) Register(w Writer) {
	if err := w.Init(); err != nil {
		panic(err)
	}
	l.writers = append(l.writers, w)
}

// SetLevel Logger set level
func (l *Logger) SetLevel(lvl int) {
	// l.level = lvl
}

// SetLayout Logger set the time data format, layout
func (l *Logger) SetLayout(layout string) {
	l.layout = layout
}

// Debug Logger deliver record to writer
func (l *Logger) Debug(fmt string, args ...interface{}) {
	l.deliverRecordToWriter(DEBUG, fmt, args...)
}

// Warn Logger deliver record to writer
func (l *Logger) Warn(fmt string, args ...interface{}) {
	l.deliverRecordToWriter(WARNING, fmt, args...)
}

// Info Logger deliver record to writer
func (l *Logger) Info(fmt string, args ...interface{}) {
	l.deliverRecordToWriter(INFO, fmt, args...)
}

// Error Logger deliver record to writer
func (l *Logger) Error(fmt string, args ...interface{}) {
	l.deliverRecordToWriter(ERROR, fmt, args...)
}

// Fatal Logger deliver record to writer
func (l *Logger) Fatal(fmt string, args ...interface{}) {
	l.deliverRecordToWriter(FATAL, fmt, args...)
}

// Close Logger close buffer, flush and stop logger
func (l *Logger) Close() {
	close(l.tunnel)
	<-l.c

	for _, w := range l.writers {
		if f, ok := w.(Flusher); ok {
			if err := f.Flush(); err != nil {
				log.Println(err)
			}
		}
	}
}

func (l *Logger) deliverRecordToWriter(level int, format string, args ...interface{}) {
	var inf, code string

	/*	if level < l.level {
		return
	}*/

	if format != "" {
		inf = fmt.Sprintf(format, args...)
	} else {
		inf = fmt.Sprint(args...)
	}

	// source code, file and line num
	_, file, line, ok := runtime.Caller(2)
	if ok {
		if l.fullPath {
			code = file + ":" + strconv.Itoa(line)
		} else {
			code = path.Base(file) + ":" + strconv.Itoa(line)
		}
	}

	// format time
	now := time.Now()
	l.lock.Lock() // avoid data race
	if now.Unix() != l.lastTime {
		l.lastTime = now.Unix()
		l.lastTimeStr = now.Format(l.layout)
	}
	lastTimeStr := l.lastTimeStr
	l.lock.Unlock()

	r := recordPool.Get().(*Record)
	r.info = inf
	r.code = code
	// r.time = l.lastTimeStr
	r.time = lastTimeStr
	r.level = level

	l.tunnel <- r
}

func bootstrapLogWriter(logger *Logger) {
	if logger == nil {
		panic("logger is nil")
	}

	var (
		r  *Record
		ok bool
	)

	if r, ok = <-logger.tunnel; !ok {
		logger.c <- true
		return
	}

	for _, w := range logger.writers {
		if err := w.Write(r); err != nil {
			log.Println(err)
		}
	}

	flushTimer := time.NewTimer(time.Millisecond * 500)
	rotateTimer := time.NewTimer(time.Second * 10)

	for {
		select {
		case r, ok = <-logger.tunnel:
			if !ok {
				logger.c <- true
				return
			}

			for _, w := range logger.writers {
				if err := w.Write(r); err != nil {
					log.Println(err)
				}
			}

			recordPool.Put(r)

		case <-flushTimer.C:
			for _, w := range logger.writers {
				if f, ok := w.(Flusher); ok {
					if err := f.Flush(); err != nil {
						log.Println(err)
					}
				}
			}
			flushTimer.Reset(time.Millisecond * 1000)

		case <-rotateTimer.C:
			for _, w := range logger.writers {
				if r, ok := w.(Rotater); ok {
					if err := r.Rotate(); err != nil {
						log.Println(err)
					}
				}
			}
			rotateTimer.Reset(time.Second * 10)
		}
	}
}

// default logger
var (
	loggerDefault *Logger
	takeUP        = false
)

// SetLevel global set level is ignore
// logger level should be set in specific logger
func SetLevel(lvl int) {
	// loggerDefault.level = lvl
}

// SetLayout loggerDefault set the time format layout
func SetLayout(layout string) {
	loggerDefault.layout = layout
}

// Debug loggerDefault deliver record to writer
func Debug(fmt string, args ...interface{}) {
	loggerDefault.deliverRecordToWriter(DEBUG, fmt, args...)
}

// Warn loggerDefault deliver record to writer
func Warn(fmt string, args ...interface{}) {
	loggerDefault.deliverRecordToWriter(WARNING, fmt, args...)
}

// Info loggerDefault deliver record to writer
func Info(fmt string, args ...interface{}) {
	loggerDefault.deliverRecordToWriter(INFO, fmt, args...)
}

// Error loggerDefault deliver record to writer
func Error(fmt string, args ...interface{}) {
	loggerDefault.deliverRecordToWriter(ERROR, fmt, args...)
}

// Fatal loggerDefault deliver record to writer
func Fatal(fmt string, args ...interface{}) {
	loggerDefault.deliverRecordToWriter(FATAL, fmt, args...)
}

// Register loggerDefault register writer
func Register(w Writer) {
	loggerDefault.Register(w)
}

// Close loggerDefault close logger
func Close() {
	loggerDefault.Close()
}

// ShowFullPath loggerDefault show full path
func ShowFullPath(show bool) {
	loggerDefault.fullPath = show
}

func init() {
	loggerDefault = NewLogger()
	recordPool = &sync.Pool{New: func() interface{} {
		return &Record{}
	}}
}
