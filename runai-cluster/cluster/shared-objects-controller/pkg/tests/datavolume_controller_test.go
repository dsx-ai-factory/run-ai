/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/controller/datavolume"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/controller/datavolume/handlers"
	"k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSharedObjectsController(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msgf("Started service tests")
	RegisterFailHandler(Fail)
	RunSpecs(t, "shared-objects-controller Test Suite")
}

const (
	testProject1 = "team-a"
	testProject2 = "team-a-2"
	testProject3 = "team-a-3"
	testProject4 = "team-b"

	testDefaultDepartment = "default"
	testDepartment2       = "dep-a"

	pvcName1 = "pvc1"
	pvcName2 = "pvc2"
	pvcName3 = "pvc3"
	pvcName4 = "pvc4"

	pvName1 = "pv1"
	pvName2 = "pv2"
	pvName3 = "pv3"
	pvName4 = "pv4"
)

var (
	testNamespace1          = fmt.Sprintf("runai-%s", testProject1)
	testNamespace2          = fmt.Sprintf("%s-ns", testProject2) // to test non "runai-" prefixed ns
	testNamespace3          = fmt.Sprintf("runai-%s", testProject3)
	testNamespace4          = fmt.Sprintf("runai-%s", testProject4)
	testNamespaceNonProject = "some-ns"

	readOnlyMany = v1.PersistentVolumeAccessMode("ReadOnlyMany")
	// mapping each project to its namespace
	//nolint:unused
	_testProjectsToNamespace = map[string]string{
		testProject1: testNamespace1,
		testProject2: testNamespace2,
		testProject3: testNamespace3,
		testProject4: testNamespace4}
	// mapping each namespace to its project
	testNamespacesToProject = map[string]string{
		testNamespace1: testProject1,
		testNamespace2: testProject2,
		testNamespace3: testProject3,
		testNamespace4: testProject4}
	//nolint:unused
	_testDepartments = []string{testDefaultDepartment, testDepartment2}
	// mapping each project to its department
	testProjectsToDepartment = map[string]string{
		testProject1: testDepartment2,
		testProject2: testDepartment2,
		testProject3: testDepartment2,
		testProject4: testDefaultDepartment,
	}

	ns1          = buildNamespace(testNamespace1, testProject1)
	ns2          = buildNamespace(testNamespace2, testProject2)
	ns3          = buildNamespace(testNamespace3, testProject3)
	ns4          = buildNamespace(testNamespace4, testProject4)
	nsNonProject = buildNamespace(testNamespaceNonProject, "")
	pr1          = buildKaiProject(testProject1, testProjectsToDepartment[testProject1])
	pr2          = buildKaiProject(testProject2, testProjectsToDepartment[testProject2])
	pr3          = buildKaiProject(testProject3, testProjectsToDepartment[testProject3])
	pr4          = buildKaiProject(testProject4, testProjectsToDepartment[testProject4])
	defaultDep   = buildKaiDepartment(testDefaultDepartment)
	dep2         = buildKaiDepartment(testDepartment2)
)

type DatavolumeTestData struct {
	testName string

	// if you want some projects, departments, datavolumes, pv, pvc to be created before the test case..
	objectsToAddBeforeTestCase []client.Object
	dvOriginalPvc              *v1.PersistentVolumeClaim
	dvOriginalPv               *v1.PersistentVolume

	dvToReconcile                      *v1alpha1.DataVolume
	updateDvSharedScopesAfterReconcile []v1alpha1.DataVolumeSharedScope
	expectedDvStatus                   *v1alpha1.DataVolumeStatus

	shouldSucceed *bool

	// the following 2 are for dv-owned pvcs and pvs
	expectedNamespacesWherePvcAndPvShouldExist    map[string]bool
	expectedNamespacesWherePvcAndPvShouldNotExist []string

	// the following 2 are for pvc and pv that should exist, even if not owned by dv
	expectedPvcThatShouldExist []types.NamespacedName
	expectedPvThatShouldExist  []string
}

