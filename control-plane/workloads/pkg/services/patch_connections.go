/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"net"
	"net/url"
	"strings"
	"unicode"
)

const (
	dot   = "."
	colon = ":"

	TypeUrl          = string(openapi.INTERNALCONNECTIONTYPE_EXTERNAL_URL)
	TypeNodePort     = string(openapi.INTERNALCONNECTIONTYPE_NODE_PORT)
	TypeLoadBalancer = string(openapi.INTERNALCONNECTIONTYPE_LOAD_BALANCER)
	TypeServingPort  = string(openapi.INTERNALCONNECTIONTYPE_SERVING_PORT)
)

// for debugging and testing purposes function will also report the method of patching it have chosen
type PatchingMethod string

const (
	PatchingMethodUndefined               = PatchingMethod("undef")
	PatchingMethodNaive                   = PatchingMethod("naive")
	PatchingMethodNetworkStatNoPreProc    = PatchingMethod("network-stat-no-pre-proc")
	PatchingMethodNetworkStatPreProcessed = PatchingMethod("network-stat-pre-processed")
)

// patchConnections patches connections with the link that was assigned to them by the cluster.
func patchConnections(ctx context.Context, envs []api.Environment, urls []string, networkStatus *api.WorkloadNetworkStatus) (PatchingMethod, error) {
	if len(envs) == 0 {
		return PatchingMethodUndefined, nil
	}

	//
	//   load info about the urls that we received from the cluster
	//
	urlsInfo, err := loadUrlsInfo(urls)
	if err != nil {
		return PatchingMethodUndefined, err
	}

	// for backward compatibility, there are actually three methods of implementing this patching
	// Method 1: network status reported by the cluster
	//     variation a: connections already contained pre-processed information with nickname, service/port names or
	//		   url fragments. in this case we can use them for doing the patch.
	//     variation b: connections are not pre-processed. in this case we follow this method:
	//         exposedUrls -> are assumed to be received in the correct order (the composer takes care for that)
	//         ports -> are done by port numbers
	// Method 2: network status is not reported
	//      this is the old "naive" way, which we just copy 1:1 and hope for the best
	if networkStatus != nil && len(networkStatus.Services) != 0 {
		return matchConnectionsByNetworkStatus(ctx, envs, *urlsInfo, networkStatus)
	} else {
		return matchConnectionsNoNetworkStatus(ctx, envs, urls)
	}
}

// this is the old "naive" way. the match is done 1-by-1, assuming the cluster sends us
//
//	exposed-urls-of-environment-1
//	ports-of-environment-1
//	exposed-urls-of-environment-2
//	ports-of-environment-2
//	etc...
//
// it has two issues:
// - load balancers are not supported, as matching them requires the network status
// - less reliable than matching by network status
func matchConnectionsNoNetworkStatus(_ context.Context, envs []api.Environment, urls []string) (PatchingMethod, error) {
	counter := 0

	for _, env := range envs {
		for i := range env.Connections {
			// load balancer ports are included in the connections, but not in the urls. so
			// skip those urls when patching the urls into the connections
			if env.Connections[i].ConnectionType == TypeLoadBalancer {
				continue
			}

			responseConnection := env.Connections[i]
			if counter < len(urls) {
				responseConnection.Url = urls[counter]
				counter++
			} else {
				return PatchingMethodNaive, fmt.Errorf("urls count is lower than total count of connections")
			}
			env.Connections[i] = responseConnection
		}
	}

	if counter != len(urls) {
		return PatchingMethodNaive, fmt.Errorf("urls count is higher than total count of connections")
	}

	return PatchingMethodNaive, nil
}

func matchConnectionsByNetworkStatus(ctx context.Context, envs []api.Environment, urlsInfo UrlsInfo, networkStatus *api.WorkloadNetworkStatus) (PatchingMethod, error) {
	// now we load the network status info
	ni := loadNetworkStatusInfo(networkStatus)

	// first thing to do is to split the connections by type, because each type
	// is handled differently
	ci, preProcessed := loadConnectionsInfo(envs)
	if !preProcessed && ci.hasConnectionWithSingleServiceURL(ni) {
		preProcessed = true
	}
	if preProcessed {
		return matchUrlWithPreProcessingInfo(ctx, ci, urlsInfo, ni)
	} else {
		return matchUrlsNoPreprocessingDone(ctx, ci, urlsInfo, ni)
	}
}

