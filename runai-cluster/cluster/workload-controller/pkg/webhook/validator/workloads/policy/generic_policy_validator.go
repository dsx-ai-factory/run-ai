/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/common/utils/validator"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/nodepools"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	interactivepolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	policyUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/policy"
	rulescommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/utils/runai_rules"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
	admissionV1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-run-ai-v2alpha1-interactivepolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=run.ai,resources=interactivepolicy,verbs=create;update:delete,versions=v2alpha1,name=runai-interactive-workload-validation.kb.io,admissionReviewVersions=v1
// +kubebuilder:webhook:path=/validate-run-ai-v2alpha1-trainingpolicy,mutating=false,failurePolicy=fail,sideEffects=None,groups=run.ai,resources=interactivepolicy,verbs=create;update:delete,versions=v2alpha1,name=runai-interactive-workload-validation.kb.io,admissionReviewVersions=v1

type Adapter interface {
	NewPolicyList() client.ObjectList
	NewEmptyPolicy() interface{}
	GetContainerFields() *field_groups.ContainerFields
	GetAccessControlFields() *field_groups.AccessControlFields
	GetRunaiSchedulerFields() *field_groups.RunaiSchedulerFields
	GetMeta() metav1.ObjectMeta
	GetSpec() interface{}
}

type ValidatorHandler struct {
	client.Client
	PolicySyncEnabledFlag  bool
	CreatePolicyAdapterFN  func(req admission.Request) (Adapter, error)
	AllowedServiceAccounts []string
}

func NewPolicyValidator(k8sClient client.Client, policyAdapterFN func(req admission.Request) (Adapter, error), allowedServiceAccounts []string) *ValidatorHandler {
	return &ValidatorHandler{
		Client:                 k8sClient,
		CreatePolicyAdapterFN:  policyAdapterFN,
		AllowedServiceAccounts: allowedServiceAccounts,
	}
}

func (vh *ValidatorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := logf.FromContext(ctx)
	policyKind := req.RequestKind.Kind

	policyInterface, err := vh.CreatePolicyAdapterFN(req)

	log.Info(fmt.Sprintf("Handle policy validation handler for: %s, operation: %s", policyKind, req.Operation))

	requesterUsername := req.UserInfo.Username

	vh.PolicySyncEnabledFlag = utils.IsControlPlanePoliciesEnabled()

	isControlPlanePolicy, err := isControlPlanePolicy(req)

	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if isControlPlanePolicy {
		err := vh.enforceOwnership(policyKind, requesterUsername, vh.AllowedServiceAccounts)
		if err != nil {
			return admission.Errored(http.StatusUnauthorized, err)
		}
	}

	switch req.Operation {
	case admissionV1.Create:
		return vh.HandleCreate(ctx, policyInterface, req)
	case admissionV1.Update:
		return vh.HandleUpdate(ctx, policyInterface, req)
	case admissionV1.Delete:
		return vh.HandleDelete()
	default:
		return admission.Allowed("Ok")
	}
}

func (vh *ValidatorHandler) HandleCreate(ctx context.Context, policyInterface Adapter, req admission.Request) admission.Response {
	log := logf.FromContext(ctx)

	namespace := req.Namespace

	policyList := policyInterface.NewPolicyList()
	policy := policyInterface.NewEmptyPolicy()

	err := json.Unmarshal(req.Object.Raw, &policy)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	policyLabels := policyInterface.GetMeta().Labels
	if req.Operation == admissionV1.Update {
		err = vh.ValidatePolicyExists(ctx, policyList, namespace, policyLabels)
	} else {
		err = vh.ValidatePolicyNotExists(ctx, policyList, namespace, policyLabels)
	}
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	policySpec, err := utils.NewJsonQ(policyInterface.GetSpec())
	if err != nil {
		log.Error(err, "Failed to load policy as JSON", "id", workloadproject.NamespacedName(policyInterface.GetMeta()))
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("Failed to parse %s.Spec", workloadproject.NamespacedName(policyInterface.GetMeta())))
	}

	runaiRules := interactivepolicy.GetRunaiRules(ctx, nodepools.NewNameLister(vh.Client))

	runaiRulesQ, err := utils.NewJsonQ(runaiRules)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	err = enforcer.ValidateSyntaxOfPolicy("", policySpec, runaiRulesQ)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	runaiRulesContainerFields := runaiRules.ContainerFields
	runaiRulesAccessControlFields := runaiRules.AccessControlFields
	policyContainerFields := policyInterface.GetContainerFields()
	policyAccessControlFields := policyInterface.GetAccessControlFields()
	policyRunaiSchedulerFields := policyInterface.GetRunaiSchedulerFields()

	rulescommon.PatchPolicyWithRunaiDefaults(ctx, &runaiRulesContainerFields, policyContainerFields, &runaiRulesAccessControlFields,
		policyAccessControlFields, policyRunaiSchedulerFields)

	emptyPolicyQ, _ := utils.NewJsonQ(policyInterface.GetSpec())

	mergeStack := enforcer.NewMergeStack(emptyPolicyQ, "", runaiRulesQ.JumpToTop(), false)
	mergeStack.AddLevel(policySpec, enforcer.LevelPolicy, policyInterface.GetMeta().Name)

	err = mergeStack.MergeAndValidate(ctx, enforcer.MergeAndValidateOptions{})
	if err != nil {
		log.Error(err, "Validation has failed")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.Allowed("Ok")
}

