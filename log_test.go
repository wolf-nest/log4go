package log4go_test

import (
	"github.com/smartwalle/log4go"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var file = log4go.NewFileWriter(log4go.LevelTrace, log4go.WithLogDir("./logs"))
	log4go.AddWriter("file", file)
	log4go.RemoveWriter("stdout")
	os.Exit(m.Run())
}

func TestLogger_Write(t *testing.T) {
	log4go.Traceln("default logger trace", 1)
	log4go.Tracef("default logger trace fmt %d \n", 10)
	log4go.Debugln("default logger debug", 1)
	log4go.Debugf("default logger debug fmt %d \n", 10)
	log4go.Infoln("default logger info", 1)
	log4go.Infof("default logger info fmt %d \n", 10)
	log4go.Warnln("default logger warn", 1)
	log4go.Warnf("default logger warn fmt %d \n", 10)
	log4go.Errorln("default logger error", 1)
	log4go.Errorf("default logger error fmt %d \n", 10)
}

func BenchmarkPrintln(b *testing.B) {
	for i := 0; i < b.N; i++ {
		log4go.Println("1", "2", "3", "4", "5")
	}
}
