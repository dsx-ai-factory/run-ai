package projects

import (
	"sync"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

var once sync.Once

func initProjectMapperCache() {
	once.Do(func() {
		err := cache.RegisterStore[BiDirectionalMapping](
			options.Ttl(runai_config.Get[*config.ProjectsCache]().Expiration()),
			options.MaxRecordSizeKB(runai_config.Get[*config.ProjectsCache]().MaxRecordSizeKB),
		)
		if err != nil {
			logging.Warnf("Failed to register cache store for type bi directional mapping: %v", err)
		}
	})
}
