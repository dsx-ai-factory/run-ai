/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrader

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

// FieldManager is persisted in managedFields and must not be renamed.
const FieldManager = "runai-crd-upgrader"

var applyBackoff = wait.Backoff{
	Duration: 250 * time.Millisecond,
	Factor:   2,
	Jitter:   0.1,
	Steps:    5,
}

type CRDClient interface {
	Patch(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) (*apiextensionsv1.CustomResourceDefinition, error)
	Get(context.Context, string, metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error)
}

type Upgrader struct {
	client             CRDClient
	crdsDir            string
	establishedTimeout time.Duration
	pollInterval       time.Duration
	logger             *log.Logger
}

type manifest struct {
	name   string
	source string
	data   []byte
}

type loadedManifests struct {
	manifests []manifest
	ignored   []string
}

type rawDocument []byte

func (d *rawDocument) UnmarshalJSON(data []byte) error {
	*d = append((*d)[:0], data...)
	return nil
}

func New(client CRDClient, crdsDir string, establishedTimeout, pollInterval time.Duration, logger *log.Logger) *Upgrader {
	return &Upgrader{
		client:             client,
		crdsDir:            crdsDir,
		establishedTimeout: establishedTimeout,
		pollInterval:       pollInterval,
		logger:             logger,
	}
}

func (u *Upgrader) Run(ctx context.Context) error {
	if u.client == nil {
		return errors.New("CRD client is required")
	}
	if u.establishedTimeout <= 0 {
		return errors.New("established timeout must be positive")
	}
	if u.pollInterval <= 0 {
		return errors.New("poll interval must be positive")
	}
	if u.logger == nil {
		u.logger = log.New(io.Discard, "", 0)
	}

	loaded, err := loadManifests(u.crdsDir)
	if err != nil {
		return err
	}
	for _, ignored := range loaded.ignored {
		u.logger.Printf("ignoring non-CRD manifest %s", ignored)
	}
	manifests := loaded.manifests

	u.logger.Printf("applying %d CRDs with field manager %q", len(manifests), FieldManager)
	for _, crd := range manifests {
		if err := u.apply(ctx, crd); err != nil {
			return err
		}
		u.logger.Printf("applied CRD %s", crd.name)
	}

	if err := u.waitUntilEstablished(ctx, manifests); err != nil {
		return err
	}
	u.logger.Printf("all %d CRDs are Established", len(manifests))
	return nil
}

func (u *Upgrader) apply(ctx context.Context, crd manifest) error {
	force := true
	options := metav1.PatchOptions{
		FieldManager: FieldManager,
		Force:        &force,
	}

	var lastErr error
	err := wait.ExponentialBackoffWithContext(ctx, applyBackoff, func(ctx context.Context) (bool, error) {
		_, lastErr = u.client.Patch(ctx, crd.name, types.ApplyPatchType, crd.data, options)
		if lastErr == nil {
			return true, nil
		}
		if !isRetryable(lastErr) {
			return false, lastErr
		}
		u.logger.Printf("retrying CRD %s after apply error: %v", crd.name, lastErr)
		return false, nil
	})
	if err != nil {
		if wait.Interrupted(err) && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) && lastErr != nil {
			err = lastErr
		}
		return fmt.Errorf("server-side apply CRD %s from %s: %w", crd.name, crd.source, err)
	}
	return nil
}

func (u *Upgrader) waitUntilEstablished(ctx context.Context, manifests []manifest) error {
	pending := make(map[string]struct{}, len(manifests))
	lastGetErrors := make(map[string]error, len(manifests))
	for _, crd := range manifests {
		pending[crd.name] = struct{}{}
	}

	err := wait.PollUntilContextTimeout(ctx, u.pollInterval, u.establishedTimeout, true, func(ctx context.Context) (bool, error) {
		for _, crd := range manifests {
			if _, waiting := pending[crd.name]; !waiting {
				continue
			}

			current, err := u.client.Get(ctx, crd.name, metav1.GetOptions{})
			if err != nil {
				if isRetryable(err) {
					lastGetErrors[crd.name] = err
					u.logger.Printf("retrying Established check for CRD %s after error: %v", crd.name, err)
					continue
				}
				return false, fmt.Errorf("get CRD %s while waiting for Established: %w", crd.name, err)
			}
			delete(lastGetErrors, crd.name)

			established, conditionErr := establishedState(current)
			if conditionErr != nil {
				return false, conditionErr
			}
			if established {
				delete(pending, crd.name)
				u.logger.Printf("CRD %s is Established", crd.name)
			}
		}
		return len(pending) == 0, nil
	})
	if err != nil {
		if len(lastGetErrors) > 0 && wait.Interrupted(err) && !errors.Is(err, context.Canceled) {
			causes := []error{err}
			for _, name := range sortedErrorKeys(lastGetErrors) {
				causes = append(causes, fmt.Errorf("last Get error for CRD %s: %w", name, lastGetErrors[name]))
			}
			err = errors.Join(causes...)
		}
		return fmt.Errorf("wait for CRDs [%s] to become Established: %w", strings.Join(sortedKeys(pending), ", "), err)
	}
	return nil
}

