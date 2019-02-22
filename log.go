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
	K_LOG_LEVEL_TRACE   = iota // "Trace
	K_LOG_LEVEL_DEBUG          // "Debug"
	K_LOG_LEVEL_INFO           // "Info"
	K_LOG_LEVEL_WARNING        // "Warning"
	K_LOG_LEVEL_ERROR          // "Error"
	K_LOG_LEVEL_PANIC          // "Panic"
	K_LOG_LEVEL_FATAL          // "Fatal"
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

// --------------------------------------------------------------------------------
type Logger interface {
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

	Tracef(format string, args ...interface{})
	Traceln(args ...interface{})
	T(format string, args ...interface{})

	Printf(format string, args ...interface{})
	Println(args ...interface{})
	P(format string, args ...interface{})

	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	D(format string, args ...interface{})

	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	I(format string, args ...interface{})

	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	W(format string, args ...interface{})

	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	E(format string, args ...interface{})

	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})

	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})

	Output(calldepth int, s string) error
}

// --------------------------------------------------------------------------------
type Writer interface {
	io.WriteCloser

	Level() int

	WriteMessage(logTime time.Time, prefix, timeStr string, level int, levelName, file string, line int, msg string)
}

type logger struct {
	mu         sync.Mutex
	writers    map[string]Writer
	prefix     string
	printStack bool
	stackLevel int
	printPath  bool
}

func New(opts ...Option) Logger {
	var l = &logger{}
	l.writers = make(map[string]Writer)
	l.stackLevel = K_LOG_LEVEL_PANIC
	l.printStack = false
	l.printPath = true
	for _, opt := range opts {
		opt.Apply(l)
	}
	return l
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
			w.WriteMessage(now, this.prefix, nowStr, level, levelName, file, line, msg)
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

// trace
func (this *logger) Tracef(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

func (this *logger) Traceln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintln(args...))
}

func (this *logger) T(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

//print
func (this *logger) Printf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

func (this *logger) Println(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintln(args...))
}

func (this *logger) P(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

//debug
func (this *logger) Debugf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintf(format, args...))
}

func (this *logger) Debugln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintln(args...))
}

func (this *logger) D(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintf(format, args...))
}

//info
func (this *logger) Infof(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintf(format, args...))
}

func (this *logger) Infoln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintln(args...))
}

func (this *logger) I(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintf(format, args...))
}

//warn
func (this *logger) Warnf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintf(format, args...))
}

func (this *logger) Warnln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintln(args...))
}

func (this *logger) W(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintf(format, args...))
}

//error
func (this *logger) Errorf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *logger) Errorln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *logger) E(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

//panic
func (this *logger) Panicf(format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	this.WriteMessage(2, K_LOG_LEVEL_PANIC, msg)
	panic(msg)
}

func (this *logger) Panicln(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	this.WriteMessage(2, K_LOG_LEVEL_PANIC, msg)
	panic(msg)
}

//fatal
func (this *logger) Fatalf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_FATAL, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *logger) Fatalln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_FATAL, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *logger) Output(calldepth int, s string) error {
	this.WriteMessage(calldepth+1, K_LOG_LEVEL_TRACE, s)
	return nil
}

// --------------------------------------------------------------------------------
var defaultLogger Logger
var once sync.Once

func init() {
	once.Do(func() {
		defaultLogger = New()
		defaultLogger.AddWriter("stdout", NewStdWriter(K_LOG_LEVEL_TRACE))
	})
}

func SetPrefix(prefix string) {
	defaultLogger.SetPrefix(prefix)
}

func Prefix() string {
	return defaultLogger.Prefix()
}

func SetStackLevel(level int) {
	defaultLogger.SetStackLevel(level)
}

func GetStackLevel() int {
	return defaultLogger.StackLevel()
}

func EnableStack() {
	defaultLogger.PrintStack()
}

func DisableStack() {
	defaultLogger.DisableStack()
}

func PrintStack() bool {
	return defaultLogger.PrintStack()
}

func EnablePath() {
	defaultLogger.EnablePath()
}

func DisablePath() {
	defaultLogger.DisablePath()
}

func PrintPath() bool {
	return defaultLogger.PrintPath()
}

func AddWriter(name string, w Writer) {
	defaultLogger.AddWriter(name, w)
}

func RemoveWriter(name string) {
	defaultLogger.RemoveWriter(name)
}

//trace
func Tracef(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

func Traceln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintln(args...))
}

func T(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

//print
func Printf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

func Println(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintln(args...))
}

func P(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

//debug
func Debugf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintf(format, args...))
}

func Debugln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintln(args...))
}

func D(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintf(format, args...))
}

//info
func Infof(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintf(format, args...))
}

func Infoln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintln(args...))
}

func I(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintf(format, args...))
}

//error
func Errorf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintf(format, args...))
}

func Errorln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintln(args...))
}

func E(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintf(format, args...))
}

//warn
func Warnf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintf(format, args...))
}

func Warnln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintln(args...))
}

func W(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintf(format, args...))
}

//panic
func Panicf(format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_PANIC, msg)
	panic(msg)
}

func Panicln(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_PANIC, msg)
	panic(msg)
}

//fatal
func Fatalf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_FATAL, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func Fatalln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_FATAL, fmt.Sprintln(args...))
	//os.Exit(-1)
}

func Output(calldepth int, s string) error {
	defaultLogger.WriteMessage(calldepth+1, K_LOG_LEVEL_TRACE, s)
	return nil
}
