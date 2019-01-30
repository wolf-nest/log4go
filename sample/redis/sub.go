package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/log4go"
)

func main() {
	var rw = log4go.NewRedisHub("test_log", "192.168.1.99:6379", 10, 2, redis.DialDatabase(15))
	rw.Redirect(log4go.NewFileWriter(log4go.K_LOG_LEVEL_TRACE))

	fmt.Println("running")
	select {}
}
