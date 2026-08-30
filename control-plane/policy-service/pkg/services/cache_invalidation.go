package services

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// gvkComponentId returns the canonical component id for a mapping dependency:
// "group/kind/version".
func gvkComponentId(gvk api.Gvk) string {
	return fmt.Sprintf("%s/%s/%s", convutils.Deref(gvk.Group, ""), gvk.Kind, gvk.Version)
}
