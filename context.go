package log4go

import (
	"context"
	"github.com/google/uuid"
	"strings"
)

type logIdKey struct{}

func NewContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.TODO()
	}
	var id, ok = ctx.Value(logIdKey{}).(string)
	if ok == false || id == "" {
		id = uuid.NewString()
		ctx = context.WithValue(ctx, logIdKey{}, id)
	}
	return ctx
}

func ContextWithLogId(ctx context.Context, logId string) context.Context {
	if ctx == nil {
		ctx = context.TODO()
	}
	logId = strings.TrimSpace(logId)
	if logId == "" {
		logId = uuid.NewString()
	}
	return context.WithValue(ctx, logIdKey{}, logId)
}

func MustGetLogId(ctx context.Context) string {
	var logId = GetLogId(ctx)
	if logId == "" {
		logId = uuid.NewString()
	}
	return logId
}

func GetLogId(ctx context.Context) string {
	if ctx == nil {
		ctx = context.TODO()
	}
	var id, _ = ctx.Value(logIdKey{}).(string)
	return id
}
