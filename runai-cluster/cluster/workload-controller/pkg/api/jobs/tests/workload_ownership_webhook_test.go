package tests

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	authv1 "k8s.io/api/authorization/v1"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"testing"
	"unsafe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	distw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	projectName    = "project-name"
	departmentName = "department-name"
	workloadName   = "workload-name"
	accessRuleName = "access-rule"
)

var (
	namespaceName = fmt.Sprintf("runai-%s", projectName)

	objects []runtime.Object

	namespace      *corev1.Namespace
	project        *runaiv2.Project
	accessRule     *v1alpha1.AccessRule
	serviceAccount *corev1.ServiceAccount
)

type TestConf struct {
	createNamespace      bool
	createProject        bool
	createServiceAccount bool

	userId                        string
	anotherUserId                 string
	groupId                       string
	workloadControllerUID         string
	runtimeRulesServiceAccountUID string
}

func init() {
	err := runaiv2.AddToScheme(scheme.Scheme)
	if err != nil {
		Fail(err.Error())
	}
	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		Fail(err.Error())
	}
	err = twv1.AddToScheme(scheme.Scheme)
	if err != nil {
		Fail(err.Error())
	}
	err = iwv1.AddToScheme(scheme.Scheme)
	if err != nil {
		Fail(err.Error())
	}
	err = infv2.AddToScheme(scheme.Scheme)
	if err != nil {
		Fail(err.Error())
	}
	err = distw.AddToScheme(scheme.Scheme)
	if err != nil {
		Fail(err.Error())
	}
	err = kaiorg.AddToScheme(scheme.Scheme)
	if err != nil {
		Fail(err.Error())
	}
}

func TestWebhook(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Workload Ownership Protection CreateMergeStackForFrontend")
}