func matchUrlsNoPreprocessingDone(ctx context.Context, connInfo ConnectionsInfo, urlsInfo UrlsInfo, netInfo NetworkStatusInfo) (PatchingMethod, error) {
	////////////////////////////////////////////////////////////////////////////////////////////
	// NOTE!!!!
	//
	// currently, the situation is that workload controller does not report urls for distributed
	// workloads. to my understanding, the builder of distributed does not even put the related
	// annotations. so up till now, distributed were not patched with urls at all. until this whole
	// thing is fixed, we either do the patching by nicknames (if we have them), otherwise - we better
	// handle distributed with the "old" method, so there is no change of behavior. once distributed
	// are taken care of in workload-controller, we'll be able to figure out what information we receive,
	// and adopt this function to parse them correctly.
	//
	// another point to consider: if workload has two replicas (master + worker), it has two different
	// containers. meaning: they both can have the same target port. this complicates matching by
	// port number. we might prefer to match by nicknames only for distributed.
	////////////////////////////////////////////////////////////////////////////////////////////
	if len(connInfo.envs) > 1 {
		return matchConnectionsNoNetworkStatus(ctx, connInfo.envs, urlsInfo.urls)
	}

	urlIdx := 0

	// urls we should copy 1:1 because the connections are originally ordered with the order of the urls
	// (order is by nickname)
	for _, perTypeConn := range connInfo.perTypeConn[TypeUrl] {
		if urlIdx >= len(urlsInfo.urls) {
			return PatchingMethodNetworkStatNoPreProc, fmt.Errorf("urls count is lower than total count of exposedUrl connections")
		}

		perTypeConn.Url = urlsInfo.urls[urlIdx]
		urlIdx++
	}

	// node ports, we take the url -> extract the assigned port -> find the correlated
	// target port -> load the connection of this target port and patch the URL
	for i := 0; i < len(connInfo.perTypeConn[TypeNodePort]); i++ {
		if urlIdx >= len(urlsInfo.urls) {
			return PatchingMethodNetworkStatNoPreProc, fmt.Errorf("urls count is lower than total count of nodePort connections")
		}

		rawUrl := urlsInfo.urls[urlIdx]
		urlIdx++

		_, extPortStr, err := net.SplitHostPort(rawUrl)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to split url %s", rawUrl)
			continue
		}

		targetPort, portFound := netInfo.extToContainerNodePort[extPortStr]
		if !portFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot find target port of %s", rawUrl)
		} else {
			connInfo.perNodePort[targetPort].Url = rawUrl
		}
	}

	// load balancers we patch from the network status map of urls
	for _, lbPort := range connInfo.perTypeConn[TypeLoadBalancer] {
		assignedUrl, portFound := netInfo.loadBalancerUrls[fmt.Sprintf("%d", lbPort.ContainerPort)]
		if !portFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot find url of load balancer port %d", lbPort.ContainerPort)
		} else {
			lbPort.Url = assignedUrl
		}
	}

	return PatchingMethodNetworkStatNoPreProc, nil
}

type ConnectionsInfo struct {
	envs            []api.Environment
	perTypeConn     map[string][]*api.Connection
	perNodePort     map[string]*api.Connection
	perLoadBalancer map[int32]*api.Connection
}

func loadConnectionsInfo(envs []api.Environment) (ConnectionsInfo, bool) {
	connectionsPreProcessed := false

	perTypeConn := make(map[string][]*api.Connection)

	// keep also mapping b/w target port and connection, for patching node ports and load balancers
	// this works under assumption that we are not patching distributed!
	perNodePort := make(map[string]*api.Connection)
	perLoadBalancer := make(map[int32]*api.Connection)

	for _, env := range envs {
		for idx, conn := range env.Connections {
			perTypeConn[conn.ConnectionType] = append(perTypeConn[conn.ConnectionType], &env.Connections[idx])

			if conn.ConnectionType == TypeNodePort {
				perNodePort[fmt.Sprintf("%d", conn.ContainerPort)] = &env.Connections[idx]
			}

			if conn.ConnectionType == TypeLoadBalancer {
				perLoadBalancer[conn.ContainerPort] = &env.Connections[idx]
			}

			if isConnectionPreProcessed(conn) {
				connectionsPreProcessed = true
			}
		}
	}

	return ConnectionsInfo{
		envs:            envs,
		perTypeConn:     perTypeConn,
		perNodePort:     perNodePort,
		perLoadBalancer: perLoadBalancer,
	}, connectionsPreProcessed
}

func isConnectionPreProcessed(conn api.Connection) bool {
	return convutils.IsStringHasValue(conn.ExposedUrlPath) || convutils.IsStringHasValue(conn.PortResourceName)
}

func (info ConnectionsInfo) hasConnectionWithSingleServiceURL(netInfo NetworkStatusInfo) bool {
	for _, conn := range info.perTypeConn[TypeUrl] {
		if conn.Url != "" {
			continue
		}
		if _, found := netInfo.singleServiceURL(conn.ServiceResourceName); found {
			return true
		}
	}
	return false
}

type ServiceAndPortName struct {
	ServiceName string
	PortName    string
}

