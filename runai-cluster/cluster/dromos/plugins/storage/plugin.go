package storage

import (
	"context"
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	PluginName = "storage-plugin"

	pvcAPIVersion = "v1"
	pvcKind       = "PersistentVolumeClaim"

	PvcsReadyConditionType = "PVCsReady"
	ReasonInitializing     = "Initializing"
	ReasonBound            = "Bound"
	ReasonFailed           = "Failed"

	msgAllPVCsBound         = "all PVCs are Bound"
	msgPVCStatusFmt         = "PVC(s) lost: %v, not created: %v, waiting to bind: %v"
	msgPVCsDeletedByCleanup = "workload reached terminal state, PVCs are deleted by cleanup"
)

var (
	ErrPvcOwnershipConflict = fmt.Errorf("PVC not controlled by this WorkloadRunner")
)

type Plugin struct {
	cleanupOnCompletionDefault bool
}

// NewPlugin builds the storage plugin
func NewPlugin(cleanupOnCompletionDefault bool) *Plugin {
	return &Plugin{
		cleanupOnCompletionDefault: cleanupOnCompletionDefault,
	}
}

// New is the framework.PluginBuilder; when STORAGE_ENABLED is off it returns a
// validator-only sentinel.
func New(_ context.Context, _ framework.PluginDeps) (framework.Plugin, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetDefault("STORAGE_ENABLED", true)
	v.SetDefault("STORAGE_CLEANUP_ON_COMPLETION_DEFAULT", false)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse storage config: %w", err)
	}
	if !cfg.Enabled {
		return framework.DisabledValidator(PluginName, storageInstructionOf, storageInstructionPath()), nil
	}
	return NewPlugin(cfg.CleanupOnCompletionDefault), nil
}

func storageInstructionOf(runner *dromosv1alpha1.WorkloadRunner) any {
	if runner.Spec.Instructions == nil ||
		runner.Spec.Instructions.Storage == nil ||
		len(runner.Spec.Instructions.Storage.NewPvcs) == 0 {
		return nil
	}
	return runner.Spec.Instructions.Storage
}

func storageInstructionPath() *field.Path {
	return field.NewPath("spec", "instructions", "storage")
}

// cleanEnabled resovles if the pvcs should be deleted
func (p *Plugin) cleanEnabled(storage *dromosv1alpha1.StorageConfig) bool {
	if storage.CleanupOnCompletion != nil {
		return *storage.CleanupOnCompletion
	}
	return p.cleanupOnCompletionDefault
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	return bldr.Owns(&corev1.PersistentVolumeClaim{}), nil
}

func (p *Plugin) errWithIssue(ctx context.Context, state *reconcile.State, object client.Object, err error) error {
	return state.AddPluginIssue(ctx, p, object, err, true, "", issue.Storage)
}

func (p *Plugin) HandleAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	_ *unstructured.Unstructured,
	_ *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	if runner.Spec.Instructions == nil || runner.Spec.Instructions.Storage == nil {
		return nil
	}

	logger := log.FromContext(ctx).WithName(PluginName).WithValues("resource_phase", runner.Status.ResourcePhase)

	if p.shouldCleanupStorage(runner) {
		// We don't want to recreate the PVC incase it's deleted in terminal state
		return nil
	}

	for _, inputPvc := range runner.Spec.Instructions.Storage.NewPvcs {
		pvc := inputPvc.DeepCopy()
		if pvc.Namespace == "" {
			pvc.Namespace = runner.Namespace
		}

		pvc.APIVersion = pvcAPIVersion
		pvc.Kind = pvcKind

		existing := &corev1.PersistentVolumeClaim{}
		key := types.NamespacedName{Namespace: pvc.Namespace, Name: pvc.Name}

		logger.V(1).Info("handling new pvc reconcile", "pvc", key.String())

		// If PVC already exists, check if it's controlled by this WorkloadRunner
		switch err := k8sClient.Get(ctx, key, existing); {
		case apierrors.IsNotFound(err):
		case err != nil:
			return p.errWithIssue(ctx, state, pvc, fmt.Errorf("get pvc %s: %w", key, err))
		case !metav1.IsControlledBy(existing, runner):
			return p.errWithIssue(ctx, state, pvc, fmt.Errorf("pvc %s: %w", key, ErrPvcOwnershipConflict))
		}

		if err := ctrl.SetControllerReference(runner, pvc, k8sClient.Scheme()); err != nil {
			return p.errWithIssue(ctx, state, pvc, fmt.Errorf("setControllerReference on pvc %s: %w", key, err))
		}

		if err := k8sClient.Patch(ctx, pvc, client.Apply, client.ForceOwnership, client.FieldOwner(PluginName)); err != nil {
			return p.errWithIssue(ctx, state, pvc, fmt.Errorf("apply pvc %s: %w", key, err))
		}
	}

	return nil
}

