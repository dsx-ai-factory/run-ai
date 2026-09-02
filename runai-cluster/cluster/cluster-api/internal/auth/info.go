/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import "k8s.io/client-go/rest"

type Info struct {
	Cfg               *rest.Config
	Username          string
	Subject           string
	OriginalSubject   string
	Groups            []string
	Uid               *int64
	Gid               *int64
	CanSeeAllProjects bool
	PreferredUsername string
	RawToken          string
}
