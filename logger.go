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
	stage        string
	sentryDns    string
	sentryTags   map[string]string
	sentrtFields []zapcore.Field
}

func init() {
	l, _ := zap.NewDevelopment()
	logger = l.Sugar()
}

func WithSentry(sentryDNS string, tags map[string]string, fields []zapcore.Field) Option {
	return func(o *option) {
		o.sentryDns = sentryDNS
		o.sentryTags = tags
		o.sentrtFields = fields
	}
}

func WithStage(stage string) Option {
	return func(o *option) {
		o.stage = stage
	}
}

func S() *zap.SugaredLogger {
	return logger
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

func createLoagger(o *option) *zap.Logger {
	l := newzap(o.stage)
	if o.sentryDns == "" || o.sentryDns == "test" {
		return l
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

	return l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
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

	return createLoagger(o).Sugar()
}
