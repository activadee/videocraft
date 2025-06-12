package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

type logger struct {
	log *logrus.Logger
}

func New(level string) Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	
	// Set log level
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	
	// Set formatter
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
	
	return &logger{log: log}
}

func (l *logger) Debug(args ...interface{}) {
	l.log.Debug(args...)
}

func (l *logger) Info(args ...interface{}) {
	l.log.Info(args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.log.Warn(args...)
}

func (l *logger) Error(args ...interface{}) {
	l.log.Error(args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.log.Fatal(args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.log.Debugf(format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.log.Infof(format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.log.Warnf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.log.Errorf(format, args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.log.Fatalf(format, args...)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{log: l.log.WithField(key, value).Logger}
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	return &logger{log: l.log.WithFields(fields).Logger}
}