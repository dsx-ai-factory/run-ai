package inference_workloads

import (
	"net/http"
	"strings"

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

const legacyNIMWorkloadLabel = "runai.nim/workload"

type inferenceEngine string

const (
	inferenceEngineUnknown inferenceEngine = ""
	inferenceEngineNIM     inferenceEngine = "nim"
	inferenceEngineVLLM    inferenceEngine = "vllm"
)

func detectInferenceEngine(labels []workloadsApi.Label, image string) (inferenceEngine, *workloadsApi.Error) {
	nimValue, nimSpecified := explicitEngineLabel(labels, constants.LabelNIMWorkload, legacyNIMWorkloadLabel)
	vllmValue, vllmSpecified := explicitEngineLabel(labels, constants.LabelVLLMWorkload)

	if nimValue && vllmValue {
		return inferenceEngineUnknown, workloadsApi.NewError(
			http.StatusBadRequest,
			"conflicting inference engine labels: NIM and vLLM cannot both be true",
		)
	}
	if nimValue {
		return inferenceEngineNIM, nil
	}
	if vllmValue {
		return inferenceEngineVLLM, nil
	}

	normalizedImage := strings.ToLower(strings.TrimSpace(image))
	switch {
	case !nimSpecified && isNIMImage(normalizedImage):
		return inferenceEngineNIM, nil
	case !vllmSpecified && isVLLMImage(normalizedImage):
		return inferenceEngineVLLM, nil
	default:
		return inferenceEngineUnknown, nil
	}
}

func explicitEngineLabel(labels []workloadsApi.Label, names ...string) (value bool, specified bool) {
	for _, label := range labels {
		if label.GetExclude() {
			continue
		}
		labelName := label.Name.Get()
		if labelName == nil {
			continue
		}
		for _, name := range names {
			if *labelName == name {
				specified = true
				value = value || labelValueIsTrue(label.Value.Get())
			}
		}
	}
	return value, specified
}

func labelValueIsTrue(value *string) bool {
	return value != nil && *value == "true"
}

func isNIMImage(image string) bool {
	return strings.HasPrefix(image, "nvcr.io/nim/") ||
		strings.HasPrefix(image, "nvcr.io/nvidia/nim:") ||
		strings.HasPrefix(image, "nvcr.io/nvidia/nim@")
}

func isVLLMImage(image string) bool {
	if image == "" {
		return false
	}
	repository := strings.SplitN(image, "@", 2)[0]
	lastSlash := strings.LastIndex(repository, "/")
	lastColon := strings.LastIndex(repository, ":")
	if lastColon > lastSlash {
		repository = repository[:lastColon]
	}
	for _, segment := range strings.Split(repository, "/") {
		if segment == "vllm" || segment == "vllm-openai" ||
			strings.HasPrefix(segment, "vllm-") || strings.HasSuffix(segment, "-vllm") {
			return true
		}
	}
	return false
}

func (engine inferenceEngine) identityLabel() string {
	switch engine {
	case inferenceEngineNIM:
		return constants.LabelNIMWorkload
	case inferenceEngineVLLM:
		return constants.LabelVLLMWorkload
	default:
		return ""
	}
}

func (engine inferenceEngine) metricsLabel() string {
	switch engine {
	case inferenceEngineNIM:
		return constants.LabelNIMWorkloadMetricsService
	case inferenceEngineVLLM:
		return constants.LabelVLLMWorkloadMetricsService
	default:
		return ""
	}
}
