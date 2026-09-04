/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package linetail provides an io.Writer that keeps the last N complete lines
// written to it and reports the current tail on every write. It is handy for
// surfacing a live "last few lines" view of a long-running command's output
// (e.g. as UI progress) while the full output is written elsewhere.
package linetail

import (
	"bytes"
	"strings"
	"sync"
)

// Writer keeps the last n complete lines written to it and reports the current
// tail via report on every write. It is safe for concurrent use, so a command's
// stdout and stderr (copied by separate goroutines) can share one.
type Writer struct {
	n       int
	report  func([]string)
	mu      sync.Mutex
	partial []byte
	lines   []string
}

// New returns a Writer that keeps the last n lines and calls report with the
// current tail whenever a new complete line is written.
func New(n int, report func([]string)) *Writer {
	return &Writer{n: n, report: report}
}

func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.partial = append(w.partial, p...)
	for {
		i := bytes.IndexByte(w.partial, '\n')
		if i < 0 {
			break
		}
		line := strings.TrimRight(string(w.partial[:i]), "\r")
		w.partial = w.partial[i+1:]
		w.lines = append(w.lines, line)
		if len(w.lines) > w.n {
			w.lines = w.lines[len(w.lines)-w.n:]
		}
	}
	tail := append([]string(nil), w.lines...)
	w.mu.Unlock()

	if len(tail) > 0 {
		w.report(tail)
	}
	return len(p), nil
}
