package logging

import (
	"fmt"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// LogError logs the message using go-logr package on a default level as ERROR
func LogError(err error, logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	(*logger).Error(err, msg, keysAndValues...)
}

// LogWarning logs the message using go-logr package on a default level as WARNING
func LogWarning(logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	if (*logger).V(0).Enabled() {
		(*logger).V(0).Info(fmt.Sprintf("WARNING: %s", msg), keysAndValues...)
	}
}

// LogInfo logs the message using go-logr package on a default level as INFO
func LogInfo(logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	if (*logger).V(0).Enabled() {
		(*logger).V(0).Info(msg, keysAndValues...)
	}
}

// LogDebug logs the message using go-logr package on a V=1 level as DEBUG
func LogDebug(logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	if (*logger).V(1).Enabled() {
		(*logger).V(1).Info(fmt.Sprintf("DEBUG: %s", msg), keysAndValues...)
	}
}

// LogTrace logs the message using go-logr package on a V=1 level as TRACE
func LogTrace(logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	if (*logger).V(2).Enabled() {
		(*logger).V(1).Info(fmt.Sprintf("TRACE: %s", msg), keysAndValues...)
	}
}

// Logger returns an instance of a logger
func Logger(name string, keysAndValues ...interface{}) logr.Logger {
	return logf.Log.WithName(name).WithValues(keysAndValues...)
}

// SetLogger sets a concrete logging implementation for all deferred Loggers.
func SetLogger(logger logr.Logger) {
	logf.SetLogger(logger)
}
