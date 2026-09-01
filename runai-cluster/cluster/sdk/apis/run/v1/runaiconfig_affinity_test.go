/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RunaiConfig CRD Affinity Fields", func() {
	Context("Generated CRD YAML", func() {
		It("should have x-kubernetes-preserve-unknown-fields on all schemaless affinity fields", func() {
			_, thisFile, _, ok := runtime.Caller(0)
			Expect(ok).To(BeTrue())
			crdPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "..",
				"config", "crd", "bases", "run.ai_runaiconfigs.yaml")
			_, err := os.Stat(crdPath)
			Expect(err).NotTo(HaveOccurred(), "CRD file not found at %s", crdPath)

			file, err := os.Open(crdPath)
			Expect(err).NotTo(HaveOccurred())
			defer file.Close()

			scanner := bufio.NewScanner(file)
			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			Expect(scanner.Err()).NotTo(HaveOccurred())

			type affinityField struct {
				lineNum            int
				hasPreserveUnknown bool
			}

			var schemalessAffinityFields []affinityField

			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if trimmed != "affinity:" {
					continue
				}

				indent := len(line) - len(strings.TrimLeft(line, " "))

				// Scan the next few lines at indent+2 to determine if this
				// is a schemaless field (only "type: object") or a full-schema
				// field (has "properties:" with nested content).
				childIndent := indent + 2
				isSchemaless := false
				hasPreserve := false
				hasProperties := false

				for j := i + 1; j < len(lines) && j < i+20; j++ {
					nextLine := lines[j]
					if strings.TrimSpace(nextLine) == "" {
						continue
					}
					nextIndent := len(nextLine) - len(strings.TrimLeft(nextLine, " "))

					// If we've gone back to the same or lesser indent, stop
					if nextIndent <= indent {
						break
					}

					nextTrimmed := strings.TrimSpace(nextLine)

					if nextIndent == childIndent {
						if nextTrimmed == "type: object" {
							isSchemaless = true
						}
						if nextTrimmed == "properties:" {
							hasProperties = true
						}
						if nextTrimmed == "x-kubernetes-preserve-unknown-fields: true" {
							hasPreserve = true
						}
					}
				}

				// A schemaless affinity field has "type: object" but no "properties:"
				if isSchemaless && !hasProperties {
					schemalessAffinityFields = append(schemalessAffinityFields, affinityField{
						lineNum:            i + 1,
						hasPreserveUnknown: hasPreserve,
					})
				}
			}

			Expect(schemalessAffinityFields).NotTo(BeEmpty(),
				"No schemaless affinity fields found in CRD — is the CRD file up to date?")

			var missing []int
			for _, af := range schemalessAffinityFields {
				if !af.hasPreserveUnknown {
					missing = append(missing, af.lineNum)
				}
			}

			Expect(missing).To(BeEmpty(),
				"Schemaless affinity fields at CRD lines %v are missing x-kubernetes-preserve-unknown-fields: true. "+
					"Add '// +kubebuilder:pruning:PreserveUnknownFields' marker to the Go type and regenerate CRDs.",
				missing)

			// Sanity check: all current operand affinity fields should be represented.
			Expect(len(schemalessAffinityFields)).To(BeNumerically(">=", 26),
				"Expected at least 26 schemaless affinity fields, found %d", len(schemalessAffinityFields))
		})
	})
})
