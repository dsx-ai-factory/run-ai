/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structs

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
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

// OffsetToNextToken and NextTokenToOffset bridge the gap between two
// incompatible pagination systems:
//
// The CLI's pagination infrastructure (NextToken) expects timestamps for
// pagination:
//   - Used by workloads, projects, etc: "Get items created after 2024-11-11T10:30:00Z"
//   - NextToken internally stores a time.Time and encodes it as base64 RFC3339Nano
//
// But some backends (e.g. Template, Node) use traditional offset-based
// pagination instead, with no createdAt cursor to anchor a timestamp to:
//   - "Get 50 items starting at position 100" (offset=100, limit=50)
//
// This adapter lets those offset-only backends still expose --next-token for
// CLI consistency with the timestamp-based commands, by encoding the numeric
// offset as a Unix timestamp. This is a harmless encoding trick since:
//   - We only need to store/retrieve the number, not use it as an actual timestamp
//   - It allows these commands to use the same --next-token flag as other CLI commands
//
// Example workflow:
//  1. Backend returns: nextOffset = 100
//  2. We convert 100 → Unix timestamp 100 seconds after epoch = "1970-01-01T00:01:40Z"
//  3. NextToken base64-encodes this timestamp → "MTk3MC0wMS0wMVQwMDowMTo0MFo="
//  4. User sees in CLI output: NEXTTOKEN: MTk3MC0wMS0wMVQwMDowMTo0MFo=
//  5. User runs the list command again with --next-token MTk3MC0wMS0wMVQwMDowMTo0MFo=
//  6. We decode token → "1970-01-01T00:01:40Z" → Unix seconds = 100 → offset = 100
//  7. We call the backend with offset=100 to get the next page

// OffsetToNextToken encodes a plain, non-negative offset as a NextToken.
// See the package-level comment above for why this exists.
func OffsetToNextToken(offset int32) (*NextToken, error) {
	// Validate that offset is non-negative.
	// Valid tokens from our system: always positive, never hit this error.
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative: %d", offset)
	}

	// Encode offset as Unix epoch + offset seconds (e.g., 100 → Jan 1 1970, 00:01:40 UTC).
	timestamp := time.Unix(int64(offset), 0).UTC()
	return NextTokenFromTimestamp(timestamp)
}

// NextTokenToOffset extracts the numeric offset from a NextToken.
//
// This reverses the encoding done by OffsetToNextToken, converting the fake
// timestamp back to the original offset number the backend expects. Returns
// (nil, nil) if the token is unset.
//
// Example: when the user provides --next-token MTk3MC0wMS0wMVQwMDowMTo0MFo=
//  1. NextToken decodes base64 → "1970-01-01T00:01:40Z"
//  2. We extract timestamp → time.Time representing Jan 1 1970, 00:01:40 UTC
//  3. We call timestamp.Unix() → 100 seconds since epoch
//  4. Return offset=100 to pass to the backend
func NextTokenToOffset(token *NextToken) (*int32, error) {
	if token == nil || !token.IsSet() {
		return nil, nil
	}

	// Extract timestamp and convert back to offset (Unix seconds since epoch).
	timestamp := token.GetTimestamp()
	unixSeconds := timestamp.Unix()

	// Validate that the offset is non-negative.
	// Valid tokens (from our system): always positive, never hit this error.
	// This check only catches a hand-crafted or corrupted token.
	if unixSeconds < 0 {
		return nil, fmt.Errorf("invalid token: offset cannot be negative")
	}

	// Validate the offset fits in an int32 before the narrowing cast below.
	// Valid tokens (from our system): always in range, never hit this error.
	if unixSeconds > math.MaxInt32 {
		return nil, fmt.Errorf("invalid token: offset out of range")
	}

	offset := int32(unixSeconds)

	return &offset, nil
}
