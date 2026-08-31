/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type WorkloadNodeAffinity struct {
	SpecName     string       `json:"specName,omitempty"`
	NodeAffinity NodeAffinity `json:"nodeAffinity"`
}

type WorkloadNodeAffinityArray []WorkloadNodeAffinity