type NetworkStatusInfo struct {
	// url assigned to each load balancer port (by target port!)
	loadBalancerUrls map[string]string

	// url assigned to each load balancer port by service + port name
	loadBalancerURLsByServicePort map[ServiceAndPortName]string

	// external URLs assigned to each service
	serviceUrls map[string][]string

	// NodePort URLs assigned to each service + port name
	nodePortURLs map[ServiceAndPortName]string

	// map between external port to container port for node ports
	extToContainerNodePort map[string]string

	// map between service + port names, to the external port
	svcAndPortNameToNodePort map[ServiceAndPortName]string
}

func loadNetworkStatusInfo(networkStatus *api.WorkloadNetworkStatus) NetworkStatusInfo {
	result := NetworkStatusInfo{
		loadBalancerUrls:              make(map[string]string),
		loadBalancerURLsByServicePort: make(map[ServiceAndPortName]string),
		serviceUrls:                   make(map[string][]string),
		nodePortURLs:                  make(map[ServiceAndPortName]string),
		extToContainerNodePort:        make(map[string]string),
		svcAndPortNameToNodePort:      make(map[ServiceAndPortName]string),
	}

	for _, service := range networkStatus.Services {
		for _, serviceURL := range service.Urls {
			if serviceURL.Url != "" {
				result.serviceUrls[service.Name] = append(result.serviceUrls[service.Name], serviceURL.Url)
			}
		}

		if loadBalancerHost := firstLoadBalancerHost(service.LoadBalancer.Ingress); loadBalancerHost != "" {
			for _, port := range service.Ports {
				assignedURL := fmt.Sprintf("%s:%d", loadBalancerHost, port.Port)
				result.loadBalancerUrls[port.TargetPort] = assignedURL
				result.loadBalancerURLsByServicePort[ServiceAndPortName{
					ServiceName: service.Name,
					PortName:    port.Name,
				}] = assignedURL
			}
		} else {
			for _, port := range service.Ports {
				result.extToContainerNodePort[fmt.Sprintf("%d", port.Port)] = port.TargetPort

				serviceAndPortName := ServiceAndPortName{
					ServiceName: service.Name,
					PortName:    port.Name,
				}
				result.svcAndPortNameToNodePort[serviceAndPortName] = fmt.Sprintf("%d", port.Port)
				if assignedURL := serviceURLByPort(service.Urls, port.Port); assignedURL != "" {
					result.nodePortURLs[serviceAndPortName] = assignedURL
				}
			}
		}
	}

	return result
}

func firstLoadBalancerHost(ingress []api.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner) string {
	for _, entry := range ingress {
		if entry.Ip != "" {
			return entry.Ip
		}
		if entry.Hostname != "" {
			return entry.Hostname
		}
	}
	return ""
}

func (info NetworkStatusInfo) singleServiceURL(serviceName *string) (string, bool) {
	if serviceName == nil || *serviceName == "" {
		return "", false
	}

	assignedURL := ""
	// Gateway dual-mode can report the same service URL from both Ingress and HTTPRoute.
	// Treat exact duplicates as one logical URL, but keep different URLs ambiguous.
	for _, candidate := range info.serviceUrls[*serviceName] {
		if candidate == "" {
			continue
		}
		if assignedURL == "" {
			assignedURL = candidate
			continue
		}
		if assignedURL != candidate {
			return "", false
		}
	}
	return assignedURL, assignedURL != ""
}

func serviceURLByPort(urls []api.WorkloadNetworkStatusServicesInnerUrlsInner, port int32) string {
	portStr := fmt.Sprintf("%d", port)
	for _, serviceURL := range urls {
		if serviceURL.Url == "" {
			continue
		}
		if urlPort(serviceURL.Url) == portStr {
			return serviceURL.Url
		}
	}
	return ""
}

func urlPort(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil && parsed.Host != "" {
		return parsed.Port()
	}

	hostPort := rawURL
	if idx := strings.IndexAny(hostPort, "/?"); idx >= 0 {
		hostPort = hostPort[:idx]
	}
	_, port, err := net.SplitHostPort(hostPort)
	if err == nil {
		return port
	}
	if strings.Count(hostPort, colon) == 1 {
		return hostPort[strings.LastIndex(hostPort, colon)+1:]
	}
	return ""
}

