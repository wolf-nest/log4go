package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/log4go"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	var file = log4go.NewFileWriter2(log4go.LevelTrace)
	log4go.AddWriter("file", file)

	var ctx = log4go.NewContext(context.TODO())

	ctx = log4go.ContextWithId(ctx, "")

	log4go.Traceln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Println(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Debugln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Infoln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Warnln(ctx, "https://github.com/smartwalle?tab=repositories")
	log4go.Errorln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Panicln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Fatalln(ctx, "https://github.com/smartwalle?tab=repositories")

	fmt.Println(time.Now())
	file.Close()
	fmt.Println(time.Now())
}
