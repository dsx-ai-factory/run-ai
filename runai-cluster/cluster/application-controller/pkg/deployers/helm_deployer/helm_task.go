/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helm_deployer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Task performs a helm install/uninstall/upgrade
type Task interface {
	// Action returns what kind of task is this
	Action() TaskActionType

	// Execute runs the task and return status.
	// reportProgress is a function that the task can call to update the status with the task progress.
	Execute(ctx context.Context, reportProgress func(*AppStatus)) *AppStatus
}

type InstallParams struct {
	chartRef                 *v1alpha1.ChartRef
	chartUrl                 string
	values                   map[string]interface{}
	credentialBindings       []v1alpha1.HelmCredentialBinding
	authOpts                 *AuthOptions
	requiresControlPlaneAuth bool
}

type AuthOptions struct {
	Username              string // Username for basic auth
	Password              string // Password for basic auth
	CAFile                string // CAFile path to CA certificate
	CertFile              string // CertFile path to client certificate
	KeyFile               string // KeyFile path to client key
	InsecureSkipTLSVerify bool   // InsecureSkipTLSVerify skips TLS verification
	RegistryConfigFile    string // RegistryConfigFile path to docker config for OCI registries
}

type TaskActionType string

const (
	TaskActionInstall   TaskActionType = "install"
	TaskActionUpgrade   TaskActionType = "upgrade"
	TaskActionUninstall TaskActionType = "uninstall"
)

func NewTask(ctx context.Context,
	taskActionType TaskActionType,
	releaseClient *ReleaseClient,

	k8sClient client.Client,
	application *v1alpha1.Application,
) (Task, error) {
	var err error
	var installParams *InstallParams

	if taskActionType != TaskActionUninstall {
		installParams, err = NewInstallParams(ctx, k8sClient, application)
		if err != nil {
			return nil, err
		}
	}

	switch taskActionType {
	case TaskActionInstall:
		return &TaskInstall{
			ReleaseClient: releaseClient,
			InstallParams: installParams,
		}, nil
	case TaskActionUpgrade:
		return &TaskUpgrade{
			ReleaseClient: releaseClient,
			InstallParams: installParams,
		}, nil
	case TaskActionUninstall:
		return &TaskUninstall{
			ReleaseClient: releaseClient,
		}, nil
	}

	return nil, fmt.Errorf("internal: unknown helm-task action %s", taskActionType)
}

func NewInstallParams(ctx context.Context, k8sClient client.Client, application *v1alpha1.Application) (*InstallParams, error) {
	if application == nil || application.Spec.Helm == nil {
		return nil, fmt.Errorf("invalid application: spec.helm is nil")
	}

	helmSpec := application.Spec.Helm
	for _, binding := range helmSpec.CredentialBindings {
		secret := &corev1.Secret{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: application.Namespace, Name: binding.SecretName}, secret); err != nil {
			return nil, fmt.Errorf("credential Secret for %q is unavailable in namespace %q; check credential synchronization and retry: %w", binding.Variable, application.Namespace, err)
		}
	}

	values, err := convertHelmValues(helmSpec.Values)
	if err != nil {
		return nil, fmt.Errorf("invalid spec.helm.values: %w", err)
	}

	result := InstallParams{
		chartRef: helmSpec.ChartRef,
		chartUrl: helmSpec.ChartUrl,
		// Convert Helm values from JSON map to map[string]interface{}
		values:                   values,
		credentialBindings:       helmSpec.CredentialBindings,
		requiresControlPlaneAuth: helmSpec.RequiresControlPlaneAuth,
	}

	// GetRelease auth options if specified
	if helmSpec.Auth != nil {
		var err error

		result.authOpts, err = getAuthOptions(ctx, k8sClient, application.Namespace, helmSpec.Auth)
		if err != nil {
			log.Error().Err(err).Msgf("failed to get auth options for application %s/%s", application.Namespace, application.Name)
			return nil, err
		}
	}

	return &result, nil
}

type TaskInstall struct {
	*ReleaseClient
	*InstallParams
}

type TaskUpgrade struct {
	*ReleaseClient
	*InstallParams
}

type TaskUninstall struct {
	*ReleaseClient
}

func (t *TaskInstall) Action() TaskActionType {
	return TaskActionInstall
}

