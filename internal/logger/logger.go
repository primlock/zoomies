package logger

import (
	"log"
	"os"
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
}

type Log struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger
	flag  bool
}

var TLog = NewLog()

func NewLog() *Log {
	return &Log{
		debug: log.New(os.Stdout, "DEBUG:\t", log.LstdFlags),
		info:  log.New(os.Stdout, "INFO:\t", log.LstdFlags),
		warn:  log.New(os.Stdout, "WARN:\t", log.LstdFlags),
		err:   log.New(os.Stdout, "ERROR:\t", log.LstdFlags),
		flag:  false,
	}
}

func (l *Log) Debug(format string, a ...any) {
	if l.flag {
		l.debug.Printf(format, a...)
	}
}

func (l *Log) Info(format string, a ...any) {
	if l.flag {
		l.info.Printf(format, a...)
	}
}

func (l *Log) Warn(format string, a ...any) {
	if l.flag {
		l.warn.Printf(format, a...)
	}
}

func (l *Log) Error(format string, a ...any) {
	l.err.Printf(format, a...)
}

func (l *Log) Verbose() {
	l.flag = true
}
