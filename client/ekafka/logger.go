package ekafka

import "miopkg/log"

// errorLogger is an log to kafka-go Logger adapter
type logger struct {
	*log.Logger
}

func (l *logger) Printf(tmpl string, args ...interface{}) {
	l.Debugf(tmpl, args...)
}

// errorLogger is an log to kafka-go ErrorLogger adapter
type errorLogger struct {
	*log.Logger
}

func (l *errorLogger) Printf(tmpl string, args ...interface{}) {
	l.Errorf(tmpl, args...)
}

func newKafkaLogger(wrappedLogger *log.Logger) *logger {
	return &logger{wrappedLogger}
}

func newKafkaErrorLogger(wrappedLogger *log.Logger) *errorLogger {
	return &errorLogger{wrappedLogger}
}