var _ = Describe("Shared Objects Controller Tests", Ordered, func() {
	var (
		datavolumeController *controller.DataVolumeController
		pvcHandler           *handlers.PvcHandler
		pvHandler            *handlers.PvHandler

		watchClient client.WithWatch

		ctx     context.Context
		cancel  context.CancelFunc
		stopper chan struct{}

		testCreatedObjects = []client.Object{}
	)

	BeforeAll(func() {
		scheme := getTestScheme()
		// Wrap the fake client so that fetched objects keep their TypeMeta populated
		// (the upgraded fake client strips it); the reconciler stamps owner refs from
		// DataVolume.Kind, so this restores the pre-upgrade / real-cluster behavior.
		watchClient = newGvkRestoringClient(fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&v1alpha1.DataVolume{}).
			WithObjects(nsNonProject, pr1, pr2, pr3, pr4, defaultDep, dep2, ns1, ns2, ns3, ns4).
			WithIndex(&kaiorg.Project{}, datavolume.KaiProjectsParentFieldIndexName, controller.ProjectsParentIndexer).
			Build(), scheme)

		datavolumeController = controller.NewDataVolumeController(watchClient)
		pvcHandler = handlers.NewPvcHandler(watchClient)
		pvHandler = handlers.NewPvHandler(watchClient)

		stopper = make(chan struct{}, 1)
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			close(stopper)
			cancel()
		})
	})

	AfterEach(func() {
		for _, testCreatedObject := range testCreatedObjects {
			err := watchClient.Delete(ctx, testCreatedObject)
			if err != nil {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}
		testCreatedObjects = []client.Object{}

		datavolumes := &v1alpha1.DataVolumeList{}
		err := watchClient.List(ctx, datavolumes)
		Expect(err).NotTo(HaveOccurred())
		for _, dv := range datavolumes.Items {
			dv.SetFinalizers(nil)
			err = watchClient.Update(ctx, &dv)
			if err != nil {
				Expect(errors.IsNotFound(err)).To(BeTrue())
				continue
			}
			err = watchClient.Delete(ctx, &dv)
			if err != nil {
				Expect(errors.IsNotFound(err)).To(BeTrue())
			}
		}

		pvcs := &v1.PersistentVolumeClaimList{}
		err = watchClient.List(ctx, pvcs)
		Expect(err).NotTo(HaveOccurred())
		for _, pvc := range pvcs.Items {
			err = watchClient.Delete(ctx, &pvc)
			Expect(err).NotTo(HaveOccurred())
		}

		pvs := &v1.PersistentVolumeList{}
		err = watchClient.List(ctx, pvs)
		Expect(err).NotTo(HaveOccurred())
		for _, pv := range pvs.Items {
			err = watchClient.Delete(ctx, &pv)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("cleans datavolume labels and owned resources on deletion", func() {
		originalPvc := buildPvc(dv1, nil)
		originalPvc.Labels["keep-this-label"] = "true"
		originalPv := buildPv(dv1, nil)
		originalPv.Labels["keep-this-label"] = "true"
		dvToReconcile := buildDatavolumeWithShouldUseOriginalVolume(
			buildDatavolume(dv1, []string{pr2.Name}, []string{}, ""),
			[]string{pr2.Name},
			[]string{},
			"",
			true,
		)

		Expect(watchClient.Create(ctx, originalPvc)).To(Succeed())
		Expect(watchClient.Create(ctx, originalPv)).To(Succeed())
		Expect(watchClient.Create(ctx, dvToReconcile)).To(Succeed())
		testCreatedObjects = append(testCreatedObjects, originalPvc, originalPv, dvToReconcile)

		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(dvToReconcile)}
		_, err := datavolumeController.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())

		dvAfterReconcile := getDvFromClient(ctx, dvToReconcile.Name, watchClient)
		Expect(dvAfterReconcile).NotTo(BeNil())
		Expect(dvAfterReconcile.Finalizers).To(ContainElement(datavolume.FinalizerName))

		pvcAfterReconcile := _getPvcFromClient(ctx, originalPvc.Name, originalPvc.Namespace, watchClient)
		Expect(pvcAfterReconcile).NotTo(BeNil())
		Expect(pvcAfterReconcile.Labels).To(HaveKeyWithValue(v1alpha1.DataVolumeIdLabel, dvToReconcile.Spec.DatavolumeId))
		Expect(pvcAfterReconcile.Labels).To(HaveKeyWithValue(v1alpha1.DataVolumeNameLabel, dvToReconcile.Name))

		pvAfterReconcile := _getPvFromClient(ctx, originalPv.Name, watchClient)
		Expect(pvAfterReconcile).NotTo(BeNil())
		Expect(pvAfterReconcile.Labels).To(HaveKeyWithValue(v1alpha1.DataVolumeIdLabel, dvToReconcile.Spec.DatavolumeId))
		Expect(pvAfterReconcile.Labels).To(HaveKeyWithValue(v1alpha1.DataVolumeNameLabel, dvToReconcile.Name))
		Expect(pvAfterReconcile.Labels).To(HaveKeyWithValue(v1alpha1.DataVolumePvForNamespaceLabel, originalPvc.Namespace))

		pvcsBeforeDeletion, err := pvcHandler.ListPvcsOfDatavolume(ctx, dvAfterReconcile)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvcsBeforeDeletion).To(HaveLen(2))
		pvsBeforeDeletion, err := pvHandler.ListPvsOfDatavolume(ctx, dvAfterReconcile)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvsBeforeDeletion).To(HaveLen(2))

		Expect(watchClient.Delete(ctx, dvAfterReconcile)).To(Succeed())

		_, err = datavolumeController.Reconcile(ctx, req)
		Expect(err).NotTo(HaveOccurred())

		pvcAfterDeletion := _getPvcFromClient(ctx, originalPvc.Name, originalPvc.Namespace, watchClient)
		Expect(pvcAfterDeletion).NotTo(BeNil())
		Expect(pvcAfterDeletion.Labels).NotTo(HaveKey(v1alpha1.DataVolumeIdLabel))
		Expect(pvcAfterDeletion.Labels).NotTo(HaveKey(v1alpha1.DataVolumeNameLabel))
		Expect(pvcAfterDeletion.Labels).To(HaveKeyWithValue("keep-this-label", "true"))

		pvAfterDeletion := _getPvFromClient(ctx, originalPv.Name, watchClient)
		Expect(pvAfterDeletion).NotTo(BeNil())
		Expect(pvAfterDeletion.Labels).NotTo(HaveKey(v1alpha1.DataVolumeIdLabel))
		Expect(pvAfterDeletion.Labels).NotTo(HaveKey(v1alpha1.DataVolumeNameLabel))
		Expect(pvAfterDeletion.Labels).NotTo(HaveKey(v1alpha1.DataVolumePvForNamespaceLabel))
		Expect(pvAfterDeletion.Labels).To(HaveKeyWithValue("keep-this-label", "true"))

		pvcsAfterDeletion, err := pvcHandler.ListPvcsOfDatavolume(ctx, dvAfterReconcile)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvcsAfterDeletion).To(HaveLen(0))
		pvsAfterDeletion, err := pvHandler.ListPvsOfDatavolume(ctx, dvAfterReconcile)
		Expect(err).NotTo(HaveOccurred())
		Expect(pvsAfterDeletion).To(HaveLen(0))
	})

	It("does not create PVs when a target PVC name is occupied", func() {
		originalPvc := buildPvc(dv1, nil)
		originalPv := buildPv(dv1, nil)
		foreignPvc := buildPvcInner2(pvcName1, testNamespace2, "foreign-pv", nil)
		stalePv := buildPvOfDvForNs(dv1, "stale-pv", pvcName1, testNamespace3, nil)
		dvToReconcile := buildDatavolumeWithShouldUseOriginalVolume(
			buildDatavolume(dv1, []string{pr2.Name}, []string{}, ""),
			[]string{pr2.Name},
			[]string{},
			"",
			true,
		)

		for _, object := range []client.Object{originalPvc, originalPv, foreignPvc, stalePv, dvToReconcile} {
			Expect(watchClient.Create(ctx, object)).To(Succeed())
			testCreatedObjects = append(testCreatedObjects, object)
		}

		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(dvToReconcile)}
		for range 2 {
			_, err := datavolumeController.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
		}

		dvAfterReconcile := getDvFromClient(ctx, dvToReconcile.Name, watchClient)
		Expect(dvAfterReconcile.Status.Phase).To(Equal(v1alpha1.DataVolumeNotReady))
		Expect(dvAfterReconcile.Status.Message).To(Equal(string(datavolume.FailedToCreatePvcsPvs)))

		pvs := &v1.PersistentVolumeList{}
		Expect(watchClient.List(ctx, pvs)).To(Succeed())
		for i := range pvs.Items {
			claimRef := pvs.Items[i].Spec.ClaimRef
			Expect(claimRef == nil || claimRef.Namespace != testNamespace2).To(BeTrue(),
				"unexpected PV %s was created for the conflicting PVC", pvs.Items[i].Name)
		}
		Expect(_getPvFromClient(ctx, stalePv.Name, watchClient)).To(BeNil(),
			"stale PV cleanup should run even when PVC creation fails")
	})

	DescribeTable("Sanity - externalworkload exists with correct info for all supported types",
		func(testData *DatavolumeTestData) {
			// create objects before the test case
			allObjectsToCreate := append(testData.objectsToAddBeforeTestCase, testData.dvToReconcile)
			if testData.dvOriginalPvc != nil {
				allObjectsToCreate = append(allObjectsToCreate, testData.dvOriginalPvc)
			}
			if testData.dvOriginalPv != nil {
				allObjectsToCreate = append(allObjectsToCreate, testData.dvOriginalPv)
			}
			for _, obj := range allObjectsToCreate {
				Expect(watchClient.Create(ctx, obj)).To(Succeed())
				testCreatedObjects = append(testCreatedObjects, obj)
			}

			// Reoncile
			key := client.ObjectKeyFromObject(testData.dvToReconcile)
			req := ctrl.Request{NamespacedName: key}

			_, err := datavolumeController.Reconcile(ctx, req)

			// update the dv after reconcile with the requested scopes
			if len(testData.updateDvSharedScopesAfterReconcile) > 0 {
				dvAfterReconcile := getDvFromClient(ctx, key.Name, watchClient)
				Expect(dvAfterReconcile).NotTo(BeNil())

				dvAfterReconcile.Spec.SharedScopes = testData.updateDvSharedScopesAfterReconcile
				err = watchClient.Update(ctx, dvAfterReconcile)
				Expect(err).NotTo(HaveOccurred())

				// reconcile again!
				_, err = datavolumeController.Reconcile(ctx, req)
			}

			// validate success
			if testData.shouldSucceed != nil && !*testData.shouldSucceed {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}

			dvAfterReconcile := getDvFromClient(ctx, key.Name, watchClient)
			Expect(dvAfterReconcile).NotTo(BeNil())

			// check the status of the object
			validateStatus(testData.testName, dvAfterReconcile, testData.expectedDvStatus)

			// validate existing/non existing objects
			dvPvcs, err := pvcHandler.ListPvcsOfDatavolume(ctx, dvAfterReconcile)
			Expect(err).NotTo(HaveOccurred())
			dvPvs, err := pvHandler.ListPvsOfDatavolume(ctx, dvAfterReconcile)
			Expect(err).NotTo(HaveOccurred())

			dvPvcsMap := convertPvcsToNamespaceToPvcMap(dvPvcs)
			dvPvsMap := convertPvsToPvMap(dvPvs)

			Expect(len(dvPvcsMap)).To(Equal(len(dvPvcs)), "Test: %s, there are two pvcs in same namespace", testData.testName)
			Expect(len(dvPvsMap)).To(Equal(len(dvPvs)), "Test: %s, there are two pvs for same namespace", testData.testName)
			Expect(len(dvPvcsMap)).To(Equal(len(dvPvsMap)), "Test: %s, dvPvcsMap and dvPvsMap; %v // %v", testData.testName, dvPvcsMap, dvPvsMap)
			Expect(len(dvPvcsMap)).To(Equal(len(testData.expectedNamespacesWherePvcAndPvShouldExist)),
				"Test: %s, dvPvcsMap and expectedExistingPvcAndPvs; %v // %v",
				testData.testName, dvPvcsMap, testData.expectedNamespacesWherePvcAndPvShouldExist)

			for ns, pvc := range dvPvcsMap {
				_, ok := testData.expectedNamespacesWherePvcAndPvShouldExist[ns]
				Expect(ok).To(BeTrue(), "Test: %s, pvc %s/%s", testData.testName, pvc.Namespace, pvc.Name)

				pv, ok := dvPvsMap[ns]
				Expect(ok).To(BeTrue(), "Test: %s, pv %s", testData.testName, ns)

				validatePvcOfDatavolume(testData.testName, dvAfterReconcile, pvc, testData.dvOriginalPvc, pv.Name)
				validatePvOfDatavolume(testData.testName, dvAfterReconcile, pv, testData.dvOriginalPv, pvc.Name, pvc.Namespace)

				testCreatedObjects = append(testCreatedObjects, pvc, pv)
			}

			for _, ns := range testData.expectedNamespacesWherePvcAndPvShouldNotExist {
				pvcs, err := pvcHandler.ListPvcsInNamespace(ctx, ns, client.HasLabels{v1alpha1.DataVolumeNameLabel})
				Expect(err).NotTo(HaveOccurred())
				Expect(pvcs).To(HaveLen(0), "Test: %s, pvcs (by name label) in namespace %s", testData.testName, ns)

				pvcs, err = pvcHandler.ListPvcsInNamespace(ctx, ns, client.HasLabels{v1alpha1.DataVolumeIdLabel})
				Expect(err).NotTo(HaveOccurred())
				Expect(pvcs).To(HaveLen(0), "Test: %s, pvcs (by id label)  in namespace %s", testData.testName, ns)

				pvs, err := pvHandler.GetAllDatavolumePvsForNamespace(ctx, ns)
				Expect(err).NotTo(HaveOccurred())
				Expect(pvs).To(HaveLen(0), "Test: %s, pvs in namespace %s", testData.testName, ns)
			}

			for _, nsName := range testData.expectedPvcThatShouldExist {
				pvc := &v1.PersistentVolumeClaim{}
				err := watchClient.Get(ctx, nsName, pvc)
				Expect(err).NotTo(HaveOccurred())
			}
			for _, pvName := range testData.expectedPvThatShouldExist {
				pv := &v1.PersistentVolume{}
				err := watchClient.Get(ctx, client.ObjectKey{Name: pvName}, pv)
				Expect(err).NotTo(HaveOccurred())
			}
		},
		Entry("Sanity - dv with project scope", &DatavolumeTestData{
			testName:         "Sanity - dv with project scope",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{pr2.Name}, []string{}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace3, testNamespace4, testNamespaceNonProject},
		}),
		Entry("Sanity - dv with department scope", &DatavolumeTestData{
			testName:         "Sanity - dv with department scope",
			dvOriginalPvc:    buildPvc(dv2, nil),
			dvOriginalPv:     buildPv(dv2, nil),
			dvToReconcile:    buildDatavolume2(dv2, true, []string{}, []string{dep2.Name}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName2, pvName2),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject},
		}),
		Entry("Sanity - dv with cluster scope", &DatavolumeTestData{
			testName:         "Sanity - dv with cluster scope",
			dvOriginalPvc:    buildPvc(dv3, nil),
			dvOriginalPv:     buildPv(dv3, nil),
			dvToReconcile:    buildDatavolume2(dv3, true, []string{}, []string{}, "cluster"),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespaceNonProject},
		}),
		Entry("Sanity - dv with project scope - same as original", &DatavolumeTestData{
			testName:         "Sanity - dv with project scope - same as original",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{pr1.Name}, []string{}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace3, testNamespace4, testNamespaceNonProject, testNamespace2},
		}),
		Entry("Dv with department+project in different dep scopes", &DatavolumeTestData{
			testName:         "Dv with department+project in different dep scopes",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{pr2.Name}, []string{defaultDep.Name}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace3, testNamespaceNonProject},
		}),
		Entry("Dv with department+project in same dep scopes", &DatavolumeTestData{
			testName:         "Dv with department+project in same dep scopes",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{pr2.Name}, []string{dep2.Name}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject},
		}),
		Entry("Dv with cluster+project scopes", &DatavolumeTestData{
			testName:         "Dv with cluster+project scope",
			dvOriginalPvc:    buildPvc(dv3, nil),
			dvOriginalPv:     buildPv(dv3, nil),
			dvToReconcile:    buildDatavolume(dv3, []string{pr2.Name}, []string{}, "cluster"),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespaceNonProject},
		}),
		Entry("Dv with cluster+department scopes", &DatavolumeTestData{
			testName:         "Dv with cluster+department scope",
			dvOriginalPvc:    buildPvc(dv3, nil),
			dvOriginalPv:     buildPv(dv3, nil),
			dvToReconcile:    buildDatavolume(dv3, []string{}, []string{dep2.Name}, "cluster"),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespaceNonProject},
		}),
		Entry("Dv with cluster+department+project scopes", &DatavolumeTestData{
			testName:         "Dv with cluster+department+project scope",
			dvOriginalPvc:    buildPvc(dv3, nil),
			dvOriginalPv:     buildPv(dv3, nil),
			dvToReconcile:    buildDatavolume(dv3, []string{pr4.Name}, []string{dep2.Name}, "cluster"),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespaceNonProject},
		}),
		Entry("Sanity - dv with department scope - then updated to project scope", &DatavolumeTestData{
			testName:                           "Sanity - dv with department scope - then updated to project scope",
			dvOriginalPvc:                      buildPvc(dv2, nil),
			dvOriginalPv:                       buildPv(dv2, nil),
			dvToReconcile:                      buildDatavolume2(dv2, true, []string{}, []string{dep2.Name}, ""),
			updateDvSharedScopesAfterReconcile: buildSharedScopes([]string{pr1.Name}, []string{}, ""),
			expectedDvStatus:                   buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName2, pvName2),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace3},
		}),
		Entry("Sanity - dv with 2 projects scopes - then updated and one of them removed", &DatavolumeTestData{
			testName:                           "Sanity - dv with 2 projects scopes - then updated and one of them removed",
			dvOriginalPvc:                      buildPvc(dv1, nil),
			dvOriginalPv:                       buildPv(dv1, nil),
			dvToReconcile:                      buildDatavolume(dv1, []string{pr2.Name, pr3.Name}, []string{}, ""),
			updateDvSharedScopesAfterReconcile: buildSharedScopes([]string{pr3.Name}, []string{}, ""),
			expectedDvStatus:                   buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace2, testNamespace4, testNamespaceNonProject},
		}),
		Entry("Sanity - dv with project scopes - then updated to cluster scope", &DatavolumeTestData{
			testName:                           "Sanity - dv with project scopes - then updated to cluster scope",
			dvOriginalPvc:                      buildPvc(dv2, nil),
			dvOriginalPv:                       buildPv(dv2, nil),
			dvToReconcile:                      buildDatavolume(dv2, []string{pr3.Name}, []string{}, ""),
			updateDvSharedScopesAfterReconcile: buildSharedScopes([]string{}, []string{}, "cluster"),
			expectedDvStatus:                   buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName2, pvName2),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace3: true, testNamespace2: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespaceNonProject},
		}),
		Entry("Sanity - dv with department scope - then updated - dep is removed, a different dep is added", &DatavolumeTestData{
			testName:                           "Sanity - dv with department scope - then updated - dep is removed, a different dep is added",
			dvOriginalPvc:                      buildPvc(dv2, nil),
			dvOriginalPv:                       buildPv(dv2, nil),
			dvToReconcile:                      buildDatavolume2(dv2, true, []string{}, []string{dep2.Name}, ""),
			updateDvSharedScopesAfterReconcile: buildSharedScopes([]string{}, []string{defaultDep.Name}, ""),
			expectedDvStatus:                   buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName2, pvName2),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace4: true, testNamespace2: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace1, testNamespaceNonProject, testNamespace3},
		}),
		Entry("Original pvc doesn't exist", &DatavolumeTestData{
			testName:         "Original pvc doesn't exist",
			dvOriginalPvc:    buildPvc(dv4, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{}, []string{}, "cluster"),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.OriginalPvcDoesntExist, "", ""),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace1, testNamespace2, testNamespace3},
			expectedPvcThatShouldExist:                    []types.NamespacedName{{Name: pvcName4, Namespace: testNamespace4}},
			expectedPvThatShouldExist:                     []string{pvName1},
		}),
		Entry("Original pvc is not bound", &DatavolumeTestData{
			testName:         "Original pvc is not bound",
			dvOriginalPvc:    buildPvcInner(pvcName1, testNamespace1, pvName1, v1.ClaimPending, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{pr4.Name}, []string{}, ""),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.OriginalPvcNotBound, "", ""),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace1, testNamespace2, testNamespace3},
			expectedPvcThatShouldExist:                    []types.NamespacedName{{Name: pvcName1, Namespace: testNamespace1}},
			expectedPvThatShouldExist:                     []string{pvName1},
		}),
		Entry("Original pv doesn't exist", &DatavolumeTestData{
			testName:         "Original pv doesn't exist",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv3, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{}, []string{dep2.Name}, ""),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.PvOfOriginalPvcDoesntExist, "", ""),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace1, testNamespace2, testNamespace3},
			expectedPvcThatShouldExist:                    []types.NamespacedName{{Name: pvcName1, Namespace: testNamespace1}},
			expectedPvThatShouldExist:                     []string{pvName3},
		}),
		Entry("Original pv is not bound", &DatavolumeTestData{
			testName:         "Original pv is not bound",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPvWithStatus(pvName1, pvcName1, testNamespace3, v1.VolumePending, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{pr4.Name}, []string{}, ""),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.PvOfOriginalPvcNotBound, "", ""),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace1, testNamespace2, testNamespace3},
			expectedPvcThatShouldExist:                    []types.NamespacedName{{Name: pvcName1, Namespace: testNamespace1}},
			expectedPvThatShouldExist:                     []string{pvName1},
		}),
		Entry("Project of original pvc not found", &DatavolumeTestData{
			testName:         "Project of original pvc not found",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolumeInner("dv1", pvcName1, "non-existing-projjjj", false, []string{}, []string{dep2.Name}, ""),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.ProjectOfOriginalPvcNotFound, "", ""),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace1, testNamespace2, testNamespace3},
		}),
		Entry("Namespace of original pvc not found", &DatavolumeTestData{
			testName:                   "Namespace of original pvc not found",
			objectsToAddBeforeTestCase: []client.Object{buildKaiProject("some-proj", testDepartment2)},
			dvOriginalPvc:              buildPvc(dv1, nil),
			dvOriginalPv:               buildPv(dv1, nil),
			dvToReconcile:              buildDatavolumeInner("dv1", pvcName1, "some-proj", false, []string{}, []string{dep2.Name}, ""),
			shouldSucceed:              ptr.To(false),
			expectedDvStatus:           buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.NamespaceOfOriginalPvcNotFound, "", ""),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace1, testNamespace2, testNamespace3},
		}),
		Entry("Original pvc+pv don't exist; runai copied ones do - should succeed", &DatavolumeTestData{
			testName:         "Original pvc+pv don't exist; runai copied ones do - should succeed",
			dvOriginalPvc:    buildPvcOfDvInNs(dv2, "dv-dv2-1234", testNamespace2, "dv-dv2-1234", ""),
			dvOriginalPv:     buildPvOfDvForNs(dv2, "dv-dv2-1234", "dv-dv2-1234", testNamespace2, nil),
			dvToReconcile:    buildDatavolume(dv2, []string{}, []string{dep2.Name}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName2, pvName2),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject},
		}),
		Entry("Original pv+pvc don't exist; runai copied ones do - with updated status - should succeed", &DatavolumeTestData{
			testName:         "Original pv+pvc don't exist; runai copied ones do - with updated status - should succeed",
			dvOriginalPvc:    buildPvcOfDvInNs(dv1, "dv-dv1-1111", testNamespace1, "dv-dv1-1111", ""),
			dvOriginalPv:     buildPvOfDvForNs(dv1, "dv-dv1-1111", "dv-dv1-1111", testNamespace1, nil),
			dvToReconcile:    buildDatavolumeWithStatus(dv1, buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", "dv-dv1-1111", "dv-dv1-1111"), []string{}, []string{defaultDep.Name}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace2, testNamespace3, testNamespaceNonProject},
		}),
		Entry("Original pv+pvc don't exist; runai copied pvc exists; while the pv exists from before (name conflict) - should fail cause original don't exist", &DatavolumeTestData{
			testName:         "Original pv+pvc don't exist; runai copied pvc exists; while the pv exists from before (name conflict) - should fail cause original don't exist",
			dvOriginalPvc:    buildPvcOfDvInNs(dv3, "dv-dv3-3333", testNamespace3, "dv-dv3-3333", ""),
			dvOriginalPv:     buildPvWithLabels("dv-dv3-3333", "some-other-user-created-pvc", testNamespace3, v1.ReadOnlyMany, v1.PersistentVolumeReclaimDelete, nil, map[string]string{}),
			dvToReconcile:    buildDatavolume(dv3, []string{}, []string{dep2.Name}, ""),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.OriginalPvcDoesntExist, "", ""),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace1, testNamespace2, testNamespace3},
		}),
		Entry("Original pv+pvc exist; runai copied pvc exists; while the pv exists from before (name conflict) - should re-create the pvc as well with new name - should succeed", &DatavolumeTestData{
			testName:                   "Original pv+pvc exist; runai copied pvc exists; while the pv exists from before (name conflict) - should re-create the pvc as well with new name - should succeed",
			objectsToAddBeforeTestCase: []client.Object{buildPv(dv3, &readOnlyMany), buildPvc(dv3, nil)},
			dvOriginalPvc:              buildPvcOfDvInNs(dv3, "dv-dv3-3333", testNamespace3, "dv-dv3-3333", ""),
			dvOriginalPv:               buildPvWithLabels("dv-dv3-3333", "some-other-user-created-pvc", testNamespace3, v1.ReadOnlyMany, v1.PersistentVolumeReclaimDelete, nil, map[string]string{}),
			dvToReconcile:              buildDatavolume(dv3, []string{}, []string{dep2.Name}, ""),
			expectedDvStatus:           buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject},
		}),
		Entry("Original pv+pvc don't exist; runai copied pv exists; while the pvc exists from before (name conflict) - should fail cause original don't exist", &DatavolumeTestData{
			testName:         "Original pv+pvc don't exist; runai copied pv exists; while the pvc exists from before (name conflict) - should fail cause original don't exist",
			dvOriginalPvc:    buildPvcWithLabels("dv-dv1-1111", testNamespace1, "", "", v1.ReadOnlyMany, nil, map[string]string{}),
			dvOriginalPv:     buildPvOfDvForNs(dv1, "dv-dv1-1111", "dv-dv1-1111", testNamespace1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{pr2.Name}, []string{}, ""),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.OriginalPvcDoesntExist, "", ""),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace1, testNamespace2, testNamespace3},
		}),
		Entry("Original pv+pvc exist; runai copied pv exists; while the pvc exists from before (name conflict) - should re-create the pv as well with new name - should succeed", &DatavolumeTestData{
			testName:                   "Original pv+pvc exist; runai copied pv exists; while the pvc exists from before (name conflict) - should re-create the pv as well with new name - should succeed",
			objectsToAddBeforeTestCase: []client.Object{buildPv(dv1, nil), buildPvc(dv1, &readOnlyMany)},
			dvOriginalPvc:              buildPvcWithLabels("dv-dv1-1111", testNamespace1, "", "", v1.ReadOnlyMany, nil, map[string]string{}),
			dvOriginalPv:               buildPvOfDvForNs(dv1, "dv-dv1-1111", "dv-dv1-1111", testNamespace1, nil),
			dvToReconcile:              buildDatavolume(dv1, []string{pr2.Name}, []string{}, ""),
			expectedDvStatus:           buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespace3, testNamespaceNonProject},
		}),
		Entry("Original pvc+pv don't exist; runai copied ones do but they are not pointing at each other - should re-create both and succeed", &DatavolumeTestData{
			testName:                   "Original pvc+pv don't exist; runai copied ones do but they are not pointing at each other - should re-create both and succeed",
			objectsToAddBeforeTestCase: []client.Object{buildPvcOfDvInNs(dv2, "dv-dv2-2222", testNamespace2, "points-to-different-pv", ""), buildPvOfDvForNs(dv2, "dv-dv2-2222", "dv-dv2-2222", testNamespace2, nil)},
			dvOriginalPvc:              buildPvc(dv2, nil),
			dvOriginalPv:               buildPv(dv2, nil),
			dvToReconcile:              buildDatavolume(dv2, []string{}, []string{}, "cluster"),
			expectedDvStatus:           buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName2, pvName2),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespaceNonProject},
		}),
		Entry("Original pvc+pv don't exist; runai copied ones do but they are not pointing at each other 2 - should re-create both and succeed", &DatavolumeTestData{
			testName: "Original pvc+pv don't exist; runai copied ones do but they are not pointing at each other 2 - should re-create both and succeed",
			objectsToAddBeforeTestCase: []client.Object{
				buildPvcOfDvInNs(dv2, "dv-dv2-2222", testNamespace2, "dv-dv2-2222", ""),
				buildPvWithLabels("dv-dv2-2222", "dv-dv2-2222", "different-namespaceeee", v1.ReadOnlyMany,
					v1.PersistentVolumeReclaimRetain, ptr.To("dv2"),
					map[string]string{v1alpha1.DataVolumeIdLabel: "dv2", v1alpha1.DataVolumePvForNamespaceLabel: testNamespace2}),
			},
			dvOriginalPvc:    buildPvc(dv2, nil),
			dvOriginalPv:     buildPv(dv2, nil),
			dvToReconcile:    buildDatavolume(dv2, []string{}, []string{}, "cluster"),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName2, pvName2),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace2: true, testNamespace3: true, testNamespace4: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespaceNonProject},
		}),
		Entry("pv+pvc in one ns exist but they are not pointing at each other - should re-create both and succeed", &DatavolumeTestData{
			testName:                   "pv+pvc in one ns exist but they are not pointing at each other - should re-create both and succeed",
			objectsToAddBeforeTestCase: []client.Object{buildPvcOfDvInNs(dv3, "dv-dv3-1111", testNamespace1, "points-to-different-pv", ""), buildPvOfDvForNs(dv3, "dv-dv3-1111", "dv-dv3-1111", testNamespace1, nil)},
			dvOriginalPvc:              buildPvc(dv3, nil),
			dvOriginalPv:               buildPv(dv3, nil),
			dvToReconcile:              buildDatavolume(dv3, []string{pr1.Name}, []string{}, ""),
			expectedDvStatus:           buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace2, testNamespace4, testNamespaceNonProject},
		}),
		Entry("pv+pvc in one ns exist but they are not pointing at each other 2 - should re-create both and succeed", &DatavolumeTestData{
			testName:                   "pv+pvc in one ns exist but they are not pointing at each other 2 - should re-create both and succeed",
			objectsToAddBeforeTestCase: []client.Object{buildPvcOfDvInNs(dv3, "dv-dv3-1111", testNamespace1, "dv-dv3-1111", ""), buildPvOfDvForNs(dv3, "dv-dv3-1111", "dv-different-name", testNamespace1, nil)},
			dvOriginalPvc:              buildPvc(dv3, nil),
			dvOriginalPv:               buildPv(dv3, nil),
			dvToReconcile:              buildDatavolume(dv3, []string{pr1.Name}, []string{}, ""),
			expectedDvStatus:           buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace2, testNamespace4, testNamespaceNonProject},
		}),
		Entry("pvc in one ns exists; pv exists from before (name conflict) - should re-create both and succeed", &DatavolumeTestData{
			testName: "pvc in one ns exists; pv exists from before (name conflict) - should re-create both and succeed",
			objectsToAddBeforeTestCase: []client.Object{
				buildPvcOfDvInNs(dv3, "dv-dv3-1111", testNamespace1, "dv-dv3-1111", ""),
				buildPvWithLabels("dv-dv3-1111", "some-other-user-created-pvc", testNamespace1,
					v1.ReadOnlyMany, v1.PersistentVolumeReclaimDelete, nil, map[string]string{}),
			},
			dvOriginalPvc:    buildPvc(dv3, nil),
			dvOriginalPv:     buildPv(dv3, nil),
			dvToReconcile:    buildDatavolume(dv3, []string{pr1.Name}, []string{}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace2, testNamespace4, testNamespaceNonProject},
		}),
		Entry("pv in one ns exists; pvc exists from before (name conflict) - should re-create both and succeed", &DatavolumeTestData{
			testName: "pv in one ns exists; pvc exists from before (name conflict) - should re-create both and succeed",
			objectsToAddBeforeTestCase: []client.Object{
				buildPvcWithLabels("dv-dv3-1111", testNamespace1, "", "", v1.ReadOnlyMany, nil, map[string]string{}),
				buildPvOfDvForNs(dv3, "dv-dv3-1111", "dv-dv3-1111", testNamespace1, nil),
			},
			dvOriginalPvc:    buildPvc(dv3, nil),
			dvOriginalPv:     buildPv(dv3, nil),
			dvToReconcile:    buildDatavolume(dv3, []string{pr1.Name}, []string{}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace2, testNamespace4, testNamespaceNonProject},
		}),
		Entry("Double pv+pvc exist in one of the namespaces; make sure they are deleted", &DatavolumeTestData{
			testName: "Double pv+pvc exist in one of the namespaces; make sure they are deleted",
			objectsToAddBeforeTestCase: []client.Object{
				buildPvcOfDvInNs(dv3, "dv-dv3-1111", testNamespace1, "dv-dv3-1111", ""),
				buildPvcOfDvInNs(dv3, "dv-dv3-1122", testNamespace1, "dv-dv3-1122", ""),
				buildPvOfDvForNs(dv3, "dv-dv3-1111", "dv-dv3-1111", testNamespace1, nil),
				buildPvOfDvForNs(dv3, "dv-dv3-1122", "dv-dv3-1122", testNamespace1, nil),
			},
			dvOriginalPvc:    buildPvc(dv3, nil),
			dvOriginalPv:     buildPv(dv3, nil),
			dvToReconcile:    buildDatavolume(dv3, []string{pr1.Name}, []string{}, ""),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName3, pvName3),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true, testNamespace3: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace2, testNamespace4, testNamespaceNonProject},
		}),
		Entry("Scope project doesn't exist", &DatavolumeTestData{
			testName:         "Scope project doesn't exist",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{"unexisting-proj"}, []string{dep2.Name}, ""),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.ScopeNotFound, pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace2, testNamespace3},
		}),
		Entry("Scope department doesn't exist", &DatavolumeTestData{
			testName:         "Scope department doesn't exist",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolume(dv1, []string{pr4.Name}, []string{"unexisting-dep"}, ""),
			shouldSucceed:    ptr.To(false),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeNotReady, datavolume.ScopeNotFound, pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist:    map[string]bool{testNamespace1: true},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{testNamespace4, testNamespaceNonProject, testNamespace2, testNamespace3},
		}),
		Entry("Sanity - Should use original volume", &DatavolumeTestData{
			testName:      "Should use original volume - project scope",
			dvOriginalPvc: buildPvc(dv1, nil),
			dvOriginalPv:  buildPv(dv1, nil),
			dvToReconcile: buildDatavolumeWithShouldUseOriginalVolume(
				buildDatavolume(dv1, []string{pr2.Name}, []string{}, ""),
				[]string{pr2.Name},
				[]string{},
				"",
				true,
			),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist: map[string]bool{
				testNamespace1: true, // original namespace
				testNamespace2: true, // target namespace
			},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{
				testNamespace3,
				testNamespace4,
				testNamespaceNonProject,
			},
			expectedPvcThatShouldExist: []types.NamespacedName{
				{Name: pvcName1, Namespace: testNamespace1},
			},
			expectedPvThatShouldExist: []string{pvName1},
		}),
		Entry("Should use original volume - multiple scopes", &DatavolumeTestData{
			testName:         "Should use original volume - multiple scopes",
			dvOriginalPvc:    buildPvc(dv1, nil),
			dvOriginalPv:     buildPv(dv1, nil),
			dvToReconcile:    buildDatavolumeWithShouldUseOriginalVolume(buildDatavolume(dv1, []string{pr2.Name}, []string{dep2.Name}, "cluster"), []string{pr2.Name}, []string{dep2.Name}, "cluster", true),
			expectedDvStatus: buildDataVolumeStatus(v1alpha1.DataVolumeReady, "", pvcName1, pvName1),
			expectedNamespacesWherePvcAndPvShouldExist: map[string]bool{
				testNamespace1: true, // original namespace
				testNamespace2: true,
				testNamespace3: true,
				testNamespace4: true,
			},
			expectedNamespacesWherePvcAndPvShouldNotExist: []string{
				testNamespaceNonProject,
			},
			expectedPvcThatShouldExist: []types.NamespacedName{
				{Name: pvcName1, Namespace: testNamespace1}, // original PVC
			},
			expectedPvThatShouldExist: []string{pvName1}, // original PV
		}),
	)
})

