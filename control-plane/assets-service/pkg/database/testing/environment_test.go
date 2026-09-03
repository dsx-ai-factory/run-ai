/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/scope"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var UidGidFromTheImage = api.UidGidSource("fromTheImage")

var _ = Describe("Environment Asset DAO", func() {
	It("serializes the asset correctly", func() {
		imagePullPolicy := api.ImagePullPolicy("Always")

		creationReq := api.EnvironmentCreationRequest{
			Spec: api.EnvironmentAssetSpec{
				EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
					{
						Name:  convutils.Ref("key1"),
						Value: convutils.Ref("value1"),
					},
					{
						Name:  convutils.Ref("key2"),
						Value: convutils.Ref("value2"),
					},
				},
				Connections: []api.Connection{
					{
						Name: "my-jupyter",
						InternalToolInfo: &api.InternalToolInfo{
							ContainerPort:  8888,
							ConnectionType: api.INTERNALCONNECTIONTYPE_NODE_PORT,
							ToolType:       api.INTERNALTOOLTYPE_JUPYTER_NOTEBOOK,
						},
					},
				},
				Args:            convutils.Ref("--arg1 --arg2"),
				Command:         convutils.Ref("command"),
				Image:           convutils.Ref("image"),
				ImagePullPolicy: &imagePullPolicy,
				WorkingDir:      convutils.Ref("/home/user"),
				HostIpc:         convutils.Ref(true),
				HostNetwork:     convutils.Ref(true),
				CreateHomeDir:   convutils.Ref(true),
				Tty:             convutils.Ref(true),
				Stdin:           convutils.Ref(true), AllowPrivilegeEscalation: convutils.Ref(true),
				UidGidSource: &UidGidFromTheImage,
				Capabilities: &[]api.Capability{api.CAPABILITY_IPC_LOCK, api.CAPABILITY_IPC_OWNER},
			},
			Meta: api.AssetCreationRequest{
				Name:  "bla",
				Scope: api.SCOPE_TENANT,
				WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
					Workspace:     convutils.Ref(true),
					Training:      convutils.Ref(false),
					Inference:     convutils.Ref(false),
					Distributed:   convutils.Ref(false),
					DistFramework: nil,
				},
			},
		}

		nameValidationQuery, _ := scope.GetNameValidationQuery(auth.MockAuthInfo(), creationReq.Meta.Name, creationReq.Meta.Scope, nil, nil, nil, creationReq.Meta.Private, api.ASSETKIND_ENVIRONMENT, nil)
		res, _ := dao.CreateEnvironmentAsset(auth.MockAuthInfo(), creationReq, nameValidationQuery)
		Expect(res.Meta.Kind).To(BeEquivalentTo("environment"))
		Expect(res.Spec).To(Equal(creationReq.Spec))
	})

	It("handles null values", func() {
		creationReq := api.EnvironmentCreationRequest{
			Spec: api.EnvironmentAssetSpec{},
			Meta: api.AssetCreationRequest{
				Name:  "bla2",
				Scope: api.SCOPE_TENANT,
				WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
					Workspace:     convutils.Ref(true),
					Training:      convutils.Ref(false),
					Inference:     convutils.Ref(false),
					Distributed:   convutils.Ref(false),
					DistFramework: nil,
				},
			},
		}

		nameValidationQuery, _ := scope.GetNameValidationQuery(auth.MockAuthInfo(), creationReq.Meta.Name, creationReq.Meta.Scope, nil, nil, nil, creationReq.Meta.Private, api.ASSETKIND_ENVIRONMENT, nil)
		res, _ := dao.CreateEnvironmentAsset(auth.MockAuthInfo(), creationReq, nameValidationQuery)
		Expect(res.Meta.Kind).To(BeEquivalentTo("environment"))
		Expect(res.Spec).To(Equal(creationReq.Spec))
	})
})
