package ownership_test

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/run-ai/runai/runai-cluster/cluster/anyworkload-controller/internal/ownership"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
)

const (
	legacyNamespace = "runai-project"
	legacyProject   = "project"
	legacyDept      = "department"
)

var _ = Describe("Legacy ownership validation", func() {
	It("allows the stored owner to update a workload", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))
		response := validator.Handle(context.Background(), validationRequest("alice", "alice"))

		Expect(response.Allowed).To(BeTrue())
	})

	It("denies changing the stored owner when mutation is unavailable", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))
		response := validator.Handle(context.Background(), validationRequestWithOwners("alice", "bob", "alice"))

		Expect(response.Allowed).To(BeFalse())
		Expect(response.Result.Message).To(Equal("workload ownership cannot be changed"))
	})

	It("denies another Run:ai user", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))
		response := validator.Handle(context.Background(), validationRequest("alice", "bob"))

		Expect(response.Allowed).To(BeFalse())
		Expect(response.Result.Message).To(Equal("operation not allowed"))
	})

	It("uses the stored workload namespace for delete authorization", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))
		request := validationRequest("alice", "bob")
		request.Operation = admissionv1.Delete
		request.Namespace = ""

		response := validator.Handle(context.Background(), request)

		Expect(response.Allowed).To(BeFalse())
		Expect(response.Result.Message).To(Equal("operation not allowed"))
	})

	It("allows a user with a privileged department access rule", func() {
		accessRule := &runaiv1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{Name: "department-editor"},
			Spec: runaiv1alpha1.AccessRuleSpec{
				SubjectId: "bob", SubjectType: runaiv1alpha1.UserSubject,
				ScopeType: runaiv1alpha1.DepartmentScope, ScopeName: legacyDept,
				RoleName: "editor",
			},
		}
		validator := ownership.NewValidator(fakeOwnershipClient(false, accessRule))
		response := validator.Handle(context.Background(), validationRequest("alice", "bob"))

		Expect(response.Allowed).To(BeTrue())
	})

	It("allows a user through a privileged group access rule", func() {
		accessRule := &runaiv1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{Name: "cluster-editor-group"},
			Spec: runaiv1alpha1.AccessRuleSpec{
				SubjectId: "researchers", SubjectType: runaiv1alpha1.GroupSubject,
				ScopeType: runaiv1alpha1.ClusterScope,
				RoleName:  "editor",
			},
		}
		validator := ownership.NewValidator(fakeOwnershipClient(false, accessRule))
		request := validationRequest("alice", "bob")
		request.UserInfo.Groups = []string{"researchers"}

		response := validator.Handle(context.Background(), request)

		Expect(response.Allowed).To(BeTrue())
	})

	It("allows the Run:ai agent", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))
		response := validator.Handle(context.Background(), validationRequest("alice", "agent-uid"))

		Expect(response.Allowed).To(BeTrue())
	})

	It("reads internal service-account identities through the uncached reader", func() {
		cachedClient := fakeOwnershipClient(false)
		serviceAccount := &corev1.ServiceAccount{}
		Expect(cachedClient.Get(context.Background(), client.ObjectKey{
			Name: "runai-agent", Namespace: "runai",
		}, serviceAccount)).To(Succeed())
		Expect(cachedClient.Delete(context.Background(), serviceAccount)).To(Succeed())

		validator := ownership.NewValidator(cachedClient, fakeOwnershipClient(false))
		response := validator.Handle(context.Background(), validationRequest("alice", "agent-uid"))

		Expect(response.Allowed).To(BeTrue())
	})

	It("allows a Kubernetes administrator through SubjectAccessReview", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(true))
		response := validator.Handle(context.Background(), validationRequest("alice", "kubeadmin"))

		Expect(response.Allowed).To(BeTrue())
	})

	It("passes the complete authenticated identity to SubjectAccessReview", func() {
		k8sClient := fakeOwnershipClientWithSubjectAccessReview(func(review *authorizationv1.SubjectAccessReview) {
			review.Status.Allowed = review.Spec.User == "kubeadmin" &&
				review.Spec.UID == "kubeadmin-uid" &&
				len(review.Spec.Groups) == 1 && review.Spec.Groups[0] == "system:masters"
		})
		validator := ownership.NewValidator(k8sClient)
		request := validationRequest("alice", "kubeadmin-uid")
		request.UserInfo = authenticationv1.UserInfo{
			Username: "kubeadmin",
			UID:      "kubeadmin-uid",
			Groups:   []string{"system:masters"},
		}

		response := validator.Handle(context.Background(), request)

		Expect(response.Allowed).To(BeTrue())
	})

	It("allows cluster-scoped authorization when the namespace has no project label", func() {
		accessRule := &runaiv1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{Name: "cluster-editor"},
			Spec: runaiv1alpha1.AccessRuleSpec{
				SubjectId: "bob", SubjectType: runaiv1alpha1.UserSubject,
				ScopeType: runaiv1alpha1.ClusterScope,
				RoleName:  "editor",
			},
		}
		k8sClient := fakeOwnershipClient(false, accessRule)
		namespace := &corev1.Namespace{}
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: legacyNamespace}, namespace)).To(Succeed())
		namespace.Labels = nil
		Expect(k8sClient.Update(context.Background(), namespace)).To(Succeed())

		response := ownership.NewValidator(k8sClient).Handle(
			context.Background(), validationRequest("alice", "bob"))

		Expect(response.Allowed).To(BeTrue())
	})

	It("allows the legacy controller during the upgrade handoff", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))
		response := validator.Handle(context.Background(), validationRequest("alice", "workload-controller-uid"))

		Expect(response.Allowed).To(BeTrue())
	})
})

