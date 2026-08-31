//go:build prometheus_rules
// +build prometheus_rules


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package definitions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"go.yaml.in/yaml/v2"
)

// Tester wraps promtool operations for testing Prometheus rule groups.
// It writes rule groups to a temp directory and provides methods to check syntax
// and run unit tests against them.
type Tester struct {
	t            *testing.T
	promtoolPath string
	dir          string
	rulesFile    string
}

func NewTester(t *testing.T, ruleGroups []monitoringv1.RuleGroup) *Tester {
	t.Helper()

	promtoolPath, err := exec.LookPath("promtool")
	if err != nil {
		t.Fatal("promtool not found in PATH")
	}

	dir := t.TempDir()
	rulesFile := writeRulesYAML(t, dir, ruleGroups)

	return &Tester{
		t:            t,
		promtoolPath: promtoolPath,
		dir:          dir,
		rulesFile:    rulesFile,
	}
}

// CheckSyntax runs "promtool check rules" to validate rule syntax.
func (pt *Tester) CheckSyntax() {
	pt.t.Helper()
	pt.run("check", "rules", pt.rulesFile)
}

// Test runs "promtool test rules" with the given test cases.
func (pt *Tester) Test(tests ...TestCase) {
	pt.t.Helper()

	tf := testFile{
		RuleFiles:          []string{filepath.Base(pt.rulesFile)},
		EvaluationInterval: "1m",
		Tests:              tests,
	}

	data, err := yaml.Marshal(tf)
	if err != nil {
		pt.t.Fatalf("failed to marshal test file: %v", err)
	}

	path := filepath.Join(pt.dir, fmt.Sprintf("test-%s.yml", uuid.NewString()))
	if err := os.WriteFile(path, data, 0644); err != nil {
		pt.t.Fatalf("failed to write %s: %v", path, err)
	}

	pt.run("test", "rules", path)
}

func (pt *Tester) run(args ...string) {
	pt.t.Helper()
	output, err := exec.Command(pt.promtoolPath, args...).CombinedOutput()
	if err != nil {
		pt.t.Fatalf("promtool %v failed:\n%s", args, output)
	}
}

// --- promtool test file structures ---

type testFile struct {
	RuleFiles          []string   `yaml:"rule_files"`
	EvaluationInterval string     `yaml:"evaluation_interval"`
	Tests              []TestCase `yaml:"tests"`
}

type TestCase struct {
	Interval string  `yaml:"interval"`
	Inputs   []Input `yaml:"input_series"`
	Tests    []Test  `yaml:"promql_expr_test"`
}

type Input struct {
	Series string `yaml:"series"`
	Values string `yaml:"values"`
}

type Test struct {
	Expr     string   `yaml:"expr"`
	EvalTime string   `yaml:"eval_time"`
	Samples  []Sample `yaml:"exp_samples"`
}

type Sample struct {
	Labels string  `yaml:"labels,omitempty"`
	Value  float64 `yaml:"value"`
}

// --- rule group conversion ---

type promRule struct {
	Record      string            `yaml:"record,omitempty"`
	Alert       string            `yaml:"alert,omitempty"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type promGroup struct {
	Name  string     `yaml:"name"`
	Rules []promRule `yaml:"rules"`
}

type promRuleFile struct {
	Groups []promGroup `yaml:"groups"`
}

func toPromRulesFile(rgs []monitoringv1.RuleGroup) promRuleFile {
	var prf promRuleFile
	for _, rg := range rgs {
		pg := promGroup{Name: rg.Name}
		for _, r := range rg.Rules {
			pg.Rules = append(pg.Rules, toPromRule(r))
		}
		prf.Groups = append(prf.Groups, pg)
	}
	return prf
}

func toPromRule(r monitoringv1.Rule) promRule {
	pr := promRule{
		Record:      r.Record,
		Alert:       r.Alert,
		Expr:        r.Expr.String(),
		Labels:      r.Labels,
		Annotations: r.Annotations,
	}
	if r.For != nil {
		pr.For = string(*r.For)
	}
	return pr
}

func writeRulesYAML(t *testing.T, dir string, rgs []monitoringv1.RuleGroup) string {
	t.Helper()
	data, err := yaml.Marshal(toPromRulesFile(rgs))
	if err != nil {
		t.Fatalf("failed to marshal rules to YAML: %v", err)
	}
	path := filepath.Join(dir, "rules.yml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
	return path
}
