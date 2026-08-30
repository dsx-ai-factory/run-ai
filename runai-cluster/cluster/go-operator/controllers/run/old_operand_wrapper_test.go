package run

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/stretchr/testify/assert"
)

func TestOldResourceWrapperPreservesDisabledState(t *testing.T) {
	resource := &oldResourceWrapper{
		ResourceOld: operands.ResourceOld{Enabled: false},
	}

	assert.False(t, resource.Enabled())
}
