package queuesManager

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/logger"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/validation"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/test"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/test/mocks"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	mpi "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v1alpha2"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var eventList = []int{50, 75, 90}

func SetEnvironment(t *testing.T) {
	t.Setenv("EVENT_LIST", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(eventList)), ","), "[]"))
}

func TestRuntimeRules(t *testing.T) {
	logger.InitLogger(int(zap.DebugLevel))
	zap.S().Infof("Started service tests")
	RegisterFailHandler(Fail)
	SetEnvironment(t)
	RunSpecs(t, "runtime-rules service Test Suite")
}

var _ = Describe("getMessage Test Suite", Ordered, func() {
	Context("getMessage function test", func() {
		timeNow := time.Now()
		It("Interactive Workload: check idle flow of events (naive)", func() {

			eventClock1 := EventClock{NextEventType: TaskTimerIdle}
			eventClock2 := EventClock{NextEventType: TaskTimerTimeout}

			eventClock1.DeactivateEvent = timeNow.Add(time.Minute * 5)
			eventClock2.DeactivateEvent = timeNow.Add(time.Minute * 5)
			msg := getMessage("test-message", eventClock1)
			Expect(msg).To(Equal("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by test-message remain idle, the job will be terminated in 4 minutes."))
			msg = getMessage("test-message", eventClock2)
			Expect(msg).To(Equal("Based on the time limit configured for your project, the job test-message will be terminated in 4 minutes."))

			eventClock1.DeactivateEvent = eventClock1.DeactivateEvent.Add(time.Hour * 3)
			eventClock2.DeactivateEvent = eventClock2.DeactivateEvent.Add(time.Hour * 3)
			msg = getMessage("test-message", eventClock1)
			Expect(msg).To(Equal("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by test-message remain idle, the job will be terminated in 3 hours - 4 minutes."))
			msg = getMessage("test-message", eventClock2)
			Expect(msg).To(Equal("Based on the time limit configured for your project, the job test-message will be terminated in 3 hours - 4 minutes."))

			eventClock1.DeactivateEvent = timeNow.Add(time.Hour * 30)
			eventClock2.DeactivateEvent = timeNow.Add(time.Hour * 30)
			msg = getMessage("test-message", eventClock1)
			Expect(msg).To(Equal("Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by test-message remain idle, the job will be terminated in 1 days - 5 hours - 59 minutes."))
			msg = getMessage("test-message", eventClock2)
			Expect(msg).To(Equal("Based on the time limit configured for your project, the job test-message will be terminated in 1 days - 5 hours - 59 minutes."))
		})
	})
})

var _ = Describe("getLimitsByWorkloadType — WorkloadRunner category routing", func() {
	const (
		iwTime    = int64(100)
		iwIdle    = int64(200)
		iwPreempt = int64(250)
		trainTime = int64(300)
		trainIdle = int64(400)
	)
	pq := &ProjectQueue{
		InteractiveJobTimeLimitSecs:                  iwTime,
		InteractiveJobMaxIdleDurationSecs:            iwIdle,
		InteractivePreemptibleJobMaxIdleDurationSecs: iwPreempt,
		TrainingJobTimeLimitSecs:                     trainTime,
		TrainingJobMaxIdleDurationSecs:               trainIdle,
	}

	It("routes an Interactive-category runner to interactive limits", func() {
		task := &WorkloadTask{WorkloadType: constants.KindWorkloadRunner, Category: constants.KindInteractiveWorkload}
		timeLimit, idleLimit, err := getLimitsByWorkloadType(pq, task)
		Expect(err).To(BeNil())
		Expect(timeLimit).To(Equal(iwTime))
		Expect(idleLimit).To(Equal(iwIdle))
	})

	It("uses the preemptible idle limit for a preemptible Interactive runner", func() {
		task := &WorkloadTask{WorkloadType: constants.KindWorkloadRunner, Category: constants.KindInteractiveWorkload, IsPreemptible: true}
		_, idleLimit, err := getLimitsByWorkloadType(pq, task)
		Expect(err).To(BeNil())
		Expect(idleLimit).To(Equal(iwPreempt))
	})

	It("routes Training and Distributed runners to training limits", func() {
		for _, category := range []string{constants.KindTrainingWorkload, constants.KindDistributedWorkload} {
			task := &WorkloadTask{WorkloadType: constants.KindWorkloadRunner, Category: category}
			timeLimit, idleLimit, err := getLimitsByWorkloadType(pq, task)
			Expect(err).To(BeNil())
			Expect(timeLimit).To(Equal(trainTime), "category %s", category)
			Expect(idleLimit).To(Equal(trainIdle), "category %s", category)
		}
	})

	It("yields zero limits (observed but not enforced) when Category is empty", func() {
		task := &WorkloadTask{WorkloadType: constants.KindWorkloadRunner}
		timeLimit, idleLimit, err := getLimitsByWorkloadType(pq, task)
		Expect(err).To(BeNil())
		Expect(timeLimit).To(BeZero())
		Expect(idleLimit).To(BeZero())
	})

	It("keeps legacy CRD limit routing unchanged", func() {
		task := &WorkloadTask{WorkloadType: constants.KindTrainingWorkload}
		timeLimit, idleLimit, err := getLimitsByWorkloadType(pq, task)
		Expect(err).To(BeNil())
		Expect(timeLimit).To(Equal(trainTime))
		Expect(idleLimit).To(Equal(trainIdle))
	})
})

