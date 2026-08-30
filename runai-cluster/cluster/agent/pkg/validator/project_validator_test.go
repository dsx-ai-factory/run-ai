package validator

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"

	. "github.com/onsi/gomega"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func init() {
	_ = v1alpha1.AddToScheme(scheme.Scheme)
}

func TestWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Workload Ownership Protection Webhook")
}

var _ = Describe("Test Workload Ownership Protection Webhook", Ordered, func() {
	var (
		k8sClient                       client.Client
		agentServiceAccount             *corev1.ServiceAccount
		CSServiceAccount                *corev1.ServiceAccount
		projectControllerServiceAccount *corev1.ServiceAccount
		runaiOperatorServiceAccount     *corev1.ServiceAccount
		objects                         []runtime.Object

		runaiAgentUID        string
		runaiCSUID           string
		projectControllerUID string
		runaiOperatorUID     string
		projectName          string
		req                  admission.Request

		operations       = []admissionv1.Operation{admissionv1.Create, admissionv1.Update, admissionv1.Delete}
		projectValidator *ProjectValidator
	)

	BeforeAll(func() {
		runaiAgentUID = uuid.New().String()
		runaiCSUID = uuid.New().String()
		projectControllerUID = uuid.New().String()
		runaiOperatorUID = uuid.New().String()

		agentServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runaiv1.DefaultNamespace,
				Name:      operands.RunaiAgentName,
				UID:       types.UID(runaiAgentUID),
			},
		}

		CSServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runaiv1.DefaultNamespace,
				Name:      operands.RunaiClusterSyncName,
				UID:       types.UID(runaiCSUID),
			},
		}

		projectControllerServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runaiv1.DefaultNamespace,
				Name:      operands.RunaiProjectControllerName,
				UID:       types.UID(projectControllerUID),
			},
		}

		runaiOperatorServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runaiv1.DefaultNamespace,
				Name:      operands.RunaiOperatorName,
				UID:       types.UID(runaiOperatorUID),
			},
		}
	})

	BeforeEach(func() {
		projectName = fmt.Sprintf("project-%s", rand.String(3))
		req = admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Name:     projectName,
				UserInfo: authenticationv1.UserInfo{},
			},
		}

		objects = []runtime.Object{agentServiceAccount, CSServiceAccount, projectControllerServiceAccount, runaiOperatorServiceAccount}
	})

	Context("Test ProjectValidator with service account", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			projectValidator = NewProjectValidator(k8sClient, []string{})
		})

		It("Requester is runai-agent - allow", func() {
			req.UserInfo.UID = runaiAgentUID

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester is project-controller - allow", func() {
			req.UserInfo.UID = projectControllerUID

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester is runai-operator - allow", func() {
			req.UserInfo.UID = runaiOperatorUID

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester is not runai-agent or project-controller or runai-operator - don't allow", func() {
			req.UserInfo.UID = uuid.New().String()

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(res.Result.Reason).To(Not(BeEmpty()))
				Expect(res.Result.Message).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			}
		})
	})

	Context("Test ProjectValidator - service account doesn't exist", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects().Build()
			projectValidator = NewProjectValidator(k8sClient, []string{})
		})

		It("Don't allow requests", func() {
			req.UserInfo.UID = runaiAgentUID

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(ContainSubstring("not found"))
				Expect(res.Result.Message).To(ContainSubstring("serviceaccount"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusInternalServerError)))
			}
		})
	})
	Context("Test ProjectValidator with allowedServiceAccounts", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			projectValidator = NewProjectValidator(k8sClient, []string{"test-service", "system:serviceaccount:kube-system"})
		})

		It("Requester username contains allowed pattern - allow without checking service account UIDs", func() {
			// Set a random UID that doesn't match any service account
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "system:serviceaccount:test-namespace:test-service-account"

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester username contains second allowed pattern - allow", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "system:serviceaccount:kube-system:coredns"

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester username doesn't contain any allowed pattern - fall back to UID check", func() {
			// Set the UID to runai-agent so it should be allowed by UID check
			req.UserInfo.UID = runaiAgentUID
			req.UserInfo.Username = "system:serviceaccount:default:some-other-service"

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester username contain system:serviceaccount pattern - allow", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "system:serviceaccount:default:unauthorized-service"

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})
		It("Requester username  doesn't contain any allowed pattern  and uid is not allowed - deny", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "system:not-allowed"

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(res.Result.Reason).To(Not(BeEmpty()))
				Expect(res.Result.Message).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			}
		})
	})

	Context("Test ProjectValidator with empty allowedServiceAccounts", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			projectValidator = NewProjectValidator(k8sClient, []string{})
		})

		It("Should fall back to UID check when allowedServiceAccounts is empty", func() {
			req.UserInfo.UID = runaiAgentUID
			req.UserInfo.Username = "system:serviceaccount:test:test-service"

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})
	})

	Context("Test ProjectValidator with single allowedServiceAccount pattern", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			projectValidator = NewProjectValidator(k8sClient, []string{"admin"})
		})

		It("Username contains the single allowed pattern - allow", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "system:serviceaccount:kube-system:admin-controller"

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Username doesn't contain the pattern - deny if UID also not allowed", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "system:not-allowed"

			for _, operation := range operations {
				req.Operation = operation
				res := projectValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(res.Result.Reason).To(Not(BeEmpty()))
				Expect(res.Result.Message).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			}
		})
	})
})
