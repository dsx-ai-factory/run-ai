/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package setup

import (
	"net/http"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/logs/download"

	"github.com/MicahParks/keyfunc"
	"github.com/go-chi/chi/v5"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/rs/zerolog/log"
	authzclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	workloadsapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/attach"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/auth"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz"
	control_plane "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz/authorizers/control-plane"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz/authorizers/control-plane/cached"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/exec"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/k8sobject"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/logs"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/metrics"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/portforward"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/project"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/router"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
	schemeinit "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	rjv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	distinfwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	distw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	mpiv2beta1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	twv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/common/auth/providers/oidc/configuration"
	"github.com/run-ai/runai/runai-cluster/common/auth/providers/oidc/userinfo"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func RouterContainer(locator workload.Locator, urlHandler map[string]router.UrlHandler, jwksFunc *keyfunc.JWKS,
	oidcConfigurationProvider configuration.Provider, userInfoProvider userinfo.Provider,
	config *config.ClusterAPI, mgr manager.Manager,
) (router.Container, error) {
	workloadMiddleware := workload.NewExistMiddleware(locator)
	allowedOrigins := buildAllowedOrigins(config.ControlPlaneURL)
	authMiddleware := auth.CreateAuthMiddleware(jwksFunc, oidcConfigurationProvider, userInfoProvider, allowedOrigins)
	authzMiddleware := InitAuthzMiddleware(config.ControlPlaneURL, locator)
	namepsaceMiddleware := InitProjectMiddleware(mgr.GetClient())
	attributesMiddleware := InitAttributeMiddleware()
	metricsMiddleware := metrics.NewConnectionMiddleware(config.EnableMetrics)

	workloadsAPIClient := newWorkloadsAPIClient(config.ControlPlaneURL)
	k8sObjectMiddleware, err := InitK8sObjectMiddleware(workloadsAPIClient, config.WorkloadAuthCacheTTL)
	if err != nil {
		return router.Container{}, err
	}
	k8sObjectService, err := InitK8sObjectService(workloadsAPIClient, mgr.GetConfig(), mgr.GetScheme())
	if err != nil {
		return router.Container{}, err
	}

	containerBuilder := router.NewContainerBuilder()
	containerBuilder.WithRouter(chi.NewRouter())
	containerBuilder.WithWorkloadMiddleware(workloadMiddleware.Middleware)
	containerBuilder.WithAuthMiddleware(authMiddleware)
	containerBuilder.WithAuthzMiddleware(authzMiddleware.Middleware)
	containerBuilder.WithProjectMiddleware(namepsaceMiddleware.Middleware)
	containerBuilder.WithWorkloadAttributesMiddleware(attributesMiddleware.Middleware)
	containerBuilder.WithMetricsMiddleware(metricsMiddleware.Middleware)
	containerBuilder.WithHandler(urlHandler)
	containerBuilder.WithK8sObjectMiddleware(k8sObjectMiddleware)
	containerBuilder.WithK8sObjectService(k8sObjectService)

	return containerBuilder.Build(), nil
}

func InitJWKSFunction(jwksURL string) (*keyfunc.JWKS, error) {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshErrorHandler: func(err error) {
			log.Printf("there was a failure refreshing JWKS data: %v", err)
		},
		RefreshInterval:   time.Minute,
		RefreshUnknownKID: true,
	})
	if err != nil {
		return nil, err
	}

	return jwks, nil
}

func Scheme() *runtime.Scheme {
	scheme := schemeinit.Scheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(ipv1.AddToScheme(scheme))         // interactive policy
	utilruntime.Must(iwv1.AddToScheme(scheme))         // interactive workload
	utilruntime.Must(twv1.AddToScheme(scheme))         // training workload
	utilruntime.Must(distw.AddToScheme(scheme))        // distributed workload
	utilruntime.Must(rjv1.AddToScheme(scheme))         // runai job
	utilruntime.Must(runaiv2.AddToScheme(scheme))      // runai project
	utilruntime.Must(mpiv2beta1.AddToScheme(scheme))   // mpi job
	utilruntime.Must(kubeflowv1.AddToScheme(scheme))   // pytorchjob, tfjob, xgboostjob
	utilruntime.Must(infwv2a1.AddToScheme(scheme))     // inference workload
	utilruntime.Must(distinfwv2a1.AddToScheme(scheme)) // distributed inference workload

	return scheme
}

func GoClient(kubeConfig *rest.Config) kubernetes.Interface {
	coreV1Client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	return coreV1Client
}

func InitProjectMiddleware(client client.Client) *project.NamespaceMiddleware {
	projectTranslator := project.NewNamespaceTranslator(client)
	return project.NewNamespaceMiddleware(projectTranslator)
}

func InitAttributeMiddleware() *api.Middleware {
	return api.NewMiddleware()
}

func InitAuthzMiddleware(authorizationServiceURL string, locator workload.Locator) *authz.Middleware {
	authzclient.Init(&authzclient.Config{
		Url: authorizationServiceURL,
	})

	authzClient := authzclient.NewAuthorizationExternalClient()
	cpAuthorizer := cached.NewCachedControlPlaneAuthorizer(control_plane.NewControlPlaneAuthorizerWithClient(authzClient))

	return authz.NewMiddleware(cpAuthorizer, locator)
}