func validatePvcOfDatavolume(testName string, dv *v1alpha1.DataVolume, pvcAfterReconcile, originalDvPvc *v1.PersistentVolumeClaim, pvName string) {
	Expect(pvcAfterReconcile.Spec.VolumeName).To(Equal(pvName), "Test: %s, pvc %s/%s volume name", testName, pvcAfterReconcile.Namespace, pvcAfterReconcile.Name)
	Expect(pvcAfterReconcile.Spec.AccessModes).To(Equal(originalDvPvc.Spec.AccessModes), "Test: %s, pvc %s/%s AccessModes", testName, pvcAfterReconcile.Namespace, pvcAfterReconcile.Name)

	// Check ownership based on ShouldUseOriginalVolume
	if dv.Spec.ShouldUseOriginalVolume {
		// The original PVC should not be owned by the datavolume
		if pvcAfterReconcile.Name == originalDvPvc.Name && pvcAfterReconcile.Namespace == originalDvPvc.Namespace {
			Expect(dv.IsOwner(pvcAfterReconcile)).To(Equal(-1), "Test: %s, pvc %s/%s owner", testName, pvcAfterReconcile.Namespace, pvcAfterReconcile.Name)
		} else {
			Expect(dv.IsOwner(pvcAfterReconcile)).To(Equal(0), "Test: %s, pvc %s/%s owner", testName, pvcAfterReconcile.Namespace, pvcAfterReconcile.Name)
		}
	} else {
		Expect(dv.IsOwner(pvcAfterReconcile)).To(Equal(0), "Test: %s, pvc %s/%s owner", testName, pvcAfterReconcile.Namespace, pvcAfterReconcile.Name)
	}

	Expect(pvcAfterReconcile.Labels[v1alpha1.DataVolumeNameLabel]).To(Equal(dv.Name), "Test: %s, pvc %s/%s DataVolumeNameLabel label", testName, pvcAfterReconcile.Namespace, pvcAfterReconcile.Name)
	Expect(pvcAfterReconcile.Labels[v1alpha1.DataVolumeIdLabel]).To(Equal(dv.Spec.DatavolumeId), "Test: %s, pvc %s/%s DataVolumeIdLabel label", testName, pvcAfterReconcile.Namespace, pvcAfterReconcile.Name)

	pvcAfterReconcile.Spec.VolumeName = ""
	originalDvPvc.Spec.VolumeName = ""
	Expect(pvcAfterReconcile.Spec).To(Equal(originalDvPvc.Spec), "Test: %s, pvc %s/%s spec", testName, pvcAfterReconcile.Namespace, pvcAfterReconcile.Name)
}

