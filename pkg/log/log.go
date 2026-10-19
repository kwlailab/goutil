package log

import (
	"context"
	"sync"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

var (
	_globalMu sync.RWMutex
	_globalL  *otelzap.Logger
	_globalS  *otelzap.SugaredLogger
)

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	setGlobals(otelzap.New(logger))
}

func setGlobals(logger *otelzap.Logger) {
	_globalL = logger
	_globalS = logger.WithOptions(zap.AddCallerSkip(1)).Sugar()
}

func ReplaceGlobals(logger *otelzap.Logger) func() {
	_globalMu.Lock()
	prev := _globalL
	setGlobals(logger)
	_globalMu.Unlock()
	return func() { ReplaceGlobals(prev) }
}

func L() *otelzap.Logger {
	_globalMu.RLock()
	l := _globalL
	_globalMu.RUnlock()
	return l
}

func S() *otelzap.SugaredLogger {
	_globalMu.RLock()
	s := _globalS
	_globalMu.RUnlock()
	return s
}

func Ctx(ctx context.Context) otelzap.LoggerWithCtx {
	return L().Ctx(ctx)
}

func Debug(msg string, kvp ...interface{}) {
	_globalS.Debugw(msg, kvp...)
}

func Debugf(template string, args ...interface{}) {
	_globalS.Debugf(template, args...)
}

func DebugCtx(ctx context.Context, msg string, kvp ...interface{}) {
	_globalS.Ctx(ctx).Debugw(msg, kvp...)
}

func Info(msg string, kvp ...interface{}) {
	_globalS.Infow(msg, kvp...)
}

func Infof(template string, args ...interface{}) {
	_globalS.Infof(template, args...)
}

func InfoCtx(ctx context.Context, msg string, kvp ...interface{}) {
	_globalS.InfowContext(ctx, msg, kvp...)
}

func Warn(msg string, kvp ...interface{}) {
	_globalS.Warnw(msg, kvp...)
}

func Warnf(template string, args ...interface{}) {
	_globalS.Warnf(template, args...)
}

func WarnCtx(ctx context.Context, msg string, kvp ...interface{}) {
	_globalS.WarnwContext(ctx, msg, kvp...)
}

func Error(msg string, kvp ...interface{}) {
	_globalS.Errorw(msg, kvp...)
}

func Errorf(template string, args ...interface{}) {
	_globalS.Errorf(template, args...)
}

func ErrorCtx(ctx context.Context, msg string, kvp ...interface{}) {
	_globalS.Ctx(ctx).Errorw(msg, kvp...)
}

func Fatal(msg string, kvp ...interface{}) {
	_globalS.Fatalw(msg, kvp...)
}

func Fatalf(template string, args ...interface{}) {
	_globalS.Fatalf(template, args...)
}

func FatalCtx(ctx context.Context, msg string, kvp ...interface{}) {
	_globalS.Ctx(ctx).Fatalw(msg, kvp...)
}

func Panic(msg string, kvp ...interface{}) {
	_globalS.Panicw(msg, kvp...)
}

func Panicf(template string, args ...interface{}) {
	_globalS.Panicf(template, args...)
}

func PanicCtx(ctx context.Context, msg string, kvp ...interface{}) {
	_globalS.Ctx(ctx).Panicw(msg, kvp...)
}
