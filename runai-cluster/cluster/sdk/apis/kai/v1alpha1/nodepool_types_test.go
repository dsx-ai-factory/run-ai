package v1alpha1

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
)

func TestKaiV1alpha1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "kai.resources/v1alpha1 Custom Resources")
}

// crdPath points at the generated NodePool CRD, relative to this package dir.
const crdPath = "../../../config/crd/bases/kai.resources_nodepools.yaml"

// immutabilityRuleFromCRD reads the labelKey/labelValue immutability rule
// straight out of the generated CRD so this test always exercises whatever rule
// actually ships. If the marker is ever changed to a form that breaks the
// default nodepool (e.g. a plain "self.labelKey == oldSelf.labelKey" that errors
// on absent fields), the specs below fail.
func immutabilityRuleFromCRD() string {
	data, err := os.ReadFile(filepath.Clean(crdPath))
	Expect(err).NotTo(HaveOccurred())

	var crd apiextensionsv1.CustomResourceDefinition
	Expect(yaml.Unmarshal(data, &crd)).To(Succeed())

	for _, v := range crd.Spec.Versions {
		if v.Schema == nil || v.Schema.OpenAPIV3Schema == nil {
			continue
		}
		spec, ok := v.Schema.OpenAPIV3Schema.Properties["spec"]
		if !ok {
			continue
		}
		for _, r := range spec.XValidations {
			if strings.Contains(r.Message, "immutable") {
				return r.Rule
			}
		}
	}
	Fail("no labelKey/labelValue immutability rule found in " + crdPath)
	return ""
}

// evaluateSpecRule runs the given CEL rule (as an x-kubernetes-validations rule
// on the spec) through the same validator the API server uses, and reports
// whether the transition from oldObj to newObj is allowed.
func evaluateSpecRule(rule string, oldObj, newObj map[string]interface{}) (allowed bool, msg string) {
	props := &apiextensions.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensions.JSONSchemaProps{
			"labelKey":   {Type: "string"},
			"labelValue": {Type: "string"},
		},
		XValidations: apiextensions.ValidationRules{{Rule: rule, Message: "labelKey and labelValue are immutable"}},
	}
	s, err := structuralschema.NewStructural(props)
	Expect(err).NotTo(HaveOccurred())

	v := cel.NewValidator(s, true, math.MaxUint32)
	errs, _ := v.Validate(context.Background(), field.NewPath("spec"), s, newObj, oldObj, math.MaxInt64)
	if len(errs) == 0 {
		return true, ""
	}
	return false, errs.ToAggregate().Error()
}

// Field representations. With omitempty, an unset field is absent from the
// stored object (empty map); a field sent explicitly as "" is present-but-empty.
var (
	labelsAbsent  = map[string]interface{}{}
	labelsGPUA100 = map[string]interface{}{"labelKey": "gpu", "labelValue": "a100"}
	labelsGPUH100 = map[string]interface{}{"labelKey": "gpu", "labelValue": "h100"}
)

var _ = Describe("NodePool labelKey/labelValue immutability CRD rule", func() {
	var rule string

	BeforeEach(func() {
		rule = immutabilityRuleFromCRD()
	})

	It("allows a no-op update of the default nodepool (both fields absent)", func() {
		allowed, msg := evaluateSpecRule(rule, labelsAbsent, labelsAbsent)
		Expect(allowed).To(BeTrue(), "default no-op must be allowed, got: %s", msg)
	})

	It("allows a no-op update of a labeled nodepool", func() {
		allowed, msg := evaluateSpecRule(rule, labelsGPUA100, labelsGPUA100)
		Expect(allowed).To(BeTrue(), "labeled no-op must be allowed, got: %s", msg)
	})

	It("rejects changing the labelValue", func() {
		allowed, _ := evaluateSpecRule(rule, labelsGPUA100, labelsGPUH100)
		Expect(allowed).To(BeFalse(), "changing the pair must be rejected")
	})

	It("rejects clearing a labeled nodepool's pair", func() {
		allowed, _ := evaluateSpecRule(rule, labelsGPUA100, labelsAbsent)
		Expect(allowed).To(BeFalse(), "clearing the pair must be rejected")
	})

	It("rejects adding a label pair to the default nodepool", func() {
		allowed, _ := evaluateSpecRule(rule, labelsAbsent, labelsGPUA100)
		Expect(allowed).To(BeFalse(), "adding a pair to the default nodepool must be rejected")
	})
})
