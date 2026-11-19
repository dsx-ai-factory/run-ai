/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apimachinery/pkg/util/validation/field"
	celconfig "k8s.io/apiserver/pkg/apis/cel"
	"sigs.k8s.io/yaml"
)

const runaiConfigCRDPath = "../../../../sdk/config/crd/bases/run.ai_runaiconfigs.yaml"

// findGatewayRefSchema walks the schema tree and returns the first object schema
// carrying the GatewayRef CEL validation (identified by its message).
func findGatewayRefSchema(s *apiextensionsv1.JSONSchemaProps) *apiextensionsv1.JSONSchemaProps {
	if s == nil {
		return nil
	}
	for _, rule := range s.XValidations {
		if strings.Contains(rule.Message, "gatewayAPI override") {
			return s
		}
	}
	for _, child := range s.Properties {
		if found := findGatewayRefSchema(&child); found != nil {
			return found
		}
	}
	if s.Items != nil {
		return findGatewayRefSchema(s.Items.Schema)
	}
	return nil
}

// validateGatewayRef runs the generated GatewayRef CEL rules against obj.
func validateGatewayRef(obj map[string]interface{}) field.ErrorList {
	raw, err := os.ReadFile(runaiConfigCRDPath)
	Expect(err).NotTo(HaveOccurred())
	crd := &apiextensionsv1.CustomResourceDefinition{}
	Expect(yaml.Unmarshal(raw, crd)).To(Succeed())

	v1Schema := findGatewayRefSchema(crd.Spec.Versions[0].Schema.OpenAPIV3Schema)
	Expect(v1Schema).NotTo(BeNil())

	internal := &apiextensions.JSONSchemaProps{}
	Expect(apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(v1Schema, internal, nil)).To(Succeed())
	structural, err := structuralschema.NewStructural(internal)
	Expect(err).NotTo(HaveOccurred())

	validator := cel.NewValidator(structural, false, celconfig.PerCallLimit)
	Expect(validator).NotTo(BeNil())
	errs, _ := validator.Validate(context.Background(), field.NewPath("gatewayAPI"), structural, obj, nil, celconfig.RuntimeCELCostBudget)
	return errs
}

var _ = Describe("GatewayRef CEL validation", func() {
	It("accepts an empty override", func() {
		Expect(validateGatewayRef(map[string]interface{}{})).To(BeEmpty())
	})

	It("accepts a full override with sectionName", func() {
		Expect(validateGatewayRef(map[string]interface{}{
			"name": "gw", "namespace": "ns", "sectionName": "https",
		})).To(BeEmpty())
	})

	It("rejects sectionName without name and namespace", func() {
		Expect(validateGatewayRef(map[string]interface{}{"sectionName": "https"})).NotTo(BeEmpty())
	})
})
