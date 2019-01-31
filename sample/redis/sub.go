package main

import (
	"fmt"
	"github.com/smartwalle/log4go"
)

func main() {
	var rw = log4go.NewRedisHub("youle_log", "192.168.1.99:6379", 10, 2)
	rw.Redirect(log4go.NewFileWriter(log4go.K_LOG_LEVEL_TRACE, log4go.WithLogDir("./test_log")))

	fmt.Println("running")
	select {}
}
