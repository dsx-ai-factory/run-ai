/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	schemaupgrades "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/schema-upgrades"
	controlplanedetailsdownloader "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/control-plane-details-downloader"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	imagepullsecret "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/image-pull-secret"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	external_workload_integrator "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/external-workload-integrator"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runaiv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	RunaiSchedulerConfig = "runai-scheduler-config"

	oAuthClientIDKey = "OAUTH2_PROXY_CLIENT_ID"
)

func installCPSecrets(k8s client.Client, installationConfiguration *controlplanedetailsdownloader.InstallationConfigurationData) error {
	logger := log.FromContext(context.Background())
	logger.Info("found control-plane details secret")

	if dockerConfigJSON, ok := installationConfiguration.Global.ControlPlane.Config.GetDockerConfigJsonOk(); ok {
		// need to act on mgmt and tenant cluster (maybe done in engine)
		err := imagepullsecret.CreateImagePullSecret(k8s, runaiv1.ReservationNamespace, hooks.RunAIControlPlaneImagePullSecretName, *dockerConfigJSON)
		if err != nil {
			return fmt.Errorf("failed creating image pull secrets in the %s namespace: %v", runaiv1.ReservationNamespace, err)
		}

		// need to act on mgmt and tenant cluster (maybe done in engine)
		err = imagepullsecret.CreateImagePullSecret(k8s, runaiv1.ScaleAdjustNamespace, hooks.RunAIControlPlaneImagePullSecretName, *dockerConfigJSON)
		if err != nil {
			return fmt.Errorf("failed creating image pull secrets in the %s namespace: %v", runaiv1.ScaleAdjustNamespace, err)
		}
	}

	if err := seedWorkloadsAuthClientSecret(k8s, installationConfiguration); err != nil {
		return err
	}

	return nil
}

// seedWorkloadsAuthClientSecret writes the workloads-auth OAuth client credentials taken from the
// (still un-sanitized) control-plane config into the workloads-auth secret. RunaiConfig.Sanitize()
// strips Workloads.Client.Secret before it is persisted to the CRD, so on a fresh install the
// operator has no source for the client secret and no pre-existing workloads-auth secret to fall
// back to. Seeding it here lets the AnyWorkload operand preserve the value via its existing
// fallback. Existing keys (cookie secret, redis password) are preserved.
func seedWorkloadsAuthClientSecret(k8s client.Client, installationConfiguration *controlplanedetailsdownloader.InstallationConfigurationData) error {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	workloadsClient := installationConfiguration.Global.ControlPlane.Config.GetWorkloads().Client
	clientSecret := workloadsClient.GetSecret()
	if clientSecret == "" {
		// Nothing to seed; the operator's fallback / fail-fast guard will handle it.
		logger.Info("no workloads auth client secret in control-plane config, skipping seeding")
		return nil
	}
	clientID := workloadsClient.GetName()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      constants.WorkloadAuthSecretName,
		},
	}
	err := k8s.Get(ctx, client.ObjectKeyFromObject(secret), secret)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return fmt.Errorf("failed getting %s secret: %w", constants.WorkloadAuthSecretName, err)
	}
	notFound := k8sErrors.IsNotFound(err)

	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	secret.Data[constants.WorkloadAuthClientSecretKey] = []byte(clientSecret)
	secret.Data[oAuthClientIDKey] = []byte(clientID)

	if notFound {
		logger.Info("creating workloads auth secret with client credentials")
		return k8s.Create(ctx, secret)
	}
	logger.Info("updating workloads auth secret with client credentials")
	return k8s.Update(ctx, secret)
}

func updateAffinityAndTolerations(runaiConfig *runaiv1.RunaiConfig, clusterAffinity *corev1.Affinity, clusterTolerations []corev1.Toleration) {
	if runaiConfig.Spec.Global == nil {
		runaiConfig.Spec.Global = &global.Global{}
	}
	runaiConfig.Spec.Global.Affinity = clusterAffinity
	runaiConfig.Spec.Global.Tolerations = clusterTolerations
}

func updateNMCParams(runaiConfig *runaiv1.RunaiConfig, params *hooks.ClusterHookParams) {
	if !params.NMC.Enabled {
		return
	}
	if runaiConfig.Spec.Internal == nil {
		runaiConfig.Spec.Internal = &runaiv1_internal.Internal{}
	}
	if runaiConfig.Spec.Internal.Global == nil {
		runaiConfig.Spec.Internal.Global = &runaiv1_internal.Global{}
	}
	runaiConfig.Spec.Internal.Global.NMC.RemoteAPIServerURL = params.NMC.RemoteServerAPIServerURL
	runaiConfig.Spec.Internal.Global.NMC.ServiceAccountName = params.NMC.ServiceAccountName
	runaiConfig.Spec.Internal.Global.NMC.AllowedServiceAccounts = params.NMC.AllowedServiceAccounts
	runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusIngress = params.NMC.MgmtPrometheusIngress
}

