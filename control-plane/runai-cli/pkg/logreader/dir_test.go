/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logreader

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func writeLogFile(t *testing.T, dir, name string, lines int, modTime time.Time) {
	t.Helper()

	var b strings.Builder
	for i := 1; i <= lines; i++ {
		fmt.Fprintf(&b, "%s-%d\n", name, i)
	}
	writeLogContent(t, dir, name, b.String(), modTime)
}

func writeLogContent(t *testing.T, dir, name, content string, modTime time.Time) {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}

func getLogList(t *testing.T, dir string, tail int) []string {
	t.Helper()

	reader, err := NewLogReader(dir, tail)
	if err != nil {
		t.Fatalf("new log reader: %v", err)
	}
	lines, err := reader.GetLogList()
	if err != nil {
		t.Fatalf("get log list: %v", err)
	}
	return lines
}

func TestGetLogListReturnsExactlyTailLines(t *testing.T) {
	for _, tail := range []int{0, 1, 5, 10} {
		t.Run(fmt.Sprintf("tail=%d", tail), func(t *testing.T) {
			dir := t.TempDir()
			writeLogFile(t, dir, "metrics.log", 10, time.Now())

			if got := getLogList(t, dir, tail); len(got) != tail {
				t.Fatalf("got %d lines, want %d: %v", len(got), tail, got)
			}
		})
	}
}

func TestGetLogListReturnsEveryLineWhenTailExceedsThem(t *testing.T) {
	dir := t.TempDir()
	writeLogFile(t, dir, "metrics.log", 10, time.Now())

	if got := getLogList(t, dir, 11); len(got) != 10 {
		t.Fatalf("got %d lines, want 10: %v", len(got), got)
	}
}

// A log file grows at its end, so the lines a tail reports are the last ones.
func TestGetLogListReturnsTheLastLinesOfAFile(t *testing.T) {
	for _, tc := range []struct {
		tail int
		want []string
	}{
		{tail: 1, want: []string{"metrics.log-5"}},
		{tail: 2, want: []string{"metrics.log-4", "metrics.log-5"}},
		{tail: 5, want: []string{"metrics.log-1", "metrics.log-2", "metrics.log-3", "metrics.log-4", "metrics.log-5"}},
	} {
		t.Run(fmt.Sprintf("tail=%d", tc.tail), func(t *testing.T) {
			dir := t.TempDir()
			writeLogFile(t, dir, "metrics.log", 5, time.Now())

			if got := getLogList(t, dir, tc.tail); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// The budget covers the whole list, not each file: a tail of 4 over two files
// of 3 lines must return 4 lines, the newest ones first.
func TestGetLogListSpansFilesWithoutExceedingTail(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	writeLogFile(t, dir, "old.log", 3, now.Add(-time.Hour))
	writeLogFile(t, dir, "new.log", 3, now)

	want := []string{"new.log-1", "new.log-2", "new.log-3", "old.log-3"}
	if got := getLogList(t, dir, 4); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v (the newest file is read first and the budget spans the remaining files)", got, want)
	}
}

// lumberjack names a rotated file metrics-<timestamp>.log, which sorts before
// the active metrics.log, so selecting files in directory order and only then
// cutting the list off reported the rotated file instead of the active one.
func TestGetLogListKeepsTheActiveFileWhenARotatedNameSortsFirst(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	writeLogFile(t, dir, "metrics-20200101000000.log", 200, now.Add(-time.Hour))
	writeLogFile(t, dir, "metrics.log", 10, now)

	want := []string{"metrics.log-6", "metrics.log-7", "metrics.log-8", "metrics.log-9", "metrics.log-10"}
	if got := getLogList(t, dir, 5); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v (the active file holds the newest lines)", got, want)
	}
}

// A stray directory inside the log directory is not a log file; reading it used
// to fail the whole command, which prints nothing when the read fails.
func TestGetLogListSkipsEntriesThatAreNotLogFiles(t *testing.T) {
	dir := t.TempDir()
	writeLogFile(t, dir, "metrics.log", 2, time.Now())
	if err := os.Mkdir(filepath.Join(dir, "archive"), 0o700); err != nil {
		t.Fatalf("mkdir archive: %v", err)
	}

	want := []string{"metrics.log-1", "metrics.log-2"}
	if got := getLogList(t, dir, 2); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// A line longer than the 64KiB bufio.Scanner default used to end the read.
func TestGetLogListReadsLinesLongerThanTheScanBuffer(t *testing.T) {
	dir := t.TempDir()
	long := strings.Repeat("x", 100*1024)
	writeLogContent(t, dir, "metrics.log", long+"\nlast\n", time.Now())

	got := getLogList(t, dir, 2)
	if len(got) != 2 || got[0] != long || got[1] != "last" {
		t.Fatalf("got %d lines, want the %d byte line followed by %q", len(got), len(long), "last")
	}
}
