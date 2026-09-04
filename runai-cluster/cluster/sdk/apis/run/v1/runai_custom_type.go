/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConditionTypeReconciled               ConditionType = "Reconciled"
	ConditionTypeReconciledSuccessMessage string        = "Reconciliation completed successfully"
	ConditionTypeReconciledLoopHaveNotRun string        = "Reconciliation have not completed first run"
)

type ConditionType string
type ConditionReason string

// +kubebuilder:object:generate=false
type RunaiCustomTypeI interface {
	client.Object
	GetRunaiCustomStatus() *RunaiCustomTypeStatus
}

type RunaiCustomTypeStatus struct {
	// Conditions Defines the observed state of the CRD in the cluster
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration specifies the latest generation of CRD
	// the operator was aware of
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

func (r *RunaiCustomTypeStatus) GetCondition(typ ConditionType) *metav1.Condition {
	for i, condition := range r.Conditions {
		if condition.Type == string(typ) {
			return &r.Conditions[i]
		}
	}
	return nil
}
