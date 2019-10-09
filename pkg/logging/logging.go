package logging

import (
	"fmt"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

/*
Log logs messages to various levels.

Based on the go-logr package.

The Log type enables various verbosity levels to be logged based on to the --zap-level argument to the operator.

To create an instance of the logger:

    log := logging.Logger("main")

When --zap-level is not provided, the logging defaults to INFO level where log.Warning() and log.Info() are logged,
if --zap-level is set to 1 then the logging goes to the DEBUG level where log.Debug() is logged, while the
--zap-level set to 2 is reserved for more finer DEBUG-level logging provided by the log.Trace() function.
*/
type Log struct {
	logger *logr.Logger //logger instance
}

// Error logs the message using go-logr package on a default level as ERROR
func (l *Log) Error(err error, msg string, keysAndValues ...interface{}) {
	(*l.logger).Error(err, msg, keysAndValues...)
}

// Warning logs the message using go-logr package on a default level as WARNING
func (l *Log) Warning(msg string, keysAndValues ...interface{}) {
	if (*l.logger).V(0).Enabled() {
		(*l.logger).V(0).Info(fmt.Sprintf("WARNING: %s", msg), keysAndValues...)
	}
}

// Info logs the message using go-logr package on a default level as INFO
func (l *Log) Info(msg string, keysAndValues ...interface{}) {
	if (*l.logger).V(0).Enabled() {
		(*l.logger).V(0).Info(msg, keysAndValues...)
	}
}

// Debug logs the message using go-logr package on a V=1 level as DEBUG
func (l *Log) Debug(msg string, keysAndValues ...interface{}) {
	if (*l.logger).V(1).Enabled() {
		(*l.logger).V(1).Info(fmt.Sprintf("DEBUG: %s", msg), keysAndValues...)
	}
}

// Trace logs the message using go-logr package on a V=1 level as TRACE
func (l *Log) Trace(msg string, keysAndValues ...interface{}) {
	if (*l.logger).V(2).Enabled() {
		// The V(1) level here is intentional: TRACE = finer DEBUG logging but is only enabled by setting V=2
		(*l.logger).V(1).Info(fmt.Sprintf("TRACE: %s", msg), keysAndValues...)
	}
}

// Logger returns an instance of a logger
func Logger(name string, keysAndValues ...interface{}) *Log {
	logger := logf.Log.WithName(name).WithValues(keysAndValues...)
	l := &Log{
		logger: &logger,
	}
	return l
}

// SetLogger sets a concrete logging implementation for all deferred Loggers.
func SetLogger(logger logr.Logger) {
	logf.SetLogger(logger)
}

// WithValues adds some key-value pairs of context to a logger.
// See Info for documentation on how key/value pairs work.
func (l *Log) WithValues(keysAndValues ...interface{}) *Log {
	lgr := ((*l.logger).WithValues(keysAndValues...))
	log := &Log{
		logger: &lgr,
	}
	return log
}

// WithName adds a new element to the logger's name.
// Successive calls with WithName continue to append
// suffixes to the logger's name.  It's strongly reccomended
// that name segments contain only letters, digits, and hyphens
// (see the package documentation for more information).
func (l *Log) WithName(name string) *Log {
	lgr := ((*l.logger).WithName(name))
	log := &Log{
		logger: &lgr,
	}
	return log
}
