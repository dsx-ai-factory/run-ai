package composer

import (
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
)

// MNNVLModeNone is the MNNVL plugin mode meaning MNNVL is not requested.
const MNNVLModeNone = mnnvl.None

// ComposeMNNVL composes the WorkloadRunner MNNVL plugin mode for a workload.
//
// A concrete run.ai/MNNVL label on the workload, if given, should dictate the final value (required/none).
// if the label is absent, a default value is sent to the setter, if there is a default value.
func (fc *FieldsComposer) ComposeMNNVL(
	labels []workloadsApi.Label,
	defaultMode string,
	target any,
	setter FnStringSetter,
	cookie any,
) func() *workloadsApi.Error {
	return func() *workloadsApi.Error {
		mode := ""

		for _, label := range labels {
			if convutils.BoolIsTrue(label.Exclude.Get()) || convutils.Deref(label.Name.Get(), "") != mnnvl.FeatureLabelKey {
				continue
			}
			value := convutils.Deref(label.Value.Get(), "")
			if value == "" {
				continue
			}
			switch value {
			// The WorkloadRunner plugin treats Preferred the same as Required, so normalize
			// it to Required at composition time and only emit the two modes the plugin API
			// exposes (Required and None).
			case mnnvl.Required, mnnvl.Preferred:
				mode = string(workloadsManagerApi.MNNVLMODE_REQUIRED)
			case mnnvl.None:
				mode = string(workloadsManagerApi.MNNVLMODE_NONE)
			default:
				return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("Invalid MNNVL mode '%s'", value))
			}
		}

		if mode == "" {
			mode = defaultMode
		}

		if mode != "" {
			return setter(&mode, target, cookie)
		}

		return nil
	}
}

// the setter
func MNNVLToWorkloadRunner(src *string, target any, _ any) *workloadsApi.Error {
	config, ok := target.(*workloadsManagerApi.WorkloadV2Configuration)
	if !ok {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("MNNVLToWorkloadRunner: target type %T is not *openapi.WorkloadV2Configuration", target))
	}
	if src == nil {
		return nil
	}
	config.SetMnnvl(workloadsManagerApi.MnnvlMode(*src))
	return nil
}