func validatePvOfDatavolume(testName string, dv *v1alpha1.DataVolume, pvAfterReconcile, originalDvPv *v1.PersistentVolume, pvcName, pvcNamespace string) {
	Expect(pvAfterReconcile.Spec.ClaimRef).ToNot(BeNil(), "Test: %s, pv %s claim ref", testName, pvAfterReconcile.Name)
	Expect(pvAfterReconcile.Spec.ClaimRef.Name).To(Equal(pvcName), "Test: %s, pv %s claim ref name", testName, pvAfterReconcile.Name)
	Expect(pvAfterReconcile.Spec.ClaimRef.Namespace).To(Equal(pvcNamespace), "Test: %s, pv %s claim ref namespace", testName, pvAfterReconcile.Name)
	Expect(pvAfterReconcile.Spec.AccessModes).To(Equal(originalDvPv.Spec.AccessModes), "Test: %s, pv %s AccessModes", testName, pvAfterReconcile.Name)

	// Check reclaim policy based on ShouldUseOriginalVolume
	if dv.Spec.ShouldUseOriginalVolume {
		if pvAfterReconcile.Name != originalDvPv.Name {
			// Only copies should have Retain policy
			Expect(pvAfterReconcile.Spec.PersistentVolumeReclaimPolicy).To(Equal(v1.PersistentVolumeReclaimRetain),
				"Test: %s, pv %s PersistentVolumeReclaimPolicy", testName, pvAfterReconcile.Name)
		}
	} else {
		// For non-original volume, use Retain policy
		if dv.Spec.ShouldDeleteOriginalVolume && testNamespacesToProject[pvcNamespace] == dv.Spec.ProjectName {
			Expect(pvAfterReconcile.Spec.PersistentVolumeReclaimPolicy).To(Equal(v1.PersistentVolumeReclaimDelete),
				"Test: %s, pv %s PersistentVolumeReclaimPolicy", testName, pvAfterReconcile.Name)
		} else {
			Expect(pvAfterReconcile.Spec.PersistentVolumeReclaimPolicy).To(Equal(v1.PersistentVolumeReclaimRetain),
				"Test: %s, pv %s PersistentVolumeReclaimPolicy", testName, pvAfterReconcile.Name)
		}
	}

	// Check ownership based on ShouldUseOriginalVolume
	if !dv.Spec.ShouldUseOriginalVolume || (pvAfterReconcile.Name != originalDvPv.Name) {
		Expect(dv.IsOwner(pvAfterReconcile)).To(Equal(0), "Test: %s, pv %s owner", testName, pvAfterReconcile.Name)
	}

	Expect(pvAfterReconcile.Labels[v1alpha1.DataVolumeNameLabel]).To(Equal(dv.Name), "Test: %s, pv %s DataVolumeNameLabel label", testName, pvAfterReconcile.Name)
	Expect(pvAfterReconcile.Labels[v1alpha1.DataVolumeIdLabel]).To(Equal(dv.Spec.DatavolumeId), "Test: %s, pv %s DataVolumeIdLabel label", testName, pvAfterReconcile.Name)
	Expect(pvAfterReconcile.Labels[v1alpha1.DataVolumePvForNamespaceLabel]).To(Equal(pvcNamespace), "Test: %s, pv %s DataVolumePvForNamespaceLabel label", testName, pvAfterReconcile.Name)

	pvAfterReconcile.Spec.ClaimRef = nil
	pvAfterReconcile.Spec.PersistentVolumeReclaimPolicy = ""
	originalDvPv.Spec.ClaimRef = nil
	originalDvPv.Spec.PersistentVolumeReclaimPolicy = ""
	Expect(pvAfterReconcile.Spec).To(Equal(originalDvPv.Spec), "Test: %s, pv %s spec", testName, pvAfterReconcile.Name)
}

