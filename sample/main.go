package main

import "github.com/smartwalle/log4go"

func main() {
	log4go.Debugln("debugln")
	log4go.Println("println")
	log4go.Infoln("infoln")
	log4go.Warnln("warnln")
	log4go.Panicln("panicln")
	log4go.Fatalln("fatalln")
}
