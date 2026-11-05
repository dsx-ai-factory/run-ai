/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logreader

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// maxLogLineBytes is the buffer a single line may need. bufio.Scanner defaults
// to 64KiB, and a longer line ends the read early - silently before the scan
// error was reported, and with an error once it is.
const maxLogLineBytes = 1024 * 1024

type LogFileInfo struct {
	Name       string
	ModifyTime int64
	LineCount  int
}

type LogReader struct {
	path        string
	tail        int
	logListInfo []LogFileInfo
}

// GetLogList returns the newest lines of the log directory, at most tail of
// them, spanning the rotated files newest first.
func (lr *LogReader) GetLogList() ([]string, error) {
	var lines []string
	remaining := lr.tail
	for _, log := range lr.logListInfo {
		if remaining <= 0 {
			break
		}

		fileLines, err := readLastLines(filepath.Join(lr.path, log.Name), remaining)
		if err != nil {
			return lines, fmt.Errorf("read log file %q: %w", log.Name, err)
		}

		lines = append(lines, fileLines...)
		remaining -= len(fileLines)
	}
	return lines, nil
}

// readLastLines returns the last n lines of the file. A log file is appended
// to, so the lines worth reporting are the ones at its end; reading from the
// start returned the oldest lines of the newest file.
func readLastLines(filePath string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}

	readFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	// Ring buffer: keep the last n lines without holding the whole file.
	ring := make([]string, n)
	start, count := 0, 0

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	fileScanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLogLineBytes)
	for fileScanner.Scan() {
		ring[(start+count)%n] = fileScanner.Text()
		if count < n {
			count++
		} else {
			start = (start + 1) % n
		}
	}

	// A failed read leaves the result short without saying so, which is exactly
	// what this function is supposed to report accurately. A close failure is
	// reported alongside the scan failure rather than dropped because the scan
	// failed first.
	if err := fileScanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", errors.Join(err, readFile.Close()))
	}
	if err := readFile.Close(); err != nil {
		return nil, fmt.Errorf("close: %w", err)
	}

	lines := make([]string, 0, count)
	for i := 0; i < count; i++ {
		lines = append(lines, ring[(start+i)%n])
	}
	return lines, nil
}

func NewLogReader(path string, tail int) (*LogReader, error) {
	logFileInfoList, err := ListLogFilesByCreateTime(path, tail)
	if err != nil {
		return nil, err
	}

	return &LogReader{
		path:        path,
		tail:        tail,
		logListInfo: logFileInfoList,
	}, nil
}

// ListLogFilesByCreateTime returns the files holding the newest lines, newest
// first, stopping once they cover limit lines (limit <= 0 keeps every file).
//
// The files are ordered by modification time before the limit is applied.
// os.ReadDir sorts by name, and a rotated name sorts before the active file
// because '-' precedes '.', so cutting the list off in directory order dropped
// the active metrics.log - the file that holds the newest lines.
func ListLogFilesByCreateTime(filePath string, limit int) ([]LogFileInfo, error) {
	entries, err := os.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	type candidate struct {
		name    string
		modTime int64
	}

	candidates := make([]candidate, 0, len(entries))
	for _, entry := range entries {
		fInfo, err := entry.Info()
		if err != nil {
			continue
		}

		candidates = append(candidates, candidate{name: entry.Name(), modTime: fInfo.ModTime().UTC().Unix()})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].modTime != candidates[j].modTime {
			return candidates[i].modTime > candidates[j].modTime
		}

		// Modification times only resolve to a second here, so files rotated
		// within the same second need a stable tie-break; the active file wins
		// it, since '.' sorts after '-'.
		return candidates[i].name > candidates[j].name
	})

	var (
		lineCount    int
		logFilesInfo []LogFileInfo
	)
	for _, c := range candidates {
		lcTemp, err := lineCounter(filepath.Join(filePath, c.name))
		if err != nil {
			// A directory or an unreadable entry is not a log file to report
			// on; keeping it in the list made the read fail on it later.
			continue
		}

		logFilesInfo = append(logFilesInfo, LogFileInfo{
			Name:       c.name,
			ModifyTime: c.modTime,
			LineCount:  lcTemp,
		})

		lineCount += lcTemp
		if limit > 0 && lineCount >= limit {
			break
		}
	}

	return logFilesInfo, nil
}

func lineCounter(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Count lines
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLogLineBytes)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("count lines: %w", err)
	}
	return lineCount, nil
}