func validateStatus(testName string, dvAfterReconcile *v1alpha1.DataVolume, expectedStatus *v1alpha1.DataVolumeStatus) {
	dvInnerName := dvAfterReconcile.Name
	if len(dvInnerName) > datavolume.MaxNameLenForPvc {
		dvInnerName = dvInnerName[:datavolume.MaxNameLenForPvc]
	}
	dvPvPvcPrefix := fmt.Sprintf("dv-%s-", dvInnerName)

	if expectedStatus.DatavolumePvcName != "" {
		if dvAfterReconcile.Spec.ShouldUseOriginalVolume {
			// When using original volume, the PVC name should be the original name
			Expect(dvAfterReconcile.Status.DatavolumePvcName).To(Equal(expectedStatus.DatavolumePvcName), "Test: %s, status.DatavolumePvcName", testName)
		} else {
			Expect(dvAfterReconcile.Status.DatavolumePvcName).To(HavePrefix(dvPvPvcPrefix), "Test: %s, status.DatavolumePvcName", testName)
		}
	} else {
		Expect(dvAfterReconcile.Status.DatavolumePvcName).To(Equal(""), "Test: %s, status.DatavolumePvcName", testName)
	}
	if expectedStatus.DatavolumePvName != "" {
		if dvAfterReconcile.Spec.ShouldUseOriginalVolume {
			// When using original volume, the PV name should be the original name
			Expect(dvAfterReconcile.Status.DatavolumePvName).To(Equal(expectedStatus.DatavolumePvName), "Test: %s, status.DatavolumePvName", testName)
		} else {
			Expect(dvAfterReconcile.Status.DatavolumePvName).To(HavePrefix(dvPvPvcPrefix), "Test: %s, status.DatavolumePvName", testName)
		}
	} else {
		Expect(dvAfterReconcile.Status.DatavolumePvName).To(Equal(""), "Test: %s, status.DatavolumePvName", testName)
	}

	dvAfterReconcile.Status.DatavolumePvcName = ""
	dvAfterReconcile.Status.DatavolumePvName = ""
	expectedStatus.DatavolumePvcName = ""
	expectedStatus.DatavolumePvName = ""
	Expect(reflect.DeepEqual(dvAfterReconcile.Status, *expectedStatus)).To(BeTrue(), "Test: %s, status; %v // %v", testName, dvAfterReconcile.Status, expectedStatus)
}
