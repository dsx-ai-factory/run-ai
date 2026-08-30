package projects

import (
	"fmt"
	"sync"
	"time"

	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func BuildCacheKey(clusterId, namespace string) string {
	return fmt.Sprintf("%s-%s", clusterId, namespace)
}

type NamespaceProjectMapping struct {
	Mapping map[string]string
}

var once sync.Once

func InitProjectMapperCache() {
	once.Do(func() {
		projectCacheConfig := runai_config.Get[*config.ProjectsCache]()
		expirationPeriod := time.Duration(projectCacheConfig.ExpirationSec) * time.Second
		err := cache.RegisterStore[NamespaceProjectMapping](
			options.Ttl(expirationPeriod),
			options.MaxRecordSizeKB(projectCacheConfig.MaxRecordSizeKB),
		)
		if err != nil {
			logging.Warnf("Failed to register cache store for type namespace project mapping: %v", err)
		}
	})
}
