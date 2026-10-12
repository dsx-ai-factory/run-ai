/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package init_containers

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// legacyOutput loads output captured from the cluster builders at 47d891f9c,
// before their removal. Keep these fixtures independent of the CP implementation.
func legacyOutput[T any](label string) T {
	GinkgoHelper()
	data, err := os.ReadFile("testdata/legacy-builder-outputs.json")
	Expect(err).NotTo(HaveOccurred())
	outputs := map[string]json.RawMessage{}
	Expect(json.Unmarshal(data, &outputs)).To(Succeed())
	key := CurrentSpecReport().FullText() + " / " + label
	output, ok := outputs[key]
	Expect(ok).To(BeTrue(), "missing legacy fixture: %s", key)
	var result T
	Expect(json.Unmarshal(output, &result)).To(Succeed())
	return result
}