var _ = Describe("Test Workload Ownership Protection CreateMergeStackForFrontend", func() {
	var (
		k8sClient client.WithWatch

		testConf *TestConf

		userId                        string
		anotherUserId                 string
		groupId                       string
		workloadControllerUID         string
		runtimeRulesServiceAccountUID string
		req                           admission.Request

		workloadOwnershipProtectionValidator *jobs.WorkloadOwnershipProtectionValidator
		workloadOwnershipMutator             *jobs.WorkloadOwnershipMutator

		trWorkload  *twv1.TrainingWorkload
		intWorkload *iwv1.InteractiveWorkload
		infWorkload *infv2.InferenceWorkload
		disWorkload *distw.DistributedWorkload
	)
	BeforeEach(func() {
		userId = uuid.New().String()
		anotherUserId = uuid.New().String()
		groupId = uuid.New().String()
		workloadControllerUID = uuid.New().String()
		runtimeRulesServiceAccountUID = uuid.New().String()

		testConf = &TestConf{
			createNamespace:      true,
			createProject:        true,
			createServiceAccount: true,

			userId:                        userId,
			groupId:                       groupId,
			workloadControllerUID:         workloadControllerUID,
			runtimeRulesServiceAccountUID: runtimeRulesServiceAccountUID,
		}

		req = admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Namespace: namespaceName,
				Name:      workloadName,
				UserInfo: authenticationv1.UserInfo{
					UID: userId,
				},
			},
		}
	})
	JustBeforeEach(func() {
		k8sClient = fakeClient.NewClientBuilder().WithScheme(scheme.Scheme).WithRuntimeObjects(objects...).
			WithInterceptorFuncs(interceptor.Funcs{
				Create: func(ctx context.Context, client client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
					switch v := obj.(type) {
					default:
						panic("type must be SubjectAccessReview")
					case *authv1.SubjectAccessReview:
						if v.Spec.User == "kubeadmin" {
							v.Status.Allowed = true
							v.Status.Denied = false
						} else {
							v.Status.Allowed = false
							v.Status.Denied = true
						}

						return nil
					}
				},
			}).Build()
		k8sClient = &client2.FakeClientWithFieldSelector{WithWatch: k8sClient}
		workloadOwnershipProtectionValidator = jobs.NewWorkloadOwnershipProtectionValidator(k8sClient, jobs.NewJobWorkloadWrapper, workloadControllerUID)
		workloadOwnershipMutator = jobs.NewWorkloadOwnershipMutator(k8sClient, jobs.NewJobWorkloadWrapper, workloadControllerUID)
		rv := reflect.ValueOf(workloadOwnershipProtectionValidator).Elem()
		rf := rv.FieldByName("agentId")
		reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem().Set(reflect.ValueOf("agent"))
	})
	Context("Happy flow - validation webhook Handle returns success", func() {
		BeforeEach(func() {
			buildObjects(testConf)
		})
		Describe("TrainingWorkload", func() {
			BeforeEach(func() {
				trWorkload = GetTrainingWorkload(namespaceName, workloadName, userId)
				objects = append(objects, trWorkload)

				marshaledObj, err := json.Marshal(trWorkload)
				Expect(err).To(BeNil())

				populateReq(&req, marshaledObj, trWorkload.GroupVersionKind(), admissionv1.Delete)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
		Describe("InteractiveWorkload", func() {
			BeforeEach(func() {
				intWorkload = GetInteractiveWorkload(namespaceName, workloadName, userId)
				objects = append(objects, intWorkload)

				marshaledObj, err := json.Marshal(intWorkload)
				Expect(err).To(BeNil())

				populateReq(&req, marshaledObj, intWorkload.GroupVersionKind(), admissionv1.Update)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
		Describe("InferenceWorkload", func() {
			BeforeEach(func() {
				infWorkload = GetInferenceWorkload(namespaceName, workloadName, userId)
				objects = append(objects, infWorkload)

				marshaledObj, err := json.Marshal(infWorkload)
				Expect(err).To(BeNil())

				populateReq(&req, marshaledObj, infWorkload.GroupVersionKind(), admissionv1.Delete)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
		Describe("DistributedWorkload", func() {
			BeforeEach(func() {
				disWorkload = GetDistributedWorkload(namespaceName, workloadName, userId)
				objects = append(objects, disWorkload)

				marshaledObj, err := json.Marshal(disWorkload)
				Expect(err).To(BeNil())

				populateReq(&req, marshaledObj, disWorkload.GroupVersionKind(), admissionv1.Delete)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
	})
	Context("Validation Failures - because some objects are missing / bad request", func() {
		BeforeEach(func() {
			trWorkload = GetTrainingWorkload(namespaceName, workloadName, anotherUserId)
			objects = append(objects, trWorkload)

			marshaledObj, err := json.Marshal(trWorkload)
			Expect(err).To(BeNil())

			populateReq(&req, marshaledObj, trWorkload.GroupVersionKind(), admissionv1.Delete)
		})
		Describe("Missing Namespace", func() {
			BeforeEach(func() {
				testConf.createNamespace = false
				buildObjects(testConf)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Message).To(ContainSubstring("not found"))
				Expect(res.Result.Message).To(ContainSubstring("namespace"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusInternalServerError)))
			})
		})
		Describe("Missing Project", func() {
			BeforeEach(func() {
				testConf.createProject = false
				buildObjects(testConf)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Message).To(ContainSubstring("not found"))
				Expect(res.Result.Message).To(ContainSubstring("project"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusInternalServerError)))
			})
		})
		Describe("Missing ServiceAccount", func() {
			BeforeEach(func() {
				testConf.createServiceAccount = false
				buildObjects(testConf)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Message).To(ContainSubstring("not found"))
				Expect(res.Result.Message).To(ContainSubstring("serviceaccount"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusInternalServerError)))
			})
		})
		Describe("Fail to Decode OldObject", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				req.OldObject = runtime.RawExtension{
					Raw: []byte{},
				}
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(string(res.Result.Message)).To(Not(BeEmpty()))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Code).To(Equal(int32(http.StatusBadRequest)))
			})
		})
		Describe("Empty requester ID", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				req.UserInfo.UID = ""
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).ToNot(BeEmpty())
				Expect(string(res.Result.Reason)).To(ContainSubstring("Forbidden"))
				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(string(res.Result.Message)).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			})
		})
	})
	Context("Validation - Allow request even though different requesterId", func() {
		Describe("Empty Workload Owner Id", func() {
			BeforeEach(func() {
				intWorkload = GetInteractiveWorkload(namespaceName, workloadName, "")
				intWorkload.Spec.UserId = nil
				objects = append(objects, intWorkload)

				marshaledObj, err := json.Marshal(intWorkload)
				Expect(err).To(BeNil())

				populateReq(&req, marshaledObj, intWorkload.GroupVersionKind(), admissionv1.Delete)
			})
			Describe("Empty Workload Owner Id", func() {
				BeforeEach(func() {
					buildObjects(testConf)
				})
				It("CreateMergeStackForFrontend Handle", func() {
					res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

					Expect(res.Result.Reason).To(BeEmpty())
					Expect(res.Result.Message).To(BeEmpty())
					Expect(res.Allowed).To(BeTrue())
					Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
				})
			})
		})
		Describe("Non-Empty Workload Owner Id", func() {
			BeforeEach(func() {
				intWorkload = GetInteractiveWorkload(namespaceName, workloadName, anotherUserId)
				objects = append(objects, intWorkload)

				marshaledObj, err := json.Marshal(intWorkload)
				Expect(err).To(BeNil())

				populateReq(&req, marshaledObj, intWorkload.GroupVersionKind(), admissionv1.Delete)
			})
			Describe("Requester ID is workload-controller ID", func() {
				BeforeEach(func() {
					buildObjects(testConf)

					req.UserInfo.UID = workloadControllerUID
				})
				It("CreateMergeStackForFrontend Handle", func() {
					res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

					Expect(res.Result.Reason).To(BeEmpty())
					Expect(res.Result.Message).To(BeEmpty())
					Expect(res.Allowed).To(BeTrue())
					Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
				})
			})
			Describe("Requester ID is runtime-rules ID", func() {
				BeforeEach(func() {
					buildObjects(testConf)

					req.UserInfo.UID = runtimeRulesServiceAccountUID
				})
				It("CreateMergeStackForFrontend Handle", func() {
					res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

					Expect(res.Result.Reason).To(BeEmpty())
					Expect(res.Result.Message).To(BeEmpty())
					Expect(res.Allowed).To(BeTrue())
					Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
				})
			})
			Describe("Requester ID is not runai authenticated user", func() {
				BeforeEach(func() {
					buildObjects(testConf)

					req.UserInfo.UID = ""
					req.UserInfo.Username = "kubeadmin"
				})
				It("CreateMergeStackForFrontend Handle", func() {
					res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

					Expect(res.Result.Reason).To(BeEmpty())
					Expect(res.Result.Message).To(BeEmpty())
					Expect(res.Allowed).To(BeTrue())
					Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
				})
			})
			Describe("Requester ID is agent", func() {
				agentName := "agent"
				BeforeEach(func() {
					buildObjects(testConf)

					req.UserInfo.UID = ""
					req.UserInfo.Username = agentName
				})
				It("CreateMergeStackForFrontend Handle", func() {
					res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

					Expect(res.Result.Reason).To(BeEmpty())
					Expect(res.Result.Message).To(BeEmpty())
					Expect(res.Allowed).To(BeTrue())
					Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
				})
			})
			Describe("Sanity - request should fail cause group is authenticated, even though Requester ID is not runai authenticated user", func() {
				BeforeEach(func() {
					buildObjects(testConf)

					differentUserId := uuid.New().String()

					req.UserInfo.UID = differentUserId
					req.UserInfo.Groups = []string{testConf.groupId, uuid.New().String(), uuid.New().String()}
				})
				It("CreateMergeStackForFrontend Handle", func() {
					res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

					Expect(res.Result.Reason).ToNot(BeEmpty())
					Expect(string(res.Result.Reason)).To(ContainSubstring("Forbidden"))
					Expect(res.Result.Message).To(Not(BeEmpty()))
					Expect(string(res.Result.Message)).To(ContainSubstring("operation not allowed"))
					Expect(res.Allowed).To(BeFalse())
					Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
				})
			})
			Describe("Sanity - request should fail if not changing anything special", func() {
				BeforeEach(func() {
					buildObjects(testConf)
				})
				It("CreateMergeStackForFrontend Handle", func() {
					res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)
					Expect(res.Result.Reason).ToNot(BeEmpty())
					Expect(string(res.Result.Reason)).To(ContainSubstring("Forbidden"))
					Expect(res.Result.Message).To(Not(BeEmpty()))
					Expect(string(res.Result.Message)).To(ContainSubstring("operation not allowed"))
					Expect(res.Allowed).To(BeFalse())
					Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
				})
			})
		})
		Describe("Sanity - request should succeed cause Usage is not Submit - even though Missing Project", func() {
			BeforeEach(func() {
				disWorkload = GetDistributedWorkload(namespaceName, workloadName, userId)
				disWorkload.Spec.Usage = common.UsedAsTemplate
				objects = append(objects, disWorkload)

				marshaledObj, err := json.Marshal(disWorkload)
				Expect(err).To(BeNil())

				populateReq(&req, marshaledObj, disWorkload.GroupVersionKind(), admissionv1.Update)

				testConf.createProject = false
				buildObjects(testConf)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
	})
	Context("doesUserHaveWorkloadCrudPermissionsInHigherScope logic tests", func() {
		BeforeEach(func() {
			intWorkload = GetInteractiveWorkload(namespaceName, workloadName, anotherUserId)
			objects = append(objects, intWorkload)

			marshaledObj, err := json.Marshal(intWorkload)
			Expect(err).To(BeNil())

			populateReq(&req, marshaledObj, intWorkload.GroupVersionKind(), admissionv1.Delete)
		})
		Describe("Requester ID has PrivilegedWorkloadCrudPermissions - cluster scope", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				accessRule.Spec.ScopeType = v1alpha1.ClusterScope
				accessRule.Spec.SubjectType = v1alpha1.UserSubject
				accessRule.Spec.SubjectId = userId
				accessRule.Spec.RoleName = "system-administrator"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
		Describe("Requester ID has PrivilegedWorkloadCrudPermissions - department scope", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				accessRule.Spec.ScopeType = v1alpha1.DepartmentScope
				accessRule.Spec.ScopeName = departmentName
				accessRule.Spec.SubjectType = v1alpha1.UserSubject
				accessRule.Spec.SubjectId = userId
				accessRule.Spec.RoleName = "department-administrator"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
		Describe("Group has PrivilegedWorkloadCrudPermissions - cluster scope", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				groupId := uuid.New().String()
				req.UserInfo.Groups = []string{groupId, uuid.New().String(), uuid.New().String()}
				accessRule.Spec.ScopeType = v1alpha1.ClusterScope
				accessRule.Spec.SubjectType = v1alpha1.GroupSubject
				accessRule.Spec.SubjectId = groupId
				accessRule.Spec.RoleName = "department-administrator"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
		Describe("Group has PrivilegedWorkloadCrudPermissions - department scope", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				groupId := uuid.New().String()
				req.UserInfo.Groups = []string{groupId, uuid.New().String(), uuid.New().String()}
				accessRule.Spec.ScopeType = v1alpha1.DepartmentScope
				accessRule.Spec.ScopeName = departmentName
				accessRule.Spec.SubjectType = v1alpha1.GroupSubject
				accessRule.Spec.SubjectId = groupId
				accessRule.Spec.RoleName = "editor"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).To(BeEmpty())
				Expect(res.Result.Message).To(BeEmpty())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Result.Code).To(Equal(int32(http.StatusOK)))
			})
		})
		Describe("Requester ID has PrivilegedWorkloadCrudPermissions - but wrong department scope - should fail", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				accessRule.Spec.ScopeType = v1alpha1.DepartmentScope
				accessRule.Spec.ScopeName = departmentName + "-random-suffix-different-department"
				accessRule.Spec.SubjectType = v1alpha1.UserSubject
				accessRule.Spec.SubjectId = userId
				accessRule.Spec.RoleName = "system-administrator"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).ToNot(BeEmpty())
				Expect(string(res.Result.Reason)).To(ContainSubstring("Forbidden"))
				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(string(res.Result.Message)).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			})
		})
		Describe("Group has PrivilegedWorkloadCrudPermissions - but wrong department scope - should fail", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				groupId := uuid.New().String()
				req.UserInfo.Groups = []string{groupId, uuid.New().String(), uuid.New().String()}
				accessRule.Spec.ScopeType = v1alpha1.DepartmentScope
				accessRule.Spec.ScopeName = departmentName + "-different-department"
				accessRule.Spec.SubjectType = v1alpha1.GroupSubject
				accessRule.Spec.SubjectId = groupId
				accessRule.Spec.RoleName = "system-administrator"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).ToNot(BeEmpty())
				Expect(string(res.Result.Reason)).To(ContainSubstring("Forbidden"))
				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(string(res.Result.Message)).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			})
		})
		Describe("Requester ID or Group do not have PrivilegedWorkloadCrudPermissions - should fail", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				groupId := uuid.New().String()
				req.UserInfo.Groups = []string{groupId, uuid.New().String(), uuid.New().String()}
				randomUserId := uuid.New().String()
				accessRule.Spec.ScopeType = v1alpha1.ClusterScope
				accessRule.Spec.SubjectType = v1alpha1.UserSubject
				accessRule.Spec.SubjectId = randomUserId
				accessRule.Spec.RoleName = "system-administrator"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).ToNot(BeEmpty())
				Expect(string(res.Result.Reason)).To(ContainSubstring("Forbidden"))
				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(string(res.Result.Message)).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			})
		})
		Describe("Requester ID does not have PrivilegedWorkloadCrudPermissions (bad scope) - should fail", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				accessRule.Spec.ScopeType = v1alpha1.ProjectScope
				accessRule.Spec.SubjectType = v1alpha1.UserSubject
				accessRule.Spec.SubjectId = userId
				accessRule.Spec.RoleName = "system-administrator"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).ToNot(BeEmpty())
				Expect(string(res.Result.Reason)).To(ContainSubstring("Forbidden"))
				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(string(res.Result.Message)).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			})
		})
		Describe("Group does not have PrivilegedWorkloadCrudPermissions (bad role name) - should fail", func() {
			BeforeEach(func() {
				buildObjects(testConf)
				groupId := uuid.New().String()
				req.UserInfo.Groups = []string{groupId, uuid.New().String(), uuid.New().String()}
				accessRule.Spec.ScopeType = v1alpha1.ClusterScope
				accessRule.Spec.SubjectType = v1alpha1.GroupSubject
				accessRule.Spec.SubjectId = groupId
				accessRule.Spec.RoleName = "viewer"
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipProtectionValidator.Handle(context.TODO(), req)

				Expect(res.Result.Reason).ToNot(BeEmpty())
				Expect(string(res.Result.Reason)).To(ContainSubstring("Forbidden"))
				Expect(res.Result.Message).To(Not(BeEmpty()))
				Expect(string(res.Result.Message)).To(ContainSubstring("operation not allowed"))
				Expect(res.Allowed).To(BeFalse())
				Expect(res.Result.Code).To(Equal(int32(http.StatusForbidden)))
			})
		})
	})
	Context("Mutating CreateMergeStackForFrontend Tests", func() {
		BeforeEach(func() {
			buildObjects(testConf)
		})
		Describe("Create request - nil UserId - should patch", func() {
			BeforeEach(func() {
				intWorkload = GetInteractiveWorkload(namespaceName, workloadName, userId)
				intWorkload.Spec.UserId = nil
				objects = append(objects, intWorkload)

				marshaledObj, err := json.Marshal(intWorkload)
				Expect(err).To(BeNil())

				populateMutationReq(&req, marshaledObj, []byte{}, intWorkload.GroupVersionKind(), admissionv1.Create)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipMutator.Handle(context.TODO(), req)

				Expect(res.Result).To(BeNil())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Patches).To(HaveLen(1))
				Expect(res.Patches[0].Operation).To(Equal("add"))
				Expect(res.Patches[0].Path).To(Equal("/spec/userId"))

				value, ok := res.Patches[0].Value.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(value).To(HaveKey("value"))

				innerValue, ok := value["value"].(string)
				Expect(ok).To(BeTrue())
				Expect(innerValue).To(Equal(userId))
			})
		})
		Describe("Create request - wrong UserId - should patch", func() {
			BeforeEach(func() {
				infWorkload = GetInferenceWorkload(namespaceName, workloadName, userId)
				infWorkload.Spec.UserId = &flds.StringField{Value: &anotherUserId}
				objects = append(objects, infWorkload)

				marshaledObj, err := json.Marshal(infWorkload)
				Expect(err).To(BeNil())

				populateMutationReq(&req, marshaledObj, []byte{}, infWorkload.GroupVersionKind(), admissionv1.Create)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipMutator.Handle(context.TODO(), req)

				Expect(res.Result).To(BeNil())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Patches).To(HaveLen(1))
				Expect(res.Patches[0].Operation).To(Equal("replace"))
				Expect(res.Patches[0].Path).To(Equal("/spec/userId/value"))

				value, ok := res.Patches[0].Value.(string)
				Expect(ok).To(BeTrue())
				Expect(value).To(Equal(userId))
			})
		})
		Describe("Create request - correct UserId - should not patch", func() {
			BeforeEach(func() {
				infWorkload = GetInferenceWorkload(namespaceName, workloadName, userId)
				infWorkload.Spec.UserId = &flds.StringField{Value: &userId}
				objects = append(objects, infWorkload)

				marshaledObj, err := json.Marshal(infWorkload)
				Expect(err).To(BeNil())

				populateMutationReq(&req, marshaledObj, []byte{}, infWorkload.GroupVersionKind(), admissionv1.Create)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipMutator.Handle(context.TODO(), req)

				Expect(res.Result).To(BeNil())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Patches).To(HaveLen(0))
			})
		})
		Describe("Update request - nil UserId - should patch", func() {
			BeforeEach(func() {
				disWorkload = GetDistributedWorkload(namespaceName, workloadName, userId)
				objects = append(objects, disWorkload)

				oldMarshaledObj, err := json.Marshal(disWorkload)
				Expect(err).To(BeNil())

				disWorkload.Spec.UserId = nil
				marshaledObj, err := json.Marshal(disWorkload)
				Expect(err).To(BeNil())

				populateMutationReq(&req, marshaledObj, oldMarshaledObj, disWorkload.GroupVersionKind(), admissionv1.Update)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipMutator.Handle(context.TODO(), req)

				Expect(res.Result).To(BeNil())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Patches).To(HaveLen(1))
				Expect(res.Patches[0].Operation).To(Equal("add"))
				Expect(res.Patches[0].Path).To(Equal("/spec/userId"))

				value, ok := res.Patches[0].Value.(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(value).To(HaveKey("value"))

				innerValue, ok := value["value"].(string)
				Expect(ok).To(BeTrue())
				Expect(innerValue).To(Equal(userId))
			})
		})
		Describe("Update request - correct UserId - should not patch", func() {
			BeforeEach(func() {
				trWorkload = GetTrainingWorkload(namespaceName, workloadName, userId)
				trWorkload.Spec.UserId = &flds.StringField{Value: &userId}
				objects = append(objects, trWorkload)

				oldMarshaledObj, err := json.Marshal(trWorkload)
				Expect(err).To(BeNil())

				trWorkload.Spec.Labels = &flds.DictionaryField{}
				marshaledObj, err := json.Marshal(trWorkload)
				Expect(err).To(BeNil())

				populateMutationReq(&req, marshaledObj, oldMarshaledObj, trWorkload.GroupVersionKind(), admissionv1.Update)
			})
			It("CreateMergeStackForFrontend Handle", func() {
				res := workloadOwnershipMutator.Handle(context.TODO(), req)

				Expect(res.Result).To(BeNil())
				Expect(res.Allowed).To(BeTrue())
				Expect(res.Patches).To(HaveLen(0))
			})
		})
	})
})

