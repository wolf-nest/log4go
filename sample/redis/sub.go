package main

import (
	"fmt"
	"github.com/smartwalle/log4go"
)

func main() {
	var rw = log4go.NewRedisHub("test_log", "192.168.1.99:6379", 10, 2)
	rw.Redirect(log4go.NewFileWriter(log4go.LevelTrace, log4go.WithLogDir("./logs")))

	fmt.Println("running")
	select {}
}