var _ = Describe("WorkloadRunner ownership validation", func() {
	It("allows the stored owner to update a runner", func() {
		response := ownership.NewValidator(fakeOwnershipClient(false)).Handle(
			context.Background(), workloadRunnerValidationRequest(admissionv1.Update, "alice", "alice", "alice"))

		Expect(response.Allowed).To(BeTrue())
	})

	It("denies another user from deleting a runner", func() {
		response := ownership.NewValidator(fakeOwnershipClient(false)).Handle(
			context.Background(), workloadRunnerValidationRequest(admissionv1.Delete, "alice", "", "bob"))

		Expect(response.Allowed).To(BeFalse())
	})

	It("denies changing the stored owner when mutation is unavailable", func() {
		response := ownership.NewValidator(fakeOwnershipClient(false)).Handle(
			context.Background(), workloadRunnerValidationRequest(admissionv1.Update, "alice", "bob", "alice"))

		Expect(response.Allowed).To(BeFalse())
		Expect(response.Result.Message).To(Equal("workload ownership cannot be changed"))
	})

	It("denies updates and deletes of an ownerless control-plane-managed runner", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))

		updateResponse := validator.Handle(context.Background(), controlPlaneManagedWorkloadRunnerValidationRequest(
			admissionv1.Update, "", "", "bob"))
		deleteResponse := validator.Handle(context.Background(), controlPlaneManagedWorkloadRunnerValidationRequest(
			admissionv1.Delete, "", "", "bob"))

		Expect(updateResponse.Allowed).To(BeFalse())
		Expect(updateResponse.Result.Message).To(Equal("control-plane-managed workload ownership has not been initialized"))
		Expect(deleteResponse.Allowed).To(BeFalse())
		Expect(deleteResponse.Result.Message).To(Equal("control-plane-managed workload ownership has not been initialized"))
	})

	It("allows the agent to backfill an ownerless control-plane-managed runner", func() {
		response := ownership.NewValidator(fakeOwnershipClient(false)).Handle(
			context.Background(), controlPlaneManagedWorkloadRunnerValidationRequest(
				admissionv1.Update, "", "alice", "agent-uid"))

		Expect(response.Allowed).To(BeTrue())
	})

	It("allows only the agent to initialize ownership during create", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))

		Expect(validator.Handle(context.Background(), workloadRunnerValidationRequest(
			admissionv1.Create, "", "alice", "agent-uid")).Allowed).To(BeTrue())
		Expect(validator.Handle(context.Background(), workloadRunnerValidationRequest(
			admissionv1.Create, "", "alice", "bob")).Allowed).To(BeFalse())
	})

	It("allows only the agent to backfill an unowned runner", func() {
		validator := ownership.NewValidator(fakeOwnershipClient(false))

		Expect(validator.Handle(context.Background(), workloadRunnerValidationRequest(
			admissionv1.Update, "", "alice", "agent-uid")).Allowed).To(BeTrue())
		Expect(validator.Handle(context.Background(), workloadRunnerValidationRequest(
			admissionv1.Update, "", "alice", "bob")).Allowed).To(BeFalse())
	})
})

func validationRequest(owner, requester string) admission.Request {
	return validationRequestWithOwners(owner, owner, requester)
}

