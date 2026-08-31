/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeQuatedprintable_ShortLine(t *testing.T) {
	input := "Hello, world!"
	want := "Hello, world!"
	encoder := BodyEncoder{}
	got := encoder.encodeQuatedprintable(input)
	assert.Equal(t, want, got)
}

func TestEncodeQuatedprintable_LongLine(t *testing.T) {
	input := "" +
		"1234567890123456789012345678901234567890123456789012345678901234567890123456789"
	want := "123456789012345678901234567890123456789012345678901234567890123456789012345=\r\n6789"
	encoder := BodyEncoder{}
	got := encoder.encodeQuatedprintable(input)
	assert.Equal(t, want, got)
}

func TestEncodeQuatedprintable_MultipleLines(t *testing.T) {
	input := "Line1\nLine2\nLine3"
	want := "Line1\r\nLine2\r\nLine3"
	encoder := BodyEncoder{}
	got := encoder.encodeQuatedprintable(input)
	assert.Equal(t, want, got)
}

func TestEncodeQuatedprintable_DotTransparency(t *testing.T) {
	input := ".Hello\n..World"
	want := ".Hello\r\n..World"
	encoder := BodyEncoder{}
	got := encoder.encodeQuatedprintable(input)
	assert.Equal(t, want, got)
}

func TestEncodeText_ShortLine(t *testing.T) {
	input := "Hello, world!"
	want := "Hello, world!\r\n"
	encoder := BodyEncoder{}
	got := encoder.encodeText(input)
	assert.Equal(t, want, got)
}

func TestEncodeText_LongLine(t *testing.T) {
	input := "" +
		"1234567890123456789012345678901234567890123456789012345678901234567890123456789"
	want := "123456789012345678901234567890123456789012345678901234567890123456789012345678\r\n9\r\n"
	encoder := BodyEncoder{}
	got := encoder.encodeText(input)
	assert.Equal(t, want, got)
}

func TestEncodeText_MultipleLines(t *testing.T) {
	input := "Line1\nLine2\nLine3"
	want := "Line1\r\nLine2\r\nLine3\r\n"
	encoder := BodyEncoder{}
	got := encoder.encodeText(input)
	assert.Equal(t, want, got)
}

func TestEncodeText_DotTransparency(t *testing.T) {
	input := ".Hello\n..World"
	want := ".Hello\r\n..World\r\n"
	encoder := BodyEncoder{}
	got := encoder.encodeText(input)
	assert.Equal(t, want, got)
}

func TestEncodeText_SplitsAtWordBoundary(t *testing.T) {
	// 79-char input with a space at index 70; should split at the space, not at 78
	input := "123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789"
	want := "123456789 123456789 123456789 123456789 123456789 123456789 123456789\r\n123456789\r\n"
	encoder := BodyEncoder{}
	got := encoder.encodeText(input)
	assert.Equal(t, want, got)
}

func TestEncodeText_HtmlTagNotSplit(t *testing.T) {
	// prefix is 73 chars; appending </title> (8 chars) = 81 total. 
	// Hard cut at 78 would split </title> as "</titl" | "e>". We want to avoid splitting the tag.
	prefix := "    <title>Workload fourt3en-chars status changed from Running to Stopped"
	input := prefix + "</title>"
	lastSpace := strings.LastIndex(input[:78], " ")
	want := input[:lastSpace] + "\r\n" + input[lastSpace+1:] + "\r\n"
	encoder := BodyEncoder{}
	got := encoder.encodeText(input)
	assert.Equal(t, want, got)
}

func TestEncodeText_SpaceAtHardCutBoundaryConsumed(t *testing.T) {
	input := strings.Repeat("x", 78) + " next"
	want := strings.Repeat("x", 78) + "\r\nnext\r\n"
	encoder := BodyEncoder{}
	got := encoder.encodeText(input)
	assert.Equal(t, want, got)
}

func TestEncodeText_NoSpaceFallsBackToHardCut(t *testing.T) {
	// 80-char word with no spaces: must hard-cut at 78
	input := "12345678901234567890123456789012345678901234567890123456789012345678901234567890"
	want := "123456789012345678901234567890123456789012345678901234567890123456789012345678\r\n90\r\n"
	encoder := BodyEncoder{}
	got := encoder.encodeText(input)
	assert.Equal(t, want, got)
}

func TestGetEncodingType_Text(t *testing.T) {
	input := `
    MIME-Version: 1.0
    Content-Type: text/html; charset=UTF-8`
	want := "text/html"
	encoder := BodyEncoder{}
	got := encoder.getEncodingType(input)
	assert.Equal(t, want, got)
}

func TestGetEncodingType_Quatedprintable(t *testing.T) {
	input := `
    Content-Type: text/html; charset=UTF-8
    Content-Transfer-Encoding: quoted-printable
    `
	want := "quoted-printable"
	encoder := BodyEncoder{}
	got := encoder.getEncodingType(input)
	assert.Equal(t, want, got)
}
