package ui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{})

func Info(msg string, keyvals ...interface{})  { logger.Info(msg, keyvals...) }
func Warn(msg string, keyvals ...interface{})  { logger.Warn(msg, keyvals...) }
func Error(msg string, keyvals ...interface{}) { logger.Error(msg, keyvals...) }
func Ok(msg string, keyvals ...interface{})    { logger.Info("✓ "+msg, keyvals...) }

func Errorf(format string, args ...interface{}) {
	logger.Error(fmt.Sprintf(format, args...))
}

func Print(msg string) {
	fmt.Fprintln(os.Stdout, msg)
}

func Printf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format, args...)
}
