package main

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	"time"
)

var ZapLogger *zap.Logger

func main() {
	hook := lumberjack.Logger{
		Filename:   "./logs/xx.log", // 日志文件路径
		MaxSize:    10,              // 每个日志文件保存的大小 单位:M
		MaxAge:     7,               // 文件最多保存多少天
		MaxBackups: 30,              // 日志文件最多保存多少个备份
		Compress:   false,           // 是否压缩
	}
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "file",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)
	var writes = []zapcore.WriteSyncer{zapcore.AddSync(&hook)}
	// 如果是开发环境，同时在控制台上也输出
	//if debug {
	//	writes = append(writes, zapcore.AddSync(os.Stdout))
	//}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writes...),
		atomicLevel,
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()
	// 开启文件及行号
	development := zap.Development()

	// 设置初始化字段
	field := zap.Fields(zap.String("appName", "xx"))

	// 构造日志
	ZapLogger = zap.New(core, caller, development, field)

	var begin = time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i := 0; i < 1000000; i++ {
			ZapLogger.Info(fmt.Sprintln(i))
		}
		wg.Done()
	}()
	wg.Wait()
	fmt.Println(time.Now().Sub(begin))
	ZapLogger.Sync()
	fmt.Println(time.Now().Sub(begin))

}
