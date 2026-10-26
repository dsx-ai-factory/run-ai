/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"context"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/s3"
	"github.com/minio/minio-go/v7"
)

type S3Client struct {
	s3Client *minio.Client
}

func NewS3Client() (S3Deleter, error) {
	client, err := s3.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client. err: %v", err)
	}

	return S3Client{
		s3Client: client,
	}, nil
}

func (d S3Client) Delete(id string) error {
	conf := config.Get[*ReportsConfiguration]()

	err := d.s3Client.RemoveObject(context.Background(), conf.S3Config.Bucket, id, minio.RemoveObjectOptions{})
	if err == nil || d.isObjectNotFound(err) {
		return nil
	}

	return err
}

func (d S3Client) isObjectNotFound(err error) bool {
	var errResponse minio.ErrorResponse
	return errors.As(err, &errResponse) && errResponse.Code == minio.NoSuchKey
}
