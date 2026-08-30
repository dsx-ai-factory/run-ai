package pre_hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	dynamicupgrades "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/dynamic-schema-upgrades"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	nodepoolControllerName = "nodepool-controller"
	schedulerName          = "runai-scheduler"
)

var legacyEngineSchedulingShardGVR = schema.GroupVersionResource{
	Group: "engine.run.ai", Version: "v1alpha1", Resource: "schedulingshards",
}

type PreCrdUpgradeHook struct {
	params *hooks.ClusterHookParams
}

func NewPreCrdUpgradeHook(params *hooks.ClusterHookParams) *PreCrdUpgradeHook {
	return &PreCrdUpgradeHook{params: params}
}

func (p *PreCrdUpgradeHook) GetIdentifier() string {
	return hooks.PreCrdUpgradeHookIdentifier
}

func (p *PreCrdUpgradeHook) Call(ctx context.Context, resources *clients.Clients) error {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("Starting %s", p.GetIdentifier()))
	defer logger.Info(fmt.Sprintf("Done %s", p.GetIdentifier()))

	shouldRunHook, err := p.shouldTearDownEngineSchedulers(
		ctx,
		resources.RemoteClusterClient,
		resources.RemoteClusterDynamicClient,
	)
	if err != nil {
		return fmt.Errorf("failed to check if pre-crd-upgrade hook should run: %w", err)
	}
	if !shouldRunHook {
		logger.Info("Skipping pre-crd-upgrade hook: no queue v1 CRD to migrate and no engine SchedulingShards to tear down")
		return nil
	}

	err = p.tearDownEngineSchedulers(ctx, resources,
		p.params.OperandDownIntervalMillis)
	if err != nil {
		return fmt.Errorf("failed to tear down engine schedulers: %w", err)
	}

	return nil
}

func (p *PreCrdUpgradeHook) shouldTearDownEngineSchedulers(
	ctx context.Context,
	k8sClient client.Client,
	dynamicClient dynamic.Interface,
) (bool, error) {
	queueVersionUpgrade, err := p.shouldPerformQueueVersionPreCrdUpgrade(ctx, k8sClient)
	if err != nil {
		return false, err
	}
	if queueVersionUpgrade {
		return true, nil
	}

	if p.params.EngineToKAIMigrationEnabled {
		return p.engineSchedulingShardsExist(ctx, dynamicClient)
	}

	return false, nil
}

func (p *PreCrdUpgradeHook) engineSchedulingShardsExist(ctx context.Context, dynamicClient dynamic.Interface) (bool, error) {
	schedulingShards, err := dynamicClient.Resource(legacyEngineSchedulingShardGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		if meta.IsNoMatchError(err) || errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to list scheduling shards: %w", err)
	}
	return len(schedulingShards.Items) > 0, nil
}

func (p *PreCrdUpgradeHook) tearDownEngineSchedulers(ctx context.Context, clients *clients.Clients,
	operandDownIntervalMillis int) error {
	// since in the "previous" (pre-crd-upgrade) operator version, runaiconfig doesn't support disabling the scheduler directly,
	// we will turn it off by disabling the nodepool-controller
	// and then deleting the scheduling shards.

	runaiConfigUpgrader := dynamicupgrades.NewRunaiConfigDisableNodepoolController()
	err := runaiConfigUpgrader.Upgrade(ctx, clients)
	if err != nil {
		return err
	}

	err = waitForOperandsToBeDown(ctx, clients.LocalClusterClient, nodepoolControllerName, operandDownIntervalMillis)
	if err != nil {
		return err
	}

	// now that the nodepool-controller is down, we can proceed with deleting the scheduling shards
	err = p.deleteSchedulingShards(ctx, clients)
	if err != nil {
		return fmt.Errorf("failed to delete scheduling shards: %w", err)
	}

	// once scheduling shards are deleted, the schedulers should be down as well
	return waitForOperandsToBeDown(ctx, clients.RemoteClusterClient, schedulerName, operandDownIntervalMillis)
}

func waitForOperandsToBeDown(ctx context.Context, k8s client.Client,
	operandName string, operandDownIntervalMillis int) error {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("Waiting for all %s operands to be down before proceeding with CRD upgrade", operandName))

	operandDownInterval := time.Duration(operandDownIntervalMillis) * time.Millisecond

	ticker := time.NewTicker(operandDownInterval)
	defer ticker.Stop()

	logger.Info(fmt.Sprintf("Waiting for all %s operands to be down, interval %d milliseconds",
		operandName, operandDownIntervalMillis))

	for {
		select {
		case <-ticker.C:
			allOperandsDown, err := areAllOperandsDown(ctx, k8s, operandName)
			if err != nil {
				return fmt.Errorf("error checking if all %s operands are down: %w", operandName, err)
			}
			if allOperandsDown {
				logger.Info(fmt.Sprintf("All %s operands are down", operandName))
				return nil
			}
			logger.Info(fmt.Sprintf("Not all %s operands are down yet", operandName))
		case <-ctx.Done():
			return fmt.Errorf("timeout reached while waiting for all %s operands to be down", operandName)
		}
	}
}

func areAllOperandsDown(ctx context.Context, k8s client.Client, operandName string) (bool, error) {
	pods := &v1.PodList{}
	err := k8s.List(ctx, pods, &client.ListOptions{Namespace: runaiv1.DefaultNamespace})
	if err != nil {
		return false, fmt.Errorf("failed to list pods in runai namespace: %w", err)
	}

	for _, pod := range pods.Items {
		appName, ok := pod.Labels["app"]
		if !ok || !strings.HasPrefix(appName, operandName) {
			continue
		}

		// found operand
		return false, nil
	}

	return true, nil
}

func (p *PreCrdUpgradeHook) deleteSchedulingShards(ctx context.Context, clients *clients.Clients) error {
	logger := log.FromContext(ctx)
	logger.Info("Deleting scheduling shards")

	shardResource := clients.RemoteClusterDynamicClient.Resource(legacyEngineSchedulingShardGVR)
	schedulingShards, err := shardResource.List(ctx, metav1.ListOptions{})
	if err != nil {
		if meta.IsNoMatchError(err) || errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to list scheduling shards: %w", err)
	}

	for _, shard := range schedulingShards.Items {
		err = shardResource.Delete(ctx, shard.GetName(), metav1.DeleteOptions{})
		if err != nil && !meta.IsNoMatchError(err) && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete scheduling shard %s: %w", shard.GetName(), err)
		}
	}
	return nil
}

func (p *PreCrdUpgradeHook) shouldPerformQueueVersionPreCrdUpgrade(ctx context.Context, k8sClient client.Client) (bool, error) {
	logger := log.FromContext(ctx)

	queueCrd := &apiextensionsv1.CustomResourceDefinition{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: "queues.scheduling.run.ai"}, queueCrd)
	if err != nil {
		if errors.IsNotFound(err) {
			// no queues CRD found, no need to perform the hook logic
			return false, nil
		}
		logger.Error(err, "Failed to get CRD queues.scheduling.run.ai")
		return false, err
	}

	for _, version := range queueCrd.Spec.Versions {
		if version.Storage == true && version.Name == "v2" {
			// if the stored version is already v2, we don't need to perform the hook logic
			return false, nil
		}
	}
	// if other version is stored, we need to perform the hook logic
	return true, nil
}