func matchUrlWithPreProcessingInfo(ctx context.Context, connInfo ConnectionsInfo, urls UrlsInfo, netInfo NetworkStatusInfo) (PatchingMethod, error) {
	for _, env := range connInfo.envs {
		if len(env.Connections) == 0 {
			continue
		}

		for idx, conn := range env.Connections {
			switch conn.ConnectionType {
			case TypeServingPort:
				// Inference workloads report the Knative serving URL directly on
				// the workload. An independently exposed URL can make the
				// connection set preprocessed, but it must not prevent the serving
				// connection from receiving that workload URL.
				if len(urls.urls) != 0 {
					// The Knative status producer appends the serving URL after the
					// optional job-URL annotation, so select the last reported URL.
					env.Connections[idx].Url = urls.urls[len(urls.urls)-1]
				}

			case TypeUrl:
				if conn.Url != "" {
					continue // custom url, no need to recalculate
				}
				if assignedURL, found := netInfo.singleServiceURL(conn.ServiceResourceName); found {
					env.Connections[idx].Url = assignedURL
					continue
				}
				if urls.domain == nil {
					continue
				}
				if urls.subdomainSupport && conn.ExposedUrlSubDomain != nil {
					copiedDomain := *urls.domain
					copiedDomain.Host = fmt.Sprintf("%s.%s", *conn.ExposedUrlSubDomain, copiedDomain.Host)
					env.Connections[idx].Url = copiedDomain.String()
				}
				if !urls.subdomainSupport && conn.ExposedUrlPath != nil {
					copiedDomain := *urls.domain
					copiedDomain.Path = *conn.ExposedUrlPath
					env.Connections[idx].Url = copiedDomain.String()
				}

			case TypeLoadBalancer:
				serviceAndPortName := ServiceAndPortName{
					ServiceName: convutils.Deref(conn.ServiceResourceName, ""),
					PortName:    convutils.Deref(conn.PortResourceName, ""),
				}
				if assignedURL, found := netInfo.loadBalancerURLsByServicePort[serviceAndPortName]; found {
					env.Connections[idx].Url = assignedURL
					continue
				}

				assignedURL, portFound := netInfo.loadBalancerUrls[fmt.Sprintf("%d", conn.ContainerPort)]
				if portFound {
					env.Connections[idx].Url = assignedURL
					continue
				}
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot find url of load balancer port %d", conn.ContainerPort)

			case TypeNodePort:
				serviceAndPortName := ServiceAndPortName{
					ServiceName: convutils.Deref(conn.ServiceResourceName, fmt.Sprintf("unknown-service-name-conn-%s", conn.CorrelationName)),
					PortName:    convutils.Deref(conn.PortResourceName, fmt.Sprintf("unknown-port-name-conn-%s", conn.CorrelationName)),
				}
				if assignedURL, found := netInfo.nodePortURLs[serviceAndPortName]; found {
					env.Connections[idx].Url = assignedURL
					continue
				}

				if urls.ipAddress == nil {
					continue
				}
				assignedPort, portFound := netInfo.svcAndPortNameToNodePort[serviceAndPortName]
				if !portFound {
					logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("cannot find url of node port %+v", serviceAndPortName)
					continue
				}

				if urls.ipAddress != nil {
					env.Connections[idx].Url = fmt.Sprintf("%s:%s", *urls.ipAddress, assignedPort)
				}
			}
		}
	}

	return PatchingMethodNetworkStatPreProcessed, nil
}

type UrlsInfo struct {
	urls []string

	ipAddress *string

	domain           *url.URL
	subdomainSupport bool
}

// in the list of urls there are two parts:
// exposed urls -> in format http://domain/path or http://subdomain.domain
// exposed ports -> in format 123.456.789.012:5678
func loadUrlsInfo(urls []string) (*UrlsInfo, error) {
	result := UrlsInfo{
		urls: urls,
	}

	if len(urls) == 0 {
		return &result, nil
	}

	// exposedUrls are always on top, so if first entry is not of exposedUrls then it indicates
	// that we do not have any urls in the list
	firstUrl := urls[0]
	if !unicode.IsDigit(rune(firstUrl[0])) {
		parsed, err := url.Parse(firstUrl)
		if err != nil {
			return nil, err
		}

		// determine subdomain or path. empty path, or path containing single slash is not considered a real path
		if len(parsed.Path) < 2 {
			result.subdomainSupport = true

			// expecting subdomain.domain.com, need to extract "domain.com"
			hostParts := strings.Split(parsed.Host, dot)
			if len(hostParts) < 2 {
				return nil, fmt.Errorf("url '%s' is malformed for cluster supporting subdomain", firstUrl)
			}

			parsed.Host = strings.Replace(parsed.Host, hostParts[0]+dot, "", 1 /*only the first match*/)
		}

		parsed.Path = "" // we only need the domain, no need for the path
		result.domain = parsed
	}

	// now look for the first entry having ip:port, load the ip address
	for _, aUrl := range urls {
		if unicode.IsDigit(rune(aUrl[0])) {
			result.ipAddress = convutils.Ref(strings.Split(aUrl, colon)[0])
			break
		}
	}

	return &result, nil
}
