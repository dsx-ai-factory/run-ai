package filesystem

import (
	"fmt"
	"os"
)

func IsDirWritable(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	file, err := os.CreateTemp(dirPath, ".write-test")
	if err != nil {
		return fmt.Errorf("directory is not writable %s: %w", dirPath, err)
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	return nil
}
