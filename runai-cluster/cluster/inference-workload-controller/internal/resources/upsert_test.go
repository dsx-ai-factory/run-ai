package resources

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	resourceutils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"net/http"
	"time"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	issue2 "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/issue"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/events"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Additional constants that should be defined
const (
	TestServiceName1      = "test-service-1"
	TestIngressName1      = "test-ingress-1"
	TestPvc1              = "test-pvc-1"
	workloadName          = "workload-name"
	projectNamespace      = "runai-test-project"
	projectName           = "test-project"
	pvcSize               = "10Gi"
	TestComputeDomainName = "test-compute-domain"
)

var _ = Describe("Resource manager", func() {
	var (
		k8sClient           k8sclient.Client
		scheme              *runtime.Scheme
		K8sResourcesManager *K8sResourceManager
		ctx                 context.Context
		cancel              context.CancelFunc
		recorder            *record.FakeRecorder
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
		ctx = workloadproject.SetProjectInContext(context.Background(), projectNamespace)
		ctx, cancel = context.WithCancel(ctx)

		recorder = record.NewFakeRecorder(200)
	})
	AfterEach(func() {
		defer cancel()
	})

	Context("UpsertResources", func() {
		var (
			workload *infw.InferenceWorkload
		)
		ownerUid := types.UID("test-uid")
		kind := constants.KindInferenceWorkload
		Context("No resources create it", func() {
			var (
				objects    []k8sclient.Object
				serviceOne *v1.Service
				ingressOne *network.Ingress
				pvc        *v1.PersistentVolumeClaim
				cd         *nvapi.ComputeDomain
			)
			BeforeEach(func() {
				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme)
				K8sResourcesManager = NewResourceManager(k8sClient, &issue2.InferenceResourceGrouper{})
				workload = NewInferenceWorkload(workloadName, projectNamespace, ownerUid)
				labels := map[string]string{constants.LabelRelease: workloadName}
				objects = []k8sclient.Object{}
				serviceOne = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, serviceOne)
				ingressOne = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, ingressOne)
				pvc = resourceutils.CreatePvc(TestPvc1, projectNamespace, resource.MustParse(pvcSize), ownerUid, kind, "", nil, nil, labels)
				objects = append(objects, pvc)
				cd = resourceutils.CreateDumbComputeDomain(TestComputeDomainName, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, cd)

				ctx = injectWorkloadEventRecorder(ctx, workload, recorder)
			})
			It("Should create all resources", func() {
				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).To(BeNil())
				service := &v1.Service{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(serviceOne), service)
				Expect(err).ToNot(HaveOccurred())
				Expect(service.GetName()).To(Equal(serviceOne.GetName()))
				ingress := &network.Ingress{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(ingressOne), ingress)
				Expect(err).ToNot(HaveOccurred())
				Expect(ingress.GetName()).To(Equal(ingressOne.GetName()))

				getPvc := &v1.PersistentVolumeClaim{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(pvc), getPvc)
				Expect(err).ToNot(HaveOccurred())
				Expect(getPvc.GetName()).To(Equal(pvc.GetName()))

				expectedCd := &nvapi.ComputeDomain{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), expectedCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(expectedCd.GetName()).To(Equal(cd.GetName()))
				Eventually(recorder.Events).WithTimeout(3 * time.Second).
					Should(Receive(ContainSubstring(fmt.Sprintf("Successfully created %s %s", constants.KindComputeDomain, cd.Name))))
			})
			// TODO add this test when conditions are available for inference workloads
			//It("Ready condition should be false", func() {
			//	issues := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
			//	Expect(issues).To(BeNil())
			//	Expect(workload.Status.Conditions).To(HaveLen(1))
			//	Expect(string(workload.Status.Conditions[0].Type)).To(Equal(common.Ready))
			//	Expect(workload.Status.Conditions[0].Status).To(Equal(v1.ConditionFalse))
			//})
		})
		Context("Resources exists update it", func() {

			var (
				objects    []k8sclient.Object
				serviceOne *v1.Service
				ingressOne *network.Ingress
				cd         *nvapi.ComputeDomain
			)
			BeforeEach(func() {
				workload = NewInferenceWorkload(workloadName, projectNamespace, ownerUid)
				objects = []k8sclient.Object{}
				labels := map[string]string{constants.LabelRelease: workloadName}
				serviceOne = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, serviceOne)
				ingressOne = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, ingressOne)
				cd = resourceutils.CreateDumbComputeDomain(TestComputeDomainName, projectNamespace, ownerUid, kind, labels)
				objects = append(objects, cd)

				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, serviceOne, ingressOne, cd)
				K8sResourcesManager = NewResourceManager(k8sClient, &issue2.InferenceResourceGrouper{})

				ctx = injectWorkloadEventRecorder(ctx, workload, recorder)
			})
			It("Should copy finalizer from the existing resource before upsert", func() {
				newCd := &nvapi.ComputeDomain{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				newCd.SetFinalizers([]string{"test-finalizer"})
				err = k8sClient.Update(ctx, newCd)
				Expect(err).ToNot(HaveOccurred())
				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).To(BeNil())
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(newCd.GetFinalizers()).To(ContainElement("test-finalizer"))
			})
			It("Should copy runtime status labels before upsert", func() {
				newCd := &nvapi.ComputeDomain{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				newCd.Labels = map[string]string{
					"test": "test",
				}
				err = k8sClient.Update(ctx, newCd)
				Expect(err).ToNot(HaveOccurred())
				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).To(BeNil())
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(newCd.GetLabels()).To(HaveKeyWithValue("test", "test"))
			})
			It("Should copy runtime status annotation before upsert", func() {
				newCd := &nvapi.ComputeDomain{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				newCd.Annotations = map[string]string{
					common.RunaiCurrentAllocatedGpusAnnotation:      "2",
					common.RunaiCurrentAllocatedGpuMemoryAnnotation: "2400",
					"test": "test",
				}
				err = k8sClient.Update(ctx, newCd)
				Expect(err).ToNot(HaveOccurred())
				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).To(BeNil())
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(newCd.GetAnnotations()).To(HaveLen(3))
				Expect(newCd.GetAnnotations()).To(HaveKeyWithValue(common.RunaiCurrentAllocatedGpuMemoryAnnotation, "2400"))
				Expect(newCd.GetAnnotations()).To(HaveKeyWithValue(common.RunaiCurrentAllocatedGpusAnnotation, "2"))
				Expect(newCd.GetAnnotations()).To(HaveKeyWithValue("test", "test"))

			})
			It("Should upsert all resources", func() {
				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).To(BeNil())
				newCd := &nvapi.ComputeDomain{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(newCd.GetName()).To(Equal(cd.GetName()))
				service := &v1.Service{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(serviceOne), service)
				Expect(err).ToNot(HaveOccurred())
				Expect(service.GetName()).To(Equal(serviceOne.GetName()))
				ingress := &network.Ingress{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(ingressOne), ingress)
				Expect(err).ToNot(HaveOccurred())
				Expect(ingress.GetName()).To(Equal(ingressOne.GetName()))

				// making sure we don't send event for update to not spam
				Consistently(recorder.Events).ShouldNot(Receive())
			})
			// TODO add this test when conditions are available for inference workloads
			//It("Ready condition should not exists, no changes to resources", func() {
			//	issues := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
			//	Expect(issues).To(BeNil())
			//	Expect(workload.Status.Conditions).To(HaveLen(0))
			//})
			It("should failed to update owner reference is not equal", func() {
				workload.SetUID("other-uid")
				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				var msgList []string
				Expect(issues).ToNot(BeNil())
				Expect(len(issues)).To(Equal(len(objects)))

				for _, iss := range issues {
					msgList = append(msgList, iss.Message())
				}

				Expect(msgList).To(ContainElement("[Service:test-service-1]  \"test-service-1\" already exists"))
				Expect(msgList).To(ContainElement("[Ingress:test-ingress-1]  \"test-ingress-1\" already exists"))
			})
		})
		Context("Backward compatibility, resources exits no owner ref", func() {
			var (
				objects    []k8sclient.Object
				serviceOne *v1.Service
				ingressOne *network.Ingress
				cd         *nvapi.ComputeDomain
			)

			BeforeEach(func() {
				workload = NewInferenceWorkload("test-job", "test-namespace", ownerUid)
				labels := map[string]string{constants.LabelRelease: workload.Name}
				objects = []k8sclient.Object{}
				serviceOne = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, "", "", labels)
				serviceOne.SetLabels(map[string]string{
					constants.LabelRelease: workload.Name,
				})
				objects = append(objects, serviceOne)
				ingressOne = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, "", "", labels)
				ingressOne.SetLabels(map[string]string{
					constants.LabelRelease: workload.Name,
				})
				objects = append(objects, ingressOne)
				cd = resourceutils.CreateDumbComputeDomain(TestComputeDomainName, projectNamespace, "", "", labels)
				cd.SetLabels(map[string]string{
					constants.LabelRelease: workload.Name,
				})
				objects = append(objects, cd)

				k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, cd, serviceOne, ingressOne)
				K8sResourcesManager = NewResourceManager(k8sClient, &issue2.InferenceResourceGrouper{})
			})

			It("Should copy runtime status annotation before upsert", func() {
				newCd := &nvapi.ComputeDomain{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				newCd.Annotations = map[string]string{
					common.RunaiCurrentAllocatedGpusAnnotation:      "2",
					common.RunaiCurrentAllocatedGpuMemoryAnnotation: "2400",
				}
				err = k8sClient.Update(ctx, newCd)
				Expect(err).ToNot(HaveOccurred())
				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).To(BeNil())
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(newCd.GetAnnotations()).To(HaveKeyWithValue(common.RunaiCurrentAllocatedGpusAnnotation, "2"))

			})
			It("Should upsert all resources", func() {
				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).To(BeNil())
				newCd := &nvapi.ComputeDomain{}
				err := k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(cd), newCd)
				Expect(err).ToNot(HaveOccurred())
				Expect(newCd.GetName()).To(Equal(cd.GetName()))
				service := &v1.Service{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(serviceOne), service)
				Expect(err).ToNot(HaveOccurred())
				Expect(service.GetName()).To(Equal(serviceOne.GetName()))
				ingress := &network.Ingress{}
				err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(ingressOne), ingress)
				Expect(err).ToNot(HaveOccurred())
				Expect(ingress.GetName()).To(Equal(ingressOne.GetName()))
			})
			// TODO add this test when conditions are available for inference workloads
			//It("Ready condition should not exists, no changes to resources", func() {
			//	issues := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
			//	Expect(issues).To(BeNil())
			//	Expect(workload.Status.Conditions).To(HaveLen(0))
			//})
			It("should failed to update release label is not equal", func() {
				otherWorkload := NewInferenceWorkload("other-test-workload", "test-namespace", ownerUid)
				issues, _ := K8sResourcesManager.UpsertResources(ctx, otherWorkload.ObjectMeta, objects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).ToNot(BeNil())
				Expect(issues[0].Error()).To(ContainSubstring("already exists"))
			})
		})
	})
	Context("Upsert Issues", func() {
		var (
			fakeClientPtr *client.FakeClientWithErrorRules

			workload       *infw.InferenceWorkload
			allTestObjects []k8sclient.Object
			service        *v1.Service
			ingress        *network.Ingress
			pvc            *v1.PersistentVolumeClaim
			cd             *nvapi.ComputeDomain

			labels   map[string]string
			ownerUid = types.UID("test-uid")
			kind     = constants.KindInferenceWorkload
		)

		BeforeEach(func() {
			fakeClientPtr = client.NewFakeClientWithErrorRules(scheme, []k8sclient.Object{&nvapi.ComputeDomain{}, &infw.InferenceWorkload{}}, nil)
			k8sClient = fakeClientPtr
			K8sResourcesManager = NewResourceManager(k8sClient, &issue2.InferenceResourceGrouper{})

			workload = NewInferenceWorkload(workloadName, projectNamespace, ownerUid)
			labels = map[string]string{constants.LabelRelease: workload.Name}

			service = resourceutils.CreateDumbService(TestServiceName1, projectNamespace, ownerUid, kind, labels)
			ingress = resourceutils.CreateDumbIngress(TestIngressName1, projectNamespace, ownerUid, kind, labels)
			pvc = resourceutils.CreatePvc(TestPvc1, projectNamespace, resource.MustParse(pvcSize), ownerUid, kind, "", nil, nil, labels)
			cd = resourceutils.CreateDumbComputeDomain(TestComputeDomainName, projectNamespace, ownerUid, kind, labels)

			allTestObjects = []k8sclient.Object{service, ingress, pvc, cd}

			ctx = injectWorkloadEventRecorder(ctx, workload, recorder)
		})

		It("Create issues", func() {
			rules := []client.ResourceFailingRule{
				{
					ClientAction: client.Create,
					ResourceKind: constants.KindIngress,
					ResourceName: ingress.GetName(),
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: client.UrlAlreadyExistsMsg, Code: http.StatusBadRequest}},
				},
				{
					ClientAction: client.Create,
					ResourceKind: constants.KindPvc,
					ResourceName: pvc.GetName(),
				},
				{
					ClientAction: client.Create,
					ResourceKind: constants.KindComputeDomain,
					ResourceName: cd.GetName(),
					StatusError: &k8serrors.StatusError{ErrStatus: metav1.Status{
						Code:    http.StatusInternalServerError,
						Message: "Internal error occurred: failed calling webhook \"bla bla bla\"",
						Details: &metav1.StatusDetails{Causes: []metav1.StatusCause{{
							Message: "failed calling webhook \"bla bla bla\": failed to call webhook: the server could not find the requested resource",
						}}}},
					},
				},
			}
			fakeClientPtr.UpdateRules(rules)

			issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, allTestObjects, &workload.Status.OwnedResources, nil, ResourceOptions{})
			Expect(issues).NotTo(BeNil())
			Expect(issues).To(HaveLen(len(rules)))

			var ingressIssue, pvcIssue, cdIssue *issue.Issue
			for _, iss := range issues {
				if iss.Kind() == constants.KindPvc {
					pvcIssue = iss
				}
				if iss.Kind() == constants.KindIngress {
					ingressIssue = iss
				}
				if iss.Kind() == constants.KindComputeDomain {
					cdIssue = iss
				}
			}

			Expect(ingressIssue).NotTo(BeNil())
			Expect(ingressIssue.IsCritical()).To(BeTrue())
			Expect(ingressIssue.Message()).To(ContainSubstring(client.UrlAlreadyExistsMsg))
			Expect(ingressIssue.GetGroup()).To(Equal(issue.Network))

			Expect(pvcIssue).NotTo(BeNil())
			Expect(pvcIssue.IsCritical()).To(BeTrue())
			Expect(pvcIssue.Message()).To(ContainSubstring(fmt.Sprintf(client.GeneralErrorMsg, pvc.GetName())))
			Expect(pvcIssue.GetGroup()).To(Equal(issue.Storage))

			Expect(cdIssue).NotTo(BeNil())
			Expect(cdIssue.IsCritical()).To(BeTrue()) //Compute upsert issues with webhook error are critical
			Expect(cdIssue.Message()).To(Equal("[ComputeDomain:test-compute-domain] Internal error occurred: failed calling webhook \"bla bla bla\""))
			Expect(cdIssue.GetGroup()).To(Equal(issue.Compute))

			Eventually(recorder.Events).WithTimeout(3 * time.Second).
				Should(Receive(ContainSubstring(fmt.Sprintf("Failed to create %s %s", constants.KindComputeDomain, cd.Name))))
		})

		It("Update issues", func() {
			for _, obj := range allTestObjects {
				// The newer controller-runtime fake client strips TypeMeta on Create.
				// allTestObjects is reused as UpsertResources input, where rule matching
				// relies on GroupVersionKind().Kind, so restore the GVK after creating.
				gvk := obj.GetObjectKind().GroupVersionKind()
				Expect(k8sClient.Create(ctx, obj)).To(Succeed())
				obj.GetObjectKind().SetGroupVersionKind(gvk)
			}

			rules := []client.ResourceFailingRule{
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindService,
					ResourceName: service.GetName(),
				},
				{
					ClientAction: client.Update,
					ResourceKind: constants.KindComputeDomain,
					ResourceName: cd.GetName(),
				},
			}
			fakeClientPtr.UpdateRules(rules)

			issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, allTestObjects, &workload.Status.OwnedResources, nil, ResourceOptions{})
			Expect(issues).NotTo(BeNil())
			// we currently don't update compute domain so it won't produce an issue
			Expect(issues).To(HaveLen(1))

			var serviceIssue, cdIssue *issue.Issue
			for _, iss := range issues {
				if iss.Kind() == constants.KindService {
					serviceIssue = iss
				}
				if iss.Kind() == constants.KindComputeDomain {
					cdIssue = iss
				}
			}

			Expect(serviceIssue).NotTo(BeNil())
			Expect(serviceIssue.IsCritical()).To(BeTrue())
			Expect(serviceIssue.Message()).To(ContainSubstring(fmt.Sprintf(client.GeneralErrorMsg, service.GetName())))
			Expect(serviceIssue.GetGroup()).To(Equal(issue.Network))

			// we currently don't update compute domain
			Expect(cdIssue).To(BeNil())

			// making sure we don't send event for update to not spam
			Consistently(recorder.Events).ShouldNot(Receive())
		})

		DescribeTable("non-critical errors",
			func(errorCode int, errorMsg string, errorReason metav1.StatusReason) {
				rules := []client.ResourceFailingRule{
					{
						ClientAction: client.Create,
						ResourceKind: constants.KindIngress,
						ResourceName: ingress.GetName(),
						StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: errorMsg, Code: int32(errorCode), Reason: errorReason}},
					},
				}
				fakeClientPtr.UpdateRules(rules)

				issues, _ := K8sResourcesManager.UpsertResources(ctx, workload.ObjectMeta, allTestObjects, &workload.Status.OwnedResources, nil, ResourceOptions{})
				Expect(issues).NotTo(BeNil())
				Expect(issues).To(HaveLen(len(rules)))

				var ingressIssue *issue.Issue
				for _, iss := range issues {
					if iss.Kind() == constants.KindIngress {
						ingressIssue = iss
					}
				}

				Expect(ingressIssue).NotTo(BeNil())
				Expect(ingressIssue.IsCritical()).To(BeFalse())
				Expect(ingressIssue.Message()).To(ContainSubstring(errorMsg))
				Expect(ingressIssue.GetGroup()).To(Equal(issue.Network))
			},
			Entry("429 - too many requests", http.StatusTooManyRequests, "too many, try again later", metav1.StatusReasonTooManyRequests),
			Entry("409 - conflict - the object has been modified", http.StatusConflict, "the object has been modified; please apply your changes to the latest version and try again", metav1.StatusReasonConflict),
			Entry("500 - internal server error", http.StatusInternalServerError, "ERROR", metav1.StatusReasonInternalError),
			Entry("303", http.StatusSeeOther, "ERROR", metav1.StatusReasonUnknown),
		)
	})
})

func NewInferenceWorkload(name, namespace string, uid types.UID) *infw.InferenceWorkload {
	return &infw.InferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       uid,
		},
	}
}

func injectWorkloadEventRecorder(ctx context.Context, workload *infw.InferenceWorkload, recorder record.EventRecorder) context.Context {
	workloadRecorder := events.NewWorkloadRecorder(workload, recorder)
	return events.ToContext(ctx, workloadRecorder)
}
