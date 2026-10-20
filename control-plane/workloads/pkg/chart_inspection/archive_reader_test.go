// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type archiveEntry struct {
	name, body string
	kind       byte
}

func chartArchive(entries ...archiveEntry) []byte {
	GinkgoHelper()

	var data bytes.Buffer
	gz := gzip.NewWriter(&data)
	tarWriter := tar.NewWriter(gz)
	for _, entry := range entries {
		kind := entry.kind
		if kind == 0 {
			kind = tar.TypeReg
		}
		header := &tar.Header{Name: entry.name, Mode: 0600, Size: int64(len(entry.body)), Typeflag: kind}
		if kind != tar.TypeReg {
			header.Size = 0
			header.Linkname = "elsewhere"
		}
		{
			err := tarWriter.WriteHeader(header)
			Expect(err).NotTo(HaveOccurred())
		}

		if kind == tar.TypeReg {
			{
				_, err := tarWriter.Write([]byte(entry.body))
				Expect(err).NotTo(HaveOccurred())
			}
		}
	}
	{
		err := tarWriter.Close()
		Expect(err).NotTo(HaveOccurred())
	}
	{
		err := gz.Close()
		Expect(err).NotTo(HaveOccurred())
	}

	return data.Bytes()
}

var _ = Describe("Chart archive parsing", func() {
	It("uses the standard API error mapping for archive diagnostics", func() {
		archive := chartArchive(archiveEntry{
			name: fixtureChartMetadataPath,
			body: fixtureChartMetadataYAML + "record not found: first\nrecord not found: second\n",
		})
		_, err := ParseArchive(archive)
		Expect(err).To(HaveOccurred())
		Expect(err.Code).To(BeNumerically("==", http.StatusNotFound))
		Expect(err.Details).To(Equal("requested resource does not exist"))
	})

	It("returns bad request for ordinary malformed archive diagnostics", func() {
		archive := chartArchive(archiveEntry{
			name: fixtureChartMetadataPath,
			body: fixtureChartMetadataYAML + "duplicate: first\nduplicate: second\n",
		})
		_, err := ParseArchive(archive)
		Expect(err).To(HaveOccurred())
		Expect(err.Code).To(BeNumerically("==", http.StatusBadRequest))
		Expect(err.Message).To(Equal("invalid Helm chart archive"))
		Expect(err.Details).To(ContainSubstring("duplicate YAML key"))
	})

	It("Inspector Bounds Expanded YAML Aliases", func() {
		for _, filename := range []string{fixtureQuestionsFilename, fixtureYMLQuestionsFilename} {
			for _, tc := range []struct{ name, value string }{
				{"scalar bytes", "[&large " + strings.Repeat("x", 32<<10) + strings.Repeat(", *large", 1000) + "]"},
				{"nodes", "[&item x" + strings.Repeat(", *item", 26000) + "]"},
			} {
				{
					By(filename + "/" + tc.name)
					questions := "questions:\n- variable: data\n  label: Data\n  type: array[string]\n  default: " + tc.value + "\n"
					archive := chartArchive(
						archiveEntry{name: fixtureChartMetadataPath, body: fixtureChartMetadataYAML},
						archiveEntry{name: "demo/" + filename, body: questions},
					)
					inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) { return archive, nil }, 0)
					result, err := inspector.Inspect(context.Background(), Selection{ChartID: fixtureUploadID}, "tenant", "")
					Expect(err).NotTo(HaveOccurred())
					Expect(!result.HasQuestions || result.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE || len(result.Questions) != 0 || !hasWarning(result, "UNUSABLE_QUESTIONS")).To(BeFalse(), func() string {
						return fmt.Sprintf("expanded aliases were accepted: %#v", result)
					})

					encoded, marshalErr := json.Marshal(result)
					Expect(marshalErr).NotTo(HaveOccurred())
					Expect(len(encoded) > 4096).To(BeFalse(), func() string {
						return fmt.Sprintf("expanded aliases leaked into response: %d bytes", len(encoded))
					})
				}
			}
		}
	})
	It("Parse Archive Returns Metadata Without Questions", func() {
		got, err := ParseArchive(chartArchive(archiveEntry{name: fixtureChartMetadataPath, body: "apiVersion: v2\nname: example\nversion: 1.2.3\n"}, archiveEntry{name: fixtureValuesPath, body: "replicas: 3\n"}))
		Expect(err).NotTo(HaveOccurred())
		Expect(got == nil || got.ChartName != fixtureExampleChartName || got.ChartVersion != fixtureVersion || got.HasQuestions || got.QuestionsStatus != "absent").To(BeFalse(), func() string {
			return fmt.Sprintf("metadata = %#v", got)
		})

		encoded, marshalErr := json.Marshal(got)
		Expect(marshalErr).NotTo(HaveOccurred())
		Expect(bytes.Contains(encoded, []byte("defaultValues"))).To(BeFalse(), func() string {
			return fmt.Sprintf("default values leaked: %s", encoded)
		})
	})
	It("Parse Archive Normalizes Either Questions Filename", func() {
		for _, filename := range []string{fixtureQuestionsFilename, fixtureYMLQuestionsFilename} {
			{
				By(filename)
				got, err := ParseArchive(chartArchive(
					archiveEntry{name: fixtureChartMetadataPath, body: "name: example\nversion: 1.2.3\n"},
					archiveEntry{name: fixtureValuesPath, body: "replicas: 3\n"},
					archiveEntry{name: "demo/" + filename, body: "groups: [{name: Compute, label: Compute settings}]\nquestions: [{variable: replicas, type: int, default: 4, group: Compute}]\n"},
					archiveEntry{name: "demo/charts/child/questions.yaml", body: fixtureMalformedQuestions},
					archiveEntry{name: "demo/charts/child/questions.yml", body: fixtureMalformedQuestions},
				))
				Expect(err).NotTo(HaveOccurred())
				Expect(got.ChartName).To(Equal(fixtureExampleChartName))
				Expect(got.ChartVersion).To(Equal(fixtureVersion))
				Expect(got.HasQuestions).To(BeTrue())
				Expect(got.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE))
				Expect(got.Questions).To(HaveLen(1))
				Expect(got.Groups).To(HaveLen(1))

				question := got.Questions[0]
				Expect(question["variable"] != "replicas" || question["label"] != "replicas" || question["type"] != "int" || question["default"] != 4 || question["group"] != "compute").To(BeFalse(), func() string {
					return fmt.Sprintf("normalized question = %#v", question)
				})
				Expect(got.Groups[0]["name"] != "compute" || got.Groups[0]["label"] != "Compute settings" || !hasWarning(got, "LABEL_DEFAULTED")).To(BeFalse(), func() string {
					return fmt.Sprintf("normalized groups or warnings = %#v / %#v", got.Groups, got.Warnings)
				})
			}
		}
	})
	It("Parse Archive Unusable Questions Warnings Name Their File", func() {
		for _, filename := range []string{fixtureQuestionsFilename, fixtureYMLQuestionsFilename} {
			for _, tc := range []struct{ name, body string }{
				{"malformed", fixtureMalformedQuestions},
				{"root array", "- variable: replicas\n"},
				{"empty", ""},
				{"invalid form", "groups: [{name: Compute}]\nquestions: [{variable: replicas, type: int, optional_hint: ignored}, {type: string}]\n"},
			} {
				{
					By(filename + "/" + tc.name)
					got, err := ParseArchive(chartArchive(
						archiveEntry{name: fixtureChartMetadataPath, body: "name: example\nversion: 1.2.3\n"},
						archiveEntry{name: "demo/" + filename, body: tc.body},
					))
					Expect(err).NotTo(HaveOccurred())
					Expect(got.ChartName).To(Equal(fixtureExampleChartName))
					Expect(got.ChartVersion).To(Equal(fixtureVersion))
					Expect(got.HasQuestions).To(BeTrue())
					Expect(got.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE))
					Expect(got.Questions).To(HaveLen(0))
					Expect(got.Groups).To(HaveLen(0))

					found := false
					for _, warning := range got.Warnings {
						if warning.Code == "UNUSABLE_QUESTIONS" {
							found = true
							Expect(warning.Path != "$" || !strings.Contains(warning.Message, filename)).To(BeFalse(), func() string {
								return fmt.Sprintf("unusable warning lost its source filename: %#v", warning)
							})
						}
					}
					Expect(found).To(BeTrue(), func() string {
						return fmt.Sprintf("missing unusable warning: %#v", got.Warnings)
					})
				}
			}
		}
	})
	It("Parse Archive Rejects Ambiguous Questions Filenames", func() {
		for _, first := range []string{fixtureQuestionsFilename, fixtureYMLQuestionsFilename} {
			{
				By(first + " first")
				second := fixtureQuestionsFilename
				if first == second {
					second = fixtureYMLQuestionsFilename
				}
				got, err := ParseArchive(chartArchive(
					archiveEntry{name: fixtureChartMetadataPath, body: fixtureChartMetadataYAML},
					archiveEntry{name: "demo/" + first, body: fixtureEmptyQuestions},
					archiveEntry{name: "demo/" + second, body: fixtureMalformedQuestions},
				))
				Expect(err == nil || got != nil).To(BeFalse(), func() string {
					return fmt.Sprintf("ambiguous declarations accepted: result=%#v error=%v", got, err)
				})
			}
		}
	})
	It("Parse Archive Rejects Unsafe Packages", func() {
		chart := archiveEntry{name: fixtureChartMetadataPath, body: fixtureChartMetadataYAML}
		for _, tc := range []struct {
			name    string
			entries []archiveEntry
		}{
			{"traversal", []archiveEntry{chart, {name: "demo/../outside", body: "x"}}},
			{"absolute", []archiveEntry{chart, {name: "/outside", body: "x"}}},
			{"backslash traversal", []archiveEntry{chart, {name: "demo\\..\\outside", body: "x"}}},
			{"symlink", []archiveEntry{chart, {name: "demo/link", kind: tar.TypeSymlink}}},
			{"hardlink", []archiveEntry{chart, {name: "demo/link", kind: tar.TypeLink}}},
			{"duplicate metadata", []archiveEntry{chart, chart}},
			{"duplicate questions.yml", []archiveEntry{chart, {name: "demo/questions.yml", body: fixtureEmptyQuestions}, {name: "demo/questions.yml", body: fixtureEmptyQuestions}}},
			{"two chart roots", []archiveEntry{chart, {name: "other/Chart.yaml", body: "name: other\nversion: 1.0.0\n"}}},
			{"missing metadata", []archiveEntry{{name: fixtureValuesPath, body: "a: 1\n"}}},
			{"metadata scalar", []archiveEntry{{name: fixtureChartMetadataPath, body: "plain"}}},
			{"duplicate YAML keys", []archiveEntry{{name: fixtureChartMetadataPath, body: "name: demo\nname: other\nversion: 1.0.0\n"}}},
			{"oversized values", []archiveEntry{chart, {name: fixtureValuesPath, body: "value: " + string(bytes.Repeat([]byte("x"), (1<<20)+1))}}},
		} {
			{
				By(tc.name)
				{
					_, err := ParseArchive(chartArchive(tc.entries...))
					Expect(err).NotTo(BeNil(), "unsafe chart accepted")
				}
			}
		}
	})
	It("Parse Archive Rejects Corrupt Gzip", func() {
		data := chartArchive(archiveEntry{name: fixtureChartMetadataPath, body: fixtureChartMetadataYAML})
		data[len(data)-8] ^= 0xff
		{
			_, err := ParseArchive(data)
			Expect(err).NotTo(BeNil(), "corrupt gzip trailer accepted")
		}
	})
	It("Parse Archive Ignores Subchart Metadata", func() {
		data := chartArchive(
			archiveEntry{name: fixtureChartMetadataPath, body: "name: root\nversion: 1.0.0\n"},
			archiveEntry{name: "demo/charts/child/Chart.yaml", body: "name: child\nversion: 2.0.0\n"},
			archiveEntry{name: "demo/charts/child/questions.yaml", body: "questions: [{variable: x, type: string}]\n"},
			archiveEntry{name: "demo/charts/child/questions.yml", body: "questions: [{variable: y, type: string}]\n"},
		)
		got, err := ParseArchive(data)
		Expect(err).NotTo(HaveOccurred())
		Expect(got == nil || got.ChartName != "root" || got.HasQuestions).To(BeFalse(), func() string {
			return fmt.Sprintf("subchart metadata used: %#v", got)
		})
	})
	It("Parse Archive Enforces Size And Entry Bounds", func() {
		chart := archiveEntry{name: fixtureChartMetadataPath, body: fixtureChartMetadataYAML}
		{
			By("compressed")
			{
				_, err := ParseArchive(make([]byte, (32<<20)+1))
				Expect(err).NotTo(BeNil(), "compressed bound ignored")
			}
		}

		{
			By("expanded")
			data := chartArchive(chart, archiveEntry{name: "demo/data", body: strings.Repeat("x", 65<<20)})
			{
				_, err := ParseArchive(data)
				Expect(err).NotTo(BeNil(), "expanded bound ignored")
			}
		}

		{
			By("entries")
			entries := []archiveEntry{chart}
			for n := 0; n < 4096; n++ {
				entries = append(entries, archiveEntry{name: "demo/unused"})
			}
			{
				_, err := ParseArchive(chartArchive(entries...))
				Expect(err).NotTo(BeNil(), "entry bound ignored")
			}
		}

		for _, filename := range []string{fixtureQuestionsFilename, fixtureYMLQuestionsFilename} {
			{
				By(filename)
				data := chartArchive(chart, archiveEntry{name: "demo/" + filename, body: strings.Repeat("x", (1<<20)+1)})
				{
					_, err := ParseArchive(data)
					Expect(err).NotTo(BeNil(), "questions bound ignored")
				}
			}
		}
	})
	It("Parse Archive Rejects Hidden Archive Suffixes", func() {
		first := chartArchive(archiveEntry{name: fixtureChartMetadataPath, body: fixtureChartMetadataYAML})
		second := chartArchive(archiveEntry{name: "other/Chart.yaml", body: "name: other\nversion: 2.0.0\n"})
		{
			By("gzip members")
			{
				_, err := ParseArchive(append(first, second...))
				Expect(err).NotTo(BeNil(), "concatenated gzip charts accepted")
			}
		}

		{
			By("tar suffix")
			var raw bytes.Buffer
			for _, data := range [][]byte{first, second} {
				gz, err := gzip.NewReader(bytes.NewReader(data))
				Expect(err).NotTo(HaveOccurred())
				{
					_, err := raw.ReadFrom(gz)
					Expect(err).NotTo(HaveOccurred())
				}
				{
					err := gz.Close()
					Expect(err).NotTo(HaveOccurred())
				}
			}
			var compressed bytes.Buffer
			gz := gzip.NewWriter(&compressed)
			{
				_, err := gz.Write(raw.Bytes())
				Expect(err).NotTo(HaveOccurred())
			}
			{
				err := gz.Close()
				Expect(err).NotTo(HaveOccurred())
			}
			{
				_, err := ParseArchive(compressed.Bytes())
				Expect(err).NotTo(BeNil(), "hidden second tar chart accepted")
			}
		}
	})
	It("Parse Archive Unreadable Defaults Do Not Fail Questions Inspection", func() {
		var manyNodes strings.Builder
		for index := 0; index < 50001; index++ {
			_, _ = fmt.Fprintf(&manyNodes, "key%d: value\n", index)
		}
		for _, values := range []struct{ name, body string }{
			{"malformed YAML", "replicas: [\n"},
			{"non-string key", "80: http\n"},
			{"duplicate keys", "replicas: 1\nreplicas: 2\n"},
			{"custom tag", "replicas: !custom 1\n"},
			{"node budget", manyNodes.String()},
			{"alias budget", "large: &large " + strings.Repeat("x", 32<<10) + "\ncopies: [" + strings.TrimSuffix(strings.Repeat("*large,", 130), ",") + "]\n"},
		} {
			for _, questions := range []struct {
				name, body string
				status     api.AiApplicationChartQuestionsStatus
			}{
				{"absent", "", api.AIAPPLICATIONCHARTQUESTIONSSTATUS_ABSENT},
				{"usable", "questions:\n- variable: replicas\n  type: int\n  default: 1\n  show_if: enabled=true\n  ignored_property: example\n", api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE},
				{"unusable", "questions:\n- variable: replicas\n  type: int\n  default: invalid\n", api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE},
			} {
				{
					By(values.name + "/" + questions.name)
					entries := []archiveEntry{
						{name: fixtureChartMetadataPath, body: fixtureChartMetadataYAML},
						{name: fixtureValuesPath, body: values.body},
					}
					if questions.body != "" {
						entries = append(entries, archiveEntry{name: "demo/questions.yaml", body: questions.body})
					}
					got, err := ParseArchive(chartArchive(entries...))
					Expect(err).To(BeNil(), func() string {
						return fmt.Sprintf("advisory defaults failed inspection: %v", err)
					})
					Expect(got.ChartName != "demo" || got.QuestionsStatus != questions.status || got.HasQuestions != (questions.body != "") || !hasWarning(got, "UNREADABLE_DEFAULTS")).To(BeFalse(), func() string {
						return fmt.Sprintf("wrong inspection outcome: %#v", got)
					})
					Expect(hasWarning(got, "MISSING_DEFAULT_PATH") || hasWarning(got, "UNKNOWN_CONDITION_REFERENCE")).To(BeFalse(), func() string {
						return fmt.Sprintf("unreadable defaults caused speculative path warnings: %#v", got.Warnings)
					})
					Expect(questions.status == api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE && (len(got.Questions) != 1 || !hasWarning(got, "IGNORED_PROPERTY"))).To(BeFalse(), func() string {
						return fmt.Sprintf("usable question diagnostics were lost: %#v", got)
					})
					Expect(questions.status == api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE && (len(got.Questions) != 0 || len(got.Groups) != 0 || !hasWarning(got, "UNUSABLE_QUESTIONS"))).To(BeFalse(), func() string {
						return fmt.Sprintf("invalid question defaults were accepted: %#v", got)
					})
				}
			}
		}
	})
})
