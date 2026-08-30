package artifactlock

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/build/airgap/pkg/lock"
)

type Config struct {
	Name               string
	Version            string
	ImageLocks         []string
	ChartSources       []string
	ChartVersions      []string
	CLIRoots           []string
	CLISources         map[string]CLISource
	HelmUsername       string
	HelmPassword       string
	NGCResourceBaseURL string
	NGCToken           string
}

type Dependencies struct {
	HelmRunner HelmRunner
	Downloader Downloader
}

func Build(ctx context.Context, config Config, dependencies Dependencies) (lock.ArtifactLock, error) {
	if config.Name == "" {
		return lock.ArtifactLock{}, fmt.Errorf("artifact lock name is required")
	}
	if config.Version == "" || config.Version == "latest" {
		return lock.ArtifactLock{}, fmt.Errorf("invalid artifact lock version %q", config.Version)
	}
	images, err := CollectImageLocks(config.ImageLocks)
	if err != nil {
		return lock.ArtifactLock{}, err
	}
	if dependencies.HelmRunner == nil {
		return lock.ArtifactLock{}, fmt.Errorf("helm runner is required")
	}
	charts, err := LockCharts(
		ctx,
		config.ChartSources,
		config.ChartVersions,
		config.HelmUsername,
		config.HelmPassword,
		dependencies.HelmRunner,
	)
	if err != nil {
		return lock.ArtifactLock{}, err
	}
	if len(config.CLIRoots) > 0 && dependencies.Downloader == nil {
		return lock.ArtifactLock{}, fmt.Errorf("CLI downloader is required")
	}
	clis, err := ResolveCLIs(
		ctx,
		config.CLIRoots,
		config.Version,
		config.NGCResourceBaseURL,
		config.NGCToken,
		config.CLISources,
		dependencies.Downloader,
	)
	if err != nil {
		return lock.ArtifactLock{}, err
	}
	result := lock.ArtifactLock{
		APIVersion: lock.APIVersion,
		Kind:       lock.ArtifactKind,
		Metadata: lock.Metadata{
			Name:    config.Name,
			Version: config.Version,
		},
		Spec: lock.ArtifactLockSpec{
			Profile:    images.Profile,
			Platform:   images.Platform,
			ImageLocks: images.Locks,
			Charts:     charts,
			CLIs:       clis,
		},
	}
	if err := result.Validate(); err != nil {
		return lock.ArtifactLock{}, err
	}
	return result, nil
}
