package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/rs/zerolog/log"

	runai_constants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	schemeinit "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/config"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller/distributedinference"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller/inference"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/controller/revision"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/migration"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/policy"
	webhookv "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/webhook"
	webhookmutator "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/webhook/mutator"
	webhookvalidator "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/webhook/validator"
	webhookvalidatorpolicy "github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/webhook/validator/policy"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/pkg/apis"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	infp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	workloadpolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/policy"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/mutators"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	workloadspolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads/policy"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

var (
	scheme *runtime.Scheme
)

type restMappingResolver interface {
	RESTMapping(schema.GroupKind, ...string) (*meta.RESTMapping, error)
}

type apiAvailability struct {
	inferenceWorkload            bool
	inferencePolicy              bool
	distributedInferenceWorkload bool
}

func resolveAPIAvailability(resolver restMappingResolver) (apiAvailability, error) {
	availability := apiAvailability{}
	candidates := []struct {
		gvk       schema.GroupVersionKind
		available *bool
	}{
		{gvk: infw.GroupVersion.WithKind("InferenceWorkload"), available: &availability.inferenceWorkload},
		{gvk: infp.GroupVersion.WithKind("InferencePolicy"), available: &availability.inferencePolicy},
		{gvk: dinfw.GroupVersion.WithKind("DistributedInferenceWorkload"), available: &availability.distributedInferenceWorkload},
	}

	var resolutionErr error
	for _, candidate := range candidates {
		_, err := resolver.RESTMapping(candidate.gvk.GroupKind(), candidate.gvk.Version)
		if meta.IsNoMatchError(err) {
			log.Warn().Str("gvk", candidate.gvk.String()).Msg("Legacy API is not served; dependent paths will be skipped")
			continue
		}
		if err != nil {
			resolutionErr = errors.Join(resolutionErr, fmt.Errorf("resolve %s: %w", candidate.gvk, err))
			continue
		}
		*candidate.available = true
	}

	return availability, resolutionErr
}

const (
	inferenceTypeForWebhook            = "inference"
	distributedInferenceTypeForWebhook = "distributed-inference"
	inferencePolicyType                = "inferencepolicy"
	validationPolicyWebhooksPrefix     = "validate-run-ai-v2alpha1"
	EnvVarClusterVersion               = "CLUSTER_VERSION"
)

func init() {
	scheme = schemeinit.Scheme()
	utilruntime.Must(infp.AddToScheme(scheme))        // inference policy
	utilruntime.Must(infw.AddToScheme(scheme))        // inference workload
	utilruntime.Must(dinfp.AddToScheme(scheme))       // distributed inference policy
	utilruntime.Must(dinfw.AddToScheme(scheme))       // distributed inference workload
	utilruntime.Must(servingv1.AddToScheme(scheme))   // knativeservice
	utilruntime.Must(lwsv1.AddToScheme(scheme))       // leaderworkerset
	utilruntime.Must(v1alpha1.AddToScheme(scheme))    // external workload (needed for the migration)
	utilruntime.Must(kaiv2alpha2.AddToScheme(scheme)) // podgroup (needed for the migration)
	utilruntime.Must(nvapi.AddToScheme(scheme))       // computedomain (needed for the mnnvl)
	utilruntime.Must(gatewayv1.Install(scheme))
}

const (
	logLevelInfo  zapcore.Level = 0
	logLevelDebug zapcore.Level = -4
)

func initLogging() {
	// Set up a new logger
	// to get logr verbosity of 4, set this field to zapcore.Level(-4) ...

	var logLevel zapcore.Level
	if config.Config().UseDebugLogLevel {
		logLevel = logLevelDebug
	} else {
		logLevel = logLevelInfo
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Level: logLevel,
	})))
}

