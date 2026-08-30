package test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	pkg "github.com/run-ai/runai/runai-cluster/cluster/runtime-rules"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/controllers"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/validation"
	dwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("A runtime-rules test", func() {
	var stopper chan struct{}
	var promClient = prometheus.GetQuerierClient(false, "")

	const (
		RecorderBufferSize = 100
	)

	Context("-------------test-runtime-rules--------------", Serial, func() {
		ctx, cancel := context.WithCancel(context.TODO())

		var watchClient client.WithWatch
		BeforeEach(func() {
			queuesManager.SetKillWorker(false, 0)
			ns := BuildNamespace(testNamespaceName, map[string]string{RunaiQueueLabel: testProjectName})
			projects := BuildProjectFixtures(testProjectName, testNamespaceName, 5, 5)

			////////////////////////////////////////////////////////////////////////////////////////////////////////////////
			// init the fake runtime-rules environment
			////////////////////////////////////////////////////////////////////////////////////////////////////////////////
			scheme := initSchema()

			// init the kube client
			watchClient = fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(append(projects, ns)...).Build()

			nsList := corev1.NamespaceList{}
			err := watchClient.List(ctx, &nsList)
			Expect(err).To(BeNil())
			Expect(len(nsList.Items)).To(Equal(1))

			projectList := kaiorg.ProjectList{}
			err = watchClient.List(ctx, &projectList)
			Expect(err).To(BeNil())
			Expect(len(projectList.Items)).To(Equal(1))
			// init the queues manager
			qm := &queuesManager.QueuesManager{}

			// init the runtime-rules (WorkloadRunner rule enforcement disabled at
			// this pre-3 test version; these legacy-CRD tests don't exercise it).
			rr := pkg.NewRuntimeRules(watchClient, qm, promClient, record.NewFakeRecorder(RecorderBufferSize), []int{queuesManager.UnknownEvent, queuesManager.DeactivateEvent}, validation.NewWorkloadValidator(), true)

			////////////////////////////////////////////////////////////////////////////////////////////////////////////////

			workloadCache := util.WorkloadCache{}
			rr.InitQueueManager(ctx, &workloadCache)
			projectController := controllers.NewProjectController(rr.K8sClient, rr.K8sClient, rr.QueuesManager,
				rr.PrometheusQuerier, &workloadCache, rr.Recorder, rr.EventList, validation.NewWorkloadValidator(),
				rr.SupportedJobKinds, true)

			workloadHandler := handlers.NewWorkloadHandler(rr.K8sClient, rr.QueuesManager, rr.PrometheusQuerier, &workloadCache, rr.SupportedJobKinds)
			twController := controllers.NewWorkloadController(rr.K8sClient, workloadHandler, constants.KindTrainingWorkload)
			iwController := controllers.NewWorkloadController(rr.K8sClient, workloadHandler, constants.KindInteractiveWorkload)
			dwController := controllers.NewWorkloadController(rr.K8sClient, workloadHandler, constants.KindDistributedWorkload)
			stopper = make(chan struct{})

			go projectController.Start(ctx, stopper)
			go iwController.Start(ctx, stopper)
			go twController.Start(ctx, stopper)
			go dwController.Start(ctx, stopper)

			time.Sleep(5 * time.Second)
		})

		AfterEach(func() {
			cancel()
			close(stopper)

			queuesManager.SetKillWorker(true, time.Duration(workerSleepSeconds+1))
		})

		It("Tests TrainingWorkload TrainingJobTimeLimitSecs", func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Error: %+v", fmt.Errorf("%v", r))
				}
			}()
			twName := "test-training-workload"
			jobStartTime := metav1.Time{Time: time.Now()}
			tw2 := BuildTrainingWorkload(testNamespaceName, twName, &jobStartTime)
			err := watchClient.Create(ctx, tw2)
			Expect(err).To(BeNil())

			time.Sleep(20 * time.Second)
			twUpdated := twv2a1.TrainingWorkload{}
			err = watchClient.Get(ctx, client.ObjectKey{Namespace: testNamespaceName, Name: twName}, &twUpdated)
			Expect(err).To(BeNil())

			Expect(twUpdated).NotTo(BeNil())
			Expect(twUpdated.Name).To(Equal(twName))

			Expect(twUpdated.Spec.Active).NotTo(BeNil())
			fmt.Printf("twUpdated.Spec.Active.Value: %v\n", twUpdated.Spec.Active.Deref())
			Expect(ptr.Deref(twUpdated.Spec.Active.Value, true)).To(Equal(false))
		})

		It("Tests InteractiveWorkload InteractiveJobTimeLimitSecs", func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Error: %+v", fmt.Errorf("%v", r))
				}
			}()
			iwName := "test-interactive-workload"
			jobStartTime := metav1.Time{Time: time.Now()}
			iw2 := BuildInteractiveWorkload(testNamespaceName, iwName, &jobStartTime)
			err := watchClient.Create(ctx, iw2)
			Expect(err).To(BeNil())

			time.Sleep(20 * time.Second)

			iwUpdated := iwv2a1.InteractiveWorkload{}
			err = watchClient.Get(ctx, client.ObjectKey{Namespace: testNamespaceName, Name: iwName}, &iwUpdated)
			Expect(err).To(BeNil())
			Expect(iwUpdated).NotTo(BeNil())
			Expect(iwUpdated.Name).To(Equal(iwName))

			Expect(iwUpdated.Spec.Active).NotTo(BeNil())
			fmt.Printf("iwUpdated.Spec.Active.Value: %v\n", iwUpdated.Spec.Active.Deref())
			Expect(ptr.Deref(iwUpdated.Spec.Active.Value, true)).To(Equal(false))

		})

		It("Tests DistributedWorkload TrainingJobTimeLimitSecs", func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Error: %+v", fmt.Errorf("%v", r))
				}
			}()
			dwName := "test-distributed-workload"
			jobStartTime := metav1.Time{Time: time.Now()}
			dw2 := BuildDistributedWorkload(testNamespaceName, dwName, &jobStartTime)
			err := watchClient.Create(ctx, dw2)
			Expect(err).To(BeNil())

			time.Sleep(20 * time.Second)

			dwUpdated := dwv2a1.DistributedWorkload{}
			err = watchClient.Get(ctx, client.ObjectKey{Namespace: testNamespaceName, Name: dwName}, &dwUpdated)
			Expect(err).To(BeNil())
			Expect(dwUpdated).NotTo(BeNil())
			Expect(dwUpdated.Name).To(Equal(dwName))

			Expect(dwUpdated.Spec.Active).NotTo(BeNil())
			fmt.Printf("dwUpdated.Spec.Active.Value: %v\n", dwUpdated.Spec.Active.Deref())
			Expect(ptr.Deref(dwUpdated.Spec.Active.Value, true)).To(Equal(false))

		})
	})
})
