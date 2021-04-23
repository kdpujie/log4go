package log4go

import (
	"fmt"
	"os"
)

type colorRecord Record

func (r *colorRecord) String() string {
	switch r.level {
	case DEBUG:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[34m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LevelFlags[r.level], r.code, r.info)

	case INFO:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[32m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LevelFlags[r.level], r.code, r.info)

	case WARNING:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[33m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LevelFlags[r.level], r.code, r.info)

	case ERROR:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[31m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LevelFlags[r.level], r.code, r.info)

	case FATAL:
		return fmt.Sprintf("\033[36m%s\033[0m [\033[35m%s\033[0m] \033[47;30m%s\033[0m %s\n",
			r.time, LevelFlags[r.level], r.code, r.info)
	}

	return ""
}

// ConsoleWriter console writer define
type ConsoleWriter struct {
	config *ConfConsoleWriter
	level  int
}

// NewConsoleWriter create new console writer
func NewConsoleWriter(conf *ConfConsoleWriter) *ConsoleWriter {
	return &ConsoleWriter{config: conf, level: getLevel(conf.Level)}
}

// NewConsoleWriterWithLevel create new console writer with level
func NewConsoleWriterWithLevel(level int, conf *ConfConsoleWriter) *ConsoleWriter {
	defaultLevel := DEBUG
	maxLevel := len(LevelFlags)
	// maxLevel >= 1 always true
	maxLevel = maxLevel - 1

	if level >= defaultLevel && level <= maxLevel {
		defaultLevel = level
	}
	return &ConsoleWriter{
		level:  defaultLevel,
		config: conf,
	}
}

// Write console write
func (w *ConsoleWriter) Write(r *Record) (err error) {
	if r.level < w.level {
		return nil
	}
	if w.config.Color {
		_, err = fmt.Fprint(os.Stdout, ((*colorRecord)(r)).String())
	} else {
		_, err = fmt.Fprint(os.Stdout, r.String())
	}
	return nil
}

// Init console init without implement
func (w *ConsoleWriter) Init() error {
	return nil
}
