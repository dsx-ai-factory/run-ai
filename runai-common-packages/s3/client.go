/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package s3

import (
	"fmt"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	initOnce     sync.Once
	globalClient *minio.Client
)

func Init(cfg *S3Config) {
	initOnce.Do(func() {
		var err error
		if cfg != nil {
			_ = config.SetWithDefaults(cfg)
		} else {
			err = InitConfig()
		}
		if err != nil {
			logging.Warnf("failed to initialize S3 client. error: %v", err)
			return
		}
		conf := config.Get[*S3Config]()

		useIam := conf.AccessKeyId == ""
		var client *minio.Client
		if useIam {
			client, err = minio.New(conf.Endpoint, &minio.Options{
				Creds:  credentials.NewIAM(""),
				Secure: conf.UseSSL,
			})
		} else {
			client, err = minio.New(conf.Endpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(conf.AccessKeyId, conf.AccessKey, ""),
				Secure: conf.UseSSL,
			})
		}

		if err != nil || client == nil {
			logging.Warnf("failed to initialize S3 client. Err: %v", err)
		}
		globalClient = client
		logging.Debugf("s3 client is ready")
	})
}

// Client returns the shared client by pointer. minio.Client is not safe to
// copy -- it embeds a singleflight.Group guarding the credential cache with a
// sync.Mutex, so a copy duplicates in-flight state and its own lock.
func Client() (*minio.Client, error) {
	if globalClient != nil {
		return globalClient, nil
	}

	return nil, fmt.Errorf("s3 client is not initialized")
}
