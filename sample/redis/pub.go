package main

import (
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/log4go"
	"time"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
		}
	}()

	log4go.SetPrefix("[test-2]")
	log4go.AddWriter("redis", log4go.NewRedisWriter(log4go.K_LOG_LEVEL_TRACE, "test_log", "192.168.1.99:6379", 10, 2, redis.DialDatabase(15)))

	for {
		time.Sleep(time.Second * 1)
		log4go.Traceln("https://github.com/smartwalle?tab=repositories")
	}
}
