/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package admissionpolicy

import (
	"fmt"

	admissionv1 "k8s.io/api/admissionregistration/v1"
)

type Definition struct {
	Policy  *admissionv1.ValidatingAdmissionPolicy
	Binding *admissionv1.ValidatingAdmissionPolicyBinding
	Enabled func() bool
}

func (d Definition) deepCopy() Definition {
	copy := d
	if d.Policy != nil {
		copy.Policy = d.Policy.DeepCopy()
	}
	if d.Binding != nil {
		copy.Binding = d.Binding.DeepCopy()
	}
	return copy
}

func (d Definition) validate() error {
	if d.Policy == nil {
		return fmt.Errorf("validating admission policy is required")
	}
	if d.Binding == nil {
		return fmt.Errorf("validating admission policy binding is required")
	}
	if d.Policy.Name == "" {
		return fmt.Errorf("validating admission policy name is required")
	}
	if d.Binding.Name == "" {
		return fmt.Errorf("validating admission policy binding name is required")
	}
	if d.Policy.Namespace != "" {
		return fmt.Errorf("validating admission policy [%s] must be cluster-scoped", d.Policy.Name)
	}
	if d.Binding.Namespace != "" {
		return fmt.Errorf("validating admission policy binding [%s] must be cluster-scoped", d.Binding.Name)
	}
	if d.Binding.Spec.PolicyName != d.Policy.Name {
		return fmt.Errorf("validating admission policy binding [%s] references policy [%s], expected [%s]",
			d.Binding.Name, d.Binding.Spec.PolicyName, d.Policy.Name)
	}
	return nil
}

func (d Definition) enabled() bool {
	return d.Enabled == nil || d.Enabled()
}
