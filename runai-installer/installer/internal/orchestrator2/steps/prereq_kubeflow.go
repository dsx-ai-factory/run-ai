/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	trainingOperatorNamespace = "kubeflow"
	trainingOperatorSentinel  = "training-operator"
	trainingOperatorOverlay   = "manifests/overlays/standalone"
	// trainingOperatorTarball is the GitHub source archive for a tag. The tag
	// already carries the "v" prefix (e.g. "v1.8.1"), matching the compat lock.
	trainingOperatorTarball = "https://codeload.github.com/kubeflow/training-operator/tar.gz/refs/tags/%s"

	mpiOperatorNamespace = "mpi-operator"
	mpiOperatorSentinel  = "mpi-operator"
	mpiOperatorManifest  = "https://raw.githubusercontent.com/kubeflow/mpi-operator/%s/deploy/v2beta1/mpi-operator.yaml"

	mpiJobsCRD = "mpijobs.kubeflow.org"
)

// trainingOperatorMPIPatch disables the training operator's built-in MPI v1
// reconciler so it does not fight the standalone MPI operator over MPIJob. It is
// applied only when both are installed (see the coexistence step). It is a
// strategic-merge patch targeting the "training-operator" container by name, so
// it does not depend on the container's position in the pod spec.
var trainingOperatorMPIPatch = []byte(`{"spec":{"template":{"spec":{"containers":[{"name":"training-operator","args":["--enable-scheme=tfjob","--enable-scheme=pytorchjob","--enable-scheme=xgboostjob"]}]}}}}`)

// NewKubeflowTraining builds the Kubeflow Training Operator install step
// (TFJob/PyTorchJob/XGBoostJob, distributed via a kustomize overlay rather than
// a Helm chart). Detect is the standard Deployment sentinel.
func NewKubeflowTraining(version string, deps orchestrator2.Deps) orchestrator2.Step {
	action := actions.ApplyKustomize{
		TarballURL: fmt.Sprintf(trainingOperatorTarball, version),
		SubPath:    trainingOperatorOverlay,
	}
	return newPrereqStep("install-training-operator", deploymentSentinel(trainingOperatorSentinel, trainingOperatorNamespace), deps, action)
}

// NewMPIOperator builds the standalone MPI Operator install step (MPIJob
// v2beta1), applied from its flat upstream manifest.
func NewMPIOperator(version string, deps orchestrator2.Deps) orchestrator2.Step {
	action := actions.ApplyRemoteManifest{URL: fmt.Sprintf(mpiOperatorManifest, version)}
	return newPrereqStep("install-mpi-operator", deploymentSentinel(mpiOperatorSentinel, mpiOperatorNamespace), deps, action)
}

// mpiCoexistenceStep reconciles the Training Operator and the standalone MPI
// Operator so they do not both own MPIJob: it disables MPI v1 in the training
// operator and removes the conflicting v1 mpijobs CRD. It is planned only when
// both operators are requested, and runs between them (the v1 CRD must be gone
// before the MPI operator installs its v2beta1 CRD). Detect keys off the
// presence of a v1 version in the mpijobs CRD, so the step is idempotent: once
// reconciled, the only mpijobs CRD left is the MPI operator's v2beta1 one, which
// must not be deleted.
type mpiCoexistenceStep struct {
	*orchestrator2.BaseStep
	deps orchestrator2.Deps
}

// NewMPICoexistence builds the Training/MPI coexistence reconciliation step.
func NewMPICoexistence(deps orchestrator2.Deps) orchestrator2.Step {
	return &mpiCoexistenceStep{BaseStep: orchestrator2.NewBaseStep("reconcile-training-mpi-coexistence"), deps: deps}
}

func mpiJobsCRDObject() *unstructured.Unstructured {
	crd := &unstructured.Unstructured{}
	crd.SetAPIVersion("apiextensions.k8s.io/v1")
	crd.SetKind("CustomResourceDefinition")
	crd.SetName(mpiJobsCRD)
	return crd
}

func (s *mpiCoexistenceStep) Detect(ctx context.Context, sp *spec.Spec) (orchestrator2.DetectResult, error) {
	// In dry-run, surface the (potentially destructive) reconciliation actions
	// instead of short-circuiting: on a fresh/dry cluster the training operator's
	// v1 CRD has not actually been created, so the CRD-based check below would
	// otherwise report "nothing to do" and hide the v1 mpijobs CRD deletion from
	// the dry-run preview.
	if sp.DryRun {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
	}

	crd := mpiJobsCRDObject()
	err := s.deps.K8sExecutor.Get(ctx, client.ObjectKey{Name: mpiJobsCRD}, crd)
	if apierrors.IsNotFound(err) {
		// No mpijobs CRD at all: nothing to reconcile.
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled}, nil
	}
	if err != nil {
		return orchestrator2.DetectResult{}, err
	}
	// Reconcile only while the training operator's v1 mpijobs CRD is present.
	// Once reconciled, the only mpijobs CRD is the MPI operator's v2beta1 one,
	// which must be left alone.
	if crdHasVersion(crd, "v1") {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
	}
	return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled}, nil
}

// crdHasVersion reports whether the CRD declares the given served version name
// in spec.versions.
func crdHasVersion(crd *unstructured.Unstructured, name string) bool {
	versions, found, err := unstructured.NestedSlice(crd.Object, "spec", "versions")
	if !found || err != nil {
		return false
	}
	for _, v := range versions {
		if m, ok := v.(map[string]any); ok && m["name"] == name {
			return true
		}
	}
	return false
}

func (s *mpiCoexistenceStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	return []orchestrator2.PerformAction{
		actions.PatchDeployment{Namespace: trainingOperatorNamespace, Name: trainingOperatorSentinel, StrategicMergePatch: trainingOperatorMPIPatch},
		actions.DeleteCRD{Name: mpiJobsCRD},
		actions.WaitForCRDDeleted{Name: mpiJobsCRD},
	}, nil
}
