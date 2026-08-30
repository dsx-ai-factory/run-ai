package integrity

import (
	"path/filepath"
	"testing"
)

const chartDir = "../.."

type scenario struct {
	name   string
	values []string
}

var scenarios = []scenario{
	{name: "defaults", values: []string{"../scenarios/defaults.yaml"}},
	{name: "nats-disabled", values: []string{"../scenarios/nats-disabled.yaml"}},
	{name: "existing-secrets", values: []string{"../scenarios/existing-secrets.yaml"}},
	{name: "external-postgres", values: []string{"../scenarios/external-postgres.yaml"}},
	{name: "custom-secret-keys", values: []string{"../scenarios/custom-secret-keys.yaml"}},
	{name: "nats-extra-env-vars", values: []string{"../scenarios/nats-extra-env-vars.yaml"}},
}

// TestChartIntegrity asserts that every secretRef in a control-plane service
// resolves to a (secret, key) pair the chart declares -- either in a rendered
// Secret manifest or in the secret-generator-config ConfigMap's secrets.json.
// Catches the class of bug where a service mounts a key the chart no longer
// promises to produce.
func TestChartIntegrity(t *testing.T) {
	chartAbs, err := filepath.Abs(chartDir)
	if err != nil {
		t.Fatalf("resolve chart dir: %v", err)
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			values := make([]string, len(sc.values))
			for i, v := range sc.values {
				abs, err := filepath.Abs(v)
				if err != nil {
					t.Fatalf("resolve values %q: %v", v, err)
				}
				values[i] = abs
			}
			rendered := renderChart(t, chartAbs, values)
			manifests := parseManifests(t, rendered)
			producers := collectProducers(t, manifests)
			for name := range collectExternalsFromValues(t, values) {
				producers.External[name] = struct{}{}
			}
			refs := collectReferences(manifests)

			violations := validateReferences(refs, producers)
			if len(violations) == 0 {
				return
			}
			t.Errorf("scenario %q: %d violation(s)", sc.name, len(violations))
			for _, v := range violations {
				t.Errorf("  [%s] %s", sc.name, v)
			}
		})
	}
}
