// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package chart_inspection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func inspectQuestions(questions, values string) *api.AiApplicationChartInspectionResponse {
	GinkgoHelper()

	got, err := ParseArchive(chartArchive(
		archiveEntry{name: fixtureChartMetadataPath, body: fixtureChartMetadataYAML},
		archiveEntry{name: "demo/questions.yaml", body: questions},
		archiveEntry{name: fixtureValuesPath, body: values},
	))
	Expect(err).NotTo(HaveOccurred())

	return got
}

func hasWarning(got *api.AiApplicationChartInspectionResponse, code string) bool {
	for _, warning := range got.Warnings {
		if warning.Code == code {
			return true
		}
	}
	return false
}

var _ = Describe("Question declarations", func() {
	It("Declaration Normalizes Compatible Metadata", func() {
		got := inspectQuestions(`apiVersion: ignored/v1
groups:
  - name: Advanced
    label: Advanced settings
    description: Declared but unused
questions:
  - variable: enabled
    type: boolean
    default: "false"
    show_subquestions_if: true
    vendor_hint: ignored
    subquestions:
      - variable: workers[0].name
        label: Worker
        type: unfamiliar
  - variable: scale
    label: Scale
    type: number
    default: 1.25
`, "enabled: true\nworkers: [{name: original}]\nscale: 1\n")
		Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE || len(got.Questions) != 2).To(BeFalse(), func() string {
			return fmt.Sprintf("declaration = %#v", got)
		})
		Expect(got.Questions[0]["label"] != "enabled" || got.Questions[0]["default"] != false || got.Questions[0]["show_subquestion_if"] != true).To(BeFalse(), func() string {
			return fmt.Sprintf("normalization = %#v", got.Questions[0])
		})

		child := got.Questions[0]["subquestions"].([]any)[0].(map[string]any)
		Expect(child["type"] != "string" || got.Questions[1]["type"] != "float").To(BeFalse(), func() string {
			return fmt.Sprintf("type normalization = %#v", got.Questions)
		})
		{
			_, exists := got.Questions[0]["vendor_hint"]
			Expect(exists).To(BeFalse(),
				"unknown property retained",
			)
		}
		Expect(len(got.Groups) != 2 || got.Groups[0]["name"] != "advanced").To(BeFalse(), func() string {
			return fmt.Sprintf("declared groups not preserved: %#v", got.Groups)
		})

		for _, code := range []string{"LABEL_DEFAULTED", "UNKNOWN_TYPE", "IGNORED_PROPERTY"} {
			Expect(hasWarning(got, code)).To(BeTrue(), func() string {
				return fmt.Sprintf("missing %s warning: %#v", code, got.Warnings)
			})
		}
	})
	It("Declaration Accepts Supported Field Defaults", func() {
		for _, tc := range []struct{ kind, value, extra string }{
			{"string", `hello`, ""}, {"multiline", `hello`, ""}, {"password", `'<your-password>'`, ""},
			{"storageclass", `fast`, ""}, {"pvc", `data`, ""}, {"configmap", `config`, ""},
			{"int", `3`, "min: 1\n      max: 5"}, {"float", `1.5`, ""}, {"boolean", `true`, ""},
			{"enum", `[1, two]`, "options: [1, two]\n      multiple: true"},
			{"enum", `1`, "options: [1, two]"}, {"radio", `false`, "options: [false, true]"},
			{"hostname", `example.com`, ""}, {"ipaddr", `::1`, ""}, {"cidr", `10.1.2.3/24`, ""},
			{"cron", `'*/5 * * * *'`, ""}, {"array[string]", `[a, b]`, ""}, {"map[string]", `{a: b}`, ""},
		} {
			{
				By(tc.kind)
				body := fmt.Sprintf("questions:\n  - variable: setting\n    label: Setting\n    type: '%s'\n    default: %s\n", tc.kind, tc.value)
				if tc.extra != "" {
					body += "    " + strings.ReplaceAll(tc.extra, "\n      ", "\n    ") + "\n"
				}
				got := inspectQuestions(body, "setting: null\n")
				Expect(got.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE), func() string {
					return fmt.Sprintf("supported field rejected: %#v", got.Warnings)
				})
			}
		}
	})
	It("Declaration Rejects Whole Invalid Form", func() {
		for _, tc := range []struct{ name, body string }{
			{"malformed YAML", fixtureMalformedQuestions}, {"root array", "- variable: x\n"},
			{"duplicate keys", "questions: []\nquestions: []\n"}, {"missing questions", "groups: []\n"},
			{"missing variable", "questions: [{type: string}]"}, {"missing type", "questions: [{variable: x}]"},
			{"secret", "questions: [{variable: x, type: secret}]"}, {"credential", "questions: [{variable: x, type: cloud-credential}]"},
			{"numeric dotted path", "questions: [{variable: workers.0.name, type: string}]"},
			{"prototype path", "questions: [{variable: 'a[\"__proto__\"]', type: string}]"},
			{"duplicate destination", "questions: [{variable: x, type: string}, {variable: x, type: string}]"},
			{"overlapping destination", "questions: [{variable: x, type: string}, {variable: x.y, type: string}]"},
			{"bad default", "questions: [{variable: x, type: int, default: nope}]"},
			{"default below bound", "questions: [{variable: x, type: int, min: 2, default: 1}]"},
			{"bad array default", "questions: [{variable: x, type: 'array[string]', default: [one, 2]}]"},
			{"bad IP", "questions: [{variable: x, type: ipaddr, default: 999.1.1.1}]"},
			{"bad hostname", "questions: [{variable: x, type: hostname, default: a_b}]"},
			{"bad cron", "questions: [{variable: x, type: cron, default: '61 * * * *'}]"},
			{"missing enum options", "questions: [{variable: x, type: enum}]"},
			{"duplicate typed options", "questions: [{variable: x, type: enum, options: [1, 1.0]}]"},
			{"duplicate multiple default", "questions: [{variable: x, type: enum, options: [1, 2], multiple: true, default: [1, 1.0]}]"},
			{"unavailable default", "questions: [{variable: x, type: enum, options: [a, b], default: c}]"},
			{"bad option", "questions: [{variable: x, type: enum, options: [{value: a}]}]"},
			{"bad constraint type", "questions: [{variable: x, type: boolean, min: 1}]"},
			{"conflicting gates", "questions: [{variable: x, type: boolean, show_subquestion_if: true, show_subquestions_if: false}]"},
			{"bad condition", "questions: [{variable: x, type: string, show_if: 'y && true'}]"},
			{"condition cycle", "questions: [{variable: x, type: string, show_if: 'y=ready'}, {variable: y, type: string, show_if: 'x=ready'}]"},
			{"parent visibility cycle", "questions: [{variable: x, type: boolean, show_if: 'y=true', subquestions: [{variable: y, type: boolean}]}]"},
		} {
			{
				By(tc.name)
				got := inspectQuestions(tc.body, "{}\n")
				Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE || !got.HasQuestions || len(got.Questions) != 0 || len(got.Groups) != 0 || !hasWarning(got, "UNUSABLE_QUESTIONS")).To(BeFalse(), func() string {
					return fmt.Sprintf("invalid declaration was salvaged: %#v", got)
				})
			}
		}
	})
	It("Declaration Preserves Four Child Levels", func() {
		makeTree := func(last int) string {
			child := "{variable: p" + fmt.Sprint(last) + ", type: boolean, subquestions: []}"
			for n := last - 1; n >= 0; n-- {
				child = "{variable: p" + fmt.Sprint(n) + ", type: boolean, group: Basics, subquestions: [" + child + "]}"
			}
			return "questions: [" + child + "]"
		}
		got := inspectQuestions(makeTree(4), "{}")
		Expect(got.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE), func() string {
			return fmt.Sprintf("four levels rejected: %#v", got.Warnings)
		})

		child := got.Questions[0]
		for depth := 0; depth < 4; depth++ {
			Expect(child["group"]).To(Equal("basics"), func() string {
				return fmt.Sprintf("group not inherited at %d", depth)
			})

			child = child["subquestions"].([]any)[0].(map[string]any)
		}
		Expect(child["group"]).To(Equal("basics"), "leaf group not inherited")
		{
			got := inspectQuestions(makeTree(5), "{}")
			Expect(got.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE), func() string {
				return fmt.Sprintf("fifth child level accepted: %#v", got)
			})
		}
	})
	It("Declaration Warns For Dynamic Values And Conditions", func() {
		got := inspectQuestions("questions: [{variable: dynamic, type: string, show_if: 'name= || ready=true'}]", "ready: true\n")
		Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE || !hasWarning(got, "MISSING_DEFAULT_PATH") || !hasWarning(got, "UNKNOWN_CONDITION_REFERENCE")).To(BeFalse(), func() string {
			return fmt.Sprintf("dynamic input rejected or unreported: %#v", got)
		})
	})
	It("Declaration Rejects Null Groups And Complex Gates", func() {
		for _, body := range []string{
			"groups: null\nquestions: []",
			"questions: [{variable: x, type: boolean, show_subquestion_if: [true], show_subquestions_if: [true]}]",
			"questions: [{variable: x, type: boolean, subquestions: null}]",
			"questions: [{variable: x, type: boolean, default: 'True'}]",
		} {
			got := inspectQuestions(body, "{}")
			Expect(got.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE), func() string {
				return fmt.Sprintf("invalid structure accepted: %s", body)
			})
		}
	})
	It("Declaration Ignores Optional Option Metadata", func() {
		got := inspectQuestions("questions: [{variable: mode, type: enum, options: [{label: Basic, value: basic, description: extra}]}]", "mode: basic")
		Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE || !hasWarning(got, "IGNORED_PROPERTY")).To(BeFalse(), func() string {
			return fmt.Sprintf("optional option metadata rejected: %#v", got)
		})

		option := got.Questions[0]["options"].([]any)[0].(map[string]any)
		{
			_, exists := option["description"]
			Expect(exists).To(BeFalse(),
				"optional option metadata retained",
			)
		}
	})
	It("Declaration Warns When Whitespace Label Is Defaulted", func() {
		got := inspectQuestions("questions: [{variable: setting, label: '   ', type: string}]", "setting: original")
		Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE || len(got.Questions) != 1 || got.Questions[0]["label"] != "setting").To(BeFalse(), func() string {
			return fmt.Sprintf("whitespace label was not normalized: %#v", got)
		})

		for _, warning := range got.Warnings {
			if warning.Code == "LABEL_DEFAULTED" && warning.Path == "questions[0].label" {
				return
			}
		}
		Fail(fmt.Sprintf("normalized label has no warning: %#v", got.Warnings))
	})
	It("Declaration Rejects Incompatible Destination Containers", func() {
		for _, pair := range [][2]string{
			{fixtureWorkerNamePath, fixtureWorkerEnabledPath},
			{fixtureWorkerEnabledPath, fixtureWorkerNamePath},
			{"workers[0].slots[0]", "workers[0].slots.name"},
		} {
			got := inspectQuestions(fmt.Sprintf("questions: [{variable: '%s', type: string}, {variable: '%s', type: string}]", pair[0], pair[1]), "{}")
			Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE || len(got.Questions) != 0 || !hasWarning(got, "UNUSABLE_QUESTIONS")).To(BeFalse(), func() string {
				return fmt.Sprintf("incompatible containers accepted: %v: %#v", pair, got)
			})
		}
		for _, pair := range [][2]string{{fixtureWorkerNamePath, "workers[1].name"}, {"workers.name", fixtureWorkerEnabledPath}} {
			got := inspectQuestions(fmt.Sprintf("questions: [{variable: '%s', type: string}, {variable: '%s', type: string}]", pair[0], pair[1]), "{}")
			Expect(got.QuestionsStatus).To(Equal(api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE), func() string {
				return fmt.Sprintf("compatible containers rejected: %v: %#v", pair, got.Warnings)
			})
		}
	})
	It("Declaration Accepts Numeric Prefix Text Conditions", func() {
		for _, value := range []string{fixtureVersion, "1.2.3-beta", "5m"} {
			{
				By(value)
				body := fmt.Sprintf("questions: [{variable: setting, type: string, show_if: 'image.tag=%s'}]", value)
				got := inspectQuestions(body, "image: {tag: original}")
				Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE || got.Questions[0]["show_if"] != "image.tag="+value).To(BeFalse(), func() string {
					return fmt.Sprintf("unquoted text condition rejected or changed: %#v", got)
				})
			}
		}
	})
	It("Declaration Preserves Distinct Large Integer Options", func() {
		got := inspectQuestions("questions: [{variable: mode, type: enum, options: [9007199254740992, 9007199254740993], default: 9007199254740993}]", "mode: 9007199254740993")
		Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE || len(got.Questions) != 1).To(BeFalse(), func() string {
			return fmt.Sprintf("distinct integer options rejected: %#v", got)
		})

		encoded, err := json.Marshal(got.Questions[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(bytes.Contains(encoded, []byte(`"default":9007199254740993`))).To(BeTrue(), func() string {
			return fmt.Sprintf("large integer default rounded: %s", encoded)
		})
	})
	It("Declaration Preserves Integers Beyond Uint64", func() {
		got := inspectQuestions("questions: [{variable: mode, type: enum, options: [18446744073709551616, 18446744073709551617], default: 18446744073709551617}]", "mode: 18446744073709551617")
		Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_USABLE || len(got.Questions) != 1).To(BeFalse(), func() string {
			return fmt.Sprintf("distinct integers were rounded: %#v", got.Warnings)
		})

		encoded, err := json.Marshal(got.Questions[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(bytes.Contains(encoded, []byte(`"default":18446744073709551617`))).To(BeTrue(), func() string {
			return fmt.Sprintf("integer default rounded: %s", encoded)
		})
	})
	It("Declaration Rejects Exact Large Integer Bound Violations", func() {
		for _, extra := range []string{
			"max: 9007199254740992, default: 9007199254740993",
			"min: 9007199254740993, default: 9007199254740992",
			"min: 9007199254740993, max: 9007199254740992",
			"min: -9007199254740992, default: -9007199254740993",
			"max: 18446744073709551614, default: 18446744073709551615",
		} {
			got := inspectQuestions("questions: [{variable: number, type: int, "+extra+"}]", "{}")
			Expect(got.QuestionsStatus != api.AIAPPLICATIONCHARTQUESTIONSSTATUS_UNUSABLE || len(got.Questions) != 0).To(BeFalse(), func() string {
				return fmt.Sprintf("exact integer bound ignored: %s: %#v", extra, got)
			})
		}
	})
})
