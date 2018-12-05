package main

import "github.com/smartwalle/log4go"

func main() {
	log4go.Debugln("https://github.com/smartwalle?tab=repositories")
	log4go.Println("https://github.com/smartwalle?tab=repositories")
	log4go.Infoln("https://github.com/smartwalle?tab=repositories")
	log4go.Warnln("https://github.com/smartwalle?tab=repositories")
}
