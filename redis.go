package log4go

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/dbr"
	"io"
	"time"
)

// --------------------------------------------------------------------------------
const (
	kRedisLogField = "log"
)

type RedisWriter struct {
	level int
	pool  dbr.Pool
	key   string
}

func NewRedisWriter(level int, key string, addr string, maxActive, maxIdle int, opts ...redis.DialOption) *RedisWriter {
	var pool = dbr.NewRedis(addr, maxActive, maxIdle, opts...)
	var rw = &RedisWriter{}
	rw.level = level
	rw.pool = pool
	rw.key = key
	return rw
}

func (this *RedisWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 || this.pool == nil {
		return 0, nil
	}

	var rSess = this.pool.GetSession()
	defer rSess.Close()

	rSess.XADD(this.key, 0, "*", kRedisLogField, p)

	return len(p), err
}

func (this *RedisWriter) Close() error {
	if this.pool == nil {
		return nil
	}
	return this.pool.Close()
}

func (this *RedisWriter) Level() int {
	return this.level
}

func (this *RedisWriter) WriteMessage(logTime time.Time, service, instance, prefix, timeStr string, level int, levelName, file string, line int, msg string) {
	fmt.Fprintf(this, "%s%s%s%s %s %s:%d %s", service, instance, prefix, timeStr, levelName, file, line, msg)
}

// --------------------------------------------------------------------------------
type RedisHub struct {
	pool dbr.Pool
	key  string
}

func NewRedisHub(key string, addr string, maxActive, maxIdle int, opts ...redis.DialOption) *RedisHub {
	var pool = dbr.NewRedis(addr, maxActive, maxIdle, opts...)
	var rh = &RedisHub{}
	rh.pool = pool
	rh.key = key
	return rh
}

func (this *RedisHub) Redirect(w io.Writer) {
	go this.redirect(this.key, w)
}

func (this *RedisHub) redirect(key string, w io.Writer) {
	var rSess = this.pool.GetSession()
	defer rSess.Close()

	var queue = key
	var group = fmt.Sprintf("%s_group", key)
	var consumer = fmt.Sprintf("%s_grop_consumer", key)

	rSess.XGROUPCREATE(queue, group, "$", "MKSTREAM")

	for {
		var sList, err = rSess.XREADGROUP(group, consumer, 0, 0, queue, ">").Streams()
		if err != nil {
			return
		}

		for _, s := range sList {
			var log = s.Fields[kRedisLogField]
			w.Write([]byte(log))
			rSess.XDEL(queue, s.Id)
		}
	}
}
