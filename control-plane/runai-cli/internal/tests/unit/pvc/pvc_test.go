/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/pvc"
	"gopkg.in/yaml.v3"
	"k8s.io/utils/ptr"

	openapi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI PVC Tests")
}

const (
	TestFilePath = "pvc"
)

var _ = Describe("Workload Test", Ordered, func() {
	var (
		configPath string
		err        error
		envMap     map[string]string
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, TestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		clients.MockPvc(mockCtrl, generateProjectPvcList())
	})

	BeforeEach(func() {
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
	})

	Context("pvc command validation", Ordered, func() {
		It("pvc list - missing project", func() {
			output, err := cli.RunCommand("pvc list", envMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("value is empty"))
			Expect(output).To(HavePrefix("Error: project is missing"))
		})

		It("pvc list project p301", func() {
			output, err := cli.RunCommand("pvc list --project=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("exists      No"))
			Expect(output).To(ContainSubstring("assets      Yes"))
		})

		It("pvc list project p301 json", func() {
			output, err := cli.RunCommand("pvc list --project=p301 --json", envMap)
			Expect(err).To(BeNil())
			var data pvc.Data
			err = json.Unmarshal([]byte(output), &data)
			Expect(err).To(BeNil())
			pvcList := data.Pvcs
			Expect(pvcList).To(HaveLen(3))
			Expect(pvcList).To(ContainElement(pvc.Brief{
				Name:    "exists",
				Managed: ptr.To(false),
			}))
			Expect(pvcList).To(ContainElement(pvc.Brief{
				Name:    "assets",
				Managed: ptr.To(true),
			}))
			Expect(pvcList).To(ContainElement(pvc.Brief{
				Name:    "notfound",
				Managed: nil,
			}))
		})

		It("pvc list project p301 json", func() {
			output, err := cli.RunCommand("pvc list --project=p301 --yaml", envMap)
			Expect(err).To(BeNil())
			var data pvc.Data
			err = yaml.Unmarshal([]byte(output), &data)
			pvcList := data.Pvcs
			Expect(err).To(BeNil())
			Expect(pvcList).To(HaveLen(3))
			Expect(pvcList).To(ContainElement(pvc.Brief{
				Name:    "exists",
				Managed: ptr.To(false),
			}))
			Expect(pvcList).To(ContainElement(pvc.Brief{
				Name:    "assets",
				Managed: ptr.To(true),
			}))
			Expect(pvcList).To(ContainElement(pvc.Brief{
				Name:    "notfound",
				Managed: nil,
			}))
		})

		It("pvc list project p302 - err", func() {
			output, err := cli.RunCommand("pvc list --project=p302", envMap)
			Expect(err).To(HaveOccurred())
			Expect(output).To(ContainSubstring("Error: PVC list failed. list API call error: generic error"))
		})
	})
})

func generateProjectPvcList() map[string]clients.MockPvcResponse {
	var projPvcMap = make(map[string]clients.MockPvcResponse)
	existing, _ := openapi.NewOriginFromValue("existing")
	assets, _ := openapi.NewOriginFromValue("assets")

	proj1 := &openapi.PVCListResponse{
		Pvcs: []openapi.PVCListInfo{{
			ClusterId: "cluster-proj-1",
			ScopeType: "project",
			ScopeId:   clients.ProjectOneID,
			Name:      "exists",
			Origin:    *openapi.NewNullableOrigin(existing),
		},
			{
				ClusterId: "cluster-proj-1",
				ScopeType: "cluster",
				ScopeId:   "cluster-proj-1-id",
				Name:      "assets",
				Origin:    *openapi.NewNullableOrigin(assets),
			},
			{
				ClusterId: "cluster-proj-1",
				ScopeType: "cluster",
				ScopeId:   "cluster-proj-1-id",
				Name:      "notfound",
			}},
	}

	projPvcMap[clients.ProjectOneID] = clients.MockPvcResponse{
		Response: proj1,
		Err:      nil,
	}
	projPvcMap[clients.ProjectTwoID] = clients.MockPvcResponse{Response: nil, Err: fmt.Errorf("generic error")}
	return projPvcMap
}
