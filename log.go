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

//30 black		黑色
//31 red		红色
//32 green		绿色
//33 yellow		黄色
//34 blue		蓝色
//35 magenta    洋红
//36 cyan		天蓝色
//37 white		白色

func black(c string) string {
	return fmt.Sprintf("\033[1;30m%s\033[0m", c)
}

func red(c string) string {
	return fmt.Sprintf("\033[1;31m%s\033[0m", c)
}

func green(c string) string {
	return fmt.Sprintf("\033[1;32m%s\033[0m", c)
}

func yellow(c string) string {
	return fmt.Sprintf("\033[1;33m%s\033[0m", c)
}

func blue(c string) string {
	return fmt.Sprintf("\033[1;34m%s\033[0m", c)
}

func magenta(c string) string {
	return fmt.Sprintf("\033[1;35m%s\033[0m", c)
}

func skyBlue(c string) string {
	return fmt.Sprintf("\033[1;36m%s\033[0m", c)
}

func white(c string) string {
	return fmt.Sprintf("\033[1;37m%s\033[0m", c)
}

var (
	levelShortNames = []string{
		"[T]",
		"[D]",
		"[I]",
		"[W]",
		"[E]",
		"[P]",
		"[F]",
	}

	levelWithColors = []string{
		white(levelShortNames[0]),
		green(levelShortNames[1]),
		blue(levelShortNames[2]),
		yellow(levelShortNames[3]),
		magenta(levelShortNames[4]),
		red(levelShortNames[5]),
		red(levelShortNames[6]),
	}
)

// --------------------------------------------------------------------------------
type Option interface {
	Apply(*Logger)
}

type optionFunc func(*Logger)

func (f optionFunc) Apply(l *Logger) {
	f(l)
}

func WithPrefix(p string) Option {
	return optionFunc(func(l *Logger) {
		l.prefix = p
	})
}

// --------------------------------------------------------------------------------
type Writer interface {
	io.WriteCloser
	Level() int
	EnableColor() bool
}

type Logger struct {
	mu         sync.Mutex
	writers    map[string]Writer
	prefix     string
	printStack bool
	stackLevel int
	printPath  bool
	printColor bool
}

func New(opts ...Option) *Logger {
	var l = &Logger{}
	l.writers = make(map[string]Writer)
	l.stackLevel = K_LOG_LEVEL_PANIC
	l.printStack = false
	l.printPath = true
	l.printColor = true
	for _, opt := range opts {
		opt.Apply(l)
	}
	return l
}

func (this *Logger) SetPrefix(prefix string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.prefix = prefix
}

func (this *Logger) Prefix() string {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.prefix
}

func (this *Logger) SetStackLevel(level int) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.stackLevel = level
}

func (this *Logger) GetStackLevel() int {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.stackLevel
}

func (this *Logger) EnableStack() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printStack = true
}

func (this *Logger) DisableStack() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printStack = false
}

func (this *Logger) PrintStack() bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.printStack
}

func (this *Logger) EnablePath() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printPath = true
}

func (this *Logger) DisablePath() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printPath = false
}

func (this *Logger) PrintPath() bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.printPath
}

func (this *Logger) EnableColor() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printColor = true
}

func (this *Logger) DisableColor() {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.printColor = false
}

func (this *Logger) PrintColor() bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.printColor
}

func (this *Logger) WriteMessage(callDepth, level int, msg string) {
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
	var levelName string

	for _, w := range this.writers {
		if w.Level() <= level {
			if this.printColor && w.EnableColor() {
				levelName = levelWithColors[level]
			} else {
				levelName = levelShortNames[level]
			}
			fmt.Fprintf(w, "%s%s %s %s:%d %s", this.prefix, now.Format("2006/01/02 15:04:05"), levelName, file, line, msg)
		}
	}
}

func (this *Logger) AddWriter(name string, w Writer) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.writers[name] = w
}

func (this *Logger) RemoveWriter(name string) {
	this.mu.Lock()
	defer this.mu.Unlock()
	var w = this.writers[name]
	if w != nil {
		w.Close()
	}
	delete(this.writers, name)
}

// trace
func (this *Logger) Tracef(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

func (this *Logger) Traceln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintln(args...))
}

//print
func (this *Logger) Printf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

func (this *Logger) Println(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintln(args...))
}

//debug
func (this *Logger) Debugf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintf(format, args...))
}

func (this *Logger) Debugln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintln(args...))
}

//info
func (this *Logger) Infof(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintf(format, args...))
}

func (this *Logger) Infoln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintln(args...))
}

//warn
func (this *Logger) Warnf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintf(format, args...))
}

func (this *Logger) Warnln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintln(args...))
}

//error
func (this *Logger) Errorf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *Logger) Errorln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintln(args...))
	os.Exit(-1)
}

//panic
func (this *Logger) Panicf(format string, args ...interface{}) {
	var msg = fmt.Sprintf(format, args...)
	this.WriteMessage(2, K_LOG_LEVEL_PANIC, msg)
	panic(msg)
}

func (this *Logger) Panicln(args ...interface{}) {
	var msg = fmt.Sprintln(args...)
	this.WriteMessage(2, K_LOG_LEVEL_PANIC, msg)
	panic(msg)
}

//fatal
func (this *Logger) Fatalf(format string, args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_FATAL, fmt.Sprintf(format, args...))
	os.Exit(-1)
}

func (this *Logger) Fatalln(args ...interface{}) {
	this.WriteMessage(2, K_LOG_LEVEL_FATAL, fmt.Sprintln(args...))
	os.Exit(-1)
}

func (this *Logger) Output(calldepth int, s string) error {
	this.WriteMessage(calldepth+1, K_LOG_LEVEL_TRACE, s)
	return nil
}

// --------------------------------------------------------------------------------
var defaultLogger *Logger
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
	defaultLogger.stackLevel = level
}

func GetStackLevel() int {
	return defaultLogger.stackLevel
}

func EnableStack() {
	defaultLogger.printStack = true
}

func DisableStack() {
	defaultLogger.printStack = false
}

func PrintStack() bool {
	return defaultLogger.printStack
}

func EnablePath() {
	defaultLogger.printPath = true
}

func DisablePath() {
	defaultLogger.printPath = false
}

func PrintPath() bool {
	return defaultLogger.printPath
}

func EnableColor() {
	defaultLogger.printColor = true
}

func DisableColor() {
	defaultLogger.printColor = false
}

func PrintColor() bool {
	return defaultLogger.printColor
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

//print
func Printf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintf(format, args...))
}

func Println(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_TRACE, fmt.Sprintln(args...))
}

//debug
func Debugf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintf(format, args...))
}

func Debugln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_DEBUG, fmt.Sprintln(args...))
}

//info
func Infof(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintf(format, args...))
}

func Infoln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_INFO, fmt.Sprintln(args...))
}

//error
func Errorf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintf(format, args...))
}

func Errorln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_ERROR, fmt.Sprintln(args...))
}

//warn
func Warnf(format string, args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintf(format, args...))
}

func Warnln(args ...interface{}) {
	defaultLogger.WriteMessage(2, K_LOG_LEVEL_WARNING, fmt.Sprintln(args...))
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
