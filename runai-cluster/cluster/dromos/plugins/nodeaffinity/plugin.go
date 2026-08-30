// Package nodeaffinity implements the dromos node-affinity plugin: it merges the
// node affinity declared in spec.instructions.nodeAffinity into every component's
// pod affinity on the rendered child manifest. The instruction is how an external
// controller constrains scheduling without knowing the workload's framework.
package nodeaffinity

import (
	"context"
	"fmt"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	kartautils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/karta"
	templateutils "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/template"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PluginName identifies the plugin in the registry and in status.plugins.
const PluginName = "node-affinity"

type Config struct {
	Enabled bool `mapstructure:"NODE_AFFINITY_ENABLED"`
}

type Plugin struct {
	// kartas resolves the Karta for a template's GVK at admission time. Nil skips
	// that check; MutateResource still checks at reconcile time.
	kartas *kartautils.Factory
	parser *templateutils.Parser
}

// NewPlugin builds the node-affinity plugin. Pass a nil reader to skip the
// admission-time Karta capability check; useful in tests.
func NewPlugin(kartaReader client.Reader) *Plugin {
	p := &Plugin{parser: templateutils.NewParser()}
	if kartaReader != nil {
		p.kartas = kartautils.NewFactory(kartaReader)
	}
	return p
}

// New is the framework.PluginBuilder. Loads Config from env and returns
// framework.ErrPluginDisabled when disabled.
func New(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetDefault("NODE_AFFINITY_ENABLED", true)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse node affinity config: %w", err)
	}
	if !cfg.Enabled {
		return nil, framework.ErrPluginDisabled
	}
	return NewPlugin(deps.Client), nil
}

func (p *Plugin) Name() string {
	return PluginName
}

// extractNodeAffinity returns the affinity to apply, or nil when the instruction
// is absent or carries no terms.
func extractNodeAffinity(runner *dromosv1alpha1.WorkloadRunner) *corev1.NodeAffinity {
	if runner.Spec.Instructions == nil || runner.Spec.Instructions.NodeAffinity == nil {
		return nil
	}

	affinity := runner.Spec.Instructions.NodeAffinity.NodeAffinity
	if affinity == nil {
		return nil
	}

	required := affinity.RequiredDuringSchedulingIgnoredDuringExecution
	hasNoTerms := (required == nil || len(required.NodeSelectorTerms) == 0) &&
		len(affinity.PreferredDuringSchedulingIgnoredDuringExecution) == 0
	if hasNoTerms {
		return nil
	}
	return affinity
}

var (
	_ framework.ResourceMutatePlugin = &Plugin{}
	_ framework.ValidatorPlugin      = &Plugin{}
	_ framework.StatusPlugin         = &Plugin{}
)