func (p *PostUpgradeHook) removeOutdatedOwnershipMetadata(k8s client.Client) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      RunaiSchedulerConfig,
		},
	}

	err := k8s.Get(context.Background(), client.ObjectKeyFromObject(cm), cm)
	if err != nil && k8sErrors.IsNotFound(err) {
		fmt.Printf("%s cm not found, nothing to do\n", RunaiSchedulerConfig)
		return nil
	}
	if err != nil {
		return err
	}

	ownerRefs := cm.OwnerReferences
	if len(ownerRefs) == 0 {
		fmt.Printf("%s cm has no owner references, nothing to do\n", RunaiSchedulerConfig)
		return nil
	}
	if len(ownerRefs) == 1 {
		ownerRef := ownerRefs[0]
		if ownerRef.Kind != runaiv1.RunaiConfigKind {
			fmt.Printf("owner reference of %s cm is %s/%s, we want to remove it only if it's runaiconfig. nothing to do\n",
				RunaiSchedulerConfig, ownerRef.Kind, ownerRef.Name)
			return nil
		}
	}

	cm.OwnerReferences = []metav1.OwnerReference{}
	err = k8s.Update(context.Background(), cm)
	if err != nil {
		return err
	}

	return nil
}

// RecoverRunaiConfig recovers a RunaiConfig from a configmap named 'runai' in the 'runai' namespace
// to a RunaiConfig named 'runai' in the 'runai' namespace
func (p *PostUpgradeHook) recoverRunaiConfig(ctx context.Context, k8s client.Client) error {
	logger := log.FromContext(ctx)
	logger.Info("restoring runaiconfig from backup in post upgrade")

	// Wait for runaiConfig to be deleted if needed
	err := ra.WaitForRunaiConfigToBeDeletedIfNeeded(ctx, k8s, false)
	if err != nil {
		return err
	}

	// Preload inputs that do not change between attempts
	var installationConfiguration *controlplanedetailsdownloader.InstallationConfigurationData
	var runaiConfigBackup *runaiv1.RunaiConfig
	runaiConfigBackup, err = ra.GetRunaiConfigFromBackupConfigMapOrDefault(k8s)
	if err != nil {
		return err
	}
	installationConfiguration, err = controlplanedetailsdownloader.GetInstallationConfiguration(k8s)
	if err != nil {
		return err
	}
	var previousVersion string
	if runaiConfigBackup.Spec.Internal != nil && runaiConfigBackup.Spec.Internal.Global != nil {
		previousVersion = runaiConfigBackup.Spec.Internal.Global.Version
	}
	installationConfiguration.Internal.Global.ControlPlane.Config.Sanitize()
	runaiConfigBackup.Spec.Internal = &installationConfiguration.Internal

	// conflict-aware retry: if the write fails due to a modification after read,
	// rebuild from fresh state and retry. For other errors, retry commit without reread.
	var lastErr error
	found := false
	var lastPreparedRunaiconfig *runaiv1.RunaiConfig
	for attempt := 1; ; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		logger.Info("attempting to set runaiconfig", "attempt", attempt)
		// Refresh desired object on first attempt or after conflicts
		needRefresh := attempt == 1 || (lastErr != nil && k8sErrors.IsConflict(lastErr))
		if needRefresh {
			var errPrepare error
			found, lastPreparedRunaiconfig, errPrepare = prepareDesiredRunaiConfig(k8s, runaiConfigBackup)
			if errPrepare != nil {
				return errPrepare
			}
			if !found {
				logger.Info("runaiconfig does not exist, creating a new one from backup")
			} else {
				logger.Info("runaiconfig found, updating it from backup")
			}
		}

		lastErr = commitRunaiconfigToCluster(ctx, k8s, lastPreparedRunaiconfig, p.params, installationConfiguration,
			found, previousVersion)
		if lastErr == nil {
			break
		}
		logger.Error(lastErr, "failed to set runaiconfig", "attempt", attempt, "conflict", k8sErrors.IsConflict(lastErr), "reason", k8sErrors.ReasonForError(lastErr), "error", lastErr.Error())
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}
	return nil
}

func commitRunaiconfigToCluster(ctx context.Context, k8s client.Client, runaiconfig *runaiv1.RunaiConfig,
	params *hooks.ClusterHookParams, icd *controlplanedetailsdownloader.InstallationConfigurationData,
	foundInCluster bool, previousVersion string) error {
	var err error
	logger := log.FromContext(ctx)
	logger.Info("updating runaiconfig")
	updateAffinityAndTolerations(runaiconfig, params.Affinity, params.Tolerations)
	updateRbacEscalation(runaiconfig, params.EnableRbacEscalation)
	updateNMCParams(runaiconfig, params)
	updateFipsMode(runaiconfig, params.FipsMode)
	persistVersion(runaiconfig, params.Version)

	if foundInCluster {
		runaiconfig, err = schemaRunaiConfigUpgrade(context.Background(), k8s, runaiconfig, icd, previousVersion)
		if err != nil {
			return fmt.Errorf("failed to upgrade runaiconfig schema, err:[%v]", err)
		}
	}

	err = ra.PatchRunaiConfig(runaiconfig, params.ClusterConfig)
	if err != nil {
		return fmt.Errorf("failed to patch runai config with cluster config: %w", err)
	}
	return createOrUpdateRunaiConfig(ctx, k8s, runaiconfig, foundInCluster)
}

