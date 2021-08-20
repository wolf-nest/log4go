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

type Level int

const (
	LevelTrace   Level = iota // "Trace
	LevelDebug                // "Debug"
	LevelInfo                 // "Info"
	LevelWarning              // "Warning"
	LevelError                // "Error"
	LevelPanic                // "Panic"
	LevelFatal                // "Fatal"
)

var (
	LevelNames = []string{
		"[T]",
		"[D]",
		"[I]",
		"[W]",
		"[E]",
		"[P]",
		"[F]",
	}
)

type Option interface {
	Apply(Logger)
}

type optionFunc func(Logger)

func (f optionFunc) Apply(l Logger) {
	f(l)
}

func WithPrefix(p string) Option {
	return optionFunc(func(l Logger) {
		l.SetPrefix(p)
	})
}

func WithService(s string) Option {
	return optionFunc(func(l Logger) {
		l.SetService(s)
	})
}

func WithInstance(s string) Option {
	return optionFunc(func(l Logger) {
		l.SetInstance(s)
	})
}

type Logger interface {
	SetService(service string)
	Service() string

	SetInstance(instance string)
	Instance() string

	SetPrefix(prefix string)
	Prefix() string

	SetStackLevel(level Level)
	StackLevel() Level

	EnableStack()
	DisableStack()
	PrintStack() bool

	EnablePath()
	DisablePath()
	PrintPath() bool

	WriteMessage(callDepth int, level Level, msg string)

	AddWriter(name string, w Writer)
	RemoveWriter(name string)

	Logf(format string, args ...interface{})
	Logln(args ...interface{})
	Log(args ...interface{})
	L(format string, args ...interface{})

	Tracef(format string, args ...interface{})
	Traceln(args ...interface{})
	Trace(args ...interface{})
	T(format string, args ...interface{})

	Printf(format string, args ...interface{})
	Println(args ...interface{})
	Print(args ...interface{})
	P(format string, args ...interface{})

	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Debug(args ...interface{})
	D(format string, args ...interface{})

	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Info(args ...interface{})
	I(format string, args ...interface{})

	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Warn(args ...interface{})
	W(format string, args ...interface{})

	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Error(args ...interface{})
	E(format string, args ...interface{})

	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
	Panic(args ...interface{})

	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Fatal(args ...interface{})

	Output(callDepth int, s string) error
}

type Writer interface {
	io.WriteCloser

	Level() Level

	WriteMessage(service, instance, prefix, logTime string, level Level, file string, line int, msg string)
}

type logger struct {
	mu         sync.Mutex
	writers    map[string]Writer
	prefix     string
	service    string
	instance   string
	printStack bool
	stackLevel Level
	printPath  bool
}

func New(opts ...Option) Logger {
	var l = &logger{}
	l.writers = make(map[string]Writer)
	l.stackLevel = LevelPanic
	l.printStack = false
	l.printPath = true
	for _, opt := range opts {
		opt.Apply(l)
	}
	return l
}

func (this *logger) SetService(service string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.service = service
}

func (this *logger) Service() string {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.service
}

func (this *logger) SetInstance(instance string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.instance = instance
}

func (this *logger) Instance() string {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.instance
}

func (this *logger) SetPrefix(prefix string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.prefix = prefix
}

func (this *logger) Prefix() string {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.prefix
}

func (this *logger) SetStackLevel(level Level) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.stackLevel = level
}

func (this *logger) StackLevel() Level {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.stackLevel
}

func (this *logger) EnableStack() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printStack = true
}

func (this *logger) DisableStack() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printStack = false
}

func (this *logger) PrintStack() bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.printStack
}

func (this *logger) EnablePath() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printPath = true
}

func (this *logger) DisablePath() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printPath = false
}

func (this *logger) PrintPath() bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.printPath
}

func (this *logger) WriteMessage(callDepth int, level Level, msg string) {
	this.mu.Lock()
	defer this.mu.Unlock()

	var file string
	var line int

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
	var logTime = now.Format("2006/01/02 15:04:05.000000")

	for _, w := range this.writers {
		if w.Level() <= level {
			w.WriteMessage(this.service, this.instance, this.prefix, logTime, level, file, line, msg)
		}
	}
}

func (this *logger) AddWriter(name string, w Writer) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.writers[name] = w
}

func (this *logger) RemoveWriter(name string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	var w = this.writers[name]
	if w != nil {
		w.Close()
	}
	delete(this.writers, name)
}

