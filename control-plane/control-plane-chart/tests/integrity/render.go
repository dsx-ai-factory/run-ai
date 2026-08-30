package integrity

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/logger"
)

const releaseName = "runai-backend"

// renderChart wraps terratest's helm.RenderTemplate. Subchart deps are vendored
// under charts/, so no network or `helm dep up` is needed at test time.
// logger.Discard silences terratest's per-line render logging; failures still
// surface via t.Errorf in the test driver.
func renderChart(t *testing.T, chartDir string, valuesFiles []string) string {
	t.Helper()
	opts := &helm.Options{
		ValuesFiles: valuesFiles,
		Logger:      logger.Discard,
	}
	return helm.RenderTemplate(t, opts, chartDir, releaseName, nil)
}
