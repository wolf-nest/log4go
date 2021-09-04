package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/log4go"
	"sync"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	var file = log4go.NewFileWriter2(log4go.LevelTrace)
	log4go.AddWriter("file", file)
	log4go.RemoveWriter("stdout")

	//var ctx = log4go.NewContext(context.TODO())
	//ctx = log4go.ContextWithId(ctx, "")

	//log4go.Traceln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Println(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Debugln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Infoln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Warnln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Errorln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Panicln(ctx, "https://github.com/smartwalle?tab=repositories")
	//log4go.Fatalln(ctx, "https://github.com/smartwalle?tab=repositories")

	var begin = time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < 1000000; i++ {
			log4go.Println(context.Background(), i)
		}
		wg.Done()
	}()
	wg.Wait()
	fmt.Println(time.Now().Sub(begin))
	log4go.Sync()
	fmt.Println(time.Now().Sub(begin))
}