func main() {
	initLogging()

	// Wait for Knative or LWS to become available
	knativeAvailable, lwsAvailable := waitForKnativeOrLws()

	log.Info().Msg("Starting Run:AI inference-workload-controller")

	clientConfig := ctrl.GetConfigOrDie()
	clientConfig.QPS = float32(config.Config().K8sClientConfigQPS)
	clientConfig.Burst = config.Config().K8sClientConfigBurst
	clientConfig.UserAgent = "inference-workload-controller"

	log.Info().Msg("Starting InferenceWorkloadController")
	metricExporterBindAddress := config.Config().MetricExporterBindAddr
	if config.Config().DisableMetrics {
		log.Info().Msg("metric exporter is disabled")
		metricExporterBindAddress = "0"
	}

	// run migrations
	if config.Config().SkipMigrations {
		log.Info().Msg("Migrations are skipped per user request")
	} else {
		err := runMigrations(clientConfig)
		if err != nil {
			log.Error().Msgf("failed to complete the required migrations: %v", err)
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(clientConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricExporterBindAddress,
		},
		WebhookServer: &webhook.DefaultServer{
			Options: webhook.Options{
				Port:    config.Config().WebhookPort,
				CertDir: config.Config().CertDir,
			}},
		HealthProbeBindAddress: config.Config().HealthProbeBindAddress,
		LeaderElection:         config.Config().EnableLeaderElection,
		LeaderElectionID:       "inference-workload.run.ai",
	})
	if err != nil {
		log.Error().Msgf("Failed to create manager: %v", err)
		os.Exit(1)
	}

	availability, err := resolveAPIAvailability(mgr.GetRESTMapper())
	if err != nil {
		log.Error().Msgf("Failed to resolve legacy API availability: %v", err)
		os.Exit(1)
	}

	k8sClient, err := client.New(mgr.GetConfig(), client.Options{Scheme: scheme})
	if err != nil {
		log.Error().Msgf("Failed to create client: %v", err)
		os.Exit(1)
	}

	if knativeAvailable && availability.inferenceWorkload {
		err := setupInferenceControllers(mgr)
		if err != nil {
			log.Error().Msgf("Failed to setup inference controllers: %v", err)
			os.Exit(1)
		}
	} else if knativeAvailable {
		log.Warn().Msg("InferenceWorkload API is not served; skipping inference controllers")
	}

	if lwsAvailable && availability.distributedInferenceWorkload {
		err := setupDistributedInferenceControllers(mgr)
		if err != nil {
			log.Error().Msgf("Failed to setup distributed inference controllers: %v", err)
			os.Exit(1)
		}
	} else if lwsAvailable {
		log.Warn().Msg("DistributedInferenceWorkload API is not served; skipping distributed inference controllers")
	}

	clusterVersion := os.Getenv(workloadcommon.EnvVarClusterVersion)
	if len(clusterVersion) == 0 {
		log.Error().Msgf("Failed to create inference controller: env var %s is not set", workloadcommon.EnvVarClusterVersion)
		os.Exit(1)
	}

	log.Info().Msg("setting up webhooks")
	setupWebhook(mgr, clusterVersion, availability)

	if config.Config().EnableWorkloadOwnershipProtection && availability.inferenceWorkload {
		log.Info().Msg("enabling workload ownership protection webhooks")
		enableWorkloadOwnershipProtection(mgr, k8sClient)
	}

	log.Info().Msg("setting up health check")
	setupHealthCheck(mgr)

	log.Info().Msg("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error().Msgf("Failed to start manager: %v", err)
		os.Exit(1)
	}
}

func setupInferenceControllers(mgr manager.Manager) error {
	log.Info().Msg("Knative is available, starting inference controllers")

	inferenceController := inference.NewSimpleReconcile(mgr.GetClient(), mgr.GetAPIReader(), mgr.GetConfig())
	if err := inferenceController.SetupWithManager(mgr, config.Config().MaxConcurrentReconciles, config.Config().EnableMnnvlSupport); err != nil {
		log.Error().Msgf("Failed to create inference controller: %v", err)
		return err
	}

	revisionController := revision.NewSimpleReconcile(mgr.GetClient())
	if err := revisionController.SetupWithManager(mgr, config.Config().MaxConcurrentReconciles); err != nil {
		log.Error().Msgf("Failed to create revision controller: %v", err)
		return err
	}

	log.Info().Msg("Inference controllers started successfully")
	return nil
}

func setupDistributedInferenceControllers(mgr manager.Manager) error {
	log.Info().Msg("LWS is available, starting distributed inference controllers")

	distributedInferenceController := distributedinference.NewController(mgr.GetClient(), workloadconfig.IsOpenshift(), config.Config().ElasticEnabled)
	if err := distributedInferenceController.SetupWithManager(mgr, config.Config().MaxConcurrentReconciles); err != nil {
		log.Error().Msgf("Failed to create distributed inference controller: %v", err)
		return err
	}

	log.Info().Msg("Distributed inference controllers started successfully")
	return nil
}

