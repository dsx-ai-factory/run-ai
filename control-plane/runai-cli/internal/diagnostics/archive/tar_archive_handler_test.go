package archive

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TarArchiveHandler", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "tar_archive_handler_test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("NewTarArchiveHandler", func() {
		It("should create a new TarArchiveHandler", func() {
			testTarFilePath := filepath.Join(tempDir, "tar_file.tar")
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())
			defer closeHandler(h)

			Expect(h).NotTo(BeNil())
			Expect(h.file).NotTo(BeNil())
			Expect(h.tarWriter).NotTo(BeNil())
			Expect(h.appendMu).NotTo(BeNil())
			Expect(h.tarFilePath).To(Equal(testTarFilePath))
			Expect(h.closed).To(BeFalse())
		})

		It("should return an error if the file cannot be created", func() {
			invalidPath := "/invalid/path/to/tarfile.tar" // the dir containing the tar file must exist before file is created
			h, err := NewTarArchiveHandler(invalidPath)
			Expect(err).To(HaveOccurred())
			Expect(h).To(BeNil())
			Expect(err.Error()).To(Equal("open /invalid/path/to/tarfile.tar: no such file or directory"))
		})
	})

	Describe("AppendFile", func() {
		var testTarFilePath string

		BeforeEach(func() {
			testTarFilePath = filepath.Join(tempDir, "tar_file.tar")
		})

		It("should append file to tar", func() {
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			file := filepath.Join(tempDir, "test.txt")
			content := "content"
			err = os.WriteFile(file, []byte(content), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = h.AppendFile(file, "test.txt")
			Expect(err).NotTo(HaveOccurred())

			closeHandler(h)

			expectedFiles := map[string]string{
				"test.txt": content,
			}
			checkTarContents(testTarFilePath, expectedFiles)
		})

		It("should append file to tar with nested archive path", func() {
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())
			defer closeHandler(h)

			file := filepath.Join(tempDir, "test.txt")
			content := "content"
			err = os.WriteFile(file, []byte(content), 0644)
			Expect(err).NotTo(HaveOccurred())

			err = h.AppendFile(file, "/dir1/dir2/test.txt")
			Expect(err).NotTo(HaveOccurred())

			closeHandler(h)

			expectedFiles := map[string]string{
				"/dir1/dir2/test.txt": content,
			}
			checkTarContents(testTarFilePath, expectedFiles)
		})

		It("should append multiple files to tar", func() {
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			files := map[string]string{
				"file1.txt": "content1",
				"file2.txt": "content2",
				"file3.txt": "content3",
			}

			for filename, content := range files {
				filePath := filepath.Join(tempDir, filename)
				err = os.WriteFile(filePath, []byte(content), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = h.AppendFile(filePath, filename)
				Expect(err).NotTo(HaveOccurred())
			}

			closeHandler(h)

			checkTarContents(testTarFilePath, files)
		})

		It("should return an error when the tar archive is closed", func() {
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())
			defer closeHandler(h)

			err = h.Close()
			Expect(err).NotTo(HaveOccurred())

			err = h.AppendFile("test.txt", "test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("tar archive is closed"))
		})

		It("should return an error when the source file doesn't exist", func() {
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())
			defer closeHandler(h)

			nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
			err = h.AppendFile(nonExistentFile, "nonexistent.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to stat file"))
		})

		It("should return an error when failed to write header", func() {
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			file := filepath.Join(tempDir, "test.txt")
			err = os.WriteFile(file, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			h.file.Close() // close the underlying file to make tar writer operations fail

			err = h.AppendFile(file, "test.txt")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to write tar header"))

			h.Close() // not checking the error here because it's expected to fail - file is already closed
			err = os.Remove(h.tarFilePath)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Close", func() {
		It("should close the tar archive", func() {
			testTarFilePath := filepath.Join(tempDir, "tar_file.tar")
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			err = h.Close()
			Expect(err).NotTo(HaveOccurred())
			Expect(h.closed).To(BeTrue())
		})

		It("should return nil if the tar archive is already closed", func() {
			testTarFilePath := filepath.Join(tempDir, "tar_file.tar")
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				h.closed = false // reset the closed flag to false to allow cleanup
				closeHandler(h)
			}()

			h.closed = true

			err = h.Close()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error if fails to close", func() {
			testTarFilePath := filepath.Join(tempDir, "tar_file.tar")
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			// close for the first time - should work
			err = h.Close()
			Expect(err).NotTo(HaveOccurred())

			h.closed = false // reset the closed flag to false to allow Close to fail

			// close for the second time - should fail
			err = h.Close()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to close file"))
		})
	})

	Describe("Compress", func() {
		It("should compress the tar archive", func() {
			testTarFilePath := filepath.Join(tempDir, "tar_file.tar")
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			file := filepath.Join(tempDir, "test.txt")
			err = os.WriteFile(file, []byte("content"), 0644)
			Expect(err).NotTo(HaveOccurred())
			err = h.AppendFile(file, "test.txt")
			Expect(err).NotTo(HaveOccurred())

			err = h.Close()
			Expect(err).NotTo(HaveOccurred())

			err = h.Compress()
			Expect(err).NotTo(HaveOccurred())

			// make sure the tar.gz file is created
			fileInfo, err := os.Stat(testTarFilePath + ".gz")
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Size()).To(BeNumerically(">", 0))

			// make sure the tar file is removed
			_, err = os.Stat(testTarFilePath)
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("should return an error if failed to ", func() {
			testTarFilePath := filepath.Join(tempDir, "tar_file.tar")
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			err = h.Compress()
			Expect(err.Error()).To(Equal("tar archive need to be closed before compressing"))
		})

		It("should return an error if failed to open tar file for compression", func() {
			testTarFilePath := filepath.Join(tempDir, "tar_file_open_fail.tar")
			h, err := NewTarArchiveHandler(testTarFilePath)
			Expect(err).NotTo(HaveOccurred())

			err = h.Close()
			Expect(err).NotTo(HaveOccurred())

			err = os.Remove(testTarFilePath) // remove the tar file to cause open failure
			Expect(err).NotTo(HaveOccurred())

			err = h.Compress()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))
		})
	})
})

func closeHandler(h *TarArchiveHandler) {
	err := h.Close()
	Expect(err).NotTo(HaveOccurred())
}

func checkTarContents(testTarFilePath string, expectedFiles map[string]string) {
	tarFile, err := os.Open(testTarFilePath)
	Expect(err).NotTo(HaveOccurred())
	defer tarFile.Close()

	verifyTarContents(tarFile, expectedFiles)
}

func verifyTarContents(tarReader io.Reader, expectedContent map[string]string) {
	tr := tar.NewReader(tarReader)
	foundFiles := make(map[string]string)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		Expect(err).NotTo(HaveOccurred())

		content, exists := expectedContent[header.Name]
		Expect(exists).To(BeTrue(), "Unexpected file or directory: "+header.Name)

		// for regular files, verify content
		if header.Typeflag == tar.TypeReg {
			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, tr)
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal(content), "Content mismatch for "+header.Name)
		}

		foundFiles[header.Name] = content
	}

	Expect(foundFiles).To(Equal(expectedContent))
}
