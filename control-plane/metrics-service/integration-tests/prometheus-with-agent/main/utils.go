package main

import (
	"fmt"
	"io"
	"os"
)

func createTempFile(content string) (*os.File, error) {
	file, err := os.CreateTemp("", "openmetrics_*")
	if err != nil {
		return nil, err
	}
	if _, err := file.WriteString(content); err != nil {
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	return file, nil
}

func bodyToString(r io.ReadCloser) (string, error) {
	defer r.Close()
	bodyBytes, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read capturedData: %w", err)
	}
	return string(bodyBytes), nil
}
