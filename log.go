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

// --------------------------------------------------------------------------------
const (
	LevelTrace   = iota // "Trace
	LevelDebug          // "Debug"
	LevelInfo           // "Info"
	LevelWarning        // "Warning"
	LevelError          // "Error"
	LevelPanic          // "Panic"
	LevelFatal          // "Fatal"
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

// --------------------------------------------------------------------------------
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

// --------------------------------------------------------------------------------
type Logger interface {
	SetService(service string)
	Service() string

	SetInstance(instance string)
	Instance() string

	SetPrefix(prefix string)
	Prefix() string

	SetStackLevel(level int)
	StackLevel() int

	EnableStack()
	DisableStack()
	PrintStack() bool

	EnablePath()
	DisablePath()
	PrintPath() bool

	WriteMessage(callDepth, level int, msg string)

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

	Output(calldepth int, s string) error
}

// --------------------------------------------------------------------------------
type Writer interface {
	io.WriteCloser

	Level() int

	WriteMessage(logTime time.Time, service, instance, prefix, timeStr string, level int, levelName, file string, line int, msg string)
}

type logger struct {
	mu         sync.Mutex
	writers    map[string]Writer
	prefix     string
	service    string
	instance   string
	printStack bool
	stackLevel int
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

func (this *logger) SetStackLevel(level int) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.stackLevel = level
}

func (this *logger) StackLevel() int {
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

func (this *logger) WriteMessage(callDepth, level int, msg string) {
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
	var nowStr = now.Format("2006/01/02 15:04:05.000000")
	var levelName = LevelNames[level]

	for _, w := range this.writers {
		if w.Level() <= level {
			w.WriteMessage(now, this.service, this.instance, this.prefix, nowStr, level, levelName, file, line, msg)
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

// log
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

// trace
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

// print
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

// debug
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

// info
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

// warn
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

// error
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

//panic
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

// fatal
func (this *logger) Fatalf(format string, args ...interface{}) {
	this.WriteMessage(2, LevelFatal, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *logger) Fatalln(args ...interface{}) {
	this.WriteMessage(2, LevelFatal, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *logger) Fatal(args ...interface{}) {
	this.WriteMessage(2, LevelFatal, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *logger) Output(calldepth int, s string) error {
	this.WriteMessage(calldepth+1, LevelTrace, s)
	return nil
}

// --------------------------------------------------------------------------------
var sharedInstance Logger
var once sync.Once

func init() {
	once.Do(func() {
		sharedInstance = New()
		sharedInstance.AddWriter("stdout", NewStdWriter(LevelTrace))
	})
}

func SharedInstance() Logger {
	return sharedInstance
}

func SetPrefix(prefix string) {
	sharedInstance.SetPrefix(prefix)
}

func Prefix() string {
	return sharedInstance.Prefix()
}

func AddWriter(name string, w Writer) {
	sharedInstance.AddWriter(name, w)
}

func RemoveWriter(name string) {
	sharedInstance.RemoveWriter(name)
}

// log
func Logf(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Logln(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func Log(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func L(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

// trace
func Tracef(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Traceln(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func Trace(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func T(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

// print
func Printf(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

func Println(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func Print(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintln(args...))
}

func P(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelTrace, fmt.Sprintf(format, args...))
}

// debug
func Debugf(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelDebug, fmt.Sprintf(format, args...))
}

func Debugln(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelDebug, fmt.Sprintln(args...))
}

func Debug(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelDebug, fmt.Sprintln(args...))
}

func D(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelDebug, fmt.Sprintf(format, args...))
}

// info
func Infof(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelInfo, fmt.Sprintf(format, args...))
}

func Infoln(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelInfo, fmt.Sprintln(args...))
}

func Info(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelInfo, fmt.Sprintln(args...))
}

func I(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelInfo, fmt.Sprintf(format, args...))
}

// error
func Errorf(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelError, fmt.Sprintf(format, args...))
}

func Errorln(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelError, fmt.Sprintln(args...))
}

func Error(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelError, fmt.Sprintln(args...))
}

func E(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelError, fmt.Sprintf(format, args...))
}

// warn
func Warnf(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelWarning, fmt.Sprintf(format, args...))
}

func Warnln(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelWarning, fmt.Sprintln(args...))
}

func Warn(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelWarning, fmt.Sprintln(args...))
}

func W(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelWarning, fmt.Sprintf(format, args...))
}

// panic
func Panicf(format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	sharedInstance.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

func Panicln(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	sharedInstance.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

func Panic(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	sharedInstance.WriteMessage(2, LevelPanic, msg)
	panic(msg)
}

// fatal
func Fatalf(format string, args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelFatal, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func Fatalln(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func Fatal(args ...interface{}) {
	sharedInstance.WriteMessage(2, LevelFatal, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func Output(calldepth int, s string) error {
	sharedInstance.WriteMessage(calldepth+1, LevelTrace, s)
	return nil
}
