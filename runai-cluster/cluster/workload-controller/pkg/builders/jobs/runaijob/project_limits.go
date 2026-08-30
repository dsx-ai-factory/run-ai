package runaijob

import (
	"strconv"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

// projectLimits holds the per-project job time/idle limits the runai job builder applies.
// They are read from the kai project's annotations (run.ai/*-seconds).
type projectLimits struct {
	interactiveTimeLimitSecs          int64
	interactiveMaxIdleDurationSecs    int64
	interactivePreemptibleMaxIdleSecs int64
	trainingTimeLimitSecs             int64
	trainingMaxIdleDurationSecs       int64
}

// limitsFromProject reads the limits off a kai.resources project (annotations).
func limitsFromProject(project *kaiorg.Project) projectLimits {
	return projectLimits{
		interactiveTimeLimitSecs:          annotationSeconds(project, constants.InteractiveJobTimeLimitSecondsAnnotation),
		interactiveMaxIdleDurationSecs:    annotationSeconds(project, constants.InteractiveJobMaxIdleDurationSecondsAnnotation),
		interactivePreemptibleMaxIdleSecs: annotationSeconds(project, constants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation),
		trainingTimeLimitSecs:             annotationSeconds(project, constants.TrainingJobTimeLimitSecondsAnnotation),
		trainingMaxIdleDurationSecs:       annotationSeconds(project, constants.TrainingJobMaxIdleDurationSecondsAnnotation),
	}
}

// annotationSeconds reads an integer-seconds annotation off the kai project, returning 0
// when it is absent or unparseable (matching the writer, which omits zero-valued limits).
func annotationSeconds(project *kaiorg.Project, key string) int64 {
	value, err := strconv.ParseInt(project.Annotations[key], 10, 64)
	if err != nil {
		return 0
	}
	return value
}
