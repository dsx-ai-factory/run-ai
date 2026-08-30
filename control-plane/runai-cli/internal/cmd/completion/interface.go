package completion

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
)

type Factory interface {
	GetCompletionFn(framework api.Framework, workloadType api.WorkloadType) (completion.CompletionFn, error)
}
