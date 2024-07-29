package logz

import (
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"xorm.io/xorm/log"
)

const defaultSlowThreshold = time.Millisecond * 500

type GoZeroContextLogger struct {
	slowThreshold *syncx.AtomicDuration
	showSQL       bool
	level         log.LogLevel
}

var _ log.ContextLogger = (*GoZeroContextLogger)(nil)

func NewGoZeroContextLogger() *GoZeroContextLogger {
	z := new(GoZeroContextLogger)
	z.level = log.LOG_INFO
	z.slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)
	return z
}

func (z *GoZeroContextLogger) Debugf(format string, v ...interface{}) {
	logx.Debugf(format, v...)
}

func (z *GoZeroContextLogger) Errorf(format string, v ...interface{}) {
	logx.Errorf(format, v...)
}

func (z *GoZeroContextLogger) Infof(format string, v ...interface{}) {
	logx.Infof(format, v...)
}

func (z *GoZeroContextLogger) Warnf(format string, v ...interface{}) {
	logx.Slowf(format, v...)
}

func (z *GoZeroContextLogger) Level() log.LogLevel {
	return z.level
}

func (z *GoZeroContextLogger) SetLevel(l log.LogLevel) {
	z.level = l
	switch l {
	case log.LOG_DEBUG:
		logx.SetLevel(logx.DebugLevel)
	case log.LOG_INFO:
		logx.SetLevel(logx.InfoLevel)
	case log.LOG_WARNING:
		logx.SetLevel(logx.ErrorLevel)
	case log.LOG_ERR:
		logx.SetLevel(logx.ErrorLevel)
	case log.LOG_OFF:
		logx.SetLevel(0xff)
	}
}

func (z *GoZeroContextLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		z.showSQL = true
		return
	}
	z.showSQL = show[0]
}

func (z *GoZeroContextLogger) IsShowSQL() bool {
	return z.showSQL
}

func (z *GoZeroContextLogger) BeforeSQL(ctx log.LogContext) {
}

func (z *GoZeroContextLogger) AfterSQL(ctx log.LogContext) {
	var sessionPart string
	v := ctx.Ctx.Value(log.SessionIDKey)
	if key, ok := v.(string); ok {
		sessionPart = fmt.Sprintf(" [%s]", key)
	}
	if ctx.ExecuteTime > z.slowThreshold.Load() {
		logx.WithContext(ctx.Ctx).WithDuration(ctx.ExecuteTime).Slowf("[SQL]%s slowcall %s %v", sessionPart, ctx.SQL, ctx.Args)
	} else if ctx.ExecuteTime > 0 {
		logx.WithContext(ctx.Ctx).WithDuration(ctx.ExecuteTime).Infof("[SQL]%s %s %v", sessionPart, ctx.SQL, ctx.Args)
	} else {
		logx.WithContext(ctx.Ctx).Infof("[SQL]%s %s %v", sessionPart, ctx.SQL, ctx.Args)
	}

	if ctx.Err != nil {
		logx.WithContext(ctx.Ctx).Errorf("[SQL]%s %s %v, error: %s", sessionPart, ctx.SQL, ctx.Args, ctx.Err.Error())
	}
}