func (vh *ValidatorHandler) HandleUpdate(ctx context.Context, policyInterface Adapter, req admission.Request) admission.Response {
	return vh.HandleCreate(ctx, policyInterface, req)
}

func (vh *ValidatorHandler) HandleDelete() admission.Response {
	return admission.Allowed("Ok")
}

func (vh *ValidatorHandler) ValidatePolicyNotExists(ctx context.Context, policyList client.ObjectList, namespace string, policyLabels map[string]string) error {

	err := vh.getPolicyList(ctx, policyList, namespace, policyLabels)
	if err != nil {
		return err
	}

	policyListSize, err := policyUtils.GetNumberOfPolicies(policyList)
	if err != nil {
		return err
	}
	if policyListSize > 0 {
		return fmt.Errorf("Cant create new policies while there are %d policies with the same labels and namespace", policyListSize)
	}

	return nil
}

func (vh *ValidatorHandler) ValidatePolicyExists(ctx context.Context, policyList client.ObjectList, namespace string, policyLabels map[string]string) error {

	err := vh.getPolicyList(ctx, policyList, namespace, policyLabels)
	if err != nil {
		return err
	}

	policyListSize, err := policyUtils.GetNumberOfPolicies(policyList)
	if err != nil {
		return err
	}
	if policyListSize != 1 {
		return fmt.Errorf("Cant update policy while there are %d policies with the same labels and namespace", policyListSize)
	}

	return nil
}

func (vh *ValidatorHandler) getPolicyList(ctx context.Context, policyList client.ObjectList, namespace string, policyLabels map[string]string) error {
	if vh.PolicySyncEnabledFlag {
		// list policies with the same labels and same namespace
		err := vh.List(ctx, policyList, client.InNamespace(namespace), client.MatchingLabels(policyLabels))
		if err != nil {
			return err
		}
	} else {
		err := vh.List(ctx, policyList, client.InNamespace(namespace))
		if err != nil {
			return err
		}
	}
	return nil
}

func (vh *ValidatorHandler) enforceOwnership(policyKind string, requesterUsername string, allowedServiceAccounts []string) error {
	if vh.PolicySyncEnabledFlag {
		switch policyKind {
		case constants.KindInteractivePolicy, constants.KindTrainingPolicy, constants.KindDistributedPolicy, constants.KindInferencePolicy:
			if !validator.AllowedServiceAccounts(requesterUsername, allowedServiceAccounts) {
				return fmt.Errorf("Username %s is not authorized to perform this action on %s", requesterUsername, policyKind)
			}
		}
	}

	return nil
}

func (vh *ValidatorHandler) InjectDecoder(_ *admission.Decoder) error {
	return nil
}

type CommonPolicyObjectMeta struct {
	metav1.ObjectMeta `json:"metadata,omitempty"`
}

func isControlPlanePolicy(req admission.Request) (bool, error) {
	var policyLabels map[string]string
	var err error

	var policyRaw []byte
	if req.Operation == admissionV1.Create {
		policyRaw = req.Object.Raw
	} else {
		policyRaw = req.OldObject.Raw
	}

	common := CommonPolicyObjectMeta{}
	err = json.Unmarshal(policyRaw, &common)
	if err != nil {
		return false, err
	}
	policyLabels = common.ObjectMeta.Labels

	labelValue, ok := policyLabels[constants.PolicyLabel]
	if !ok {
		return false, nil
	}

	return labelValue == constants.BoolTrueAsString, nil
}
