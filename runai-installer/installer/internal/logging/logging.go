/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logging

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	initOnce sync.Once
	initErr  error
	logFile  *os.File
)

// getGoroutineID extracts the goroutine ID from runtime.Stack output.
// Format: "goroutine 123 [running]:\n..."
func getGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	// Skip "goroutine " prefix
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	// Find the space after the ID
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		return 0
	}
	id, _ := strconv.ParseUint(string(b[:i]), 10, 64)
	return id
}

// customFormatter adds goroutine ID and caller info to log entries
type customFormatter struct {
	logrus.TextFormatter
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Get goroutine ID
	gid := getGoroutineID()

	// Get caller info - walk up the stack to find the actual caller
	loc := "unknown"
	for skip := 1; skip < 15; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		// Skip logrus and our logging package
		if strings.Contains(file, "logrus") || strings.Contains(file, "logging/logging.go") {
			continue
		}
		// Get just the filename, not full path
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		loc = fmt.Sprintf("%s:%d", short, line)
		break
	}

	// Format: time level gid loc msg
	timestamp := entry.Time.Format(f.TimestampFormat)
	return []byte(fmt.Sprintf("time=%q level=%s gid=%-4d loc=%-25s msg=%q\n",
		timestamp, entry.Level.String(), gid, loc, entry.Message)), nil
}

// InitFromDir configures the global logger to write to runai-installer.log inside dir.
// If dir is empty, the current working directory is used.
func InitFromDir(dir string) error {
	if dir == "" {
		dir = "."
	}
	return Init(filepath.Join(dir, "runai-installer.log"))
}

// Init configures a global logrus logger.
// If filePath is empty, logs are written to stdout.
// If filePath is provided, the directory is created (if needed) and logs are written to that file.
func Init(filePath string) error {
	initOnce.Do(func() {
		var out io.Writer

		if filePath == "" {
			out = os.Stdout
		} else {
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				initErr = fmt.Errorf("failed to create logs dir %q: %w", dir, err)
				return
			}
			f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				initErr = fmt.Errorf("failed to open log file %q: %w", filePath, err)
				return
			}
			logFile = f
			out = f
		}

		logrus.SetOutput(out)
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetFormatter(&customFormatter{
			TextFormatter: logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05.000",
			},
		})

		target := "stdout"
		if filePath != "" {
			target = filePath
		}
		logrus.Infof("logging initialized at %s; target: %s", time.Now().Format(time.RFC3339), target)
	})
	return initErr
}

// Close closes the underlying log file if opened.
func Close() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// Convenience wrappers to keep call sites simple and decoupled from logrus.
func Infof(format string, args ...any)  { logrus.Infof(format, args...) }
func Errorf(format string, args ...any) { logrus.Errorf(format, args...) }
func Warnf(format string, args ...any)  { logrus.Warnf(format, args...) }
func Debugf(format string, args ...any) { logrus.Debugf(format, args...) }
func WithField(key string, value any) *logrus.Entry {
	return logrus.WithField(key, value)
}

// Writers to capture command outputs and route through the logger.
func InfoWriter() io.Writer  { return logrus.StandardLogger().WriterLevel(logrus.InfoLevel) }
func ErrorWriter() io.Writer { return logrus.StandardLogger().WriterLevel(logrus.ErrorLevel) }
func DebugWriter() io.Writer { return logrus.StandardLogger().WriterLevel(logrus.DebugLevel) }
