package log4go

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	K_LOG_LEVEL_DEBUG   = iota //= "Debug"
	K_LOG_LEVEL_INFO           //= "Info"
	K_LOG_LEVEL_WARNING        //= "Warning"
	K_LOG_LEVEL_PANIC          //= "Panic"
	K_LOG_LEVEL_FATAL          //= "Fatal"
)

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

func green(c string) string {
	return fmt.Sprintf("\033[1;32m%s\033[0m", c)
}

func blue(c string) string {
	return fmt.Sprintf("\033[1;34m%s\033[0m", c)
}

func yellow(c string) string {
	return fmt.Sprintf("\033[1;33m%s\033[0m", c)
}

func magenta(c string) string {
	return fmt.Sprintf("\033[1;35m%s\033[0m", c)
}

func red(c string) string {
	return fmt.Sprintf("\033[1;31m%s\033[0m", c)
}

var (
	levelShortNames = []string{
		"[D]",
		"[I]",
		"[W]",
		"[P]",
		"[F]",
	}

	levelWithColors = []string{
		green(levelShortNames[0]),
		blue(levelShortNames[1]),
		yellow(levelShortNames[2]),
		magenta(levelShortNames[3]),
		red(levelShortNames[4]),
	}
)

type Writer interface {
	io.WriteCloser
	Level() int
	EnableColor() bool
}

type Logger struct {
	writers    map[string]Writer
	printStack bool
	stackLevel int
	printPath  bool
	printColor bool
}

func NewLogger() *Logger {
	var l = &Logger{}
	l.writers = make(map[string]Writer)
	l.stackLevel = K_LOG_LEVEL_PANIC
	l.printStack = false
	l.printPath = true
	l.printColor = true
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

func (this *Logger) EnablePath() {
	this.printPath = true
}

func (this *Logger) DisablePath() {
	this.printPath = false
}

func (this *Logger) PrintPath() bool {
	return this.printPath
}

func (this *Logger) EnableColor() {
	this.printColor = true
}

func (this *Logger) DisableColor() {
	this.printColor = false
}

func (this *Logger) PrintColor() bool {
	return this.printColor
}

func (this *Logger) WriteMessage(level int, msg string) {
	var callDepth = 2
	if this == Default {
		callDepth = 3
	}

	_, file, line, ok := runtime.Caller(callDepth)
	if ok {
		if this.printPath == false {
			_, file = filepath.Split(file)
		}
	} else {
		file = "???"
		line = -1
	}

	if this.printStack && level >= this.stackLevel {
		var buf [4096]byte
		n := runtime.Stack(buf[:], true)
		msg += string(buf[:n])
		msg += "\n"
	}

	var now = time.Now()
	var levelName string

	for _, w := range this.writers {
		if w.Level() <= level {
			if this.printColor && w.EnableColor() {
				levelName = levelWithColors[level]
			} else {
				levelName = levelShortNames[level]
			}
			fmt.Fprintf(w, "%s %s [%s:%d] %s", now.Format("2006/01/02 15:04:05"), levelName, file, line, msg)
		}
	}
}

func (this *Logger) AddWriter(name string, w Writer) {
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

//panic
func (this *Logger) Panicf(format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	this.WriteMessage(K_LOG_LEVEL_PANIC, msg)
	panic(msg)
}

func (this *Logger) Panicln(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	this.WriteMessage(K_LOG_LEVEL_PANIC, msg)
	panic(msg)
}

//fatal
func (this *Logger) Fatalf(format string, args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_FATAL, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *Logger) Fatalln(args ...interface{}) {
	this.WriteMessage(K_LOG_LEVEL_FATAL, fmt.Sprintln(args...))
	os.Exit(-1)
}

// --------------------------------------------------------------------------------
var Default *Logger
var once sync.Once

func init() {
	once.Do(func() {
		Default = NewLogger()
		Default.AddWriter("std_out", NewStdWriter(K_LOG_LEVEL_DEBUG))
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
