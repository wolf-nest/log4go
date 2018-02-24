package log4go

import (
	"testing"
)

func TestLogger_Write(t *testing.T) {
	var log = NewLogger()
	log.AddWriter("test", NewConsoleWriter(K_LOG_LEVEL_DEBUG))

	var file = NewFileWriter(K_LOG_LEVEL_DEBUG, "./logs")
	log.AddWriter("file", file)

	for i := 0; i < 100000; i++ {
		log.Debugln("new logger debug", 1)
		log.Debugf("new logger debug fmt %d \n", 10)
		log.Infoln("new logger info", 1)
		log.Infof("new logger info fmt %d \n", 10)
		log.Warnln("new logger warn", 1)
		log.Warnf("new logger warn fmt %d \n", 1)
		log.Fatalln("new logger fatal", 1)
		log.Fatalf("new logger fatal fmt %d \n", 1)
		log.Panicln("new logger panic", 1)
		log.Panicf("new logger panic fmt %d \n", 1)
	}

	//Debugln("default logger debug", 1)
	//Debugf("default logger debug fmt %d", 10)
	//Infoln("default logger info", 1)
	//Infof("default logger info fmt %d", 10)
	//Warnln("default logger warn", 1)
	//Warnf("default logger warn fmt %d", 1)
	//Fatalln("default logger fatal", 1)
	//Fatalf("default logger fatal fmt %d", 1)
	//Panicln("default logger panic", 1)
	//Panicf("default logger panic fmt %d \n", 1)
}