func validationRequestWithOwners(storedOwner, desiredOwner, requester string) admission.Request {
	gvk := schema.GroupVersionKind{Group: "run.ai", Version: "v2alpha1", Kind: "TrainingWorkload"}
	return admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
		Operation: admissionv1.Update,
		Namespace: legacyNamespace,
		OldObject: runtime.RawExtension{Raw: legacyWorkloadJSON(gvk, storedOwner)},
		Object:    runtime.RawExtension{Raw: legacyWorkloadJSON(gvk, desiredOwner)},
		UserInfo:  authenticationv1.UserInfo{UID: requester},
	}}
}

func workloadRunnerValidationRequest(operation admissionv1.Operation, oldOwner, newOwner, requester string) admission.Request {
	return workloadRunnerValidationRequestWithManagement(operation, oldOwner, newOwner, requester, false)
}

func controlPlaneManagedWorkloadRunnerValidationRequest(
	operation admissionv1.Operation,
	oldOwner, newOwner, requester string,
) admission.Request {
	return workloadRunnerValidationRequestWithManagement(operation, oldOwner, newOwner, requester, true)
}

func workloadRunnerValidationRequestWithManagement(
	operation admissionv1.Operation,
	oldOwner, newOwner, requester string,
	controlPlaneManaged bool,
) admission.Request {
	return admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
		Operation: operation,
		Kind:      metav1.GroupVersionKind{Group: "run.ai", Version: "v1alpha1", Kind: "WorkloadRunner"},
		Namespace: legacyNamespace,
		OldObject: runtime.RawExtension{Raw: workloadRunnerJSON(oldOwner, controlPlaneManaged)},
		Object:    runtime.RawExtension{Raw: workloadRunnerJSON(newOwner, controlPlaneManaged)},
		UserInfo:  authenticationv1.UserInfo{UID: requester},
	}}
}

func workloadRunnerJSON(owner string, controlPlaneManaged ...bool) []byte {
	annotations := map[string]string{}
	if owner != "" {
		annotations[clusterconstants.WorkloadOwnerUIDAnnotation] = owner
	}
	labels := map[string]string{}
	if len(controlPlaneManaged) != 0 && controlPlaneManaged[0] {
		labels[commonconstants.ControlPlaneManagedLabel] = "true"
	}
	object := map[string]any{
		"apiVersion": "run.ai/v1alpha1",
		"kind":       "WorkloadRunner",
		"metadata": map[string]any{
			"name": "runner", "namespace": legacyNamespace, "annotations": annotations, "labels": labels,
		},
		"spec": map[string]any{"template": map[string]any{"apiVersion": "batch/v1", "kind": "Job"}},
	}
	raw, err := json.Marshal(object)
	Expect(err).NotTo(HaveOccurred())
	return raw
}

func fakeOwnershipClient(sarAllowed bool, extraObjects ...client.Object) client.Client {
	return fakeOwnershipClientWithSubjectAccessReview(func(review *authorizationv1.SubjectAccessReview) {
		review.Status.Allowed = sarAllowed
	}, extraObjects...)
}

func fakeOwnershipClientWithSubjectAccessReview(
	reviewSubjectAccess func(*authorizationv1.SubjectAccessReview),
	extraObjects ...client.Object,
) client.Client {
	scheme := runtime.NewScheme()
	Expect(corev1.AddToScheme(scheme)).To(Succeed())
	Expect(authorizationv1.AddToScheme(scheme)).To(Succeed())
	Expect(runaiv1alpha1.AddToScheme(scheme)).To(Succeed())

	project := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "kai.resources/v1alpha1",
		"kind":       "Project",
		"metadata":   map[string]any{"name": legacyProject},
		"spec":       map[string]any{"parent": legacyDept},
	}}
	project.SetGroupVersionKind(schema.GroupVersionKind{Group: "kai.resources", Version: "v1alpha1", Kind: "Project"})
	objects := []client.Object{
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: legacyNamespace, Labels: map[string]string{"runai/queue": legacyProject}}},
		project,
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "runai-agent", Namespace: "runai", UID: types.UID("agent-uid")}},
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "runtime-rules", Namespace: "runai", UID: types.UID("runtime-rules-uid")}},
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "workload-controller", Namespace: "runai", UID: types.UID("workload-controller-uid")}},
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "inference-workload-controller", Namespace: "runai", UID: types.UID("inference-controller-uid")}},
	}
	objects = append(objects, extraObjects...)

	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).WithInterceptorFuncs(interceptor.Funcs{
		Create: func(_ context.Context, _ client.WithWatch, object client.Object, _ ...client.CreateOption) error {
			review := object.(*authorizationv1.SubjectAccessReview)
			reviewSubjectAccess(review)
			return nil
		},
	}).Build()
}
