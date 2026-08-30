package database

import (
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func ClearCache() {
	err := cache.ClearCache()
	if err != nil {
		logging.Warnf("failed clearing cache. %+v", err)
	}
}