func runMigrations(cfg *rest.Config) error {
	// init manager
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "",
			CertDir:     config.Config().CertDir,
		},
		WebhookServer: &webhook.DefaultServer{
			Options: webhook.Options{
				Port:    config.Config().WebhookPort,
				CertDir: config.Config().CertDir,
			}},
	})
	if err != nil {
		return fmt.Errorf("unable to start webhooks: %w", err)
	}

	// init the required clients (see NewWorkloadMigration docs)
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	k8sClientWithFieldIndexer, err := migration.NewMigrationCachedClient(cfg, cache.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := k8sClientWithFieldIndexer.Start(ctx); err != nil {
			log.Error().Msgf("Failed to start cache: %v", err)
		}
	}()

	// Wait for the cache to sync
	if !k8sClientWithFieldIndexer.WaitForCacheSync(ctx) {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	err = setMigrationFieldIndexers(k8sClientWithFieldIndexer)
	if err != nil {
		return err
	}

	err = registerMigrationWebhooks(mgr, k8sClient)
	if err != nil {
		return err
	}

	migrator, err := migration.NewWorkloadMigration(k8sClient, k8sClientWithFieldIndexer)
	if err != nil {
		return err
	}

	err = migrator.Run(context.Background(), mgr.GetWebhookServer())
	if err != nil {
		return err
	}

	return nil
}

func registerMigrationWebhooks(mgr manager.Manager, k8sClient client.Client) error {
	migrationWebookHandler, err := migration.NewWebhookHandler(k8sClient, "")
	if err != nil {
		return fmt.Errorf("failed to create migration webhook handler: %w", err)
	}
	webhookServer := mgr.GetWebhookServer()
	registerWorkloadValidationWebhook(webhookServer, migrationWebookHandler, inferenceTypeForWebhook)
	registerWorkloadMutationWebhook(webhookServer, migrationWebookHandler, inferenceTypeForWebhook)
	return nil
}

// setMigrationFieldIndexers set required field indexers, in order to be able to list external-workload resources by their owner
func setMigrationFieldIndexers(k8sCachedClient migration.CachedClient) error {
	err := k8sCachedClient.IndexField(
		context.Background(), &v1alpha1.ExternalWorkload{},
		migration.ExternalWorkloadByOriginalUIDIndexerName, migration.ExternalWorkloadByOriginalWorkloadUIDIndexer)
	if err != nil {
		return fmt.Errorf("failed indexing ew field: %v: %w", migration.ExternalWorkloadByOriginalUIDIndexerName, err)
	}

	return nil
}

func setupWebhook(mgr manager.Manager, clusterVersion string, availability apiAvailability) {
	// register policy validation webhooks
	if availability.inferencePolicy {
		setupPolicyValidationWebhooks(mgr)
	}

	// register workload mutation webhooks
	setupWorkloadMutationWebhooks(mgr, clusterVersion, availability)

	// register workload validation webhooks
	setupWorkloadValidationWebhooks(mgr, os.Getenv(EnvVarClusterVersion), availability)
}

func setupPolicyValidationWebhooks(mgr manager.Manager) {
	k8sClient := mgr.GetClient()

	webhookServer := mgr.GetWebhookServer()

	policyValidationWebhookHandler := workloadspolicy.NewPolicyValidator(k8sClient, webhookvalidatorpolicy.InitInferencePolicyFromReq, config.Config().AllowedServiceAccounts)

	webhookServer.Register(fmt.Sprintf("/%s-%s", validationPolicyWebhooksPrefix, inferencePolicyType),
		&webhook.Admission{Handler: policyValidationWebhookHandler})
}

func enableWorkloadOwnershipProtection(mgr manager.Manager, k8sClient client.Client) {
	err := jobs.IndexAccessRules(mgr)
	if err != nil {
		log.Error().Msgf("Failed to index access rules: %v", err)
		os.Exit(1)
	}

	workloadControllerUID, err := migration.GetInferenceWorkloadControllerUID(context.Background(), k8sClient)
	if err != nil {
		log.Error().Msgf("Failed getting inference-workload-controller UID: %v", err)
		os.Exit(1)
	}

	workloadOwnershipProtectionValidator := webhookv.NewInferenceWorkloadOwnershipProtectionValidator(mgr.GetClient(), workloadControllerUID)
	workloadOwnershipMutator := webhookv.NewInferenceWorkloadOwnershipMutator(mgr.GetClient(), workloadControllerUID)

	webhookServer := mgr.GetWebhookServer()
	registerWorkloadOwnershipProtectionWebhooks(webhookServer, workloadOwnershipProtectionValidator, workloadOwnershipMutator)
}

func registerWorkloadOwnershipProtectionWebhooks(webhookServer webhook.Server, workloadOwnershipProtectionValidator, workloadOwnershipMutator admission.Handler) {
	webhookServer.Register("/validate-workload-ownership-protection", &webhook.Admission{Handler: workloadOwnershipProtectionValidator})
	webhookServer.Register("/mutate-workload-ownership", &webhook.Admission{Handler: workloadOwnershipMutator})
}