func (this *logger) Logf(format string, args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Logln(args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) Log(args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) L(format string, args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Tracef(format string, args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Traceln(args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) Trace(args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) T(format string, args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Printf(format string, args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Println(args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) Print(args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) P(format string, args ...interface{}) {
	this.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Debugf(format string, args ...interface{}) {
	this.WriteMessage(2, LevelDebug, fmt.Sprintf(format, args...))
}

func (this *logger) Debugln(args ...interface{}) {
	this.WriteMessage(2, LevelDebug, fmt.Sprintln(args...))
}

func (this *logger) Debug(args ...interface{}) {
	this.WriteMessage(2, LevelDebug, fmt.Sprintln(args...))
}

func (this *logger) D(format string, args ...interface{}) {
	this.WriteMessage(2, LevelDebug, fmt.Sprintf(format, args...))
}

func (this *logger) Infof(format string, args ...interface{}) {
	this.WriteMessage(2, LevelInfo, fmt.Sprintf(format, args...))
}

func (this *logger) Infoln(args ...interface{}) {
	this.WriteMessage(2, LevelInfo, fmt.Sprintln(args...))
}

func (this *logger) Info(args ...interface{}) {
	this.WriteMessage(2, LevelInfo, fmt.Sprintln(args...))
}

func (this *logger) I(format string, args ...interface{}) {
	this.WriteMessage(2, LevelInfo, fmt.Sprintf(format, args...))
}

func (this *logger) Warnf(format string, args ...interface{}) {
	this.WriteMessage(2, LevelWarning, fmt.Sprintf(format, args...))
}

func (this *logger) Warnln(args ...interface{}) {
	this.WriteMessage(2, LevelWarning, fmt.Sprintln(args...))
}

func (this *logger) Warn(args ...interface{}) {
	this.WriteMessage(2, LevelWarning, fmt.Sprintln(args...))
}

func (this *logger) W(format string, args ...interface{}) {
	this.WriteMessage(2, LevelWarning, fmt.Sprintf(format, args...))
}

func (this *logger) Errorf(format string, args ...interface{}) {
	this.WriteMessage(2, LevelError, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *logger) Errorln(args ...interface{}) {
	this.WriteMessage(2, LevelError, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *logger) Error(args ...interface{}) {
	this.WriteMessage(2, LevelError, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *logger) E(format string, args ...interface{}) {
	this.WriteMessage(2, LevelError, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *logger) Panicf(format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	this.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

func (this *logger) Panicln(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	this.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

func (this *logger) Panic(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	this.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

func (this *logger) Fatalf(format string, args ...interface{}) {
	this.WriteMessage(2, LevelFatal, fmt.Sprintf(format, args...))
	//os.Exit(-1)
}

func (this *logger) Fatalln(args ...interface{}) {
	this.WriteMessage(2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func (this *logger) Fatal(args ...interface{}) {
	this.WriteMessage(2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func (this *logger) Output(callDepth int, s string) error {
	this.WriteMessage(callDepth+1, LevelTrace, s)
	return nil
}

var sharedLogger Logger
var once sync.Once

func init() {
	once.Do(func() {
		sharedLogger = New()
		sharedLogger.AddWriter("stdout", NewStdWriter(LevelTrace))
	})
}

func SharedLogger() Logger {
	return sharedLogger
}

func SetPrefix(prefix string) {
	sharedLogger.SetPrefix(prefix)
}

func Prefix() string {
	return sharedLogger.Prefix()
}

func AddWriter(name string, w Writer) {
	sharedLogger.AddWriter(name, w)
}

func RemoveWriter(name string) {
	sharedLogger.RemoveWriter(name)
}

func Logf(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Logln(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func Log(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func L(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Tracef(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Traceln(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func Trace(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func T(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Printf(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Println(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func Print(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func P(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelDebug, fmt.Sprintf(format, args...))
}

func Debugln(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelDebug, fmt.Sprintln(args...))
}

func Debug(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelDebug, fmt.Sprintln(args...))
}

func D(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelDebug, fmt.Sprintf(format, args...))
}

func Infof(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelInfo, fmt.Sprintf(format, args...))
}

func Infoln(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelInfo, fmt.Sprintln(args...))
}

func Info(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelInfo, fmt.Sprintln(args...))
}

func I(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelInfo, fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelError, fmt.Sprintf(format, args...))
}

func Errorln(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelError, fmt.Sprintln(args...))
}

func Error(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelError, fmt.Sprintln(args...))
}

func E(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelError, fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelWarning, fmt.Sprintf(format, args...))
}

func Warnln(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelWarning, fmt.Sprintln(args...))
}

func Warn(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelWarning, fmt.Sprintln(args...))
}

func W(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelWarning, fmt.Sprintf(format, args...))
}

func Panicf(format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	sharedLogger.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

func Panicln(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	sharedLogger.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

func Panic(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	sharedLogger.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

func Fatalf(format string, args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelFatal, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func Fatalln(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func Fatal(args ...interface{}) {
	sharedLogger.WriteMessage(2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func Output(callDepth int, s string) error {
	sharedLogger.WriteMessage(callDepth+1, LevelTrace, s)
	return nil
}
