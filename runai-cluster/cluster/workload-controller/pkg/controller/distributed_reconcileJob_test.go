package controller_test

import (
	"bytes"
	"context"
	"fmt"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	fields "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
	"net/http"
	"reflect"
	"strings"

	nodepoolbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders/nodepool"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/jax"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftV1 "github.com/openshift/api/route/v1"
	clusterCommonConstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	controllerutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	jobs2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/jobs"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	mpiv2beta1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/mpi"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/pytorch"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/tfjob"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/xgboost"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/managers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/aggregator"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	workloadUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/convertors"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	kubeResource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	projectName  = "test"
	nodepoolName = "test"

	jobName = "distributed-job"

	testUid = "test-uid"

	nodepoolLabelKey   = "runai/node-pool"
	nodepoolLabelValue = "test"
	storageName        = "storage"
	pvcSize            = "1Gi"
)

type testJobPhase int

const (
	creating testJobPhase = iota
	running
	succeeded
	failed
	pending
	initializing
	degraded
	suspended
	suspending
)

const (
	nonMPIMasterReplicaType kubeflowv1.ReplicaType = "master"
	workerReplicaType                              = "worker"
	MpiMasterReplicaType                           = "launcher"
	JaxMasterReplicaType                           = "none"
)

var (
	TestNamespace = fmt.Sprintf("runai-%s", projectName)
)

type TestEntryContainer struct {
	workloadControllerFactory func(client3 client.Client, directReader client.Reader, isOpenShift bool,
		eventRecorder record.EventRecorder, podLabels map[string]string,
		podAnnotations map[string]string, controllerOptions controller.Options) *controller.DistributedWorkloadControllerFactory
	getJobWorkloadFn  func() *jobs.JobWorkload
	jobConvertor      convertors.Convertor
	getWorkloadFn     func() client.Object
	getJobFn          func() client.Object
	masterReplicaType kubeflowv1.ReplicaType
	getStatusFn       func(object client.Object) *kubeflowv1.JobStatus
	isOpenShift       bool
}

func baseDistributedWorkload(jobType string) *jobs.JobWorkload {
	pvc := resource.CreatePvc(jobName, TestNamespace,
		kubeResource.MustParse(pvcSize), testUid, clusterCommonConstants.KindDistributedWorkload, storageName, nil, nil, nil)

	distributeBuilder := workloadUtils.NewJobWorkloadBuilder().WithWorkers(10).WithPorts([]string{"8888"}).
		WithExternalUrl([]string{"http://test.com"}, []int{1000}).
		WithDistribution().WithJobType(jobType).WithDefaultServiceType(clusterCommonConstants.ServiceTypeNodePort).
		WithUid(testUid).WithName(jobName).WithNodePools(nodepoolName).WithPVC(pvc, false, false).
		WithName(jobName).WithNamespace(TestNamespace)

	// jax does not support master
	if jobType != clusterCommonConstants.KindJAXJob {
		masterBuilder := workloadUtils.NewJobWorkloadBuilder().WithImage("alpine").
			WithJobType(jobType).WithNamespace(TestNamespace).WithPorts([]string{"8888"}).
			WithExternalUrl([]string{"http://test.com"}, []int{1000}).
			WithNodePools(nodepoolName).WithCommand("sleep 1000").
			WithDefaultServiceType(clusterCommonConstants.ServiceTypeNodePort)

		distributeBuilder = distributeBuilder.WithMasterSpec(&(masterBuilder.Build().Spec))
	}

	return distributeBuilder.Build()
}

func pytorchDistributedWorkload() *jobs.JobWorkload {
	return baseDistributedWorkload(clusterCommonConstants.KindPyTorchJob)
}
func mpiDistributedWorkload() *jobs.JobWorkload {
	return baseDistributedWorkload(clusterCommonConstants.KindMpiJob)
}

func tfDistributedWorkload() *jobs.JobWorkload {
	return baseDistributedWorkload(clusterCommonConstants.KindTFJob)
}

func xgboostDistributedWorkload() *jobs.JobWorkload {
	return baseDistributedWorkload(clusterCommonConstants.KindXGBoostJob)
}

func jaxDistributedWorkload() *jobs.JobWorkload {
	return baseDistributedWorkload(clusterCommonConstants.KindJAXJob)
}

func pytorchGetStatus(pytorchJob client.Object) *kubeflowv1.JobStatus {
	return &pytorchJob.(*kubeflowv1.PyTorchJob).Status

}

func mpiGetStatus(mpiJob client.Object) *kubeflowv1.JobStatus {
	return &mpiJob.(*mpiv2beta1.MPIJob).Status
}

func tfGetStatus(tfJob client.Object) *kubeflowv1.JobStatus {
	return &tfJob.(*kubeflowv1.TFJob).Status
}

func xgboostGetStatus(xgboostJob client.Object) *kubeflowv1.JobStatus {
	return &xgboostJob.(*kubeflowv1.XGBoostJob).Status
}

func jaxGetStatus(jaxJob client.Object) *kubeflowv1.JobStatus {
	return &jaxJob.(*kubeflowv1.JAXJob).Status
}

func getEventRecorder() record.EventRecorder {
	return record.NewFakeRecorder(100)
}

