package main

import (
	"github.com/smartwalle/log4go"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	log4go.Traceln("https://github.com/smartwalle?tab=repositories")
	log4go.Println("https://github.com/smartwalle?tab=repositories")
	log4go.Debugln("https://github.com/smartwalle?tab=repositories")
	log4go.Infoln("https://github.com/smartwalle?tab=repositories")
	log4go.Warnln("https://github.com/smartwalle?tab=repositories")
	log4go.Errorln("https://github.com/smartwalle?tab=repositories")
	log4go.Panicln("https://github.com/smartwalle?tab=repositories")
	//log4go.Fatalln("https://github.com/smartwalle?tab=repositories")
}
