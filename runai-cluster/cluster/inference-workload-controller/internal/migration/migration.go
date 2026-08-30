package migration

import (
	"context"
	"errors"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	workloadMigratorConfigName         = "inference-workload-controller"
	workloadMigratorConfigNamespace    = "runai"
	defaultGracePeriod                 = 6 * time.Second
	webserverGracePeriodTimeoutMessage = "webserver did not start during the grace period"
)

var log = logger.Log.WithName("migration")

type WorkloadMigration struct {
	k8sClient                client.Client
	migrationConfigMapName   string
	migrationConfigNamespace string
	migrationSteps           []Migrator
	webserverGraceDelay      time.Duration
}

type WebhookServer interface {
	Start(ctx context.Context) error
	StartedChecker() healthz.Checker
}

type Migrator interface {
	Migrate(ctx context.Context) error
	String() string
	RequiresWebhooks() bool
}

func (wm *WorkloadMigration) GetMigrationConfigMap(ctx context.Context) (corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{}
	err := wm.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: wm.migrationConfigNamespace,
		Name:      wm.migrationConfigMapName,
	}, &configMap)
	return configMap, err
}

func (wm *WorkloadMigration) Run(ctx context.Context, webServer WebhookServer) error {
	log.Info("Running migration steps")
	timeNow := time.Now()

	defer func() {
		log.Info("migration took", "time", time.Since(timeNow))
	}()

	configMap, err := wm.GetMigrationConfigMap(ctx)
	if err != nil {
		log.Error(err, "failed to get migration config map, running migration steps without config map")
	}

	var create bool
	if apierrors.IsNotFound(err) {
		create = true
	}

	defer func() {
		if err = wm.updateMigrationConfigMap(ctx, configMap, create); err != nil {
			log.Error(err, "failed to update migration config map")
		}
	}()

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	startedWebserver := false
	if wm.webserverGraceDelay == 0 {
		wm.webserverGraceDelay = defaultGracePeriod
	}
	webserverCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		if startedWebserver {
			time.Sleep(wm.webserverGraceDelay)
		}
	}()

	for _, migrator := range wm.migrationSteps {
		migratorCp := migrator
		migratorName := migratorCp.String()
		if _, ok := configMap.Data[migratorName]; !ok {
			log.Info("Running migration", "step", migratorName)

			if migratorCp.RequiresWebhooks() && !startedWebserver {
				err = wm.startWebServer(webServer, webserverCtx, webServer.StartedChecker())
				if err != nil {
					return err
				}
				startedWebserver = true
			}

			err = migratorCp.Migrate(ctx)
			if err != nil {
				log.Error(err, "failed to run a migration step", "step", migratorName)
				return err
			}
			configMap.Data[migratorName] = "true"

		} else {
			log.Info("Skipping an already handled migration", "step", migratorName)
		}
	}

	log.Info("migration steps completed successfully")

	return nil
}

func ExternalWorkloadByOriginalWorkloadUIDIndexer(obj client.Object) []string {
	externalWorkload, ok := obj.(*v1alpha1.ExternalWorkload)
	if !ok {
		log.Error(fmt.Errorf("failed converting object to ExternalWorkload in ExternalWorkloadOfObjectIndexer: %v", obj), "")
		return []string{""}
	}

	return []string{string(externalWorkload.Spec.OriginalWorkloadUID)}
}

func (wm *WorkloadMigration) startWebServer(webServer WebhookServer, webserverCtx context.Context, webhookHealthFn healthz.Checker) error {
	startupErrChan := make(chan error)

	go func() {
		err := webServer.Start(webserverCtx)
		if err != nil {
			log.Error(err, "failed to start webserver")
			startupErrChan <- err
		}
	}()

	log.Info("waiting for the webserver to start")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), wm.webserverGraceDelay)
	defer cancel()
	go func() {
		for {
			select {
			case <-ctxTimeout.Done():
				startupErrChan <- errors.New(webserverGracePeriodTimeoutMessage)
				return
			default:
				if err := webhookHealthFn(&http.Request{}); err == nil {
					log.Info("webserver started and reachable")
					startupErrChan <- nil
					return
				}
			}
		}
	}()

	return <-startupErrChan
}

func (wm *WorkloadMigration) updateMigrationConfigMap(ctx context.Context, configMap corev1.ConfigMap, create bool) error {
	if create {
		configMap.Name = wm.migrationConfigMapName
		configMap.Namespace = wm.migrationConfigNamespace
		return wm.k8sClient.Create(ctx, &configMap)
	}
	return wm.k8sClient.Update(ctx, &configMap)
}

func NewWorkloadMigration(k8sClient client.Client, k8sClientWithFieldIndexer CachedClient) (*WorkloadMigration, error) {
	// the migrations use two different clients:
	// 1. regular (uncached) client: for operations that are not supported in
	//    the cached client (e.g. Delete).
	//    we can't just use the manager's client here, as it's only started when
	//    starting the webhook server, which is inside the migration loop itself,
	//    while we need the client before that (e.g. to get ConfigMap)
	// 2. cached client: for the top-owner migration, which needs a field indexer,
	//    a feature of cached clients only

	networkCleaner := NewNetworkCleaner(k8sClient, workloadconfig.IsOpenshift())
	topOwner := NewTopOwnerMigrator(k8sClient, k8sClientWithFieldIndexer)
	nimMetricsMigration := NewNIMMetricsServiceMigration(k8sClient)
	checksumSHA256 := NewChecksumSHA256Migration(k8sClient)

	return &WorkloadMigration{
		k8sClient:                k8sClient,
		migrationConfigMapName:   workloadMigratorConfigName,
		migrationConfigNamespace: workloadMigratorConfigNamespace,
		migrationSteps:           []Migrator{networkCleaner, topOwner, nimMetricsMigration, checksumSHA256},
		webserverGraceDelay:      defaultGracePeriod,
	}, nil
}
