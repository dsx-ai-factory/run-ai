package parsers

import (
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/test"
)

func TestParseBoolLabel(t *testing.T) {
	ans, err := parseBoolLabel(nil, constants.LabelJupyterJob)
	if err == nil || ans != false {
		t.Errorf("parse float allocation was incorrect. got: %v expected: %v", ans, 0.0)
	}
	expectedTest := true
	ansTest, err2 := parseBoolLabel(map[string]string{constants.LabelJupyterJob: "true"}, constants.LabelJupyterJob)
	if err2 != nil || expectedTest != ansTest {
		t.Errorf("parse string allocation was incorrect. got: %v xpected: %v", ansTest, expectedTest)
	}

	notABoolTest := "abc"

	ansNotAFloatTest, err3 := parseBoolLabel(map[string]string{constants.LabelJupyterJob: notABoolTest}, constants.LabelJupyterJob)
	if err3 == nil || ansNotAFloatTest != false {
		t.Errorf("parse string allocation was incorrect. got: %v expected: %v", ansNotAFloatTest, 0.0)
	}
}

func TestGetTotalRequestedMigDevices(t *testing.T) {
	calculatedRequestedMigDevices := GetTotalRequestedMigDevices(test.BuildMigPods())
	const expected = "1g.5gb,1g.5gb,1g.5gb,2g.10gb"
	if calculatedRequestedMigDevices != expected {
		t.Errorf("Calculated total requested Mig devices is incorrect, got: %s, wanted %s.", calculatedRequestedMigDevices, expected)
	}
}

func TestGetTotalRequestedCPU(t *testing.T) {
	calculatedRequestedMigDevices := GetTotalRequestedCPU(test.BuildMigPods())
	resources := test.BuildResourceList("5000m", "5G")
	expected := resources.Cpu().AsApproximateFloat64()
	if calculatedRequestedMigDevices != expected {
		t.Errorf("Calculated total requested Mig devices is incorrect, got: %f, wanted %f.", calculatedRequestedMigDevices, expected)
	}
}

func TestGetTotalRequestedCpuMemory(t *testing.T) {
	calculatedRequestedMigDevices := GetTotalRequestedCpuMemory(test.BuildMigPods())
	resources := test.BuildResourceList("5000m", "5G")
	expected := resources.Memory().AsApproximateFloat64()
	if calculatedRequestedMigDevices != expected {
		t.Errorf("Calculated total requested Mig devices is incorrect, got: %f, wanted %f.", calculatedRequestedMigDevices, expected)
	}
}

func TestGetTotalLimitCpuMemory(t *testing.T) {
	calculatedRequestedMigDevices := GetTotalLimitCpuMemory(test.BuildMigPods())
	expected := 0.0
	if calculatedRequestedMigDevices != expected {
		t.Errorf("Calculated total requested Mig devices is incorrect, got: %f, wanted %f.", calculatedRequestedMigDevices, expected)
	}
}

func TestGetParentWorkloadName(t *testing.T) {
	job := &runaiJobV1.RunaiJob{
		ObjectMeta: metav1.ObjectMeta{
			UID:               "UID",
			Labels:            map[string]string{"workloadName": "parentWorkloadName"},
			Annotations:       map[string]string{},
			CreationTimestamp: metav1.Now(),
			Name:              "Name",
			Namespace:         "Namespace",
		},
	}
	parentWorkloadName := GetParentWorkloadName(job.Labels)
	expectedParentWorkloadName := "parentWorkloadName"
	if parentWorkloadName != expectedParentWorkloadName {
		t.Errorf("expectedParentWorkloadName: %s, parentWorkloadName: %s", expectedParentWorkloadName, parentWorkloadName)
	}
}

func TestParseStringAnnotation(t *testing.T) {
	ans := parseStringAnnotation(nil, nil, "annotation", "myObject")
	if ans != "" {
		t.Errorf("parse string allocation was incorrect. got: %s expected: %s", ans, "")
	}
	expectedBackupTest := "backupValue"
	ansBackupTest := parseStringAnnotation(map[string]string{"abc": "value"}, map[string]string{"backup": expectedBackupTest}, "backup", "myObject")
	if ansBackupTest != expectedBackupTest {
		t.Errorf("parse string allocation was incorrect. got: %s expected: %s", ans, expectedBackupTest)
	}

	expectedTest := "Value"
	ansTest := parseStringAnnotation(map[string]string{"abc": expectedTest}, map[string]string{"backup": expectedBackupTest}, "abc", "myObject")
	if ansTest != expectedTest {
		t.Errorf("parse string allocation was incorrect. got: %s expected: %s", ans, expectedBackupTest)
	}
}

func TestParseFloatAnnotation(t *testing.T) {
	ans := parseFloatAnnotation(nil, nil, "annotation", "myObject")
	if ans != 0 {
		t.Errorf("parse float allocation was incorrect. got: %f expected: %f", ans, 0.0)
	}
	expectedBackupTest := 23.3
	ansBackupTest := parseFloatAnnotation(map[string]string{"abc": "value"}, map[string]string{"backup": "23.3"}, "backup", "myObject")
	if ansBackupTest != expectedBackupTest {
		t.Errorf("parse string allocation was incorrect. got: %v expected: %f", ansBackupTest, expectedBackupTest)
	}

	expectedTest := 11.1
	ansTest := parseFloatAnnotation(map[string]string{"abc": "11.1"}, map[string]string{"backup": "23.3"}, "abc", "myObject")
	if ansTest != expectedTest {
		t.Errorf("parse string allocation was incorrect. got: %f expected: %f", ansTest, expectedTest)
	}

	notAFloatTest := "abc"

	ansNotAFloatTest := parseFloatAnnotation(map[string]string{"abc": notAFloatTest}, map[string]string{"backup": "22233.4"}, "abc", "myObject")
	if ansNotAFloatTest != 0.0 {
		t.Errorf("parse string allocation was incorrect. got: %f expected: %f", ansNotAFloatTest, 0.0)
	}
}
