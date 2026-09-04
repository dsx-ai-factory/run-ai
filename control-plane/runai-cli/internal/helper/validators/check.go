/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
)

// AlphanumericOrHyphenRegex allow first letter lowercase alphanumeric, after lowercase alphanumeric or hyphen and ends with lowercase alphanumeric
const AlphanumericOrHyphenRegex = "^[a-z0-9]([a-z0-9-]*)?$"

var (
	alphanumericRegex    = regexp.MustCompile(AlphanumericOrHyphenRegex)
	instance             = config.GetInstance()
	errEmptyValue        = errors.New("value is empty")
	errEndWithHyphen     = errors.New("cannot end with hyphen")
	errInvalidCharacters = errors.New("string contain other characters other then lower case alphanumeric or hyphen")
)

type Check interface {
	check() error
}

type Url struct {
	url string
}

func (uc *Url) check() error {
	_, err := url.ParseRequestURI(uc.url)
	if err != nil {
		return err
	}
	return nil
}

type Alphanumeric struct {
	length     int
	msg        string
	validEmpty bool
}

func (ac *Alphanumeric) check() error {
	msgLen := len(ac.msg)
	if ac.validEmpty && msgLen == 0 {
		return nil
	}
	if msgLen == 0 {
		return errEmptyValue
	}
	if ac.msg[msgLen-1] == '-' {
		return errEndWithHyphen
	}
	if ac.length > 0 {
		if len(ac.msg) > ac.length {
			return fmt.Errorf("exceeded max length (%d)", ac.length)
		}
	}
	if !alphanumericRegex.MatchString(ac.msg) {
		return errInvalidCharacters
	}
	return nil
}

type Token struct {
}

func (tc *Token) check() error {
	_, err := auth.GetToken(instance.GetTxn().GetCtx())
	if err != nil {
		errPathType := &os.PathError{}
		if errors.As(err, &errPathType) {
			return fmt.Errorf("%s. %s", consts.TokenErrorIfMissingAuth, consts.ReLoginMsg)
		}
		return err
	}
	return nil
}
