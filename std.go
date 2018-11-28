package log4go

import (
	"github.com/mattn/go-isatty"
	"io"
	"os"
	"sync"
)

type StdWriter struct {
	level       int
	out         io.Writer
	mutex       sync.Mutex
	enableColor bool
}

func NewStdWriter(level int) *StdWriter {
	var sw = &StdWriter{}
	sw.level = level
	sw.out = os.Stdout
	sw.enableColor = true
	if w, ok := sw.out.(*os.File); !ok || (os.Getenv("TERM") == "dumb" || (!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd()))) {
		sw.enableColor = false
	}
	return sw
}

func (this *StdWriter) Level() int {
	return this.level
}

func (this *StdWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	return this.out.Write(p)
}

func (this *StdWriter) Close() error {
	return nil
}

func (this *StdWriter) EnableColor() bool {
	return this.enableColor
}
