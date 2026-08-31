/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"fmt"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

const (
	GeneralErrorMsg = "Encountered general error for %s"
)

type ClientAction int

const (
	// Unspecified is the default, will apply rule for every action
	Unspecified ClientAction = iota
	Create
	Update
	Delete
)

type ResourceFailingRule struct {
	ResourceName string
	ResourceKind string
	ClientAction ClientAction
	StatusError  *k8serrors.StatusError
}

func (rfr *ResourceFailingRule) GenerateDefaultError(name string) error {
	return &k8serrors.StatusError{ErrStatus: metav1.Status{Message: fmt.Sprintf(GeneralErrorMsg, name), Code: http.StatusBadRequest}}
}
