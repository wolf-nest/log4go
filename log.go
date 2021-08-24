package log4go

import (
	"context"
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

	WriteMessage(ctx context.Context, callDepth int, level Level, msg string)

	AddWriter(name string, w Writer)
	RemoveWriter(name string)

	Logf(ctx context.Context, format string, args ...interface{})
	Logln(ctx context.Context, args ...interface{})
	Log(ctx context.Context, args ...interface{})
	L(ctx context.Context, format string, args ...interface{})

	Tracef(ctx context.Context, format string, args ...interface{})
	Traceln(ctx context.Context, args ...interface{})
	Trace(ctx context.Context, args ...interface{})
	T(ctx context.Context, format string, args ...interface{})

	Printf(ctx context.Context, format string, args ...interface{})
	Println(ctx context.Context, args ...interface{})
	Print(ctx context.Context, args ...interface{})
	P(ctx context.Context, format string, args ...interface{})

	Debugf(ctx context.Context, format string, args ...interface{})
	Debugln(ctx context.Context, args ...interface{})
	Debug(ctx context.Context, args ...interface{})
	D(ctx context.Context, format string, args ...interface{})

	Infof(ctx context.Context, format string, args ...interface{})
	Infoln(ctx context.Context, args ...interface{})
	Info(ctx context.Context, args ...interface{})
	I(ctx context.Context, format string, args ...interface{})

	Warnf(ctx context.Context, format string, args ...interface{})
	Warnln(ctx context.Context, args ...interface{})
	Warn(ctx context.Context, args ...interface{})
	W(ctx context.Context, format string, args ...interface{})

	Errorf(ctx context.Context, format string, args ...interface{})
	Errorln(ctx context.Context, args ...interface{})
	Error(ctx context.Context, args ...interface{})
	E(ctx context.Context, format string, args ...interface{})

	Panicf(ctx context.Context, format string, args ...interface{})
	Panicln(ctx context.Context, args ...interface{})
	Panic(ctx context.Context, args ...interface{})

	Fatalf(ctx context.Context, format string, args ...interface{})
	Fatalln(ctx context.Context, args ...interface{})
	Fatal(ctx context.Context, args ...interface{})

	Output(ctx context.Context, callDepth int, s string) error
}

type Writer interface {
	io.WriteCloser

	Level() Level

	WriteMessage(logId, service, instance, prefix, logTime string, level Level, file string, line int, msg string)
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

func (this *logger) WriteMessage(ctx context.Context, callDepth int, level Level, msg string) {
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

	var logId = MustGetId(ctx)

	for _, w := range this.writers {
		if w.Level() <= level {
			w.WriteMessage(logId, this.service, this.instance, this.prefix, logTime, level, file, line, msg)
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

func (this *logger) Logf(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Logln(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) Log(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) L(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Tracef(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Traceln(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) Trace(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) T(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Printf(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Println(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) Print(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func (this *logger) P(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func (this *logger) Debugf(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelDebug, fmt.Sprintf(format, args...))
}

func (this *logger) Debugln(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelDebug, fmt.Sprintln(args...))
}

func (this *logger) Debug(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelDebug, fmt.Sprintln(args...))
}

func (this *logger) D(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelDebug, fmt.Sprintf(format, args...))
}

func (this *logger) Infof(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelInfo, fmt.Sprintf(format, args...))
}

func (this *logger) Infoln(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelInfo, fmt.Sprintln(args...))
}

func (this *logger) Info(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelInfo, fmt.Sprintln(args...))
}

func (this *logger) I(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelInfo, fmt.Sprintf(format, args...))
}

func (this *logger) Warnf(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelWarning, fmt.Sprintf(format, args...))
}

func (this *logger) Warnln(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelWarning, fmt.Sprintln(args...))
}

func (this *logger) Warn(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelWarning, fmt.Sprintln(args...))
}

func (this *logger) W(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelWarning, fmt.Sprintf(format, args...))
}

func (this *logger) Errorf(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelError, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *logger) Errorln(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelError, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *logger) Error(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelError, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *logger) E(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelError, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *logger) Panicf(ctx context.Context, format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	this.WriteMessage(ctx, 2, LevelPanic, msg)
	panic(msg)
}

func (this *logger) Panicln(ctx context.Context, args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	this.WriteMessage(ctx, 2, LevelPanic, msg)
	panic(msg)
}

func (this *logger) Panic(ctx context.Context, args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	this.WriteMessage(ctx, 2, LevelPanic, msg)
	panic(msg)
}

func (this *logger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelFatal, fmt.Sprintf(format, args...))
	//os.Exit(-1)
}

func (this *logger) Fatalln(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func (this *logger) Fatal(ctx context.Context, args ...interface{}) {
	this.WriteMessage(ctx, 2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func (this *logger) Output(ctx context.Context, callDepth int, s string) error {
	this.WriteMessage(ctx, callDepth+1, LevelTrace, s)
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

func Logf(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func Logln(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func Log(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func L(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func Tracef(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func Traceln(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func Trace(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func T(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func Printf(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func Println(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func Print(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintln(args...))
}

func P(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelTrace, fmt.Sprintf(format, args...))
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelDebug, fmt.Sprintf(format, args...))
}

func Debugln(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelDebug, fmt.Sprintln(args...))
}

func Debug(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelDebug, fmt.Sprintln(args...))
}

func D(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelDebug, fmt.Sprintf(format, args...))
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelInfo, fmt.Sprintf(format, args...))
}

func Infoln(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelInfo, fmt.Sprintln(args...))
}

func Info(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelInfo, fmt.Sprintln(args...))
}

func I(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelInfo, fmt.Sprintf(format, args...))
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelError, fmt.Sprintf(format, args...))
}

func Errorln(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelError, fmt.Sprintln(args...))
}

func Error(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelError, fmt.Sprintln(args...))
}

func E(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelError, fmt.Sprintf(format, args...))
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelWarning, fmt.Sprintf(format, args...))
}

func Warnln(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelWarning, fmt.Sprintln(args...))
}

func Warn(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelWarning, fmt.Sprintln(args...))
}

func W(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelWarning, fmt.Sprintf(format, args...))
}

func Panicf(ctx context.Context, format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	sharedLogger.WriteMessage(ctx, 2, LevelPanic, msg)
	panic(msg)
}

func Panicln(ctx context.Context, args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	sharedLogger.WriteMessage(ctx, 2, LevelPanic, msg)
	panic(msg)
}

func Panic(ctx context.Context, args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	sharedLogger.WriteMessage(ctx, 2, LevelPanic, msg)
	panic(msg)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelFatal, fmt.Sprintf(format, args...))
	//os.Exit(-1)
}

func Fatalln(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func Fatal(ctx context.Context, args ...interface{}) {
	sharedLogger.WriteMessage(ctx, 2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func Output(ctx context.Context, callDepth int, s string) error {
	sharedLogger.WriteMessage(ctx, callDepth+1, LevelTrace, s)
	return nil
}
