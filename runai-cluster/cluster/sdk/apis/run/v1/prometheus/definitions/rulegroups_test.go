//go:build prometheus_rules
// +build prometheus_rules

package definitions

import (
	"strings"
	"testing"
)

func testRuleGroupsConfig() *RuleGroupsConfig {
	return &RuleGroupsConfig{
		AgentPullPushRateMinForAlert:             0.001,
		ClusterSyncQueueDepthIncreaseMaxForAlert: 100,
	}
}

func newTester(t *testing.T) *Tester {
	t.Helper()
	return NewTester(t, DefaultRuleGroups(testRuleGroupsConfig()))
}

func TestAllRulesSyntax(t *testing.T) {
	newTester(t).CheckSyntax()
}

func TestVLLMRulesHaveDedicatedGroup(t *testing.T) {
	const groupName = "runai-vllm-rules"

	var vllmGroupCount int
	for _, group := range DefaultRuleGroups(testRuleGroupsConfig()) {
		if group.Name == groupName {
			vllmGroupCount++
			if got, want := len(group.Rules), 53; got != want {
				t.Errorf("%s contains %d rules, want %d", groupName, got, want)
			}
		}

		for _, rule := range group.Rules {
			if strings.HasPrefix(rule.Record, "runai_vllm_") && group.Name != groupName {
				t.Errorf("vLLM rule %q is in group %q, want %q", rule.Record, group.Name, groupName)
			}
		}
	}

	if vllmGroupCount != 1 {
		t.Fatalf("found %d %q groups, want 1", vllmGroupCount, groupName)
	}
}
