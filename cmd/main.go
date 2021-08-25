package main

import (
	"context"
	"github.com/smartwalle/log4go"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	var ctx = log4go.NewContext(context.TODO())

	ctx = log4go.ContextWithLogId(ctx, "")

	log4go.Traceln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Println(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Debugln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Infoln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Warnln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Errorln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Panicln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Fatalln(ctx, "https://github.com/smartwalle?tab=repositories")
}
