package zapsentry

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	Development = "development"
	Production  = "production"
	Nop         = "nop"
)

var logger *zap.SugaredLogger

type Option func(o *option)

type option struct {
	state        string
	sentryDns    string
	sentryTags   map[string]string
	sentrtFields []zapcore.Field
}

func WithSentry(sentryDNS string, tags map[string]string, fields []zapcore.Field) Option {
	return func(o *option) {
		o.sentryDns = sentryDNS
		o.sentryTags = tags
		o.sentrtFields = fields
	}
}

func WithState(state string) Option {
	return func(o *option) {
		o.state = state
	}
}

func newzap(stage string) *zap.Logger {
	var l *zap.Logger
	switch stage {
	case Production:
		l, _ = zap.NewProduction()
	case Nop:
		l = zap.NewNop()
	default:
		l, _ = zap.NewDevelopment()
	}

	return l
}

func addSentry(l *zap.Logger, o *option) {
	if o.sentryDns == "" || o.sentryDns == "test" {
		return
	}

	cfg := Configuration{
		DSN:  o.sentryDns,
		Tags: o.sentryTags,
	}

	sentryCore, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	if o.sentrtFields != nil && len(o.sentrtFields) > 0 {
		sentryCore = sentryCore.With(o.sentrtFields)
	}
	l = l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, sentryCore)
	}))
}

func NewDefault(opts ...Option) {
	logger = New(opts...)
}

func New(opts ...Option) *zap.SugaredLogger {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}

	l := newzap(o.state)
	addSentry(l, o)

	return l.Sugar()
}

func With(args ...interface{}) *zap.SugaredLogger {
	return logger.With(args...)
}

// Debug uses fmt.Sprint to construct and log a message.
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Info uses fmt.Sprint to construct and log a message.
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error uses fmt.Sprint to construct and log a message.
func Error(args ...interface{}) {
	logger.Error(args...)
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanic(args ...interface{}) {
	logger.DPanic(args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	logger.Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
func DPanicf(template string, args ...interface{}) {
	logger.DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func DPanicw(msg string, keysAndValues ...interface{}) {
	logger.DPanicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...interface{}) {
	logger.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}

// Sync flushes any buffered log entries.
func Sync() error {
	return logger.Sync()
}
