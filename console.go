package log4go

import (
	"log"
	"os"
	"runtime"
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

type ConsoleWriter struct {
	logger *log.Logger
	level  int
}

func NewConsoleWriter(level int) *ConsoleWriter {
	var console = &ConsoleWriter{}
	console.level = level
	console.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return console
}

func (this *ConsoleWriter) Write(level int, file string, line int, prefix string, msg string) {
	if level < this.level {
		return
	}
	if isWindows {
		this.logger.Printf("%s [%s:%d] %s", prefix, file, line, msg)
		return
	}
	this.logger.Printf("%s [%s:%d] %s", k_CONSOLE_COLORS[level](prefix), file, line, msg)
}

func (this *ConsoleWriter) Close() {
}
