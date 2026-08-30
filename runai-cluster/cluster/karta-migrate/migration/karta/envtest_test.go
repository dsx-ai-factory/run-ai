package karta

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// TestEnrichKartaSpec_RealCRDSchemaAcceptsAllFixtures spins up envtest with
// the real karta CRD installed and asserts every captured RI fixture, after
// going through enrichKartaSpec + the migration's createKartaFromRI build
// path, is accepted by the apiserver. This catches CRD-schema bugs (like
// statusDefinition.statusMappings being required) that the in-process
// kartavalidation unit tests miss.
//
// Auto-skips when KUBEBUILDER_ASSETS is not set, so it stays opt-in until
// karta-migrate's Makefile wires up envtest. Run locally with:
//
//	KUBEBUILDER_ASSETS="$(setup-envtest use 1.33 -p path)" \
//	  go test ./migration/... -run TestEnrichKartaSpec_RealCRDSchema -v
func TestEnrichKartaSpec_RealCRDSchemaAcceptsAllFixtures(t *testing.T) {
	if os.Getenv("KUBEBUILDER_ASSETS") == "" {
		t.Skip("KUBEBUILDER_ASSETS not set; skipping envtest")
	}

	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "runai-crd-upgrader", "crds"),
		},
		ErrorIfCRDPathMissing: true,
	}
	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatalf("envtest start: %v", err)
	}
	t.Cleanup(func() { _ = testEnv.Stop() })

	scheme := runtime.NewScheme()
	if err := kartav1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}
	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	fixtures, err := filepath.Glob(filepath.Join("testdata", "ri-fixtures", "ri-*.yaml"))
	if err != nil {
		t.Fatalf("glob fixtures: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no RI fixtures found")
	}

	for _, path := range fixtures {
		name := strings.TrimSuffix(filepath.Base(path), ".yaml")
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path) // #nosec G304 -- path is from filepath.Glob over testdata
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			ri := &unstructured.Unstructured{}
			if err := yaml.Unmarshal(data, &ri.Object); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			specMap, found, err := unstructured.NestedMap(ri.Object, "spec")
			if err != nil || !found {
				t.Skipf("RI has no spec: err=%v found=%v", err, found)
			}

			// Round-trip the RI's unstructured spec into the typed KartaSpec
			// via JSON, then enrich the typed form. Matches production
			// createKartaFromRI flow.
			specJSON, err := json.Marshal(specMap)
			if err != nil {
				t.Fatalf("marshal spec: %v", err)
			}
			k := &kartav1alpha1.Karta{}
			// Use a unique name per fixture so subtests don't collide if the
			// envtest apiserver is reused.
			k.SetName(strings.ReplaceAll("envtest-"+name, "_", "-"))
			if err := json.Unmarshal(specJSON, &k.Spec); err != nil {
				t.Fatalf("unmarshal spec into typed Karta: %v", err)
			}
			enrichKartaSpec(&k.Spec)

			if err := c.Create(context.Background(), k); err != nil {
				t.Fatalf("apiserver rejected enriched Karta: %v", err)
			}
		})
	}
}
