package readiness

import (
	"fmt"
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds what is specific to the readiness gate. Topology expectations (NP, N) and
// control-plane connectivity come from scale.ClusterConfig, which the caller passes in
// separately.
type Config struct {
	// Namespaces checked for workload/pod health, restart stability and metrics coverage.
	// Empty means resource_rec.DefaultNamespaces.
	Namespaces []string `envconfig:"LIMIT_NAMESPACES"`

	// QuietPeriod is how long to wait between the two restart-count samples.
	QuietPeriod string `envconfig:"READINESS_QUIET_PERIOD" default:"5m"`

	// ProjectPrefix and ProjectCount describe the projects the provisioner created.
	ProjectPrefix string `envconfig:"PROJECT_PREFIX" default:"load-project"`
	ProjectCount  int    `envconfig:"PROJECT_COUNT" default:"100"`

	// NodepoolPrefix selects the seeded nodepools, excluding the cluster's "default" one.
	NodepoolPrefix string `envconfig:"NODEPOOL_PREFIX" default:"nodepool-"`

	// Prometheus location for the scrape-target checks.
	PrometheusPodName   string `envconfig:"PROMETHEUS_POD_NAME" default:"prometheus-runai-0"`
	PrometheusNamespace string `envconfig:"PROMETHEUS_NAMESPACE" default:"runai"`
	PrometheusPort      int    `envconfig:"PROMETHEUS_PORT" default:"9090"`

	OutputDir string `envconfig:"OUTPUT_DIR" default:"output"`
}

func ConstructConfig() *Config {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatalf("failed to load readiness config: %v", err)
	}
	if err := c.validate(); err != nil {
		log.Fatalf("invalid readiness config: %v", err)
	}
	return &c
}

func (c *Config) validate() error {
	if _, err := time.ParseDuration(c.QuietPeriod); err != nil {
		return fmt.Errorf("invalid READINESS_QUIET_PERIOD %q: %w", c.QuietPeriod, err)
	}
	if c.ProjectCount < 0 {
		return fmt.Errorf("PROJECT_COUNT must be >= 0, got %d", c.ProjectCount)
	}
	if c.PrometheusPort <= 0 {
		return fmt.Errorf("PROMETHEUS_PORT must be > 0, got %d", c.PrometheusPort)
	}
	return nil
}

func (c *Config) quietPeriod() time.Duration {
	d, _ := time.ParseDuration(c.QuietPeriod) // validated in ConstructConfig
	return d
}
