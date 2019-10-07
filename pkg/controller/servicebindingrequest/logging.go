package servicebindingrequest

import (
	"fmt"

	"github.com/go-logr/logr"
)

// LogError logs the message using go-logr package on a default level as WARNING
func LogError(err error, logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	(*logger).Error(err, msg, keysAndValues...)
}

// LogWarning logs the message using go-logr package on a default level as WARNING
func LogWarning(logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	(*logger).Info(fmt.Sprintf("WARNING: %s", msg), keysAndValues...)
}

// LogInfo logs the message using go-logr package on a default level as INFO
func LogInfo(logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	(*logger).Info(msg, keysAndValues...)
}

// LogDebug logs the message using go-logr package on a V=1 level as DEBUG
func LogDebug(logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	(*logger).V(1).Info(msg, keysAndValues...)
}

// LogTrace logs the message using go-logr package on a V=2 level as TRACE
func LogTrace(logger *logr.Logger, msg string, keysAndValues ...interface{}) {
	(*logger).V(2).Info(fmt.Sprintf("TRACE: %s", msg), keysAndValues...)
}
