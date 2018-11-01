package log4go

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
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

var NewLine = []byte("\n")

var k_LOG_LEVEL_SHORT_NAMES = []string{
	"[D]",
	"[I]",
	"[W]",
	"[F]",
	"[P]",
}

//30 black		黑色
//31 red		红色
//32 green		绿色
//33 yellow		黄色
//34 blue		蓝色
//35 magenta    洋红
//36 cyan		蓝绿色
//37 white		白色

//LevelDebug = "Debug"		绿色  	32
//LevelInfo  = "Info"		蓝色  	34
//LevelWarn  = "Warn"    	黄色  	33
//LevelFatal = "Fatal"   	洋红  	35
//LevelPanic = "Panic"   	红色  	31

type color func(string) string

func newColor(c string) color {
	return func(t string) string {
		return "\033[1;" + c + "m" + t + "\033[0m"
	}
}

var k_CONSOLE_COLORS = []color{
	newColor("32"),
	newColor("34"),
	newColor("33"),
	newColor("35"),
	newColor("31"),
}

var isWindows bool

func init() {
	if runtime.GOOS == "windows" {
		isWindows = true
	}
}

type LogMessage struct {
	level     int
	file      string
	line      int
	header    string
	levelName string
	message   string
	created   time.Time

	bytes []byte
}

func newMessage(level int, file string, line int, prefix, msg string) *LogMessage {
	var m = &LogMessage{}
	m.created = time.Now()
	m.level = level
	m.file = file
	m.line = line
	m.levelName = prefix
	m.message = msg
	month, day, year := m.created.Month(), m.created.Day(), m.created.Year()
	hour, minute, second := m.created.Hour(), m.created.Minute(), m.created.Second()
	m.header = fmt.Sprintf("%04d/%02d/%02d %02d:%02d:%02d", year, month, day, hour, minute, second)
	return m
}

func (this *LogMessage) Bytes(c bool) []byte {
	var buf bytes.Buffer
	buf.WriteString(this.header)
	buf.WriteString(" ")
	if c && isWindows == false {
		buf.WriteString(k_CONSOLE_COLORS[this.level](this.levelName))
	} else {
		buf.WriteString(this.levelName)
	}
	buf.WriteString(" [")
	buf.WriteString(this.file)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(this.line))
	buf.WriteString("] ")
	buf.WriteString(this.message)
	return buf.Bytes()
}

type LogWriter interface {
	WriteMessage(msg *LogMessage)
	Close() error
	Level() int
}

type Logger struct {
	writers    map[string]LogWriter
	printStack bool
	stackLevel int
}

func NewLogger() *Logger {
	var l = &Logger{}
	l.writers = make(map[string]LogWriter)
	l.stackLevel = K_LOG_LEVEL_WARNING
	l.printStack = false
	return l
}

func (this *Logger) SetStackLevel(level int) {
	this.stackLevel = level
}

func (this *Logger) GetStackLevel() int {
	return this.stackLevel
}

func (this *Logger) EnableStack() {
	this.printStack = true
}

func (this *Logger) DisableStack() {
	this.printStack = false
}

func (this *Logger) PrintStack() bool {
	return this.printStack
}

func (this *Logger) WriteMessage(level int, msg string) {
	var callDepth = 2
	if this == Default {
		callDepth = 3
	}

	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		file = "???"
		line = -1
	}

	var prefix = k_LOG_LEVEL_SHORT_NAMES[level]

	if this.printStack && level >= this.stackLevel {
		var buf [4096]byte
		n := runtime.Stack(buf[:], true)
		msg += string(buf[:n])
		msg += "\n"
	}

	var logMsg = newMessage(level, file, line, prefix, msg)

	for _, writer := range this.writers {
		if writer.Level() <= logMsg.level {
			writer.WriteMessage(logMsg)
		}
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
var Default *Logger
var once sync.Once

func init() {
	once.Do(func() {
		Default = NewLogger()
		Default.AddWriter("default_console", NewConsoleWriter(K_LOG_LEVEL_DEBUG))
	})
}

func DefaultLogger() *Logger {
	return Default
}

func Debugf(format string, args ...interface{}) {
	Default.Debugf(format, args...)
}

func Debugln(args ...interface{}) {
	Default.Debugln(args...)
}

func Printf(format string, args ...interface{}) {
	Default.Printf(format, args...)
}

func Println(args ...interface{}) {
	Default.Println(args...)
}

func Infof(format string, args ...interface{}) {
	Default.Infof(format, args...)
}

func Infoln(args ...interface{}) {
	Default.Infoln(args...)
}

func Warnf(format string, args ...interface{}) {
	Default.Warnf(format, args...)
}

func Warnln(args ...interface{}) {
	Default.Warnln(args...)
}

func Panicf(format string, args ...interface{}) {
	Default.Panicf(format, args...)
}

func Panicln(args ...interface{}) {
	Default.Panicln(args...)
}

func Fatalf(format string, args ...interface{}) {
	Default.Fatalf(format, args...)
}

func Fatalln(args ...interface{}) {
	Default.Fatalln(args...)
}
