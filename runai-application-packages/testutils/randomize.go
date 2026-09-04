/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testutils

import "math/rand"

// letters used for creating a random string of given length
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))] //nolint:gosec // G404: non-cryptographic test data
	}
	return string(b)
}

// RandStringTestMode - for testing that compare texts we need to have static result
func RandStringTestMode(restrictedLetters string) {
	letters = []rune(restrictedLetters)
}