var _ = Describe("DistributedReconcileJob", func() {
	var (
		testEntries           []TableEntry
		fakeClientPtr         *client2.FakeClientWithErrorRules
		k8sClient             client.Client
		nodepool              *v1alpha1.NodePool
		nodepoolKai           *kaiorg.NodePool
		ctx                   = context.Background()
		defaultPodLabels      map[string]string
		defaultPodAnnotations map[string]string
		controllerOptions     = controller.Options{FailureResourceCleanupPolicy: managers.GetDefaultFailureCleanupPolicy(), EnableDanglingResourceCleanup: true}
		controllerLogsBuffer  *bytes.Buffer
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)
		project := resource.CreateProject(projectName)
		namespace := resource.CreateNamespace(TestNamespace, projectName)

		nodepool = resource.CreateNodePool(nodepoolName, nodepoolLabelKey, nodepoolLabelValue)

		nodepoolKai = resource.CreateKaiNodePool(nodepoolName, nodepoolLabelKey, nodepoolLabelValue)

		fakeClientPtr = client2.NewFakeClientWithErrorRules(scheme, []client.Object{&v2alpha1.DistributedWorkload{}, &kubeflowv1.PyTorchJob{},
			&kubeflowv1.TFJob{}, &kubeflowv1.XGBoostJob{}, &mpiv2beta1.MPIJob{}, &kubeflowv1.JAXJob{}}, nil, project, namespace, nodepool, nodepoolKai)
		k8sClient = fakeClientPtr

		controllerLogsBuffer = bytes.NewBuffer(nil)
		fakeLogger := zap.New(zap.WriteTo(controllerLogsBuffer))
		ctrl.SetLogger(fakeLogger)
		ctx = logf.IntoContext(ctx, fakeLogger)
	})

	testEntries = []TableEntry{
		Entry("pytorch job", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          pytorchDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  pytorch.NewPytorchJobFn,
			masterReplicaType:         nonMPIMasterReplicaType,
			getStatusFn:               pytorchGetStatus,
			isOpenShift:               false,
		}, func() client.Client {
			return k8sClient
		}),
		Entry("mpi job", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          mpiDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  mpi.NewMpiJobFn,
			isOpenShift:               false,
			masterReplicaType:         MpiMasterReplicaType,
			getStatusFn:               mpiGetStatus,
		}, func() client.Client {
			return k8sClient
		}),
		Entry("tf job", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          tfDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  tfjob.NewTFJobFn,
			isOpenShift:               false,
			masterReplicaType:         nonMPIMasterReplicaType,
			getStatusFn:               tfGetStatus,
		}, func() client.Client {
			return k8sClient
		}),
		Entry("xgboost job", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          xgboostDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  xgboost.NewXGBoostJobFn,
			isOpenShift:               false,
			masterReplicaType:         nonMPIMasterReplicaType,
			getStatusFn:               xgboostGetStatus,
		}, func() client.Client {
			return k8sClient
		}),
		Entry("pytorch job openshift", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          pytorchDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  pytorch.NewPytorchJobFn,
			isOpenShift:               true,
			masterReplicaType:         nonMPIMasterReplicaType,
			getStatusFn:               pytorchGetStatus,
		}, func() client.Client {
			return k8sClient
		}),
		Entry("mpi job openshift", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          mpiDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  mpi.NewMpiJobFn,
			isOpenShift:               true,
			masterReplicaType:         MpiMasterReplicaType,
			getStatusFn:               mpiGetStatus,
		}, func() client.Client {
			return k8sClient
		}),
		Entry("tf job openshift", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          tfDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  tfjob.NewTFJobFn,
			isOpenShift:               true,
			masterReplicaType:         nonMPIMasterReplicaType,
			getStatusFn:               tfGetStatus,
		}, func() client.Client {
			return k8sClient
		}),
		Entry("xgboost job openshift", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          xgboostDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  xgboost.NewXGBoostJobFn,
			isOpenShift:               true,
			masterReplicaType:         nonMPIMasterReplicaType,
			getStatusFn:               xgboostGetStatus,
		}, func() client.Client {
			return k8sClient
		}),
		Entry("jax job", &TestEntryContainer{
			workloadControllerFactory: controller.NewDistributedWorkloadControllerFactory,
			getJobWorkloadFn:          jaxDistributedWorkload,
			jobConvertor:              convertors.NewDistributedWorkloadConvertor(),
			getWorkloadFn:             v2alpha1.GetNewDistributedWorkloadFn,
			getJobFn:                  jax.NewJAXJobFn,
			isOpenShift:               false,
			masterReplicaType:         JaxMasterReplicaType,
			getStatusFn:               jaxGetStatus,
		}, func() client.Client {
			return k8sClient
		}),
	}
	Context("First Reconcile", func() {

		DescribeTable("check workload flow", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("Creating a new DistributedWorkload")
			workload := testEntry.getJobWorkloadFn()
			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			concreteWorkload := testEntry.jobConvertor.Convert(workload)
			err = k8sClient.Create(ctx, concreteWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was updated with first ack")
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
			By("Verifying that the DistributedWorkload was updated with first ack")
			Expect(distributedWorkload.Status.FirstAcknowledgedTime.IsZero()).To(BeFalse())
			By("Verifying resources were created")
			VerifyWorkloadResources(ctx, testEntry, k8sClient, concreteWorkload)
			By("Verifying phase is 'Creating'")
			Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseCreating.String()))
		}, testEntries)

		Context("Init context", func() {
			DescribeTable("failure in first reconcile will fail workload", func(testEntry *TestEntryContainer, getClient func() client.Client) {
				By("Creating a new DistributedWorkload with the MNNVL feature label")
				workload := testEntry.getJobWorkloadFn()
				workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.Required
				k8sClient := getClient()
				workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
					testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
				workloadController, err := workloadControllerFactory.CreateController()
				Expect(err).ToNot(HaveOccurred())
				Expect(workloadController).ToNot(BeNil())
				concreteWorkload := testEntry.jobConvertor.Convert(workload)
				Expect(k8sClient.Create(ctx, concreteWorkload)).To(Succeed())
				By("Verifying that the DistributedWorkload was created")
				createdWorkload := testEntry.getWorkloadFn()
				err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
				Expect(err).ToNot(HaveOccurred())
				By("running the reconcile loop")
				_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
				Expect(err).ToNot(HaveOccurred())
				By("Verifying that the DistributedWorkload was updated with first ack")
				err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
				Expect(err).ToNot(HaveOccurred())
				distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
				By("Verifying that the DistributedWorkload was updated with first ack")
				Expect(distributedWorkload.Status.FirstAcknowledgedTime.IsZero()).To(BeFalse())

				By("Verifying phase is 'Failed'")
				Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseFailed.String()))
				Expect(distributedWorkload.Status.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to a Internal issue"))
				Expect(distributedWorkload.Status.WorkloadCommonStatus.Issues.Internal).To(HaveLen(1))
				Expect(distributedWorkload.Status.WorkloadCommonStatus.Issues.Internal[0].Description).To(ContainSubstring(mnnvl.UserFacingFeatureName))
			}, testEntries)
			DescribeTable("failure in second reconcile will not workload", func(testEntry *TestEntryContainer, getClient func() client.Client) {
				By("Creating a new DistributedWorkload")
				workload := testEntry.getJobWorkloadFn()
				k8sClient := getClient()
				workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
					testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
				workloadController, err := workloadControllerFactory.CreateController()
				Expect(err).ToNot(HaveOccurred())
				Expect(workloadController).ToNot(BeNil())
				concreteWorkload := testEntry.jobConvertor.Convert(workload)
				err = k8sClient.Create(ctx, concreteWorkload)
				Expect(err).ToNot(HaveOccurred())
				By("Verifying that the DistributedWorkload was created")
				createdWorkload := testEntry.getWorkloadFn()
				err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
				Expect(err).ToNot(HaveOccurred())
				By("running the reconcile loop")
				_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
				Expect(err).ToNot(HaveOccurred())
				By("Verifying that the DistributedWorkload was updated with first ack")
				err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
				Expect(err).ToNot(HaveOccurred())
				distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
				By("Verifying that the DistributedWorkload was updated with first ack")
				Expect(distributedWorkload.Status.FirstAcknowledgedTime.IsZero()).To(BeFalse())
				By("Verifying resources were created")
				VerifyWorkloadResources(ctx, testEntry, k8sClient, concreteWorkload)
				By("Verifying phase is 'Creating'")
				Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseCreating.String()))

				By("Adding the MNNVL feature label and run the reconcile again")
				if distributedWorkload.Labels == nil {
					distributedWorkload.Labels = make(map[string]string)
				}
				distributedWorkload.Labels[mnnvl.FeatureLabelKey] = mnnvl.Required
				distributedWorkload.CopyTo(workload)
				_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(mnnvl.UserFacingFeatureName))

				By("Verifying phase is still 'Creating' and not 'Failed")
				Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseCreating.String()))
			}, testEntries)
		})
	})
	Context("Reconcile with all existing resources", func() {
		DescribeTable("check workload flow with two recon", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("setup")
			workload := testEntry.getJobWorkloadFn()
			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			By("Creating a new DistributedWorkload")
			concreteWorkload := testEntry.jobConvertor.Convert(workload)

			err = k8sClient.Create(ctx, concreteWorkload)
			labels := map[string]string{clusterCommonConstants.LabelRelease: workload.Name}
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Creating services")
			service := resource.CreateDumbService(fmt.Sprintf("dw-%s-0-%s", jobName,
				strings.ToLower(clusterCommonConstants.ServiceTypeNodePort)), TestNamespace, concreteWorkload.GetUID(),
				workload.Kind, labels)
			Expect(k8sClient.Create(ctx, service)).To(Succeed())
			serviceIngress := resource.CreateDumbService(fmt.Sprintf("dw-%s-0", jobName), TestNamespace, concreteWorkload.GetUID(),
				workload.Kind, labels)
			Expect(k8sClient.Create(ctx, serviceIngress)).To(Succeed())
			By("Creating externalUrl")
			externalUrl := createExternalUrl(testEntry.isOpenShift, jobName, TestNamespace, workload.Kind, concreteWorkload.GetUID())
			Expect(k8sClient.Create(ctx, externalUrl)).To(Succeed())
			By("Creating job")
			job := testEntry.getJobFn()
			job.SetNamespace(TestNamespace)
			job.SetName(jobName)
			err = controllerutils.SetControllerReference(createdWorkload, job, k8sClient.Scheme())
			Expect(err).ToNot(HaveOccurred())

			Expect(k8sClient.Create(ctx, job)).To(Succeed())

			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop again")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())

			VerifyWorkloadResources(ctx, testEntry, k8sClient, concreteWorkload)
		}, testEntries)
		DescribeTable("recreate all resources after deleting them ", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("setup")
			workload := testEntry.getJobWorkloadFn()
			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			concreteWorkload := testEntry.jobConvertor.Convert(workload)

			By("Creating a new DistributedWorkload")
			err = k8sClient.Create(ctx, concreteWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Running the reconcile")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Deleting all resources")
			DeleteAllResources(k8sClient, ctx)
			By("Running the reconcile again")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			VerifyWorkloadResources(ctx, testEntry, k8sClient, concreteWorkload)

		}, testEntries)
		DescribeTable("upgrading resource without owner but with owner label", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("setup")
			workload := testEntry.getJobWorkloadFn()
			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			By("Creating a new DistributedWorkload")
			concreteWorkload := testEntry.jobConvertor.Convert(workload)

			err = k8sClient.Create(ctx, concreteWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			labels := map[string]string{clusterCommonConstants.LabelRelease: workload.Name}
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			By("Creating services")
			service := resource.CreateDumbService(fmt.Sprintf("%s-%s", jobName,
				strings.ToLower(clusterCommonConstants.ServiceTypeNodePort)), TestNamespace, "", "", labels)
			SetLabelOwnership(service, jobName)
			Expect(k8sClient.Create(ctx, service)).To(Succeed())
			serviceIngress := resource.CreateDumbService(jobName, TestNamespace, concreteWorkload.GetUID(), workload.Kind, labels)
			SetLabelOwnership(serviceIngress, jobName)
			Expect(k8sClient.Create(ctx, serviceIngress)).To(Succeed())
			By("Creating externalUrl")
			externalUrl := createExternalUrl(testEntry.isOpenShift, jobName, TestNamespace, "", "")
			SetLabelOwnership(externalUrl, jobName)
			Expect(k8sClient.Create(ctx, externalUrl)).To(Succeed())
			By("Creating job")
			job := testEntry.getJobFn()
			job.SetNamespace(TestNamespace)
			job.SetName(jobName)
			SetLabelOwnership(job, jobName)
			Expect(err).ToNot(HaveOccurred())

			Expect(k8sClient.Create(ctx, job)).To(Succeed())

			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			VerifyWorkloadResources(ctx, testEntry, k8sClient, concreteWorkload)
		}, testEntries)
		DescribeTable("Nodepool change after creation but node affinity stay the same", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("setup")
			workload := testEntry.getJobWorkloadFn()

			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			By("Creating a new DistributedWorkload")
			concreteWorkload := testEntry.jobConvertor.Convert(workload)

			err = k8sClient.Create(ctx, concreteWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			// check the job node affinity
			createdJob := testEntry.getJobFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
			Expect(err).ToNot(HaveOccurred())
			jobAdapterFactory, err := jobs2.GetJobAdapterFactory(createdJob, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			jobAdapter, err := jobAdapterFactory.CreateJobAdapter(createdJob)
			Expect(err).ToNot(HaveOccurred())

			By("Changing nodepool label")
			exitingNodepool := v1alpha1.NodePool{}
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(nodepool), &exitingNodepool)
			Expect(err).ToNot(HaveOccurred())
			originalNodepool := exitingNodepool.DeepCopy()
			originalAffinity := &v1.Affinity{}
			err = nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, k8sClient, "", originalNodepool.Name, &originalAffinity)
			Expect(err).ToNot(HaveOccurred())
			exitingNodepool.Spec.LabelKey = "modified"
			exitingNodepool.Spec.LabelValue = "modified"
			err = k8sClient.Update(ctx, &exitingNodepool)

			// Mirror the change onto the kai twin so the affinity assertions below hold under
			// both values of the compile-time UseKaiNodePools flag.
			// TODO: drop once constants.UseKaiNodePools=true and the legacy fixture goes
			existingKaiNodepool := kaiorg.NodePool{}
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(nodepoolKai), &existingKaiNodepool)).To(Succeed())
			existingKaiNodepool.Spec.LabelKey = "modified"
			existingKaiNodepool.Spec.LabelValue = "modified"
			Expect(k8sClient.Update(ctx, &existingKaiNodepool)).To(Succeed())
			modifiedNodepool := exitingNodepool.DeepCopy()
			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop again")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("varify job node affinity did not change")
			createdJob = testEntry.getJobFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
			Expect(err).ToNot(HaveOccurred())
			jobAdapterFactory, err = jobs2.GetJobAdapterFactory(createdJob, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			jobAdapter, err = jobAdapterFactory.CreateJobAdapter(createdJob)
			Expect(err).ToNot(HaveOccurred())
			modifiedAffinity := &v1.Affinity{}

			err = nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, k8sClient, "", modifiedNodepool.Name, &modifiedAffinity)
			Expect(err).ToNot(HaveOccurred())
			for _, replica := range jobAdapter.GetSpec().GetReplicaTypeSpecs() {
				spec := replica.GetPodTemplateSpec().Spec
				Expect(spec).ToNot(BeNil())
				Expect(reflect.DeepEqual(replica.GetPodTemplateSpec().Spec.Affinity.NodeAffinity, originalAffinity.NodeAffinity)).To(BeTrue())
				Expect(reflect.DeepEqual(replica.GetPodTemplateSpec().Spec.Affinity.NodeAffinity, modifiedAffinity.NodeAffinity)).To(BeFalse())
			}

		}, testEntries)
		DescribeTable("job is not changed between 4 reconciles", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("setup")
			workload := testEntry.getJobWorkloadFn()
			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			By("Creating a new DistributedWorkload")
			concreteWorkload := testEntry.jobConvertor.Convert(workload)

			err = k8sClient.Create(ctx, concreteWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			createdWorkload = testEntry.jobConvertor.Convert(workload)
			createdJob := testEntry.getJobFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
			Expect(err).ToNot(HaveOccurred())
			for i := 0; i < 4; i++ {
				By("running the reconcile loop again")
				_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
				Expect(err).ToNot(HaveOccurred())
				createdWorkload = testEntry.jobConvertor.Convert(workload)
				currentJob := testEntry.getJobFn()
				By("varify job did not change")
				err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), currentJob)
				Expect(err).ToNot(HaveOccurred())
				Expect(reflect.DeepEqual(GetJobSpec(createdJob), GetJobSpec(currentJob))).To(BeTrue())
			}
		}, testEntries)
		DescribeTable("Nodepool change after creating with status first Ack is nil, node affinity should stay the same", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("setup")
			workload := testEntry.getJobWorkloadFn()
			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			By("Creating a new DistributedWorkload")
			concreteWorkload := testEntry.jobConvertor.Convert(workload)

			err = k8sClient.Create(ctx, concreteWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Changing nodepool label")
			exitingNodepool := v1alpha1.NodePool{}
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(nodepool), &exitingNodepool)
			Expect(err).ToNot(HaveOccurred())
			originalNodepool := exitingNodepool.DeepCopy()
			originalAffinity := &v1.Affinity{}
			err = nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, k8sClient, "", originalNodepool.Name, &originalAffinity)
			Expect(err).ToNot(HaveOccurred())
			exitingNodepool.Spec.LabelKey = "modified"
			exitingNodepool.Spec.LabelValue = "modified"
			err = k8sClient.Update(ctx, &exitingNodepool)

			// Mirror the change onto the kai twin so the affinity assertions below hold under
			// both values of the compile-time UseKaiNodePools flag.
			// TODO: drop once constants.UseKaiNodePools=true and the legacy fixture goes
			existingKaiNodepool := kaiorg.NodePool{}
			Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(nodepoolKai), &existingKaiNodepool)).To(Succeed())
			existingKaiNodepool.Spec.LabelKey = "modified"
			existingKaiNodepool.Spec.LabelValue = "modified"
			Expect(k8sClient.Update(ctx, &existingKaiNodepool)).To(Succeed())
			modifiedNodepool := exitingNodepool.DeepCopy()
			Expect(err).ToNot(HaveOccurred())
			By("remove first acknowledge time")
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			workload = testEntry.getJobWorkloadFn()
			createdWorkload = testEntry.jobConvertor.Convert(workload)
			By("running the reconcile loop again")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("varify first acknowledge time is not nil")
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			Expect(createdWorkload.(*v2alpha1.DistributedWorkload).Status.FirstAcknowledgedTime.IsZero()).To(BeFalse())
			By("running the reconcile loop again")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("varify job node affinity did not change")
			createdJob := testEntry.getJobFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
			Expect(err).ToNot(HaveOccurred())
			jobAdapterFactory, err := jobs2.GetJobAdapterFactory(createdJob, k8sClient)
			Expect(err).ToNot(HaveOccurred())
			jobAdapter, err := jobAdapterFactory.CreateJobAdapter(createdJob)
			Expect(err).ToNot(HaveOccurred())
			modifiedAffinity := &v1.Affinity{}

			err = nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, k8sClient, "", modifiedNodepool.Name, &modifiedAffinity)
			Expect(err).ToNot(HaveOccurred())
			for _, replica := range jobAdapter.GetSpec().GetReplicaTypeSpecs() {
				Expect(reflect.DeepEqual(replica.GetPodTemplateSpec().Spec.Affinity.NodeAffinity, originalAffinity.NodeAffinity)).To(BeTrue())
				Expect(reflect.DeepEqual(replica.GetPodTemplateSpec().Spec.Affinity.NodeAffinity, modifiedAffinity.NodeAffinity)).To(BeFalse())
			}
		}, testEntries)
	})
	DescribeTable("job is pending", func(testEntry *TestEntryContainer, getClient func() client.Client) {
		By("setup")
		workload := testEntry.getJobWorkloadFn()
		k8sClient := getClient()
		workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
			testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
		workloadController, err := workloadControllerFactory.CreateController()
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadController).ToNot(BeNil())
		By("Creating a new DistributedWorkload")
		concreteWorkload := testEntry.jobConvertor.Convert(workload)

		err = k8sClient.Create(ctx, concreteWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("Verifying that the DistributedWorkload was created")
		createdWorkload := testEntry.getWorkloadFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)

		Expect(err).ToNot(HaveOccurred())
		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("set job to pending")
		createdJob := testEntry.getJobFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			pending, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
		for _, pod := range pods {
			err = k8sClient.Create(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status := testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("varify workload phase is pending")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhasePending.String()))
		By("varify job condition")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Degraded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Succeeded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Failed, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "job is not running")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Initialized, "")
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Pending)
	}, testEntries)
	DescribeTable("job is running", func(testEntry *TestEntryContainer, getClient func() client.Client) {
		By("setup")
		workload := testEntry.getJobWorkloadFn()
		k8sClient := getClient()
		workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
			testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
		workloadController, err := workloadControllerFactory.CreateController()
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadController).ToNot(BeNil())
		By("Creating a new DistributedWorkload")
		concreteWorkload := testEntry.jobConvertor.Convert(workload)

		err = k8sClient.Create(ctx, concreteWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("Verifying that the DistributedWorkload was created")
		createdWorkload := testEntry.getWorkloadFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)

		Expect(err).ToNot(HaveOccurred())
		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("set job to running")
		createdJob := testEntry.getJobFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			running, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
		for _, pod := range pods {
			err = k8sClient.Create(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status := testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("varify workload phase is running")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseRunning.String()))
		By("varify job condition")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Degraded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Succeeded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Failed, "")
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Ready)
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Initialized)
	}, testEntries)
	DescribeTable("job is initializing", func(testEntry *TestEntryContainer, getClient func() client.Client) {
		By("setup")

		workload := testEntry.getJobWorkloadFn()
		k8sClient := getClient()
		workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
			testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
		workloadController, err := workloadControllerFactory.CreateController()
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadController).ToNot(BeNil())
		By("Creating a new DistributedWorkload")
		concreteWorkload := testEntry.jobConvertor.Convert(workload)

		err = k8sClient.Create(ctx, concreteWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("Verifying that the DistributedWorkload was created")
		createdWorkload := testEntry.getWorkloadFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)

		Expect(err).ToNot(HaveOccurred())
		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("set job to running")
		createdJob := testEntry.getJobFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			initializing, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
		for _, pod := range pods {
			err = k8sClient.Create(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status := testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("varify workload phase is initializing")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseInitializing.String()))
		By("varify job condition")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Degraded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Succeeded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Failed, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "job is not running")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Initialized, "")
	}, testEntries)
	DescribeTable("job is suspended", func(testEntry *TestEntryContainer, getClient func() client.Client) {
		By("setup")
		workload := testEntry.getJobWorkloadFn()
		k8sClient := getClient()
		workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
			testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
		workloadController, err := workloadControllerFactory.CreateController()
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadController).ToNot(BeNil())
		By("Creating a new DistributedWorkload")
		concreteWorkload := testEntry.jobConvertor.Convert(workload)

		err = k8sClient.Create(ctx, concreteWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("Verifying that the DistributedWorkload was created")
		createdWorkload := testEntry.getWorkloadFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("set job to running")
		createdJob := testEntry.getJobFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			running, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
		for _, pod := range pods {
			err = k8sClient.Create(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status := testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("verify workload phase is running")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseRunning.String()))
		By("verify job condition")
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Initialized)
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Ready)
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Suspended, "")
		By("set workload desired state to suspended")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		distributedWorkload.Spec.Active = fields.NewOptionalBoolField(ptr.To(false))
		err = k8sClient.Update(ctx, distributedWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload.CopyTo(workload)
		By("set job to suspending")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		createdJob.SetUID(testUid)
		Expect(err).ToNot(HaveOccurred())
		modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			suspending, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())

		Expect(len(pods) > 0).To(BeTrue())
		remainingPod := pods[0]
		for _, pod := range pods[1:] {
			err = k8sClient.Delete(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status = testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("verify workload phase is not yet suspended")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseSuspending.String()))
		By("verify workload suspending")
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Initialized)
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "workload is suspending")
		suspendedCondition := workloadUtils.GetCondition(distributedWorkload.Status.Conditions, workloadcommon.Suspended)
		Expect(suspendedCondition).NotTo(BeNil())
		Expect(suspendedCondition.Status).To(Equal(v1.ConditionFalse))
		Expect(suspendedCondition.Reason).To(Equal(workloadcommon.PodsStillRunningReason))

		Expect(controllerLogsBuffer.String()).NotTo(ContainSubstring(managers.CleaningResourcesOnSuspendMessage))

		By("set job to suspended")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		createdJob.SetUID(testUid)
		Expect(err).ToNot(HaveOccurred())
		modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			suspended, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())

		err = k8sClient.Delete(ctx, remainingPod)
		Expect(err).ToNot(HaveOccurred())

		status = testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("verify workload phase is suspended")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseSuspended.String()))
		By("verify workload suspended")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Initialized, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "")
		suspendedCondition = workloadUtils.GetCondition(distributedWorkload.Status.Conditions, workloadcommon.Suspended)
		Expect(suspendedCondition).NotTo(BeNil())
		Expect(suspendedCondition.Status).To(Equal(v1.ConditionTrue))
		Expect(suspendedCondition.Reason).To(Equal(""))

		Expect(controllerLogsBuffer.String()).To(ContainSubstring(managers.CleaningResourcesOnSuspendMessage))
	}, testEntries)
	DescribeTable("job is resumed", func(testEntry *TestEntryContainer, getClient func() client.Client) {
		By("setup")
		workload := testEntry.getJobWorkloadFn()
		k8sClient := getClient()
		workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
			testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
		workloadController, err := workloadControllerFactory.CreateController()
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadController).ToNot(BeNil())
		By("Creating a new DistributedWorkload")
		concreteWorkload := testEntry.jobConvertor.Convert(workload)

		err = k8sClient.Create(ctx, concreteWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("Verifying that the DistributedWorkload was created")
		createdWorkload := testEntry.getWorkloadFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("set job to running")
		createdJob := testEntry.getJobFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			running, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
		for _, pod := range pods {
			err = k8sClient.Create(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status := testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("verify workload phase is running")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseRunning.String()))
		By("verify job condition")
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Initialized)
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Ready)
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Suspended, "")
		By("set workload desired state to suspended")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		distributedWorkload.Spec.Active = fields.NewOptionalBoolField(ptr.To(false))
		err = k8sClient.Update(ctx, distributedWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload.CopyTo(workload)
		By("set job to suspending")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		createdJob.SetUID(testUid)
		Expect(err).ToNot(HaveOccurred())
		modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			suspending, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())

		Expect(len(pods) > 0).To(BeTrue())
		remainingPod := pods[0]
		for _, pod := range pods[1:] {
			err = k8sClient.Delete(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status = testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("verify workload phase is not yet suspended")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseSuspending.String()))
		By("verify workload suspending")
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Initialized)
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "workload is suspending")
		condition := workloadUtils.GetCondition(distributedWorkload.Status.Conditions, workloadcommon.Suspended)
		Expect(condition).NotTo(BeNil())
		Expect(condition.Status).To(Equal(v1.ConditionFalse))
		Expect(condition.Reason).To(Equal(workloadcommon.PodsStillRunningReason))

		Expect(controllerLogsBuffer.String()).NotTo(ContainSubstring(managers.CleaningResourcesOnSuspendMessage))

		By("set job to suspended")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		createdJob.SetUID(testUid)
		Expect(err).ToNot(HaveOccurred())
		modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			suspended, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())

		err = k8sClient.Delete(ctx, remainingPod)
		Expect(err).ToNot(HaveOccurred())

		status = testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("verify workload phase is suspended")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseSuspended.String()))
		By("verify workload suspended")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Initialized, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "")
		condition = workloadUtils.GetCondition(distributedWorkload.Status.Conditions, workloadcommon.Suspended)
		Expect(condition).NotTo(BeNil())
		Expect(condition.Status).To(Equal(v1.ConditionTrue))
		Expect(condition.Reason).To(Equal(""))
		Expect(controllerLogsBuffer.String()).To(ContainSubstring(managers.CleaningResourcesOnSuspendMessage))
		By("set workload desired state to active")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		distributedWorkload.Spec.Active = fields.NewOptionalBoolField(ptr.To(true))
		err = k8sClient.Update(ctx, distributedWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload.CopyTo(workload)
		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("verify workload phase is resuming")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseResuming.String()))
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Initialized, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "job suspended")
		By("set job to running")
		createdJob = testEntry.getJobFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		pods = modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			running, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
		for _, pod := range pods {
			err = k8sClient.Create(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status = testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("verify workload phase is running")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseRunning.String()))
	}, testEntries)
	DescribeTable("workload is degraded due to issues", func(testEntry *TestEntryContainer, getClient func() client.Client) {
		By("setup")
		workload := testEntry.getJobWorkloadFn()
		k8sClient := getClient()
		workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
			testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
		workloadController, err := workloadControllerFactory.CreateController()
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadController).ToNot(BeNil())
		By("Creating a new DistributedWorkload")
		concreteWorkload := testEntry.jobConvertor.Convert(workload)

		err = k8sClient.Create(ctx, concreteWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("Verifying that the DistributedWorkload was created")
		createdWorkload := testEntry.getWorkloadFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("set job to running")
		createdJob := testEntry.getJobFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			running, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
		for _, pod := range pods {
			err = k8sClient.Create(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status := testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("verify workload phase is running")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseRunning.String()))
		By("verify job condition")
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Initialized)
		workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Ready)
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Suspended, "")

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		status = testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("generate issues to cause degradation")
		rules := []client2.ResourceFailingRule{
			{
				ClientAction: client2.Update,
				ResourceKind: clusterCommonConstants.KindService,
				ResourceName: fmt.Sprintf("dw-%s-0", createdJob.GetName()),
				StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Bla", Code: http.StatusBadRequest}},
			},
		}
		fakeClientPtr.UpdateRules(rules)

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("upsert workload resources error"))

		By("verify workload phase is degraded")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseDegraded.String()))
		Expect(distributedWorkload.Status.PhaseMessage).To(Equal("Workload is in degraded state due to a Network issue"))

		By("verify conditions")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Suspended, "")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Succeeded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Failed, "")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Ready, aggregator.WorkloadHasIssuesMessage)
		workloadUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Initialized)
		workloadUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Degraded)
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Pending, "")
	}, testEntries)
	DescribeTable("workload is failed due to issues", func(testEntry *TestEntryContainer, getClient func() client.Client) {
		By("setup")
		workload := testEntry.getJobWorkloadFn()
		k8sClient := getClient()
		workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
			testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
		workloadController, err := workloadControllerFactory.CreateController()
		Expect(err).ToNot(HaveOccurred())
		Expect(workloadController).ToNot(BeNil())
		By("Creating a new DistributedWorkload")
		concreteWorkload := testEntry.jobConvertor.Convert(workload)

		err = k8sClient.Create(ctx, concreteWorkload)
		Expect(err).ToNot(HaveOccurred())
		By("Verifying that the DistributedWorkload was created")
		createdWorkload := testEntry.getWorkloadFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("set job to running")
		createdJob := testEntry.getJobFn()
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
			initializing, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
		for _, pod := range pods {
			err = k8sClient.Create(ctx, pod)
			Expect(err).ToNot(HaveOccurred())
		}

		status := testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("running the reconcile loop")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).ToNot(HaveOccurred())

		By("verify workload phase is initializing")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseInitializing.String()))
		By("verify job condition")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "")
		workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Initialized, "")

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
		Expect(err).ToNot(HaveOccurred())
		createdJob.SetUID(testUid)
		status = testEntry.getStatusFn(createdJob).DeepCopy()
		err = k8sClient.Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
		Expect(err).ToNot(HaveOccurred())

		status.DeepCopyInto(testEntry.getStatusFn(createdJob))
		err = k8sClient.Status().Update(ctx, createdJob)
		Expect(err).ToNot(HaveOccurred())

		By("generate issues to cause degradation")
		rules := []client2.ResourceFailingRule{
			{
				ClientAction: client2.Update,
				ResourceKind: clusterCommonConstants.KindService,
				ResourceName: fmt.Sprintf("dw-%s-0", createdJob.GetName()),
				StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Bla", Code: http.StatusBadRequest}},
			},
		}
		fakeClientPtr.UpdateRules(rules)

		By("running the reconcile loop again")
		_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("upsert workload resources error"))

		By("verify workload phase is failed")
		err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
		Expect(err).ToNot(HaveOccurred())
		distributedWorkload = createdWorkload.(*v2alpha1.DistributedWorkload)
		Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseFailed.String()))
		Expect(distributedWorkload.Status.PhaseMessage).To(Equal("Workload failed due to a Network issue"))

		By("verify conditions")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Suspended, "")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Succeeded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Ready, aggregator.WorkloadHasIssuesMessage)
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Initialized, "")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Degraded, "")
		workloadUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, workloadcommon.Pending, "")

		failedCondition := workloadUtils.GetCondition(distributedWorkload.Status.Conditions, workloadcommon.Failed)
		Expect(failedCondition).NotTo(BeNil())
		Expect(failedCondition.Status).To(Equal(v1.ConditionTrue))
		Expect(failedCondition.Reason).To(Equal(workloadcommon.WorkloadFailedReason))
	}, testEntries)
	Context("job is completed", func() {
		DescribeTable("job is succeeded", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("setup")

			workload := testEntry.getJobWorkloadFn()
			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			By("Creating a new DistributedWorkload")
			concreteWorkload := testEntry.jobConvertor.Convert(workload)

			err = k8sClient.Create(ctx, concreteWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)

			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("set job to running")
			createdJob := testEntry.getJobFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
			Expect(err).ToNot(HaveOccurred())
			createdJob.SetUID(testUid)
			pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
				succeeded, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
			for _, pod := range pods {
				err = k8sClient.Create(ctx, pod)
				Expect(err).ToNot(HaveOccurred())
			}

			status := testEntry.getStatusFn(createdJob).DeepCopy()
			err = k8sClient.Update(ctx, createdJob)
			Expect(err).ToNot(HaveOccurred())

			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
			Expect(err).ToNot(HaveOccurred())

			status.DeepCopyInto(testEntry.getStatusFn(createdJob))
			err = k8sClient.Status().Update(ctx, createdJob)
			Expect(err).ToNot(HaveOccurred())

			By("running the reconcile loop again")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("varify workload phase is completed")
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
			Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseCompleted.String()))
			By("varify job condition")
			workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Degraded, "")
			workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Succeeded)
			workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Failed, "")
			workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "job completed")
		}, testEntries)
		DescribeTable("job is failed", func(testEntry *TestEntryContainer, getClient func() client.Client) {
			By("setup")

			workload := testEntry.getJobWorkloadFn()
			k8sClient := getClient()
			workloadControllerFactory := testEntry.workloadControllerFactory(k8sClient, k8sClient,
				testEntry.isOpenShift, getEventRecorder(), defaultPodLabels, defaultPodAnnotations, controllerOptions)
			workloadController, err := workloadControllerFactory.CreateController()
			Expect(err).ToNot(HaveOccurred())
			Expect(workloadController).ToNot(BeNil())
			By("Creating a new DistributedWorkload")
			concreteWorkload := testEntry.jobConvertor.Convert(workload)

			err = k8sClient.Create(ctx, concreteWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("Verifying that the DistributedWorkload was created")
			createdWorkload := testEntry.getWorkloadFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)

			Expect(err).ToNot(HaveOccurred())
			By("running the reconcile loop")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("set job to running")
			createdJob := testEntry.getJobFn()
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdJob)
			Expect(err).ToNot(HaveOccurred())
			createdJob.SetUID(testUid)
			pods := modifyJob(testEntry.getStatusFn(createdJob), []kubeflowv1.ReplicaType{testEntry.masterReplicaType, workerReplicaType},
				failed, createdJob.GetName(), createdJob.GetNamespace(), createdJob.GetUID())
			for _, pod := range pods {
				err = k8sClient.Create(ctx, pod)
				Expect(err).ToNot(HaveOccurred())
			}

			status := testEntry.getStatusFn(createdJob).DeepCopy()
			err = k8sClient.Update(ctx, createdJob)
			Expect(err).ToNot(HaveOccurred())

			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(createdJob), createdJob)
			Expect(err).ToNot(HaveOccurred())

			status.DeepCopyInto(testEntry.getStatusFn(createdJob))
			err = k8sClient.Status().Update(ctx, createdJob)
			Expect(err).ToNot(HaveOccurred())

			By("running the reconcile loop again")
			_, err = workloadController.Reconcile(ctx, ctrl.Request{}, workload, createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			By("varify workload phase is completed")
			err = k8sClient.Get(ctx, client.ObjectKeyFromObject(concreteWorkload), createdWorkload)
			Expect(err).ToNot(HaveOccurred())
			distributedWorkload := createdWorkload.(*v2alpha1.DistributedWorkload)
			Expect(distributedWorkload.Status.Phase).To(Equal(workloadcommon.PhaseFailed.String()))
			By("varify job condition")
			workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Degraded, "")
			workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Succeeded, "")
			workloadUtils.CheckConditionIsTrue(distributedWorkload.Status.Conditions, workloadcommon.Failed)
			workloadUtils.CheckConditionIsFalseOrUnset(distributedWorkload.Status.Conditions, workloadcommon.Ready, "job failed")
		}, testEntries)
	})

})

