/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parser

import (
	"errors"
)

var (
	ErrParserNoDic       = errors.New("parser is missing dictionary")
	ErrParseValueIsEmpty = errors.New("value is an empty string")
)

const (
	ErrMsgDictionaryMustContainAllKeys = "dictionary must contain all keys"
	ErrMsgDictionaryDoesNotContainKey  = "dictionary does not contain key"
	ErrMsgMissingMandatoryFields       = "missing mandatory fields"
)
