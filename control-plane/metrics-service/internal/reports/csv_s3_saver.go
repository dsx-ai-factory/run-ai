/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/s3"
	"github.com/minio/minio-go/v7"
)

const (
	csvContentType = "text/csv"
)

type CSVS3Saver struct {
	s3Client *minio.Client
}

func NewCSVS3Saver() (S3Saver, error) {
	conf := config.Get[*ReportsConfiguration]()

	client, err := s3.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client. err: %v", err)
	}

	exists, err := client.BucketExists(context.Background(), conf.S3Config.Bucket)
	if err != nil || !exists {
		return nil, fmt.Errorf("bucket %s does not exists. err: %v", conf.S3Config.Bucket, err)
	}

	return CSVS3Saver{
		s3Client: client,
	}, nil
}

func (c CSVS3Saver) Save(tenantId int32, id string, iterator ResultMapIterator, labels []string, metrics []string) error {
	filePath := getFilePath(tenantId, id)
	err := createCSVFile(filePath, iterator, labels, metrics)
	if err != nil {
		return err
	}

	conf := config.Get[*ReportsConfiguration]()

	_, err = c.s3Client.FPutObject(context.Background(), conf.S3Config.Bucket, id, filePath, minio.PutObjectOptions{ContentType: csvContentType})
	if err != nil {
		return err
	}

	return nil
}

func createCSVFile(path string, iterator ResultMapIterator, labels []string, metrics []string) error {
	file, err := createFileAndDirectory(path)
	if err != nil {
		logging.Errorf("Failed to create file: %s", err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writeCSVHeaders(writer, labels, metrics)
	if err != nil {
		return err
	}

	err = writeCSVData(writer, iterator)
	if err != nil {
		return err
	}

	return nil
}

func writeCSVHeaders(writer *csv.Writer, labels []string, metrics []string) error {
	if err := writer.Write(append(labels, metrics...)); err != nil {
		logging.Errorf("Failed to write headers: %s", err)
		return err
	}

	return nil
}

func writeCSVData(writer *csv.Writer, iterator ResultMapIterator) error {
	for keys, values := range iterator {
		keysSplit := strings.Split(keys, ",")
		row := append(keysSplit, values...)
		if err := writer.Write(row); err != nil {
			logging.Errorf("Failed to write row: %s", err)
			return err
		}
	}
	return nil
}

func createFileAndDirectory(path string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(path), 0o750)
	if err != nil {
		logging.Errorf("Failed to create directory: %v", err)
		return nil, err
	}

	return os.Create(path) //nolint:gosec // G304: path is constructed internally
}

func getFilePath(tenantId int32, name string) string {
	return fmt.Sprintf("/tmp/reports/%d/%s", tenantId, name)
}