var _ = Describe("HandleWorkloadTask Test Suite", Ordered, func() {

	Context("HandleWorkloadTask controller tests", func() {
		ctx, cancel := context.WithCancel(context.TODO())
		eventList := []int{0, 50, 75, 90, 100}

		const (
			noteJobName  = "notification-test-"
			noteIwName   = "note-iw-test-"
			noteProjName = "note-proj-test"

			noteIdleIWTime = 8640000
			noteIdleTWTime = 8640000

			fakeRecorderBuffSize = 100

			testNamespaceName = "test-ns-name"
			testProjectName   = "test-proj-name"

			messageCompareTimeout = "Normal WorkloadTimeoutApproaching Based on the time limit configured for your project, the job %s will be terminated in "
			messageCompareIdle    = "Normal WorkloadTimeoutApproaching Based on the idle GPU timeout configured for your project, as long as the GPU(s) used by %s remain idle, the job will be terminated in "
			messageTimeoutReached = "Normal WorkloadTimeoutReached The workload has reached the running time limitation on the project (100 days - 0 hours - 0 minutes). Stopping the workload."
		)

		var (
			k8sClient     client.WithWatch
			projectQueue  ProjectQueue
			priorityQueue *PriorityQueue
			mockCtrl      *gomock.Controller
			recorder      = record.FakeRecorder{}

			timePointer0  = time.Second * 0
			timePointer99 = time.Second * (noteIdleIWTime - 10)

			uuidCounter = 0
		)

		BeforeAll(func() {
			test.BuildNamespace(testNamespaceName, map[string]string{test.RunaiQueueLabel: testProjectName})
			// init the fake environment
			scheme := initSchema()
			k8sClient = fakerunaiclient.NewClientBuilder().WithScheme(scheme).Build()
			mockCtrl = gomock.NewController(GinkgoT())
		})

		JustBeforeEach(func() {
			priorityQueue = NewPriorityQueue()
			recorder = *record.NewFakeRecorder(fakeRecorderBuffSize)
			uuidCounter++
		})

		AfterAll(func() {
			cancel()
		})

		It("Clock initializing status check", func() {
			noteJobNameId := noteJobName + strconv.Itoa(uuidCounter)
			noteIwNameId := noteIwName + strconv.Itoa(uuidCounter)
			jobUuid := strconv.Itoa(uuidCounter)
			promQuerier := mocks.NewMockQuerier(mockCtrl)

			workloadCache := util.WorkloadCache{}

			timeNow := time.Now()
			jobStartTime := metav1.Time{Time: timeNow}
			iw2 := test.BuildInteractiveWorkload(testNamespaceName, noteIwNameId, &jobStartTime)
			_ = k8sClient.Create(ctx, iw2)

			projectQueue = ProjectQueue{
				ProjectName:                 noteProjName,
				queue:                       priorityQueue,
				Workers:                     nil,
				InteractiveJobTimeLimitSecs: noteIdleIWTime,
				TrainingJobTimeLimitSecs:    noteIdleTWTime,
				K8sClient:                   k8sClient,
				PromQuerier:                 promQuerier,
				Recorder:                    &recorder,
				EventList:                   eventList,
				WorkloadValidator:           &validation.WorkloadValidator{},
			}

			actionList1 := NewActionList(projectQueue.EventList, 0, 0, nil)

			workloadTask := &WorkloadTask{ // just a dummy WorkloadTask
				StartTime:    timeNow,
				WorkloadName: noteIwNameId,
				ProjectName:  projectQueue.ProjectName,
				Namespace:    testNamespaceName,
				JobUuid:      types.UID(jobUuid),
				JobName:      noteJobNameId,
				WorkloadType: constants.KindInteractiveWorkload,
				TaskAction:   *actionList1,
			}

			err := HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(0))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))

			timeStartMod := noteIdleIWTime/2 + 10
			workloadTask.StartTime = timeNow.Add(time.Second * time.Duration(-timeStartMod))
			workloadTask.TaskAction.SetEventType(TaskTimerTimeout).SetIndex(0)

			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(0))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(2))
			Expect(len(recorder.Events)).To(Equal(0))

			timeStartMod = noteIdleIWTime*0.75 + 10
			workloadTask.StartTime = timeNow.Add(time.Second * time.Duration(-timeStartMod))
			workloadTask.TaskAction.SetEventType(TaskTimerTimeout).SetIndex(0)

			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(0))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(3))
			Expect(len(recorder.Events)).To(Equal(0))

			timeStartMod = noteIdleIWTime*0.90 + 10
			workloadTask.StartTime = timeNow.Add(time.Second * time.Duration(-timeStartMod))
			workloadTask.TaskAction.SetEventType(TaskTimerTimeout).SetIndex(0)

			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(0))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(4))
			Expect(len(recorder.Events)).To(Equal(0))
		})

		It("Interactive Workload: check idle flow of events (naive)", func() {
			noteJobNameId := noteJobName + strconv.Itoa(uuidCounter)
			noteIwNameUnique := noteIwName + strconv.Itoa(uuidCounter)
			noteIwId := strconv.Itoa(uuidCounter)
			jobUuid := strconv.Itoa(uuidCounter)
			promQuerier := mocks.NewMockQuerier(mockCtrl)

			workloadCache := util.WorkloadCache{}

			timeNow := time.Now()
			jobStartTime := metav1.Time{Time: timeNow}
			iw2 := test.BuildInteractiveWorkload(testNamespaceName, noteIwNameUnique, &jobStartTime)
			iw2.Spec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
			_ = k8sClient.Create(ctx, iw2)
			first := promQuerier.EXPECT().GetWorkloadIdleDuration(ctx, noteIwNameUnique, noteIwId).Return(&timePointer0, nil)
			second := promQuerier.EXPECT().GetWorkloadIdleDuration(ctx, noteIwNameUnique, noteIwId).Return(&timePointer99, nil).AnyTimes() // setting duration to 99% of time
			gomock.InOrder(first, second)

			projectQueue = ProjectQueue{
				ProjectName:                       noteProjName,
				queue:                             priorityQueue,
				Workers:                           nil,
				InteractiveJobMaxIdleDurationSecs: noteIdleIWTime,
				TrainingJobMaxIdleDurationSecs:    noteIdleTWTime,
				K8sClient:                         k8sClient,
				PromQuerier:                       promQuerier,
				Recorder:                          &recorder,
				EventList:                         eventList,
				WorkloadValidator:                 &validation.WorkloadValidator{},
			}

			actionList1 := NewActionList(projectQueue.EventList, 1, 1, nil)

			job := &WorkloadTask{ // just a dummy WorkloadTask
				StartTime:            timeNow,
				WorkloadName:         noteIwNameUnique,
				WorkloadID:           types.UID(noteIwId),
				ProjectName:          projectQueue.ProjectName,
				Namespace:            testNamespaceName,
				JobUuid:              types.UID(jobUuid),
				JobName:              noteJobNameId,
				WorkloadType:         constants.KindInteractiveWorkload,
				TaskAction:           *actionList1,
				WorkloadRequestedGpu: true,
			}

			err := HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, job, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(job.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(job.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))

			job.StartTime = timeNow.Add(time.Second * -time.Duration(noteIdleIWTime*0.95))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, job, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(job.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(2))
			Expect(job.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg := <-recorder.Events
			msgCmp := fmt.Sprintf(messageCompareIdle, job.JobName)
			Expect(strings.HasPrefix(msg, msgCmp)).To(BeTrue())

			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, job, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(job.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(3))
			Expect(job.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg = <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp)).To(BeTrue())

			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, job, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(job.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(4))
			Expect(job.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg = <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp)).To(BeTrue())
		})

		It("Interactive Preemptible Workload: check idle flow of events (naive)", func() {
			noteJobNameId := noteJobName + strconv.Itoa(uuidCounter)
			noteIwNameUnique := noteIwName + strconv.Itoa(uuidCounter)
			noteIwId := strconv.Itoa(uuidCounter)
			jobUuid := strconv.Itoa(uuidCounter)
			promQuerier := mocks.NewMockQuerier(mockCtrl)

			workloadCache := util.WorkloadCache{}

			timeNow := time.Now()
			jobStartTime := metav1.Time{Time: timeNow}
			iw2 := test.BuildInteractiveWorkload(testNamespaceName, noteIwNameUnique, &jobStartTime)
			iw2.Spec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
			iw2.Spec.Preemptible = field_types.NewOptionalBoolField(ptr.To(true)) //nolint:staticcheck
			_ = k8sClient.Create(ctx, iw2)
			first := promQuerier.EXPECT().GetWorkloadIdleDuration(ctx, noteIwNameUnique, noteIwId).Return(&timePointer0, nil)
			second := promQuerier.EXPECT().GetWorkloadIdleDuration(ctx, noteIwNameUnique, noteIwId).Return(&timePointer99, nil).AnyTimes() // setting duration to 99% of time
			gomock.InOrder(first, second)

			projectQueue = ProjectQueue{
				ProjectName: noteProjName,
				queue:       priorityQueue,
				Workers:     nil,
				InteractivePreemptibleJobMaxIdleDurationSecs: noteIdleIWTime,
				K8sClient:         k8sClient,
				PromQuerier:       promQuerier,
				Recorder:          &recorder,
				EventList:         eventList,
				WorkloadValidator: &validation.WorkloadValidator{},
			}

			actionList1 := NewActionList(projectQueue.EventList, 1, 1, nil)

			job := &WorkloadTask{ // just a dummy WorkloadTask
				StartTime:            timeNow,
				WorkloadName:         noteIwNameUnique,
				WorkloadID:           types.UID(noteIwId),
				ProjectName:          projectQueue.ProjectName,
				Namespace:            testNamespaceName,
				JobUuid:              types.UID(jobUuid),
				JobName:              noteJobNameId,
				WorkloadType:         constants.KindInteractiveWorkload,
				TaskAction:           *actionList1,
				WorkloadRequestedGpu: true,
				IsPreemptible:        true,
			}

			err := HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, job, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(job.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(job.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))

			job.StartTime = timeNow.Add(time.Second * -time.Duration(noteIdleIWTime*0.95))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, job, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(job.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(2))
			Expect(job.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg := <-recorder.Events
			msgCmp := fmt.Sprintf(messageCompareIdle, job.JobName)
			Expect(strings.HasPrefix(msg, msgCmp)).To(BeTrue())

			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, job, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(job.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(3))
			Expect(job.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg = <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp)).To(BeTrue())

			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, job, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(job.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(4))
			Expect(job.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg = <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp)).To(BeTrue())
		})

		It("Interactive Workload: check idle reset of event", func() {
			noteJobNameId := noteJobName + strconv.Itoa(uuidCounter)
			noteIwNameUnique := noteIwName + strconv.Itoa(uuidCounter)
			noteIwId := strconv.Itoa(uuidCounter)
			jobUuid := strconv.Itoa(uuidCounter)
			promQuerier := mocks.NewMockQuerier(mockCtrl)

			workloadCache := util.WorkloadCache{}

			timeNow := time.Now()
			jobStartTime := metav1.Time{Time: timeNow}
			iw2 := test.BuildInteractiveWorkload(testNamespaceName, noteIwNameUnique, &jobStartTime)
			iw2.Spec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
			_ = k8sClient.Create(ctx, iw2)

			first := promQuerier.EXPECT().GetWorkloadIdleDuration(ctx, noteIwNameUnique, noteIwId).Return(&timePointer0, nil).Times(2)
			second := promQuerier.EXPECT().GetWorkloadIdleDuration(ctx, noteIwNameUnique, noteIwId).Return(&timePointer99, nil).Times(2) // setting duration to 99% of time

			third := promQuerier.EXPECT().GetWorkloadIdleDuration(ctx, noteIwNameUnique, noteIwId).Return(&timePointer0, nil).Times(2) // setting duration to 99% of time

			forth := promQuerier.EXPECT().GetWorkloadIdleDuration(ctx, noteIwNameUnique, noteIwId).Return(&timePointer99, nil).Times(2) // setting duration to 99% of time
			gomock.InOrder(first, second, third, forth)

			actionList1 := NewActionList(projectQueue.EventList, 1, 1, nil)

			projectQueue = ProjectQueue{
				ProjectName:                       noteProjName,
				queue:                             priorityQueue,
				Workers:                           nil,
				InteractiveJobMaxIdleDurationSecs: noteIdleIWTime,
				TrainingJobMaxIdleDurationSecs:    noteIdleTWTime,
				K8sClient:                         k8sClient,
				PromQuerier:                       promQuerier,
				Recorder:                          &recorder,
				EventList:                         eventList,
				WorkloadValidator:                 &validation.WorkloadValidator{},
			}

			workloadTask := &WorkloadTask{ // just a dummy WorkloadTask
				StartTime:            timeNow,
				WorkloadName:         noteIwNameUnique,
				WorkloadID:           types.UID(noteIwId),
				ProjectName:          projectQueue.ProjectName,
				Namespace:            testNamespaceName,
				JobUuid:              types.UID(jobUuid),
				JobName:              noteJobNameId,
				WorkloadType:         constants.KindInteractiveWorkload,
				TaskAction:           *actionList1,
				WorkloadRequestedGpu: true,
			}

			msgCmp := fmt.Sprintf(messageCompareIdle, workloadTask.JobName)
			err := HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))

			workloadTask.StartTime = timeNow.Add(time.Second * time.Duration(-noteIdleIWTime))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(2))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg := <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp)).To(BeTrue())

			workloadTask.StartTime = timeNow
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))

			workloadTask.StartTime = timeNow.Add(time.Second * time.Duration(-noteIdleIWTime))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(2))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg = <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp)).To(BeTrue())
		})

		It("Interactive Workload: idle limit not affecting workload without GPU request", func() {
			noteJobNameId := noteJobName + strconv.Itoa(uuidCounter)
			noteIwNameUnique := noteIwName + strconv.Itoa(uuidCounter)
			noteIwId := strconv.Itoa(uuidCounter)
			jobUuid := strconv.Itoa(uuidCounter)
			promQuerier := mocks.NewMockQuerier(mockCtrl)

			workloadCache := util.WorkloadCache{}

			timeNow := time.Now()
			jobStartTime := metav1.Time{Time: timeNow}
			iw2 := test.BuildInteractiveWorkload(testNamespaceName, noteIwNameUnique, &jobStartTime)
			_ = k8sClient.Create(ctx, iw2)

			actionList1 := NewActionList(projectQueue.EventList, 1, 1, nil)

			projectQueue = ProjectQueue{
				ProjectName:                       noteProjName,
				queue:                             priorityQueue,
				Workers:                           nil,
				InteractiveJobMaxIdleDurationSecs: noteIdleIWTime,
				TrainingJobMaxIdleDurationSecs:    noteIdleTWTime,
				K8sClient:                         k8sClient,
				PromQuerier:                       promQuerier,
				Recorder:                          &recorder,
				EventList:                         eventList,
				WorkloadValidator:                 &validation.WorkloadValidator{},
			}

			workloadTask := &WorkloadTask{ // just a dummy WorkloadTask
				StartTime:            timeNow,
				WorkloadName:         noteIwNameUnique,
				WorkloadID:           types.UID(noteIwId),
				ProjectName:          projectQueue.ProjectName,
				Namespace:            testNamespaceName,
				JobUuid:              types.UID(jobUuid),
				JobName:              noteJobNameId,
				WorkloadType:         constants.KindInteractiveWorkload,
				TaskAction:           *actionList1,
				WorkloadRequestedGpu: false,
			}

			err := HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))

			workloadTask.StartTime = timeNow.Add(time.Second * time.Duration(-noteIdleIWTime))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))

			workloadTask.StartTime = timeNow
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))

			workloadTask.StartTime = timeNow.Add(time.Second * time.Duration(-noteIdleIWTime))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, false)
			Expect(err).To(BeNil())
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTask.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(0))
		})

		It("Train Workload: check timeout of event synchronization", func() {
			noteJobNameId := noteJobName + strconv.Itoa(uuidCounter)
			noteIwNameId := noteIwName + strconv.Itoa(uuidCounter)
			jobIwUuid := strconv.Itoa(uuidCounter)
			uuidCounter++
			noteTwNameId := noteIwName + strconv.Itoa(uuidCounter)
			jobTwUuid := strconv.Itoa(uuidCounter)
			promQuerier := mocks.NewMockQuerier(mockCtrl)

			workloadCache := util.WorkloadCache{}

			timeNow := time.Now()
			jobStartTime := metav1.Time{Time: timeNow}
			tw1 := test.BuildTrainingWorkload(testNamespaceName, noteTwNameId, &jobStartTime)
			_ = k8sClient.Create(ctx, tw1)
			iw1 := test.BuildInteractiveWorkload(testNamespaceName, noteIwNameId, &jobStartTime)
			_ = k8sClient.Create(ctx, iw1)

			actionList1 := NewActionList(projectQueue.EventList, 1, 1, nil)
			actionList2 := NewActionList(projectQueue.EventList, 1, 1, nil)

			projectQueue = ProjectQueue{
				ProjectName:                 noteProjName,
				queue:                       priorityQueue,
				Workers:                     nil,
				InteractiveJobTimeLimitSecs: noteIdleIWTime,
				TrainingJobTimeLimitSecs:    noteIdleTWTime * 3,
				K8sClient:                   k8sClient,
				PromQuerier:                 promQuerier,
				Recorder:                    &recorder,
				EventList:                   eventList,
				WorkloadValidator:           &validation.WorkloadValidator{},
			}

			workloadTaskIw := &WorkloadTask{ // just a dummy WorkloadTask
				StartTime:    timeNow,
				WorkloadName: noteIwNameId,
				ProjectName:  projectQueue.ProjectName,
				Namespace:    testNamespaceName,
				JobUuid:      types.UID(jobIwUuid),
				JobName:      noteJobNameId,
				WorkloadType: constants.KindInteractiveWorkload,
				TaskAction:   *actionList1,
			}

			workloadTaskTw := &WorkloadTask{ // just a dummy WorkloadTask
				StartTime:    timeNow,
				WorkloadName: noteTwNameId,
				ProjectName:  projectQueue.ProjectName,
				Namespace:    testNamespaceName,
				JobUuid:      types.UID(jobTwUuid),
				JobName:      noteJobNameId,
				WorkloadType: constants.KindTrainingWorkload,
				TaskAction:   *actionList2,
			}

			msgCmp := fmt.Sprintf(messageCompareTimeout, workloadTaskIw.JobName)

			time50 := noteIdleIWTime/2 + 10
			workloadTaskIw.StartTime = timeNow.Add(time.Second * time.Duration(-time50))
			workloadTaskTw.StartTime = timeNow.Add(time.Second * time.Duration(-time50))
			err := HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTaskIw, &projectQueue, false)
			Expect(err).To(BeNil())
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTaskTw, &projectQueue, false)
			Expect(err).To(BeNil())

			Expect(workloadTaskIw.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTaskIw.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(2))
			Expect(workloadTaskTw.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTaskTw.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg := <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp+"49 days - 23 hours - 5")).To(BeTrue())

			time75 := noteIdleIWTime*.75 + 10
			workloadTaskIw.StartTime = timeNow.Add(time.Second * time.Duration(-time75))
			workloadTaskTw.StartTime = timeNow.Add(time.Second * time.Duration(-time75))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTaskIw, &projectQueue, false)
			Expect(err).To(BeNil())
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTaskTw, &projectQueue, false)
			Expect(err).To(BeNil())

			Expect(workloadTaskIw.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTaskIw.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(3))
			Expect(workloadTaskTw.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTaskTw.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg = <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp+"24 days - 23 hours - 5")).To(BeTrue())

			time90 := noteIdleIWTime*.90 + 10
			workloadTaskIw.StartTime = timeNow.Add(time.Second * time.Duration(-time90))
			workloadTaskTw.StartTime = timeNow.Add(time.Second * time.Duration(-time90))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTaskIw, &projectQueue, false)
			Expect(err).To(BeNil())
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTaskTw, &projectQueue, false)
			Expect(err).To(BeNil())

			Expect(workloadTaskIw.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTaskIw.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(4))
			Expect(workloadTaskTw.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTaskTw.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg = <-recorder.Events
			Expect(strings.HasPrefix(msg, msgCmp+"9 days - 23 hours - 5")).To(BeTrue())

			timedOut := noteIdleIWTime + 10
			workloadTaskIw.StartTime = timeNow.Add(time.Second * time.Duration(-timedOut))
			workloadTaskTw.StartTime = timeNow.Add(time.Second * time.Duration(-timedOut))
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTaskIw, &projectQueue, false)
			Expect(err).To(BeNil())
			err = HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTaskTw, &projectQueue, false)
			Expect(err).To(BeNil())

			Expect(workloadTaskIw.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTaskIw.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(4))
			Expect(workloadTaskTw.TaskAction.SetEventType(TaskTimerIdle).GetIndex()).To(Equal(1))
			Expect(workloadTaskTw.TaskAction.SetEventType(TaskTimerTimeout).GetIndex()).To(Equal(1))
			Expect(len(recorder.Events)).To(Equal(1))
			msg = <-recorder.Events
			Expect(msg).To(Equal(messageTimeoutReached))
		})

		It("WorkloadRunner: vanished runner at deactivation is skipped, not falsely stopped", func() {
			// Runner absent at deactivation must be skipped, not falsely stopped.
			promQuerier := mocks.NewMockQuerier(mockCtrl)
			workloadCache := util.WorkloadCache{}

			projectQueue = ProjectQueue{
				ProjectName:                 noteProjName,
				queue:                       priorityQueue,
				InteractiveJobTimeLimitSecs: noteIdleIWTime,
				K8sClient:                   k8sClient,
				PromQuerier:                 promQuerier,
				Recorder:                    &recorder,
				EventList:                   eventList,
				WorkloadValidator:           &validation.WorkloadValidator{},
			}

			// Point both indices at the DeactivateEvent percentile so the past-due
			// task takes the deactivation branch (the one that would suspend).
			lastIdx := len(eventList) - 1
			actionList := NewActionList(projectQueue.EventList, lastIdx, lastIdx, nil)

			workloadTask := &WorkloadTask{
				StartTime:    time.Now().Add(-time.Second * time.Duration(noteIdleIWTime+10)),
				WorkloadName: "vanished-runner-" + strconv.Itoa(uuidCounter),
				ProjectName:  projectQueue.ProjectName,
				Namespace:    testNamespaceName,
				JobName:      "vanished-runner-" + strconv.Itoa(uuidCounter),
				WorkloadType: constants.KindWorkloadRunner,
				Category:     constants.KindInteractiveWorkload,
				TaskAction:   *actionList,
			}

			err := HandleWorkloadTask(ctx, &workloadCache, k8sClient, promQuerier, workloadTask, &projectQueue, true)
			Expect(err).To(BeNil())
			Expect(len(recorder.Events)).To(Equal(0))
		})
	})
})

func initSchema() *apiruntime.Scheme {
	scheme := scheme_init.Scheme()
	err := mpi.AddToScheme(scheme)
	if err != nil {
		zap.S().Error("Failed to add mpijob to scheme", err)
	}
	err = iwv2a1.AddKnownTypes(scheme)
	if err != nil {
		zap.S().Error("Failed to add interactive workload types", err)
	}
	err = iwv2a1.AddToScheme(scheme)
	if err != nil {
		zap.S().Error("Failed to add interactive workload to scheme", err)
	}
	err = twv2a1.AddKnownTypes(scheme)
	if err != nil {
		zap.S().Error("Failed to add training workload types", err)
	}
	err = twv2a1.AddToScheme(scheme)
	if err != nil {
		zap.S().Error("Failed to add training workload to scheme", err)
	}
	return scheme
}
