package testutils

import (
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

// MapComparator is a gomega wrapper for comparing two maps and display the diff in readable way.
// to use just change
//
//	Expect(mapA).To(Equal(mapB)
//
// with
//
//	testing.Expect(mapA).ToEqual(mapB)
type MapComparator struct {
	actualAsYaml []byte
}

func Expect(expected interface{}) *MapComparator {
	actualAsYaml, err := yaml.Marshal(expected)
	gomega.Expect(err).To(gomega.BeNil())

	return &MapComparator{
		actualAsYaml: actualAsYaml,
	}
}

func (mc *MapComparator) ToEqual(expected interface{}) {
	expectedAsYaml, _ := yaml.Marshal(expected)

	result := cmp.Diff(expectedAsYaml, mc.actualAsYaml, cmp.Options{})
	isEqual := result == ""
	gomega.Expect(isEqual).To(gomega.Equal(true), result)
}
