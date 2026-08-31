/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

const (
	MaxProjectLength = 63
	MaxNameLength    = 50
	EmptyString      = ""
)

type Validation interface {
	Execute() (string, error)
}

func NewAuthValidator(url string) *AuthValidator {
	return &AuthValidator{check: Url{url: url}}
}

type AuthValidator struct {
	check Url
}

func (av *AuthValidator) Execute() (string, error) {
	if err := av.validate(); err != nil {
		return av.getFailMsg(), err
	}
	return EmptyString, nil
}

func (av *AuthValidator) validate() error {
	return av.check.check()
}

func (av *AuthValidator) getFailMsg() string {
	return consts.BadAuthUrl
}

func NewCpValidator(url string) *CpValidator {
	return &CpValidator{check: Url{url: url}}
}

type CpValidator struct {
	check Url
}

func (cp *CpValidator) Execute() (string, error) {
	if err := cp.validate(); err != nil {
		return cp.getFailMsg(), err
	}
	return EmptyString, nil
}

func (cp *CpValidator) validate() error {
	return cp.check.check()
}

func (cp *CpValidator) getFailMsg() string {
	return consts.BadCpUrl
}

func NewClusterValidator(url string) *ClusterValidator {
	return &ClusterValidator{check: Url{url: url}}
}

type ClusterValidator struct {
	check Url
}

func (cv *ClusterValidator) Execute() (string, error) {
	if err := cv.validate(); err != nil {
		return cv.getFailMsg(), err
	}
	return EmptyString, nil
}

func (cv *ClusterValidator) validate() error {
	return cv.check.check()
}

func (cv *ClusterValidator) getFailMsg() string {
	return consts.BadClusterUrl
}

func NewProjectValidator(project string) *ProjectValidator {
	return &ProjectValidator{check: Alphanumeric{msg: project, length: MaxProjectLength}}
}

type ProjectValidator struct {
	check Alphanumeric
}

func (pv *ProjectValidator) getFailMsg() string {
	return consts.EmptyProjectValue
}

func (pv *ProjectValidator) Execute() (string, error) {
	if err := pv.check.check(); err != nil {
		return pv.getFailMsg(), err
	}
	return EmptyString, nil
}

type NameValidator struct {
	check Alphanumeric
}

func NewNameValidator(name string) *NameValidator {
	return &NameValidator{check: Alphanumeric{msg: name, length: MaxNameLength, validEmpty: true}}
}

func (nv *NameValidator) getFailMsg() string {
	return consts.BadWorkloadName
}

func (nv *NameValidator) Execute() (string, error) {
	if err := nv.check.check(); err != nil {
		return nv.getFailMsg(), err
	}
	return EmptyString, nil
}
