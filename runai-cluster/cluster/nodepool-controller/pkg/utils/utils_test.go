package utils

import (
	"reflect"
	"testing"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteFromList(t *testing.T) {
	for testName, testData := range map[string]struct {
		listOfStrings     []string
		listOfNodePools   []v1alpha1.NodePool
		stringToRemove    string
		nodePoolToRemove  v1alpha1.NodePool
		expectedStrings   []string
		expectedNodePools []v1alpha1.NodePool
	}{
		"Sanity": {
			listOfStrings: []string{"0", "1", "2"},
			listOfNodePools: []v1alpha1.NodePool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nodepool-a",
						Labels: map[string]string{"label-a": "value-a"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nodepool-b",
						Labels: map[string]string{"label": "value"},
					},
				},
			},
			stringToRemove: "2",
			nodePoolToRemove: v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{
				Name:   "nodepool-a",
				Labels: map[string]string{"label-a": "value-a"},
			}},
			expectedStrings: []string{"0", "1"},
			expectedNodePools: []v1alpha1.NodePool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nodepool-b",
						Labels: map[string]string{"label": "value"},
					},
				},
			},
		},
		"Item does not exist": {
			listOfStrings: []string{"0", "1", "2"},
			listOfNodePools: []v1alpha1.NodePool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nodepool-a",
						Labels: map[string]string{"label-a": "value-a"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nodepool-b",
						Labels: map[string]string{"label": "value"},
					},
				},
			},
			stringToRemove: "3",
			nodePoolToRemove: v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{
				Name:   "nodepool-a",
				Labels: map[string]string{"label-wrongggg": "value-a"},
			}},
			expectedStrings: []string{"0", "1", "2"},
			expectedNodePools: []v1alpha1.NodePool{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nodepool-a",
						Labels: map[string]string{"label-a": "value-a"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "nodepool-b",
						Labels: map[string]string{"label": "value"},
					},
				},
			},
		},
		"Sanity 2": {
			listOfStrings:   []string{"0", "1", "2"},
			listOfNodePools: []v1alpha1.NodePool{},
			stringToRemove:  "1",
			nodePoolToRemove: v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{
				Name:   "nodepool-a",
				Labels: map[string]string{"label-a": "value-a"},
			}},
			expectedStrings:   []string{"0", "2"},
			expectedNodePools: []v1alpha1.NodePool{},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			outputStrings := DeleteFromList(testData.listOfStrings, testData.stringToRemove)
			outputNodePools := DeleteFromList(testData.listOfNodePools, testData.nodePoolToRemove)

			if !reflect.DeepEqual(outputStrings, testData.expectedStrings) {
				t.Errorf("Strings: Expected: %v, got: %v", testData.expectedStrings, outputStrings)
			}
			if !reflect.DeepEqual(outputNodePools, testData.expectedNodePools) {
				t.Errorf("NodePools: Expected: %v, got: %v",
					getNodePoolNames(testData.expectedNodePools), getNodePoolNames(outputNodePools))
			}
		})
	}
}

func TestGetIndex(t *testing.T) {
	for testName, testData := range map[string]struct {
		listOfStrings       []string
		stringToFind        string
		expectedStringIndex int
	}{
		"Sanity 0": {
			listOfStrings:       []string{"0", "1", "2", "3"},
			stringToFind:        "0",
			expectedStringIndex: 0,
		},
		"Sanity 1": {
			listOfStrings:       []string{"0", "1", "2", "3"},
			stringToFind:        "1",
			expectedStringIndex: 1,
		},
		"Sanity 2": {
			listOfStrings:       []string{"0", "1", "2", "3"},
			stringToFind:        "2",
			expectedStringIndex: 2,
		},
		"Sanity 3": {
			listOfStrings:       []string{"0", "1", "2", "3"},
			stringToFind:        "3",
			expectedStringIndex: 3,
		},
		"Sanity 4": {
			listOfStrings:       []string{},
			stringToFind:        "2",
			expectedStringIndex: -1,
		},
		"Sanity 5": {
			listOfStrings:       []string{"0", "1"},
			stringToFind:        "2",
			expectedStringIndex: -1,
		},
	} {
		t.Run(testName, func(t *testing.T) {
			outputIndex := indexOfStringInList(testData.listOfStrings, testData.stringToFind)

			if outputIndex != testData.expectedStringIndex {
				t.Errorf("Wrong index: Expected: %v, got: %v", testData.expectedStringIndex, outputIndex)
			}
		})
	}
}

func TestIsItemInList(t *testing.T) {
	for testName, testData := range map[string]struct {
		listOfStrings []string
		stringToCheck string
		expectedFound bool
	}{
		"Sanity 0": {
			listOfStrings: []string{"0", "1", "2", "3"},
			stringToCheck: "0",
			expectedFound: true,
		},
		"Sanity 1": {
			listOfStrings: []string{"0", "1", "2", "3"},
			stringToCheck: "1",
			expectedFound: true,
		},
		"Sanity 3": {
			listOfStrings: []string{"0", "1", "2", "3"},
			stringToCheck: "3",
			expectedFound: true,
		},
		"Sanity 4": {
			listOfStrings: []string{},
			stringToCheck: "2",
			expectedFound: false,
		},
		"Sanity 5": {
			listOfStrings: []string{"0", "1"},
			stringToCheck: "2",
			expectedFound: false,
		},
	} {
		t.Run(testName, func(t *testing.T) {
			outputFound := IsItemInList(testData.listOfStrings, testData.stringToCheck)

			if outputFound != testData.expectedFound {
				t.Errorf("Wrong is item in list result: Expected: %v, got: %v", testData.expectedFound, outputFound)
			}
		})
	}
}

func getNodePoolNames(nodePools []v1alpha1.NodePool) []string {
	nodePoolNames := []string{}
	for _, nodePool := range nodePools {
		nodePoolNames = append(nodePoolNames, nodePool.Name)
	}
	return nodePoolNames
}
