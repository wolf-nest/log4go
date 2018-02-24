package log4go

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
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
	level int
	out   io.Writer
	mutex sync.Mutex
}

func NewConsoleWriter(level int) *ConsoleWriter {
	var console = &ConsoleWriter{}
	console.level = level
	console.out = os.Stdout
	return console
}

func (this *ConsoleWriter) WriteMessage(msg *LogMessage) {
	if msg == nil {
		return
	}
	if msg.level < this.level {
		return
	}

	var buf bytes.Buffer
	buf.WriteString(msg.header)
	buf.WriteString(" ")
	if isWindows {
		buf.WriteString(msg.levelName)
	} else {
		buf.WriteString(k_CONSOLE_COLORS[msg.level](msg.levelName))
	}
	buf.WriteString(" [")
	buf.WriteString(msg.file)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(msg.line))
	buf.WriteString("] ")
	buf.WriteString(msg.message)

	this.Write(buf.Bytes())
}

func (this *ConsoleWriter) Write(p []byte) (n int, err error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.out.Write(p)
}

func (this *ConsoleWriter) Close() error {
	return nil
}