func (t *TaskInstall) Execute(ctx context.Context, reportProgress func(*AppStatus)) *AppStatus {
	log.Info().Msgf("install task of %s started", t.ReleaseId.String())

	reportProgress(NewAppStatus().TaskDownloading(t))

	chartRequested, err := t.prepareToInstall(t.InstallParams)
	if err != nil {
		return NewAppStatus().TaskFailed(t, err)
	}
	values, err := applyCredentialBindings(chartRequested, t.values, t.credentialBindings)
	if err != nil {
		return NewAppStatus().TaskFailed(t, err)
	}

	reportProgress(NewAppStatus().TaskRunning(t))

	// Create install client
	installClient := action.NewInstall(t.ActionConfig)
	installClient.ReleaseName = t.ReleaseId.Name
	installClient.Namespace = t.ReleaseId.Namespace
	installClient.CreateNamespace = false // Namespace should already exist (Application CR is in it)
	installClient.SkipCRDs = false
	installClient.Wait = false // Don't wait for resources to be ready

	// Run the installation
	rel, err := installClient.RunWithContext(ctx, chartRequested, values)
	if err != nil {
		log.Error().Err(err).Msgf("HelmTask --> failed to install release %s", t.ReleaseId.String())
		return NewAppStatus().TaskFailed(t, err)
	}

	log.Info().Msgf("HelmTask --> successfully installed release %s", t.ReleaseId.String())
	return NewAppStatus().TaskCompleted(t).WithReleaseInfo(rel)
}

func (t *TaskUpgrade) Action() TaskActionType {
	return TaskActionUpgrade
}

func (t *TaskUpgrade) Execute(ctx context.Context, reportProgress func(*AppStatus)) *AppStatus {
	log.Info().Msgf("upgrade task of %s started", t.ReleaseId.String())

	reportProgress(NewAppStatus().TaskDownloading(t))

	chartRequested, err := t.prepareToInstall(t.InstallParams)
	if err != nil {
		return NewAppStatus().TaskFailed(t, err)
	}
	values, err := applyCredentialBindings(chartRequested, t.values, t.credentialBindings)
	if err != nil {
		return NewAppStatus().TaskFailed(t, err)
	}

	reportProgress(NewAppStatus().TaskRunning(t))

	// Create upgrade client
	upgradeClient := action.NewUpgrade(t.ActionConfig)
	upgradeClient.Namespace = t.ReleaseId.Namespace
	upgradeClient.Wait = false // Don't wait for resources to be ready

	// Run the upgrade
	rel, err := upgradeClient.RunWithContext(ctx, t.ReleaseId.Name, chartRequested, values)
	if err != nil {
		log.Error().Err(err).Msgf("HelmTask --> failed to upgrade release %s", t.ReleaseId.String())
		return NewAppStatus().TaskFailed(t, err)
	}

	log.Info().Msgf("HelmTask --> successfully upgraded release %s", t.ReleaseId.String())
	return NewAppStatus().TaskCompleted(t).WithReleaseInfo(rel)
}

func (t *TaskUninstall) Action() TaskActionType {
	return TaskActionUninstall
}

func (t *TaskUninstall) Execute(_ context.Context, reportProgress func(*AppStatus)) *AppStatus {
	log.Info().Msgf("uninstall task of %s started", t.ReleaseId.String())
	reportProgress(NewAppStatus().TaskRunning(t))

	uninstallClient := action.NewUninstall(t.ActionConfig)

	_, err := uninstallClient.Run(t.ReleaseId.Name)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			log.Info().Msgf("HelmTask --> release %s not found, treat it as uninstall succeeded", t.ReleaseId.String())
			return NewAppStatus().TaskCompleted(t)
		}
		log.Error().Err(err).Msgf("HelmTask --> failed to uninstall release %s", t.ReleaseId.String())
		return NewAppStatus().TaskFailed(t, err)
	}

	log.Info().Msgf("HelmTask --> successfully uninstalled release %s", t.ReleaseId.String())
	return NewAppStatus().TaskCompleted(t)
}

// ensureWorkingDirectories creates the necessary Helm cache and config directories
func ensureWorkingDirectories(settings *cli.EnvSettings) error {
	// Create repository cache directory
	if settings.RepositoryCache != "" {
		if err := os.MkdirAll(settings.RepositoryCache, ChartsDirPermissions); err != nil {
			return fmt.Errorf("failed to create repository cache dir %s: %w", settings.RepositoryCache, err)
		}
		log.Debug().Msgf("ensured Helm repository cache directory: %s", settings.RepositoryCache)
	}

	// Create repository config directory (parent of repositories.yaml)
	if settings.RepositoryConfig != "" {
		configDir := filepath.Dir(settings.RepositoryConfig)
		if err := os.MkdirAll(configDir, ChartsDirPermissions); err != nil {
			return fmt.Errorf("failed to create repository config dir %s: %w", configDir, err)
		}
		log.Debug().Msgf("ensured Helm repository config directory: %s", configDir)
	}

	return nil
}

func convertHelmValues(values map[string]apiextensionsv1.JSON) (map[string]interface{}, error) {
	if values == nil {
		return map[string]interface{}{}, nil
	}

	result := make(map[string]interface{}, len(values))
	for k, v := range values {
		// The Raw field contains the JSON bytes
		if v.Raw != nil {
			var parsed interface{}
			if err := json.Unmarshal(v.Raw, &parsed); err != nil {
				return nil, fmt.Errorf("invalid JSON in values[%q]: %w", k, err)
			}
			result[k] = parsed
		}
	}
	return result, nil
}