func buildObjects(testConf *TestConf) {
	namespace = nil
	project = nil
	serviceAccount = nil
	objects = []runtime.Object{}

	if testConf.createNamespace {
		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
				Labels: map[string]string{
					"runai/queue": projectName,
				},
			},
		}
		objects = append(objects, namespace)
	}

	if testConf.createProject {
		project = &runaiv2.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: projectName,
			},
			Spec: runaiv2.ProjectSpec{
				Department: departmentName,
			},
		}
		// The kai.resources project carries the same parent/department so the webhook
		// resolves the department from the kai.resources project.
		kaiProject := &kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: projectName,
			},
			Spec: kaiorg.ProjectSpec{
				Parent: departmentName,
			},
		}
		accessRule1 := &v1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "access-rule1",
				Labels: map[string]string{},
			},
			Spec: v1alpha1.AccessRuleSpec{
				SubjectId:   testConf.userId,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   projectName,
				ScopeType:   v1alpha1.ProjectScope,
			},
		}
		accessRule2 := &v1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "access-rule2",
				Labels: map[string]string{},
			},
			Spec: v1alpha1.AccessRuleSpec{
				SubjectId:   testConf.anotherUserId,
				SubjectType: v1alpha1.UserSubject,
				ScopeName:   projectName,
				ScopeType:   v1alpha1.ProjectScope,
			},
		}
		accessRule3 := &v1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "access-rule3",
				Labels: map[string]string{},
			},
			Spec: v1alpha1.AccessRuleSpec{
				SubjectId:   testConf.groupId,
				SubjectType: v1alpha1.GroupSubject,
				ScopeName:   projectName,
				ScopeType:   v1alpha1.ProjectScope,
			},
		}
		accessRule4 := &v1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "access-rule4",
				Labels: map[string]string{},
			},
			Spec: v1alpha1.AccessRuleSpec{
				SubjectId:   uuid.New().String(),
				SubjectType: v1alpha1.GroupSubject,
				ScopeName:   projectName,
				ScopeType:   v1alpha1.ProjectScope,
			},
		}
		accessRule5 := &v1alpha1.AccessRule{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "access-rule5",
				Labels: map[string]string{},
			},
			Spec: v1alpha1.AccessRuleSpec{
				SubjectId:   uuid.New().String(),
				SubjectType: v1alpha1.GroupSubject,
				ScopeName:   projectName,
				ScopeType:   v1alpha1.ProjectScope,
			},
		}
		objects = append(objects, project, kaiProject, accessRule1, accessRule2, accessRule3, accessRule4, accessRule5)
	}

	if testConf.createServiceAccount {
		serviceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.WorkloadControllerOperandName,
				Namespace: constants.NamespaceRunAi,
				UID:       types.UID(testConf.workloadControllerUID),
			},
		}
		objects = append(objects, serviceAccount)

		serviceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.RuntimeRulesOperandName,
				Namespace: constants.NamespaceRunAi,
				UID:       types.UID(testConf.runtimeRulesServiceAccountUID),
			},
		}
		objects = append(objects, serviceAccount)
	}

	accessRule = GetAccessRule(accessRuleName)
	objects = append(objects, accessRule)
}

func populateReq(
	req *admission.Request,
	marshaledObj []byte,
	gvk schema.GroupVersionKind,
	operation admissionv1.Operation) {
	req.Operation = operation
	req.Kind = metav1.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
	req.OldObject = runtime.RawExtension{
		Raw: marshaledObj,
	}
}

func populateMutationReq(
	req *admission.Request,
	marshaledObj []byte,
	marshaledOldObj []byte,
	gvk schema.GroupVersionKind,
	operation admissionv1.Operation) {
	populateReq(req, marshaledOldObj, gvk, operation)
	req.Object = runtime.RawExtension{
		Raw: marshaledObj,
	}
}