func createOrUpdateRunaiConfig(ctx context.Context, k8s client.Client, runaiconfig *runaiv1.RunaiConfig, foundInCluster bool) error {
	logger := log.FromContext(ctx)
	logger.Info("creating or updating runaiconfig")
	if !foundInCluster {
		return k8s.Create(ctx, runaiconfig)
	}
	return k8s.Update(ctx, runaiconfig)
}

func schemaRunaiConfigUpgrade(ctx context.Context, k8s client.Client,
	runaiconfig *runaiv1.RunaiConfig, icd *controlplanedetailsdownloader.InstallationConfigurationData,
	previousVersion string) (*runaiv1.RunaiConfig, error) {
	if icd.Development != nil && icd.Development.Headless {
		runaiConfigUpgrader := schemaupgrades.NewHeadlessRunaiConfigUpgrade(icd, previousVersion)
		return runaiConfigUpgrader.UpgradeRunaiConfigSpec(ctx, runaiconfig)
	} else {
		clusterCR, err := ra.GetClusterCr(ctx, k8s)
		if err != nil {
			return nil, fmt.Errorf("failed to get cluster cr, should be avaialble in cluster, err:[%v]", err)
		}

		runaiConfigUpgrader := schemaupgrades.NewRunaiConfigUpgrade(icd, clusterCR, previousVersion)
		return runaiConfigUpgrader.UpgradeRunaiConfigSpec(ctx, runaiconfig)
	}
}

func deleteObsoleteClusterCR(ctx context.Context, k8s client.Client) error {
	obsoleteClusterCR := &runaiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "runai",
		},
	}
	err := k8s.Delete(ctx, obsoleteClusterCR)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete obsolete cluster CR: %w", err)
	}

	return nil
}

func persistVersion(runaiconfig *runaiv1.RunaiConfig, version string) {
	if version == "" {
		return
	}
	if runaiconfig.Spec.Internal == nil {
		runaiconfig.Spec.Internal = &runaiv1_internal.Internal{}
	}
	if runaiconfig.Spec.Internal.Global == nil {
		runaiconfig.Spec.Internal.Global = &runaiv1_internal.Global{}
	}
	runaiconfig.Spec.Internal.Global.Version = version
}

func updateFipsMode(runaiconfig *runaiv1.RunaiConfig, fipsMode string) {
	if runaiconfig == nil {
		return
	}

	if fipsMode == "" {
		fipsMode = "off"
	}

	if runaiconfig.Spec.Internal == nil {
		runaiconfig.Spec.Internal = &runaiv1_internal.Internal{}
	}
	if runaiconfig.Spec.Internal.Global == nil {
		runaiconfig.Spec.Internal.Global = &runaiv1_internal.Global{}
	}

	runaiconfig.Spec.Internal.Global.FipsMode = fipsMode
}

func updateRbacEscalation(runaiconfig *runaiv1.RunaiConfig, enableRbacEscalation bool) {
	if enableRbacEscalation == external_workload_integrator.EnableRbacEscalationDefault {
		// no need to update, it is the default value
		return
	}

	if runaiconfig == nil {
		return
	}

	if runaiconfig.Spec.ExternalWorkloadIntegrator == nil {
		runaiconfig.Spec.ExternalWorkloadIntegrator = &external_workload_integrator.ExternalWorkloadIntegrator{}
	}
	if runaiconfig.Spec.ExternalWorkloadIntegrator.RIDRbacConfig == nil {
		runaiconfig.Spec.ExternalWorkloadIntegrator.RIDRbacConfig = &external_workload_integrator.RIDRBACConfig{}
	}

	runaiconfig.Spec.ExternalWorkloadIntegrator.RIDRbacConfig.Enabled = ptr.To(enableRbacEscalation)
}

// prepareDesiredRunaiConfig prepares the desired RunaiConfig to commit, based on whether it exists in the cluster.
func prepareDesiredRunaiConfig(k8s client.Client, backup *runaiv1.RunaiConfig) (bool, *runaiv1.RunaiConfig, error) {
	existingRunaiConfig, err := ra.GetExistingRunaiConfig(k8s)
	if k8sErrors.IsNotFound(err) {
		desired := backup.DeepCopy()
		desired.ObjectMeta.UID = ""
		desired.ObjectMeta.ResourceVersion = ""
		desired.ObjectMeta.Generation = 0
		desired.ObjectMeta.DeletionTimestamp = nil
		return false, desired, nil
	}
	if err != nil {
		return false, nil, err
	}
	existingRunaiConfig.Spec = backup.Spec
	return true, existingRunaiConfig, nil
}