func modifyJob(jobStatus *kubeflowv1.JobStatus, replicaTypes []kubeflowv1.ReplicaType, jobPhase testJobPhase, name string, namespace string, jobUid types.UID) []client.Object {
	if jobStatus.ReplicaStatuses == nil {
		jobStatus.ReplicaStatuses = make(map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaStatus)
	}
	pods := make([]client.Object, 0)
	switch jobPhase {
	case creating:
		jobStatus.ReplicaStatuses = nil
		return nil
	case pending:
		for _, replicaType := range replicaTypes {
			jobStatus.ReplicaStatuses[replicaType] = &kubeflowv1.ReplicaStatus{
				Active:    0,
				Succeeded: 0,
				Failed:    0,
			}
			pods = append(pods, createPod(fmt.Sprintf("%s-%s", name,
				replicaType), namespace, "", jobUid, pending, replicaType))
		}
	case running:
		for _, replicaType := range replicaTypes {
			jobStatus.ReplicaStatuses[replicaType] = &kubeflowv1.ReplicaStatus{
				Active:    3,
				Succeeded: 0,
				Failed:    0,
			}

			for i := 0; i < 3; i++ {
				pods = append(pods, createPod(fmt.Sprintf("%s-%s-%d", name,
					replicaType, i), namespace, "", jobUid, running, replicaType))
			}
		}
		jobStatus.Conditions = []kubeflowv1.JobCondition{}
	case degraded:
		for _, replicaType := range replicaTypes {
			jobStatus.ReplicaStatuses[replicaType] = &kubeflowv1.ReplicaStatus{
				Active:    3,
				Succeeded: 0,
				Failed:    0,
			}
			for i := 0; i < 2; i++ {
				pods = append(pods, createPod(fmt.Sprintf("%s-%s-%d", name,
					replicaType, i), namespace, "", jobUid, running, replicaType))
			}
			pods = append(pods, createPod(fmt.Sprintf("%s-%s-pending", name,
				replicaType), namespace, "", jobUid, pending, replicaType))
		}
	case initializing:
		for _, replicaType := range replicaTypes {
			jobStatus.ReplicaStatuses[replicaType] = &kubeflowv1.ReplicaStatus{
				Active:    0,
				Succeeded: 0,
				Failed:    0,
			}
			pods = append(pods, createPod(fmt.Sprintf("%s-%s", name,
				replicaType), namespace, "", jobUid, initializing, replicaType))
		}
	case succeeded:
		jobStatus.CompletionTime = &metav1.Time{Time: metav1.Now().Add(-1)}
		for _, replicaType := range replicaTypes {
			jobStatus.ReplicaStatuses[replicaType] = &kubeflowv1.ReplicaStatus{
				Active:    0,
				Succeeded: 3,
				Failed:    0,
			}
		}
		jobStatus.Conditions = []kubeflowv1.JobCondition{
			{
				Type:   kubeflowv1.JobSucceeded,
				Status: v1.ConditionTrue,
			},
		}
	case failed:
		for _, replicaType := range replicaTypes {
			jobStatus.ReplicaStatuses[replicaType] = &kubeflowv1.ReplicaStatus{
				Active:    0,
				Succeeded: 0,
				Failed:    3,
			}
		}
		jobStatus.CompletionTime = &metav1.Time{Time: metav1.Now().Add(-1)}
		jobStatus.Conditions = []kubeflowv1.JobCondition{
			{
				Type:   kubeflowv1.JobFailed,
				Status: v1.ConditionTrue,
			},
		}
	case suspended:
		for _, replicaType := range replicaTypes {
			jobStatus.ReplicaStatuses[replicaType] = &kubeflowv1.ReplicaStatus{
				Active:    0,
				Succeeded: 0,
				Failed:    0,
			}
		}
		jobStatus.Conditions = []kubeflowv1.JobCondition{
			{
				Type:   kubeflowv1.JobSuspended,
				Status: v1.ConditionTrue,
			},
		}
	case suspending: // not really a phase, just simulating the situation where the job is set to suspended and there are pods that are still not deleted
		for _, replicaType := range replicaTypes {
			jobStatus.ReplicaStatuses[replicaType] = &kubeflowv1.ReplicaStatus{
				Active:    1,
				Succeeded: 0,
				Failed:    0,
			}
		}
		jobStatus.Conditions = []kubeflowv1.JobCondition{
			{
				Type:   kubeflowv1.JobSuspended,
				Status: v1.ConditionTrue,
			},
		}
	default:
		fmt.Printf("unknown job phase %d", jobPhase)
	}
	return pods
}

