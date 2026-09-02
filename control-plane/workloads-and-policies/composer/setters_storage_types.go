/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

type S3Info struct {
	Items  map[string]*S3Item
	Config ClusterStorageConfig

	// RunAsUid/RunAsGid feed goofys' --uid/--gid. They originate from the
	// container security fields, not the storage block, and are workload-wide.
	RunAsUid *int64
	RunAsGid *int64
}

type S3Item struct {
	Bucket             string
	Path               string
	Url                string
	SecretName         string
	AccessKeyIdKey     string
	SecretAccessKeyKey string
	SessionToken       string
}