func setupWorkloadMutationWebhooks(mgr manager.Manager, clusterVersion string, availability apiAvailability) {
	webserver := mgr.GetWebhookServer()

	// register inference workload merge mutator webhook
	if availability.inferenceWorkload {
		inferenceMergeStackCreator := policy.NewInferenceWorkloadMergeCreator(mgr.GetClient())
		inferenceWorkloadPolicy := workloadpolicy.NewSimplePolicyApplier(inferenceMergeStackCreator)
		inferenceHandler := mutators.NewWorkloadMutatorHandler(inferenceWorkloadPolicy, apis.GetNewInferenceWorkloadFn, clusterVersion, &webhookmutator.WorkloadStatusUpdater{}, false, false, false)

		registerWorkloadMutationWebhook(webserver, inferenceHandler, inferenceTypeForWebhook)
	}

	if availability.distributedInferenceWorkload {
		distributedInferenceMutator := webhookmutator.NewDistributedInferenceMutator(mgr.GetClient(), config.Config().EnableMnnvlSupport, config.Config().TopologyEnabled, config.Config().ElasticEnabled)
		registerWorkloadMutationWebhook(webserver, distributedInferenceMutator, distributedInferenceTypeForWebhook)
	}
}

func setupWorkloadValidationWebhooks(
	mgr manager.Manager,
	clusterVersion string,
	availability apiAvailability,
) {
	webhookServer := mgr.GetWebhookServer()

	// inference workload
	if availability.inferenceWorkload {
		knativeValidator := webhookvalidator.NewKnativeValidator(mgr.GetClient())
		infwGenericParamsValidator := workloads.NewGenericParamsValidator()
		infwParamsValidator := webhookvalidator.NewInferenceParamsValidator()
		infwGpuValidator := jobs.NewWorkloadGpuValidator()

		infwValidatorHandler := webhookv.NewInferenceValidator(mgr.GetClient(), config.Config().AllowedServiceAccounts, knativeValidator, infwGenericParamsValidator, infwParamsValidator, infwGpuValidator)
		registerWorkloadValidationWebhook(webhookServer, infwValidatorHandler, inferenceTypeForWebhook)
	}

	// distributed inference workload
	if availability.distributedInferenceWorkload {
		lwsValidator := webhookvalidator.NewLwsValidator(mgr.GetClient())
		dinfwParamsValidator := webhookvalidator.NewDistributedInferenceParamsValidator(mgr.GetClient(), config.Config().EnableMnnvlSupport, config.Config().TopologyEnabled, config.Config().ElasticEnabled)
		dinfwGpuValidator := webhookvalidator.NewDistributedInferenceGpuValidator()

		dinfwValidatorHandler := webhookv.NewDistributedInferenceValidator(mgr.GetClient(), config.Config().AllowedServiceAccounts, config.Config().ElasticEnabled, lwsValidator, dinfwParamsValidator, dinfwGpuValidator)
		registerWorkloadValidationWebhook(webhookServer, dinfwValidatorHandler, distributedInferenceTypeForWebhook)
	}
}

func registerWorkloadValidationWebhook(webhookServer webhook.Server, handler webhook.AdmissionHandler, workloadType string) {
	// register workload validation webhook
	webhookServer.Register(fmt.Sprintf("/validate-%s-workload", workloadType), &webhook.Admission{Handler: handler})
}

func registerWorkloadMutationWebhook(webhookServer webhook.Server, workloadMutationHandler webhook.AdmissionHandler, workloadType string) {
	webhookServer.Register(fmt.Sprintf("/mutate-%s-workload", workloadType), &webhook.Admission{Handler: workloadMutationHandler})
}

func setupHealthCheck(mgr manager.Manager) {
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Err(err).Msg("unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Err(err).Msg("unable to set up ready check")
		os.Exit(1)
	}
}

func waitForKnativeOrLws() (bool, bool) {
	log.Info().Msg("Waiting for Knative or LWS to become available...")

	for {
		knativeAvailable := os.Getenv(runai_constants.KnativeInstalledEnvVar) == "true"
		lwsAvailable := os.Getenv(runai_constants.LwsInstalledEnvVar) == "true"

		if knativeAvailable || lwsAvailable {
			log.Info().Msgf("Knative available: %v, LWS available: %v", knativeAvailable, lwsAvailable)
			return knativeAvailable, lwsAvailable
		}

		log.Info().Msg("Neither Knative nor LWS is available yet")
		// We can just sleep forever. Once the env vars will change the pod will get a restart from the deployment
		time.Sleep(1 * time.Hour)
	}
}