func (p *Plugin) WebhookValidate(_ context.Context, _, desiredRunner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, field.ErrorList) {
	if desiredRunner.Spec.Instructions == nil || desiredRunner.Spec.Instructions.Storage == nil {
		return nil, nil
	}

	var errs field.ErrorList
	basePath := field.NewPath("spec", "instructions", "storage", "newPvcs")
	seen := map[string]bool{}

	for i, pvc := range desiredRunner.Spec.Instructions.Storage.NewPvcs {
		path := basePath.Index(i)

		// Name not empty
		if pvc.Name == "" {
			errs = append(errs, field.Required(path.Child("metadata", "name"), "name is required"))
		}

		// Unique names
		if _, dup := seen[pvc.Name]; dup {
			errs = append(errs, field.Duplicate(path.Child("metadata", "name"), pvc.Name))
		}

		seen[pvc.Name] = true

		// PVC namespace eventually must equal to the WorkloadRunner namespace since the owner reference requires it
		if pvc.Namespace != "" && pvc.Namespace != desiredRunner.Namespace {
			errs = append(errs,
				field.Invalid(
					path.Child("metadata", "namespace"),
					pvc.Namespace,
					fmt.Sprintf("namespace must match WorkloadRunner %q", desiredRunner.Namespace)))
		}

		if len(pvc.Spec.AccessModes) == 0 {
			errs = append(errs, field.Required(path.Child("spec", "accessModes"), "at least one access mode is required"))
		}

		if _, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; !ok {
			errs = append(errs, field.Required(path.Child("spec", "resources", "requests", "storage"), "storage request is required"))
		}
	}

	return nil, errs
}

func (p *Plugin) Status(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	k8sClient client.Client,
	status framework.PluginStatusSetter,
	state *reconcile.State,
) error {
	logger := log.FromContext(ctx).WithName(PluginName)

	if runner.Spec.Instructions == nil || runner.Spec.Instructions.Storage == nil || len(runner.Spec.Instructions.Storage.NewPvcs) == 0 {
		status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)
		return nil
	}

	for _, iss := range state.PluginIssues(p) {
		if iss.IsCritical {
			logger.V(1).Info("critical issue found in state, marking plugin status as Failed", "issue", iss)
			status.SetCondition(metav1.Condition{
				Type:    PvcsReadyConditionType,
				Status:  metav1.ConditionFalse,
				Reason:  ReasonFailed,
				Message: iss.Error(),
			})
			status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
			return nil
		}
	}

	if p.shouldCleanupStorage(runner) {
		status.SetCondition(metav1.Condition{
			Type:    PvcsReadyConditionType,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonBound,
			Message: msgPVCsDeletedByCleanup,
		})
		status.SetPhase(dromosv1alpha1.PluginPhaseReady)
		return nil
	}

	var notCreated, notBound, lost []string
	for i := range runner.Spec.Instructions.Storage.NewPvcs {
		pvcSpec := &runner.Spec.Instructions.Storage.NewPvcs[i]
		ns := pvcSpec.Namespace
		if ns == "" {
			ns = runner.Namespace
		}
		key := types.NamespacedName{Namespace: ns, Name: pvcSpec.Name}

		live := &corev1.PersistentVolumeClaim{}
		err := k8sClient.Get(ctx, key, live)
		if err != nil && !apierrors.IsNotFound(err) {
			status.SetCondition(metav1.Condition{
				Type:    PvcsReadyConditionType,
				Status:  metav1.ConditionUnknown,
				Reason:  ReasonFailed,
				Message: fmt.Sprintf("get pvc %s: %v", key, err),
			})
			status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
			return fmt.Errorf("get pvc %s: %w", key, err)
		}
		switch {
		case apierrors.IsNotFound(err):
			notCreated = append(notCreated, key.String())
		case live.Status.Phase == corev1.ClaimBound:
		case live.Status.Phase == corev1.ClaimLost:
			lost = append(lost, key.String())
		default:
			notBound = append(notBound, key.String())
		}
	}
	// Phase represents the worst across all PVCs
	// Lost (Failed) > NotCreated (Initializing) > NotBound (Initializing) > Ready (Bound)
	switch {
	case len(lost) > 0:
		status.SetCondition(metav1.Condition{
			Type:    PvcsReadyConditionType,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonFailed,
			Message: fmt.Sprintf(msgPVCStatusFmt, lost, notCreated, notBound),
		})
		status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
	case len(notCreated) > 0:
		status.SetCondition(metav1.Condition{
			Type:    PvcsReadyConditionType,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonInitializing,
			Message: fmt.Sprintf(msgPVCStatusFmt, lost, notCreated, notBound),
		})
		status.SetPhase(dromosv1alpha1.PluginPhaseCreating)
	case len(notBound) > 0:
		status.SetCondition(metav1.Condition{
			Type:    PvcsReadyConditionType,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonInitializing,
			Message: fmt.Sprintf(msgPVCStatusFmt, lost, notCreated, notBound),
		})
		status.SetPhase(dromosv1alpha1.PluginPhaseInitializing)
	default:
		status.SetCondition(metav1.Condition{
			Type:    PvcsReadyConditionType,
			Status:  metav1.ConditionTrue,
			Reason:  ReasonBound,
			Message: msgAllPVCsBound,
		})
		status.SetPhase(dromosv1alpha1.PluginPhaseReady)
	}

	return nil

}

