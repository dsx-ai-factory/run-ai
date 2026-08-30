package logreader

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
)

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

func (lr *LogReader) GetLogList() ([]string, error) {
	var lines []string
	var lc = lr.tail
	for _, log := range lr.logListInfo {
		readFile, err := os.Open(filepath.Join(lr.path, log.Name))

		if err != nil {
			return lines, err
		}
		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)
		var fileLines []string

		for fileScanner.Scan() {
			fileLines = append(fileLines, fileScanner.Text())
		}

		readFile.Close()

		for _, line := range fileLines {
			lines = append(lines, line)
			lc--

			if lc < 0 {
				break
			}
		}
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

func ListLogFilesByCreateTime(filePath string, limit int) ([]LogFileInfo, error) {
	var (
		lineCount    = 0
		lcTemp       int
		logFilesInfo []LogFileInfo
		fInfo        os.FileInfo
	)

	fileList, err := os.ReadDir(filePath)
	if err != nil {
		return nil, err
	}

	for _, file := range fileList {
		fInfo, err = file.Info()
		if err != nil {
			continue
		}
		lcTemp, err = lineCounter(filepath.Join(filePath, file.Name()))
		if err != nil {
			continue
		}

		logFilesInfo = append(logFilesInfo, LogFileInfo{
			Name:       file.Name(),
			ModifyTime: fInfo.ModTime().UTC().Unix(),
			LineCount:  lcTemp,
		})

		lineCount += lcTemp
		if limit > 0 && lineCount > limit {
			break
		}
	}

	sort.Slice(logFilesInfo, func(i, j int) bool {
		return logFilesInfo[i].ModifyTime > logFilesInfo[j].ModifyTime
	})

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
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	return lineCount, nil
}
