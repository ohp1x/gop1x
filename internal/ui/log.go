package ui

import (
	"os"

	"github.com/charmbracelet/log"
)

var logger = log.NewWithOptions(os.Stderr, log.Options{})

func Info(msg string, keyvals ...interface{})  { logger.Info(msg, keyvals...) }
func Warn(msg string, keyvals ...interface{})  { logger.Warn(msg, keyvals...) }
func Error(msg string, keyvals ...interface{}) { logger.Error(msg, keyvals...) }
func Ok(msg string, keyvals ...interface{})    { logger.Info("✓ "+msg, keyvals...) }