func createPod(name, namespace, ownerKind string, ownerUID types.UID, phase testJobPhase, replicaType kubeflowv1.ReplicaType) client.Object {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         namespace,
			CreationTimestamp: metav1.Time{Time: metav1.Now().Add(-1)},
			Labels: map[string]string{
				kubeflowv1.JobNameLabel:     jobName,
				kubeflowv1.JobRoleLabel:     string(replicaType),
				kubeflowv1.ReplicaTypeLabel: string(replicaType),
			},
		},
	}
	switch phase {
	case pending:
		pod.Status.Phase = v1.PodPending
	case running:
		pod.Status.Phase = v1.PodRunning
		pod.Status.Conditions = []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionTrue,
			},
		}
	case initializing:
		pod.Status.Phase = v1.PodRunning
		pod.Status.Conditions = []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionFalse,
			},
		}
	default:
	}

	if ownerUID != "" {
		pod.SetOwnerReferences([]metav1.OwnerReference{
			{
				Kind: ownerKind,
				UID:  ownerUID,
			},
		})
	}
	return pod
}

func createExternalUrl(isOpenshift bool, name, namespace, ownerKind string, ownerUID types.UID) client.Object {
	fullName := fmt.Sprintf("dw-%s-0", name)
	if isOpenshift {
		return resource.CreateDumbRoute(fullName, namespace, ownerUID, ownerKind, nil)
	}
	return resource.CreateDumbIngress(fullName, namespace, ownerUID, ownerKind, nil)
}

