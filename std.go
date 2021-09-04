package log4go

import (
	"fmt"
	"github.com/mattn/go-isatty"
	"os"
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
	levelColors = []string{
		white(LevelNames[0]),
		green(LevelNames[1]),
		blue(LevelNames[2]),
		yellow(LevelNames[3]),
		magenta(LevelNames[4]),
		red(LevelNames[5]),
		red(LevelNames[6]),
	}
)

type StdWriter struct {
	level       Level
	out         *os.File
	enableColor bool
}

func NewStdWriter(level Level) *StdWriter {
	var w = &StdWriter{}
	w.level = level
	w.out = os.Stdout
	w.enableColor = true
	if os.Getenv("TERM") == "dumb" || (!isatty.IsTerminal(w.out.Fd()) && !isatty.IsCygwinTerminal(w.out.Fd())) {
		w.enableColor = false
	}
	return w
}

func (this *StdWriter) Write(p []byte) (n int, err error) {
	return this.out.Write(p)
}

func (this *StdWriter) Close() error {
	return nil
}

func (this *StdWriter) Level() Level {
	return this.level
}

func (this *StdWriter) Sync() error {
	return this.out.Sync()
}

func (this *StdWriter) WriteMessage(logId, service, instance, prefix, logTime string, level Level, file, line, msg string) {
	var levelName = LevelNames[level]
	if this.enableColor {
		levelName = levelColors[level]
	}
	fmt.Fprintf(this, "[%s] %s%s%s%s %s %s:%s %s", logId, service, instance, prefix, logTime, levelName, file, line, msg)
}
