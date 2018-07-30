package log4go

import (
	"io"
	"os"
	"sync"
)

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

func (this *ConsoleWriter) Level() int {
	return this.level
}

func (this *ConsoleWriter) WriteMessage(msg *LogMessage) {
	if msg == nil {
		return
	}
	this.Write(msg.Bytes(true))
}

func (this *ConsoleWriter) Write(p []byte) (n int, err error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.out.Write(p)
}

func (this *ConsoleWriter) Close() error {
	return nil
}
