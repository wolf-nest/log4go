package log4go_test

import (
	"context"
	"github.com/smartwalle/log4go"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	//var file = log4go.NewFileWriter2(log4go.LevelTrace)
	//log4go.AddWriter("file", file)
	//log4go.RemoveWriter("stdout")
	//log4go.SharedLogger()
	os.Exit(m.Run())
}

func TestLogger_Write(t *testing.T) {
	log4go.Traceln(nil, "default logger trace", 1)
	log4go.Tracef(nil, "default logger trace fmt %d \n", 10)
	log4go.Debugln(nil, "default logger debug", 1)
	log4go.Debugf(nil, "default logger debug fmt %d \n", 10)
	log4go.Infoln(nil, "default logger info", 1)
	log4go.Infof(nil, "default logger info fmt %d \n", 10)
	log4go.Warnln(nil, "default logger warn", 1)
	log4go.Warnf(nil, "default logger warn fmt %d \n", 10)
	log4go.Errorln(nil, "default logger error", 1)
	log4go.Errorf(nil, "default logger error fmt %d \n", 10)
}

func BenchmarkPrintln(b *testing.B) {
	var ctx = log4go.NewContext(context.Background())
	for i := 0; i < b.N; i++ {
		log4go.Println(ctx, i)
	}
}
