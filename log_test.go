package log4go

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var file = NewFileWriter(K_LOG_LEVEL_TRACE, WithLogDir("./logs"))
	AddWriter("file", file)
	RemoveWriter("stdout")
	DisablePath()
	os.Exit(m.Run())
}

func TestLogger_Write(t *testing.T) {
	Traceln("default logger trace", 1)
	Tracef("default logger trace fmt %d \n", 10)
	Debugln("default logger debug", 1)
	Debugf("default logger debug fmt %d \n", 10)
	Infoln("default logger info", 1)
	Infof("default logger info fmt %d \n", 10)
	Warnln("default logger warn", 1)
	Warnf("default logger warn fmt %d \n", 10)
	Errorln("default logger error", 1)
	Errorf("default logger error fmt %d \n", 10)
}

func BenchmarkPrintln(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Println("1", "2", "3", "4", "5")
	}
}
