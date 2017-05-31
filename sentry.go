package sentry

import (
	raven "github.com/getsentry/raven-go"
	"go.uber.org/zap/zapcore"
)

const (
	_platform          = "go"
	_traceContextLines = 3
	_traceSkipFrames   = 2
)

func ravenSeverity(lvl zapcore.Level) raven.Severity {
	switch lvl {
	case zapcore.DebugLevel:
		return raven.INFO
	case zapcore.InfoLevel:
		return raven.INFO
	case zapcore.WarnLevel:
		return raven.WARNING
	case zapcore.ErrorLevel:
		return raven.ERROR
	case zapcore.DPanicLevel:
		return raven.FATAL
	case zapcore.PanicLevel:
		return raven.FATAL
	case zapcore.FatalLevel:
		return raven.FATAL
	default:
		// Unrecognized levels are fatal.
		return raven.FATAL
	}
}

type client interface {
	Capture(*raven.Packet, map[string]string) (string, chan error)
	Wait()
}

// Configuration is a minimal set of parameters for Sentry integration.
type Configuration struct {
	DSN   string `yaml:"DSN"`
	Tags  map[string]string
	Trace trace
}

type trace struct {
	Disabled bool
}

// Build uses the provided configuration to construct a Sentry-backed logging
// core.
func (c Configuration) Build() (zapcore.Core, error) {
	client, err := raven.New(c.DSN)
	if err != nil {
		return zapcore.NewNopCore(), err
	}
	return newCore(c, client, zapcore.ErrorLevel), nil
}

type core struct {
	client
	zapcore.LevelEnabler
	trace

	fields map[string]interface{}
	tags   map[string]string
}

func newCore(cfg Configuration, c client, enab zapcore.LevelEnabler) *core {
	sentryCore := &core{
		client:       c,
		LevelEnabler: enab,
		trace:        cfg.Trace,
		fields:       make(map[string]interface{}),
		tags:         cfg.Tags,
	}
	return sentryCore
}

func (c *core) With(fs []zapcore.Field) zapcore.Core {
	return c.with(fs)
}

func (c *core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *core) Write(ent zapcore.Entry, fs []zapcore.Field) error {
	clone := c.with(fs)

	packet := &raven.Packet{
		Message:   ent.Message,
		Timestamp: raven.Timestamp(ent.Time),
		Level:     ravenSeverity(ent.Level),
		Platform:  _platform,
		Extra:     clone.fields,
	}

	if !c.trace.Disabled {
		trace := raven.NewStacktrace(_traceSkipFrames, _traceContextLines, nil /* app prefixes */)
		if trace != nil {
			packet.Interfaces = append(packet.Interfaces, trace)
		}
	}

	_, _ = c.Capture(packet, c.tags)

	// We may be crashing the program, so should flush any buffered events.
	if ent.Level > zapcore.ErrorLevel {
		c.Wait()
	}
	return nil
}

func (c *core) Sync() error {
	c.client.Wait()
	return nil
}

func (c *core) with(fs []zapcore.Field) *core {
	// Copy our map.
	m := make(map[string]interface{}, len(c.fields))
	for k, v := range c.fields {
		m[k] = v
	}

	// Add fields to an in-memory encoder.
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fs {
		f.AddTo(enc)
	}

	// Merge the two maps.
	for k, v := range enc.Fields {
		m[k] = v
	}

	return &core{
		client:       c.client,
		LevelEnabler: c.LevelEnabler,
		trace:        c.trace,
		fields:       m,
	}
}
