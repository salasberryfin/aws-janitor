package action

import (
	"fmt"
	"os"
)

const (
	failedExitCode = 1
)

// Log will write a log entry to stdout.
func Log(msg string, a ...interface{}) {
	message := fmt.Sprintf(msg, a...)
	fmt.Println(message) //nolint: forbidigo
}

// LogDebug will write a debug message command to stdout.
func LogDebug(msg string, a ...interface{}) {
	message := fmt.Sprintf(msg, a...)
	Log("::debug::%s", message)
}

// LogWarning will write a warning message command to stdout.
func LogWarning(msg string, a ...interface{}) {
	message := fmt.Sprintf(msg, a...)
	Log("::warning::%s", message)
}

// LogError will write a error message command to stdout.
func LogError(msg string, a ...interface{}) {
	message := fmt.Sprintf(msg, a...)
	Log("::error::%s", message)
}

// LogErrorAndExit will write a error message command to stdout and exit
// with a non-zero exit code.
func LogErrorAndExit(msg string, a ...interface{}) {
	LogError(msg, a...)
	os.Exit(failedExitCode)
}