func SetLabelOwnership(obj client.Object, name string) {
	obj.SetLabels(map[string]string{
		clusterCommonConstants.LabelRelease: name,
	})
}

func VerifyWorkloadResources(ctx context.Context, testEntry *TestEntryContainer, k8sClient client.Client, workload client.Object) {
	By("Verifying that the Job was created")
	createdJob := testEntry.getJobFn()
	err := k8sClient.Get(ctx, client.ObjectKeyFromObject(workload), createdJob)
	Expect(err).ToNot(HaveOccurred())
	By("Verifying that service all service were created")
	createdService := v1.ServiceList{}
	err = k8sClient.List(ctx, &createdService, client.InNamespace(TestNamespace))
	Expect(err).ToNot(HaveOccurred())
	Expect(len(createdService.Items)).To(Equal(2))
	By("Verifying all services owner reference")
	for _, service := range createdService.Items {
		Expect(service.OwnerReferences).ToNot(BeNil())
		Expect(service.OwnerReferences).To(HaveLen(1))
		Expect(service.OwnerReferences[0].UID).To(Equal(workload.GetUID()))
	}
	By("Verifying that externalUrl was created")
	if testEntry.isOpenShift {
		createdRoutes := openshiftV1.RouteList{}
		err = k8sClient.List(ctx, &createdRoutes, client.InNamespace(TestNamespace))
		By("Verifying that the Route was created")
		Expect(err).ToNot(HaveOccurred())
		Expect(len(createdRoutes.Items)).To(Equal(1))
		By("Verifying route owner reference")
		Expect(createdRoutes.Items[0].OwnerReferences).ToNot(BeNil())
		Expect(createdRoutes.Items[0].OwnerReferences).To(HaveLen(1))
		Expect(createdRoutes.Items[0].OwnerReferences[0].UID).To(Equal(workload.GetUID()))
	} else {
		By("Verifying that the Ingress was created")
		createdIngress := network.IngressList{}
		err = k8sClient.List(ctx, &createdIngress, client.InNamespace(TestNamespace))
		Expect(err).ToNot(HaveOccurred())
		Expect(len(createdIngress.Items)).To(Equal(1))
		By("Verifying ingress owner reference")
		Expect(createdIngress.Items[0].OwnerReferences).ToNot(BeNil())
		Expect(createdIngress.Items[0].OwnerReferences).To(HaveLen(1))
		Expect(createdIngress.Items[0].OwnerReferences[0].UID).To(Equal(workload.GetUID()))
	}
	By("Verifying that pvc was created")
	createdPvc := v1.PersistentVolumeClaimList{}
	err = k8sClient.List(ctx, &createdPvc, client.InNamespace(TestNamespace))
	Expect(err).ToNot(HaveOccurred())
	Expect(len(createdPvc.Items)).To(Equal(1))
	By("Verifying pvc owner reference")
	Expect(createdPvc.Items[0].OwnerReferences).ToNot(BeNil())
	Expect(createdPvc.Items[0].OwnerReferences).To(HaveLen(1))
	Expect(createdPvc.Items[0].OwnerReferences[0].UID).To(Equal(workload.GetUID()))
}

