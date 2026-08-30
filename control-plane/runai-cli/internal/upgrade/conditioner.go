package upgrade

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	"go.uber.org/zap"
)

const (
	cacheFilePerm = 0600
)

type Cache struct {
	Timestamp       time.Time `json:"timestamp"`
	ControlPlaneURL string    `json:"control_plane_url"`
}

// FlowCondition manages the upgrade condition checks and Cache updates.
// It checks if the last successful API check is still valid based on the TTL interval
// and the current control plane URL.
// If the last check is valid, it avoids running the upgrade flow again.
// If the last check is not valid, it allows the upgrade flow to proceed.
// It also provides methods to read and write the Cache to a file.
type FlowCondition struct {
	cachePath    string
	ttlInterval  time.Duration
	currentCPURL string
	logger       *zap.Logger
}

func NewFlowCondition(cachePath string, ttlInterval time.Duration, currentCPURL string, logger *zap.Logger) *FlowCondition {
	return &FlowCondition{
		cachePath:    cachePath,
		ttlInterval:  ttlInterval,
		currentCPURL: currentCPURL,
		logger:       logger,
	}
}

func (f *FlowCondition) readCache() (Cache, error) {
	var cache Cache
	cacheData, err := os.ReadFile(f.cachePath)
	if err != nil {
		return cache, err
	}
	err = json.Unmarshal(cacheData, &cache)
	if err != nil {
		return cache, err
	}
	return cache, nil
}

func (f *FlowCondition) writeCache(cache Cache) error {
	cacheBytes, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(f.cachePath, cacheBytes, cacheFilePerm)
}

// IsRunningUpdateFlowRequired determines if an update check should be performed.
// If outputFormat is "json" or "yaml" and autoUpdateEnabled is false, skip the check.
// Otherwise, use the cache/URL/TTL logic to decide.
func (f *FlowCondition) IsRunningUpdateFlowRequired(outputFormat string, autoUpdateEnabled bool) bool {
	if f.currentCPURL == "" {
		return false
	}
	if (outputFormat == cli.JsonOutput || outputFormat == cli.YamlOutput) && !autoUpdateEnabled {
		f.logger.Debug("Skipping update check: machine-readable output and auto-update is off")
		return false
	}
	if f.cachePath == "" {
		return true
	}
	now := time.Now()
	cache, err := f.readCache()
	if err != nil {
		f.logger.Debug("failed to read cache, will run update flow", zap.Error(err))
		return true
	}
	// Check if the time since last check is within the TTL interval
	if now.Sub(cache.Timestamp) < f.ttlInterval {
		// Check if the control plane URL has not changed since last check
		if cache.ControlPlaneURL == f.currentCPURL {
			// If the last notification is within the TTL interval and the CP URL has not changed,
			// do not need to check the API again.
			f.logger.Debug("the last version check is valid, No need to check the api again",
				zap.Time("timestamp", cache.Timestamp), zap.String("cp_url", cache.ControlPlaneURL))
			return false
		}
	}
	f.logger.Debug("the last version check is not valid, need to check the api again",
		zap.Time("timestamp", cache.Timestamp), zap.String("cp_url", cache.ControlPlaneURL))
	return true
}

func (f *FlowCondition) UpdateSuccessAPITimestamp() error {
	if f.cachePath == "" {
		return fmt.Errorf("no Cache path specified")
	}
	now := time.Now()
	cache := Cache{
		Timestamp:       now,
		ControlPlaneURL: f.currentCPURL,
	}
	if err := f.writeCache(cache); err != nil {
		return err
	}
	f.logger.Debug("updated last successful check timestamp", zap.Time("timestamp", now), zap.String("cp_url", f.currentCPURL))
	return nil
}
