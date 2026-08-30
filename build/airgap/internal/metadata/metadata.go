package metadata

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type Artifact struct {
	ServiceName  string `json:"service_name"`
	ArtifactType string `json:"artifact_type"`
	Enabled      *bool  `json:"enabled"`
	SkipAirgap   bool   `json:"skip_airgap"`
	SourcePath   string `json:"-"`
}

func Discover(roots []string, artifactType string) ([]Artifact, error) {
	artifacts := map[string]Artifact{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || entry.Name() != "artifact-metadata.json" {
				return nil
			}
			contents, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read %s: %w", path, err)
			}
			var artifact Artifact
			if err := json.Unmarshal(contents, &artifact); err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}
			if artifact.ArtifactType != artifactType || artifact.SkipAirgap || artifact.Enabled != nil && !*artifact.Enabled {
				return nil
			}
			if artifact.ServiceName == "" {
				return fmt.Errorf("%s has no service_name", path)
			}
			if previous, exists := artifacts[artifact.ServiceName]; exists {
				return fmt.Errorf("duplicate %s service_name %q: %s and %s", artifactType, artifact.ServiceName, previous.SourcePath, path)
			}
			artifact.SourcePath = path
			artifacts[artifact.ServiceName] = artifact
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("discover %s artifacts under %s: %w", artifactType, root, err)
		}
	}
	result := make([]Artifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		result = append(result, artifact)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ServiceName < result[j].ServiceName
	})
	return result, nil
}
