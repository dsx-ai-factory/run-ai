package cluster_actions

import (
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

var RunaiConfigResource = schema.GroupVersionResource{
	Group:    runaiv1.Group,
	Version:  runaiv1.Version,
	Resource: strings.ToLower(runaiv1.RunaiConfigKind + "s"),
}
