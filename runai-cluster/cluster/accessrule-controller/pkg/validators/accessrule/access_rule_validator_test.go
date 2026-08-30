package accessrule

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
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
	v1alpha1.AddToScheme(scheme.Scheme)
}

func TestWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Workload Ownership Protection Webhook")
}

var _ = Describe("Test Workload Ownership Protection Webhook", Ordered, func() {
	var (
		k8sClient                          client.Client
		agentServiceAccount                *corev1.ServiceAccount
		accessruleControllerServiceAccount *corev1.ServiceAccount
		runaiOperatorServiceAccount        *corev1.ServiceAccount
		objects                            []runtime.Object

		runaiAgentUID           string
		accessruleControllerUID string
		runaiOperatorUID        string
		accessRuleName          string
		req                     admission.Request

		operations          = []admissionv1.Operation{admissionv1.Create, admissionv1.Update, admissionv1.Delete}
		accessRuleValidator *AccessRuleValidator
	)

	BeforeAll(func() {
		runaiAgentUID = uuid.New().String()
		accessruleControllerUID = uuid.New().String()
		runaiOperatorUID = uuid.New().String()

		agentServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runaiv1.DefaultNamespace,
				Name:      common.RunaiAgentServiceAccountName,
				UID:       types.UID(runaiAgentUID),
			},
		}

		accessruleControllerServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runaiv1.DefaultNamespace,
				Name:      common.RunaiAccessRuleControllerServiceAccountName,
				UID:       types.UID(accessruleControllerUID),
			},
		}

		runaiOperatorServiceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: runaiv1.DefaultNamespace,
				Name:      common.RunaiOperatorServiceAccountName,
				UID:       types.UID(runaiOperatorUID),
			},
		}
	})

	BeforeEach(func() {
		accessRuleName = fmt.Sprintf("accessrule-%s", rand.String(3))
		req = admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Name:     accessRuleName,
				UserInfo: authenticationv1.UserInfo{},
			},
		}

		objects = []runtime.Object{agentServiceAccount, accessruleControllerServiceAccount, runaiOperatorServiceAccount}
	})

	Context("Test AccessRuleValidator with service account", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			accessRuleValidator = NewAccessRuleValidator(k8sClient, []string{})
		})

		It("Requester is runai-agent - allow", func() {
			req.UserInfo.UID = runaiAgentUID

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester is accessrule-controller - allow", func() {
			req.UserInfo.UID = accessruleControllerUID

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

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
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester is not runai-agent or accessrule-controller or runai-operator - don't allow", func() {
			req.UserInfo.UID = uuid.New().String()

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(res.Result.Reason).To(Not(BeEmpty()))
				Expect(res.Result.Message).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			}
		})
	})

	Context("Test AccessRuleValidator - service account doesn't exist", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects().Build()
			accessRuleValidator = NewAccessRuleValidator(k8sClient, []string{})
		})

		It("Don't allow requests", func() {
			req.UserInfo.UID = runaiAgentUID

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(ContainSubstring("not found"))
				Expect(res.Result.Message).To(ContainSubstring("serviceaccount"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusInternalServerError)))
			}
		})
	})

	Context("Test AccessRuleValidator with allowedServiceAccounts", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			accessRuleValidator = NewAccessRuleValidator(k8sClient, []string{"test-service", "kube-system"})
		})

		It("Requester username contains allowed pattern - allow without checking service account UIDs", func() {
			// Set a random UID that doesn't match any service account
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "serviceaccount:test-namespace:test-service-account"

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester username contains second allowed pattern - allow", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "serviceaccount:kube-system:coredns"

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester username doesn't contain any allowed pattern - fall back to UID check", func() {
			// Set the UID to runai-agent so it should be allowed by UID check
			req.UserInfo.UID = runaiAgentUID
			req.UserInfo.Username = "serviceaccount:default:some-other-service"

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Requester username doesn't contain allowed pattern and UID is not allowed - deny", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "serviceaccount:default:unauthorized-service"

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(res.Result.Reason).To(Not(BeEmpty()))
				Expect(res.Result.Message).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			}
		})
	})

	Context("Test AccessRuleValidator with empty allowedServiceAccounts", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			accessRuleValidator = NewAccessRuleValidator(k8sClient, []string{})
		})

		It("Should fall back to UID check when allowedServiceAccounts is empty", func() {
			req.UserInfo.UID = runaiAgentUID
			req.UserInfo.Username = "serviceaccount:test:test-service"

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Service account is from  system:serviceaccount - allow", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "system:serviceaccount:kube-system:coredns"

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})
	})

	Context("Test AccessRuleValidator with single allowedServiceAccount pattern", func() {
		JustBeforeEach(func() {
			k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).Build()
			accessRuleValidator = NewAccessRuleValidator(k8sClient, []string{"admin"})
		})

		It("Username contains the single allowed pattern - allow", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "serviceaccount:kube-system:admin-controller"

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			}
		})

		It("Username doesn't contain the pattern - deny if UID also not allowed", func() {
			req.UserInfo.UID = uuid.New().String()
			req.UserInfo.Username = "serviceaccount:default:user-service"

			for _, operation := range operations {
				req.Operation = operation
				res := accessRuleValidator.Handle(context.Background(), req)

				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(res.Result.Reason).To(Not(BeEmpty()))
				Expect(res.Result.Message).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			}
		})
	})
})
