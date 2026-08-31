/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structs

import (
	"encoding/base64"
	"errors"
	"time"
)

const (
	TimeType = time.RFC3339Nano
)

var (
	errorNextTokenIsZero    = errors.New("next token cannot be with 0 timestamp")
	errorEncodeValueIsEmpty = errors.New("encoded value is empty")
)

type NextToken struct {
	timestamp time.Time
}

func NextTokenFromValue(value string) (*NextToken, error) {
	timestamp, err := decode(value)
	if err != nil {
		return nil, err
	}
	return &NextToken{timestamp: timestamp}, nil
}

func NextTokenFromTimestamp(timestamp time.Time) (*NextToken, error) {
	if timestamp.IsZero() {
		return nil, errorNextTokenIsZero
	}
	return &NextToken{timestamp: timestamp}, nil
}

func (nt *NextToken) String() string {
	return nt.timestamp.Format(TimeType)
}

func (nt *NextToken) Encode() (string, error) {
	return encode(nt.String())
}

func (nt *NextToken) GetTimestamp() time.Time {
	return nt.timestamp
}

func (nt *NextToken) IsSet() bool {
	return !nt.timestamp.IsZero()
}

func decode(value string) (time.Time, error) {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return time.Time{}, err
	}
	createdBy, err := time.Parse(TimeType, string(data))
	if err != nil {
		return time.Time{}, err
	}
	return createdBy, nil
}

func encode(value string) (string, error) {
	if len(value) == 0 {
		return "", errorEncodeValueIsEmpty
	}
	return base64.StdEncoding.EncodeToString([]byte(value)), nil
}