// HandlePostPhase deletes the PVCs this WorkloadRunner created once it reaches a terminal phase (Completed/Failed)
func (p *Plugin) HandlePostPhase(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	_ *unstructured.Unstructured,
	_ *kartav1alpha1.Karta,
	k8sClient client.Client,
) error {

	if runner.Spec.Instructions == nil || runner.Spec.Instructions.Storage == nil {
		return nil
	}

	if !p.shouldCleanupStorage(runner) {
		return nil
	}

	logger := log.FromContext(ctx).WithName(PluginName).WithValues("phase", runner.Status.ResourcePhase)

	for _, inputPvc := range runner.Spec.Instructions.Storage.NewPvcs {
		namespace := inputPvc.Namespace
		if namespace == "" {
			namespace = runner.Namespace
		}
		key := types.NamespacedName{Namespace: namespace, Name: inputPvc.Name}

		pvc := &corev1.PersistentVolumeClaim{}
		switch err := k8sClient.Get(ctx, key, pvc); {
		case apierrors.IsNotFound(err):
			continue
		case err != nil:
			return fmt.Errorf("get pvc %s: %w", key, err)
		}

		// Only delete PVCs this WorkloadRunner created
		if !metav1.IsControlledBy(pvc, runner) {
			continue
		}

		logger.Info("deleting owned pvc on terminal phase", "pvc", key.String())

		if err := k8sClient.Delete(ctx, pvc); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("delete pvc %s: %w", key, err)
		}
	}

	return nil
}

// shouldCleanupStorage reports whether the runner's PVCs should be cleaned up:
// cleanup is enabled and the runner reached a terminal phase
func (p *Plugin) shouldCleanupStorage(runner *dromosv1alpha1.WorkloadRunner) bool {
	return p.cleanEnabled(runner.Spec.Instructions.Storage) && isTerminal(runner.Status.Phase, runner.Status.ResourcePhase)
}

// isTerminal reports whether the WorkloadRunner has reached a terminal phase
func isTerminal(runnerPhase, resourcePhase dromosv1alpha1.WorkloadRunnerPhase) bool {
	isResourceTerminal := resourcePhase == dromosv1alpha1.WorkloadRunnerPhaseCompleted || resourcePhase == dromosv1alpha1.WorkloadRunnerPhaseFailed
	return isResourceTerminal && runnerPhase == resourcePhase
}

var (
	_ framework.WatcherPlugin            = &Plugin{}
	_ framework.AdditionalResourcePlugin = &Plugin{}
	_ framework.ValidatorPlugin          = &Plugin{}
	_ framework.StatusPlugin             = &Plugin{}
	_ framework.PostPhasePlugin          = &Plugin{}
)
