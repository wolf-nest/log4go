package log4go

import (
	"fmt"
	"path"
	"runtime"
	"sync"
	"time"
)

const (
	K_LOG_LEVEL_DEBUG   = iota //= "Debug"
	K_LOG_LEVEL_INFO           //= "Info"
	K_LOG_LEVEL_WARNING        //= "Warning"
	K_LOG_LEVEL_FATAL          //= "Fatal"
	K_LOG_LEVEL_PANIC          //= "Panic"
)

var k_LOG_LEVEL_SHORT_NAMES = []string{
	"[D]",
	"[I]",
	"[W]",
	"[F]",
	"[P]",
}

type LogMessage struct {
	level     int
	file      string
	line      int
	header    string
	levelName string
	message   string
	created   time.Time
}

type LogWriter interface {
	WriteMessage(msg *LogMessage)
	Close() error
}

type Logger struct {
	writers     map[string]LogWriter
	enableStack bool
	stackLevel  int
}

func NewLogger() *Logger {
	var l = &Logger{}
	l.writers = make(map[string]LogWriter)
	return l
}

func (this *Logger) WriteMessage(level int, msg string) {
	var callDepth = 2
	if this == defaultLogger {
		callDepth = 3
	}

	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		file = "???"
		line = -1
	} else {
		_, file = path.Split(file)
	}

	var prefix = k_LOG_LEVEL_SHORT_NAMES[level]

	if this.enableStack && level >= this.stackLevel {
		buf := make([]byte, 1024*1024)
		n := runtime.Stack(buf, true)
		msg += string(buf[:n])
		msg += "\n"
	}

	var logMsg = &LogMessage{}
	logMsg.created = time.Now()
	logMsg.level = level
	logMsg.file = file
	logMsg.line = line
	logMsg.levelName = prefix
	logMsg.message = msg

	month, day, year := logMsg.created.Month(), logMsg.created.Day(), logMsg.created.Year()
	hour, minute, second := logMsg.created.Hour(), logMsg.created.Minute(), logMsg.created.Second()
	logMsg.header = fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d", year, month, day, hour, minute, second)

	for _, writer := range this.writers {
		writer.WriteMessage(logMsg)
	}
}

func (this *Logger) AddWriter(name string, w LogWriter) {
	this.writers[name] = w
}

func (this *Logger) RemoveWriter(name string) {
	var w = this.writers[name]
	if w != nil {
		w.Close()
	}
	delete(this.writers, name)
}

//debug
func (this *Logger) Debugf(format string, args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_DEBUG, fmt.Sprintf(format, args...))
}

func (this *Logger) Debugln(args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_DEBUG, fmt.Sprintln(args...))
}

//print
func (this *Logger) Printf(format string, args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_DEBUG, fmt.Sprintf(format, args...))
}

func (this *Logger) Println(args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_DEBUG, fmt.Sprintln(args...))
}

//info
func (this *Logger) Infof(format string, args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_INFO, fmt.Sprintf(format, args...))
}

func (this *Logger) Infoln(args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_INFO, fmt.Sprintln(args...))
}

//warn
func (this *Logger) Warnf(format string, args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_WARNING, fmt.Sprintf(format, args...))
}

func (this *Logger) Warnln(args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_WARNING, fmt.Sprintln(args...))
}

//fatal
func (this *Logger) Fatalf(format string, args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_FATAL, fmt.Sprintf(format, args...))
}

func (this *Logger) Fatalln(args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_FATAL, fmt.Sprintln(args...))
}

//panic
func (this *Logger) Panicf(format string, args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_PANIC, fmt.Sprintf(format, args...))
}

func (this *Logger) Panicln(args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_PANIC, fmt.Sprintln(args...))
}

// --------------------------------------------------------------------------------
var defaultLogger *Logger
var once sync.Once

func SharedLogger() *Logger {
	once.Do(func() {
		defaultLogger = NewLogger()
		defaultLogger.AddWriter("default_console", NewConsoleWriter(K_LOG_LEVEL_DEBUG))
	})
	return defaultLogger
}

func Debugf(format string, args ...interface{}) {
	SharedLogger().Debugf(format, args...)
}

func Debugln(args ...interface{}) {
	SharedLogger().Debugln(args...)
}

func Printf(format string, args ...interface{}) {
	SharedLogger().Printf(format, args...)
}

func Println(args ...interface{}) {
	SharedLogger().Println(args...)
}

func Infof(format string, args ...interface{}) {
	SharedLogger().Infof(format, args...)
}

func Infoln(args ...interface{}) {
	SharedLogger().Infoln(args...)
}

func Warnf(format string, args ...interface{}) {
	SharedLogger().Warnf(format, args...)
}

func Warnln(args ...interface{}) {
	SharedLogger().Warnln(args...)
}

func Panicf(format string, args ...interface{}) {
	SharedLogger().Panicf(format, args...)
}

func Panicln(args ...interface{}) {
	SharedLogger().Panicln(args...)
}

func Fatalf(format string, args ...interface{}) {
	SharedLogger().Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	SharedLogger().Fatalln(args...)
}
