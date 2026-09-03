/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"errors"
	"regexp"
)

// clean error like:
// admission webhook \"runai-admission-controller.runai.svc\" denied the request: The flags --gpu and --gpu-memory cannot be used together",
// by removing the first part which is not interesting to a customer
var admissionRegexp = regexp.MustCompile(`admission webhook .* denied the request: `)

func CleanWebhookError(err error) error {

	if admissionRegexp.FindString(err.Error()) != "" {
		return errors.New(admissionRegexp.ReplaceAllString(err.Error(), ""))
	}
	return err
}