func establishedState(crd *apiextensionsv1.CustomResourceDefinition) (bool, error) {
	for _, condition := range crd.Status.Conditions {
		if condition.Type == apiextensionsv1.Established && condition.Status == apiextensionsv1.ConditionTrue {
			return true, nil
		}
	}
	for _, condition := range crd.Status.Conditions {
		if condition.Type == apiextensionsv1.NamesAccepted && condition.Status == apiextensionsv1.ConditionFalse {
			return false, fmt.Errorf("CRD %s names were rejected: %s: %s", crd.Name, condition.Reason, condition.Message)
		}
	}
	return false, nil
}

func loadManifests(directory string) (loadedManifests, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return loadedManifests{}, fmt.Errorf("read CRD directory %s: %w", directory, err)
	}

	var loaded loadedManifests
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !isManifestFile(entry.Name()) {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		decoded, err := decodeFile(path)
		if err != nil {
			return loadedManifests{}, err
		}
		loaded.manifests = append(loaded.manifests, decoded.manifests...)
		loaded.ignored = append(loaded.ignored, decoded.ignored...)
	}
	if len(loaded.manifests) == 0 {
		return loadedManifests{}, fmt.Errorf("no CRD manifests found in %s", directory)
	}

	sort.Slice(loaded.manifests, func(i, j int) bool {
		if loaded.manifests[i].source == loaded.manifests[j].source {
			return loaded.manifests[i].name < loaded.manifests[j].name
		}
		return loaded.manifests[i].source < loaded.manifests[j].source
	})
	seen := make(map[string]string, len(loaded.manifests))
	for _, crd := range loaded.manifests {
		if previous, exists := seen[crd.name]; exists {
			return loadedManifests{}, fmt.Errorf("CRD %s is defined in both %s and %s", crd.name, previous, crd.source)
		}
		seen[crd.name] = crd.source
	}
	return loaded, nil
}

func decodeFile(path string) (loadedManifests, error) {
	// #nosec G304 -- the configured directory intentionally supplies the manifests to apply.
	contents, err := os.ReadFile(path)
	if err != nil {
		return loadedManifests{}, fmt.Errorf("read CRD manifest %s: %w", path, err)
	}

	decoder := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(contents), 4096)
	var loaded loadedManifests
	for document := 1; ; document++ {
		var raw rawDocument
		if err := decoder.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return loadedManifests{}, fmt.Errorf("decode %s document %d: %w", path, document, err)
		}
		if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			continue
		}

		var identity struct {
			APIVersion string `json:"apiVersion"`
			Kind       string `json:"kind"`
			Metadata   struct {
				Name string `json:"name"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(raw, &identity); err != nil {
			return loadedManifests{}, fmt.Errorf("decode identity from %s document %d: %w", path, document, err)
		}
		if identity.Kind != "CustomResourceDefinition" {
			loaded.ignored = append(loaded.ignored, fmt.Sprintf("%s document %d (%s %s)", path, document, identity.APIVersion, identity.Kind))
			continue
		}
		if identity.APIVersion != "apiextensions.k8s.io/v1" {
			return loadedManifests{}, fmt.Errorf("%s document %d contains unsupported CRD apiVersion %q; only apiextensions.k8s.io/v1 is supported", path, document, identity.APIVersion)
		}
		if identity.Metadata.Name == "" {
			return loadedManifests{}, fmt.Errorf("%s document %d has no metadata.name", path, document)
		}

		loaded.manifests = append(loaded.manifests, manifest{
			name:   identity.Metadata.Name,
			source: path,
			data:   append([]byte(nil), raw...),
		})
	}
	return loaded, nil
}

func isManifestFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func isRetryable(err error) bool {
	var networkError net.Error
	return errors.As(err, &networkError) ||
		apierrors.IsTimeout(err) ||
		apierrors.IsServerTimeout(err) ||
		apierrors.IsTooManyRequests(err) ||
		apierrors.IsServiceUnavailable(err) ||
		apierrors.IsInternalError(err) ||
		apierrors.IsUnexpectedServerError(err)
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedErrorKeys(values map[string]error) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
