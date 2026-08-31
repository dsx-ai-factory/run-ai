/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/application/list"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// tableHeader is the fixed header row produced by the default 8-column layout.
const tableHeader = " AI Application        Type             Status           Project          GPU Compute Req     GPU Compute Alloc   GPU Memory Req      GPU Memory Alloc   \n─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────"

func buildApp(id, name, appType, project, status string) openapi.AIApplication {
	app := openapi.AIApplication{
		Id:          ptr.To(id),
		Name:        ptr.To(name),
		Type:        ptr.To(appType),
		ProjectName: ptr.To(project),
		CreatedAt:   func() *time.Time { t := time.Unix(1, 0); return &t }(),
	}
	app.ActionStatus.Set(&status)
	return app
}

var _ = Describe("AIApplication Display", func() {

	Context("Table output", func() {
		It("single app matches exact table layout", func() {
			out, err := list.NewOutput([]openapi.AIApplication{
				buildApp("id-1", "app-alpha", "inference", "proj-a", "Running"),
			}, nil)
			Expect(err).ToNot(HaveOccurred())

			got, err := out.GenerateTable()
			Expect(err).ToNot(HaveOccurred())

			expected := tableHeader + "\n app-alpha             inference        Running          proj-a           -                   -                   -                   -                  "
			Expect(got).To(Equal(expected))
		})

		It("two apps with next token appended after table", func() {
			createdAt := time.Unix(10, 0)
			tok, err := structs.NextTokenFromTimestamp(createdAt)
			Expect(err).ToNot(HaveOccurred())
			encoded, err := tok.Encode()
			Expect(err).ToNot(HaveOccurred())

			apps := []openapi.AIApplication{
				buildApp("id-1", "app-alpha", "inference", "proj-a", "Running"),
				buildApp("id-2", "app-beta", "training", "proj-b", "Pending"),
			}
			apps[1].CreatedAt = &createdAt

			out, err := list.NewOutput(apps, tok)
			Expect(err).ToNot(HaveOccurred())

			got, err := out.GenerateTable()
			Expect(err).ToNot(HaveOccurred())

			expected := tableHeader +
				"\n app-alpha             inference        Running          proj-a           -                   -                   -                   -                  " +
				"\n app-beta              training         Pending          proj-b           -                   -                   -                   -                  " +
				fmt.Sprintf("\r\n\n next token: %s", encoded)
			Expect(got).To(Equal(expected))
		})

		It("empty app list produces only the header", func() {
			out, err := list.NewOutput([]openapi.AIApplication{}, nil)
			Expect(err).ToNot(HaveOccurred())
			got, err := out.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(tableHeader + "\n"))
		})

		It("custom columns affect header and hide others", func() {
			out, err := list.NewOutput([]openapi.AIApplication{
				buildApp("id-1", "my-app", "training", "team-x", "Pending"),
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out.SetColumns([]string{"name", "status"})).To(Succeed())

			got, err := out.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(ContainSubstring("AI Application"))
			Expect(got).To(ContainSubstring("Status"))
			Expect(got).ToNot(ContainSubstring("Type"))
			Expect(got).ToNot(ContainSubstring("Project"))
		})

		It("invalid column name returns error", func() {
			out, err := list.NewOutput([]openapi.AIApplication{}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out.SetColumns([]string{"invalid-column"})).To(MatchError(ContainSubstring("invalid-column")))
		})
	})

	Context("JSON output", func() {
		It("single app produces exact JSON", func() {
			out, err := list.NewOutput([]openapi.AIApplication{
				buildApp("id-42", "exact-app", "inference", "proj-exact", "Running"),
			}, nil)
			Expect(err).ToNot(HaveOccurred())

			got, err := out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			expected := `{
  "aiApplications": [
    {
      "name": "exact-app",
      "type": "inference",
      "actionStatus": "Running",
      "projectName": "proj-exact"
    }
  ]
}`
			Expect(got).To(Equal(expected))
		})

		It("two apps with next token produce exact JSON", func() {
			createdAt := time.Unix(10, 0)
			tok, err := structs.NextTokenFromTimestamp(createdAt)
			Expect(err).ToNot(HaveOccurred())
			encoded, err := tok.Encode()
			Expect(err).ToNot(HaveOccurred())

			apps := []openapi.AIApplication{
				buildApp("id-1", "app-alpha", "inference", "proj-a", "Running"),
				buildApp("id-2", "app-beta", "training", "proj-b", "Pending"),
			}

			out, err := list.NewOutput(apps, tok)
			Expect(err).ToNot(HaveOccurred())

			got, err := out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			expected := fmt.Sprintf(`{
  "aiApplications": [
    {
      "name": "app-alpha",
      "type": "inference",
      "actionStatus": "Running",
      "projectName": "proj-a"
    },
    {
      "name": "app-beta",
      "type": "training",
      "actionStatus": "Pending",
      "projectName": "proj-b"
    }
  ],
  "nextToken": %q
}`, encoded)
			Expect(got).To(Equal(expected))
		})

		It("empty app list produces empty JSON object", func() {
			out, err := list.NewOutput([]openapi.AIApplication{}, nil)
			Expect(err).ToNot(HaveOccurred())
			got, err := out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal(`{}`))
		})
	})

	Context("YAML output", func() {
		It("single app produces exact YAML", func() {
			out, err := list.NewOutput([]openapi.AIApplication{
				buildApp("id-99", "yaml-app", "training", "proj-yaml", "Pending"),
			}, nil)
			Expect(err).ToNot(HaveOccurred())

			got, err := out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())

			expected := "aiApplications:\n    - name: yaml-app\n      type: training\n      actionStatus: Pending\n      projectName: proj-yaml\n"
			Expect(got).To(Equal(expected))
		})

		It("two apps with next token produce exact YAML", func() {
			createdAt := time.Unix(10, 0)
			tok, err := structs.NextTokenFromTimestamp(createdAt)
			Expect(err).ToNot(HaveOccurred())
			encoded, err := tok.Encode()
			Expect(err).ToNot(HaveOccurred())

			apps := []openapi.AIApplication{
				buildApp("id-1", "app-alpha", "inference", "proj-a", "Running"),
				buildApp("id-2", "app-beta", "training", "proj-b", "Pending"),
			}

			out, err := list.NewOutput(apps, tok)
			Expect(err).ToNot(HaveOccurred())

			got, err := out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())

			expected := fmt.Sprintf("aiApplications:\n    - name: app-alpha\n      type: inference\n      actionStatus: Running\n      projectName: proj-a\n    - name: app-beta\n      type: training\n      actionStatus: Pending\n      projectName: proj-b\nnextToken: %s\n", encoded)
			Expect(got).To(Equal(expected))
		})

		It("empty app list produces YAML with empty array", func() {
			out, err := list.NewOutput([]openapi.AIApplication{}, nil)
			Expect(err).ToNot(HaveOccurred())
			got, err := out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(Equal("aiApplications: []\n"))
		})
	})

	Context("IsEmpty", func() {
		It("returns true for empty output", func() {
			out, err := list.NewOutput([]openapi.AIApplication{}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out.IsEmpty()).To(BeTrue())
		})

		It("returns false for non-empty output", func() {
			out, err := list.NewOutput([]openapi.AIApplication{
				buildApp("1", "app-one", "inference", "proj-a", "Running"),
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(out.IsEmpty()).To(BeFalse())
		})
	})
})
