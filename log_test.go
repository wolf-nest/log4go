package log4go

import (
	"testing"
)

func TestMain(m *testing.M) {
	//var file = NewFileWriter(K_LOG_LEVEL_DEBUG, WithLogDir("./logs"))
	//AddWriter("file", file)
	//os.Exit(m.Run())
}

func TestLogger_Write(t *testing.T) {
	Debugln("default logger debug", 1)
	Debugf("default logger debug fmt %d \n", 10)
	Infoln("default logger info", 1)
	Infof("default logger info fmt %d \n", 10)
	Warnln("default logger warn", 1)
	Warnf("default logger warn fmt %d \n", 1)
	Panicln("default logger panic", 1)
	Panicf("default logger panic fmt %d \n", 1)
	Fatalln("default logger fatal", 1)
	Fatalf("default logger fatal fmt %d \n", 1)
}

func BenchmarkPrintln(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Println("1", "2", "3", "4", "5")
	}
}