func DeleteAllResources(k8sClient client.Client, ctx context.Context) {
	createdService := v1.ServiceList{}
	err := k8sClient.List(ctx, &createdService, client.InNamespace(TestNamespace))
	Expect(err).ToNot(HaveOccurred())
	for _, service := range createdService.Items {
		Expect(k8sClient.Delete(ctx, &service)).To(Succeed())
	}
	createdIngress := network.IngressList{}
	err = k8sClient.List(ctx, &createdIngress, client.InNamespace(TestNamespace))
	Expect(err).ToNot(HaveOccurred())
	for _, externalUrl := range createdIngress.Items {
		Expect(k8sClient.Delete(ctx, &externalUrl)).To(Succeed())
	}
	createdPvc := v1.PersistentVolumeClaimList{}
	err = k8sClient.List(ctx, &createdPvc, client.InNamespace(TestNamespace))
	Expect(err).ToNot(HaveOccurred())
	for _, pvc := range createdPvc.Items {
		Expect(k8sClient.Delete(ctx, &pvc)).To(Succeed())
	}
	routeList := openshiftV1.RouteList{}
	err = k8sClient.List(ctx, &routeList, client.InNamespace(TestNamespace))
	Expect(err).ToNot(HaveOccurred())
	for _, route := range routeList.Items {
		Expect(k8sClient.Delete(ctx, &route)).To(Succeed())
	}
}

func GetJobSpec(one client.Object) map[kubeflowv1.ReplicaType]*kubeflowv1.ReplicaSpec {
	switch one {
	case &kubeflowv1.TFJob{}:
		return one.(*kubeflowv1.TFJob).Spec.TFReplicaSpecs
	case &kubeflowv1.PyTorchJob{}:
		return one.(*kubeflowv1.PyTorchJob).Spec.PyTorchReplicaSpecs
	case &kubeflowv1.XGBoostJob{}:
		return one.(*kubeflowv1.XGBoostJob).Spec.XGBReplicaSpecs
	case &kubeflowv1.MPIJob{}:
		return one.(*kubeflowv1.MPIJob).Spec.MPIReplicaSpecs
	}
	return nil
}