func InitK8sObjectMiddleware(apiClient *workloadsapi.APIClient, cacheTTL time.Duration) (func(http.Handler) http.Handler, error) {
	authorizer, err := k8sobject.NewCachedWorkloadAuthorizer(k8sobject.NewAPIWorkloadsGetter(apiClient.WorkloadsAPI), cacheTTL)
	if err != nil {
		return nil, err
	}
	return k8sobject.WorkloadAuthorizerMiddleware(authorizer), nil
}

func InitK8sObjectService(apiClient *workloadsapi.APIClient, restConfig *rest.Config, scheme *runtime.Scheme) (*k8sobject.Service, error) {
	nonCachedClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}
	elementGetter, podGetter := k8sobject.NewClients(apiClient)
	return k8sobject.NewService(nonCachedClient, elementGetter, podGetter), nil
}

func newWorkloadsAPIClient(controlPlaneURL string) *workloadsapi.APIClient {
	cfg := workloadsapi.NewConfiguration()
	cfg.Servers = workloadsapi.ServerConfigurations{{URL: controlPlaneURL}}
	return workloadsapi.NewAPIClient(cfg)
}

func WorkloadLocator(client client.Client) workload.Locator {
	return workload.NewK8sLocator(client)
}

func CreateHandlerList(client client.Client, rest *rest.Config, timeout time.Duration) map[string]router.UrlHandler {
	handlerMap := map[string]router.UrlHandler{}

	podGetter := workload.NewPodsSelector(client)
	coreV1Client := GoClient(rest)
	logService := LogService(podGetter, coreV1Client)

	handlerMap[router.PathLogs] = getLogHandler(logService, timeout)
	handlerMap[router.PathLogsDownload] = getLogDownloadHandler(logService)
	handlerMap[router.PathPortforward] = getPortforwardHandler(rest, coreV1Client, podGetter, timeout)
	handlerMap[router.PathExec] = getExecHandler(podGetter, rest, coreV1Client, timeout)
	handlerMap[router.PathAttach] = getAttachHandler(podGetter, rest, coreV1Client, timeout)

	return handlerMap
}

func getPortforwardHandler(
	rc *rest.Config, coreV1Client kubernetes.Interface, podGetter *workload.Pods, timeout time.Duration,
) router.UrlHandler {
	wu := websocket.NewUpgradeConnection()
	return portforward.NewHandler(PFService(rc, coreV1Client, podGetter), wu, timeout)
}

func getExecHandler(
	podGetter *workload.Pods, rc *rest.Config, coreV1Client kubernetes.Interface, timeout time.Duration,
) router.UrlHandler {
	wu := websocket.NewUpgradeConnection()
	return exec.NewHandler(ExecService(podGetter, rc, coreV1Client), wu, timeout)
}

func getAttachHandler(
	podGetter *workload.Pods, rc *rest.Config, coreV1Client kubernetes.Interface, timeout time.Duration,
) router.UrlHandler {
	wu := websocket.NewUpgradeConnection()
	return attach.NewHandler(AttachService(podGetter, rc, coreV1Client), wu, timeout)
}

func getLogHandler(logsService logs.PodLogsService, timeout time.Duration) router.UrlHandler {
	wu := websocket.NewUpgradeConnection()
	proxyFactory := logs.NewOneWayProxyFactory()

	return logs.NewHandler(logsService, wu, proxyFactory, timeout)
}

func getLogDownloadHandler(logsService logs.PodLogsService) router.UrlHandler {
	return download.NewHandler(logsService)
}

func LogService(podGetter *workload.Pods, coreV1Client kubernetes.Interface) logs.PodLogsService {
	logStreamer := logs.NewPodStreamer(coreV1Client)
	return logs.NewService(podGetter, logStreamer)
}

func PFService(rc *rest.Config, coreV1Client kubernetes.Interface, podGetter *workload.Pods) portforward.PodPFService {
	proxyFactory := portforward.NewTwoWayProxyFactory()
	forwarder := portforward.NewPodForwarder(rc, coreV1Client, proxyFactory)

	return portforward.NewService(podGetter, forwarder)
}

func ExecService(podGetter *workload.Pods, rc *rest.Config, coreV1Client kubernetes.Interface) exec.PodService {
	podExec := exec.NewPodExecute(coreV1Client)
	return exec.NewService(podGetter, podExec, rc)
}

func AttachService(podGetter *workload.Pods, rc *rest.Config, coreV1Client kubernetes.Interface) attach.PodService {
	podAttach := attach.NewPodAttach(coreV1Client)
	return attach.NewService(podGetter, podAttach, rc)
}

func buildAllowedOrigins(controlPlaneURL string) []string {
	var origins []string
	if origin := auth.OriginFromURL(controlPlaneURL); origin != "" {
		origins = append(origins, origin)
	}
	if len(origins) == 0 {
		log.Warn().Msg("No allowed CORS origins configured; all cross-origin requests will be blocked")
	}
	return origins
}
