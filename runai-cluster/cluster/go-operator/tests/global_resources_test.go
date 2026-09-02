/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"encoding/json"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	ccontroller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/controllers/run"
	controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/controllers/run"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/util"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
)

var _ = Describe("Global Sub-Resources", func() {
	Context("Resources", func() {
		It("should populate global sub-resources to relevant operands objects", func() {
			// Prepare
			// Maps a global resource name to it's test case
			cases := map[string]struct {
				// Operands of which their objects should not be populated with this global resources
				excludedOperands map[string]struct{}
				// Object kinds that should be populated with this global resources
				relevantObjects []client.Object
				// The expected resources that should be populated on the relevant objects
				expectedSubResources []interface{}
			}{
				"security-context": {
					excludedOperands: util.SliceToMap([]string{
						"cert-manager-webhook-dynu",
						"mps-server",
						"runai-scheduler",
					}),
					relevantObjects: []client.Object{
						&appsv1.Deployment{},
						&appsv1.StatefulSet{},
						&appsv1.DaemonSet{},
					},
					expectedSubResources: []interface{}{
						outils.GlobalSecurityContext(runaiConfig),
					},
				},
				"volumes": {
					excludedOperands: util.SliceToMap([]string{}),
					relevantObjects: []client.Object{
						&appsv1.Deployment{},
						&appsv1.StatefulSet{},
						&appsv1.DaemonSet{},
					},
					expectedSubResources: util.SliceToInterface(outils.GlobalVolumes(runaiConfig)),
				},
				"volume-mounts": {
					excludedOperands: util.SliceToMap([]string{}),
					relevantObjects: []client.Object{
						&appsv1.Deployment{},
						&appsv1.StatefulSet{},
						&appsv1.DaemonSet{},
					},
					expectedSubResources: util.SliceToInterface(outils.GlobalVolumes(runaiConfig)),
				},
			}

			// Act
			hybridClient := kc.New(k8s, k8s, nil)
			managedOperands := ccontroller.NewManagedOperands(apiCtx, hybridClient, hybridClient, runaiConfig)
			resources, err := controller.ResourcesForOperands(apiCtx,
				managedOperands.GetRemoteClusterEnabled())

			// Assert
			ExpectNoErr(err)
			Expect(resources.CreateOrUpdate).NotTo(BeEmpty())

			// For each case, check that the relevant objects are populated with the expected resources
			for _, c := range cases {
				for _, resource := range resources.CreateOrUpdate {
					if _, ok := c.excludedOperands[resource.OperandName()]; ok {
						continue
					}

					object, err := resource.Object(context.Background())
					ExpectNoErr(err)

					if !isRelevantObject(object, c.relevantObjects) {
						continue
					}

					for _, expectedSubResource := range c.expectedSubResources {
						expectedSubResourceJson, err := json.Marshal(expectedSubResource)
						ExpectNoErr(err)
						objectJson, err := json.Marshal(object)
						ExpectNoErr(err)

						Expect(objectJson).To(ContainSubstring(string(expectedSubResourceJson)))
					}
				}
			}

		})
	})
})

func isRelevantObject(object client.Object, relevantObjects []client.Object) bool {
	for _, relevantObject := range relevantObjects {
		if reflect.TypeOf(object) == reflect.TypeOf(relevantObject) {
			return true
		}
	}
	return false
}
