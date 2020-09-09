package logger

import (
	"fmt"
	"io"
	"log"
)

func New(w io.Writer, tag string) *Logger {
	return &Logger{
		stdlog: log.New(w, "", 0),
		tag:    tag,
	}
}

type Logger struct {
	stdlog  *log.Logger
	tag     string
	Verbose bool
}

func (l *Logger) Debug(v ...interface{}) {
	if !l.Verbose {
		return
	}
	l.print(l.tag, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.print(l.tag, v...)
}

func (l *Logger) print(tag string, v ...interface{}) {
	if l.Verbose {
		tag = fmt.Sprintf("%12v | ", tag)
		l.stdlog.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
		v = append([]interface{}{tag}, v...)
	}
	l.stdlog.Print(v...)
	if l.Verbose {
		l.stdlog.SetFlags(0)
	}
}

type InnerLogger struct {
	*Logger
	tag string
}

func NewFromLogger(l *Logger, tag string) *InnerLogger {
	return &InnerLogger{
		Logger: l,
		tag:    tag,
	}
}

func (l *InnerLogger) Debug(v ...interface{}) {
	if !l.Verbose {
		return
	}
	l.print(l.tag, v...)
}

func (l *InnerLogger) Info(v ...interface{}) {
	l.print(l.tag, v...)
}
