/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
)

// both jupyter and tensorboard use the same default container port.
const (
	DefaultIngressHttpPort               = 80
	DefaultJupyterAndTensorContainerPort = int32(8888)
	maxHostPort                          = 32767
	minHostPort                          = 30000
	maxDNSLabelLength                    = 63
	dnsLabelHashLength                   = 10
)

var (
	serviceTypeMeta     = metav1.TypeMeta{APIVersion: "v1", Kind: "Service"}
	serviceGroupVersion = metav1.GroupVersionKind(serviceTypeMeta.GetObjectKind().GroupVersionKind())

	ingressTypeMeta     = metav1.TypeMeta{APIVersion: "v1", Kind: "Ingress"}
	ingressGroupVersion = metav1.GroupVersionKind(ingressTypeMeta.GetObjectKind().GroupVersionKind())

	routeTypeMeta     = metav1.TypeMeta{APIVersion: "route.openshift.io/v1", Kind: "Route"}
	RouteGroupVersion = metav1.GroupVersionKind(routeTypeMeta.GetObjectKind().GroupVersionKind())

	httpRouteTypeMeta     = metav1.TypeMeta{APIVersion: constants.GatewayV1, Kind: constants.KindHTTPRoute}
	HTTPRouteGroupVersion = metav1.GroupVersionKind(httpRouteTypeMeta.GetObjectKind().GroupVersionKind())

	grpcRouteTypeMeta     = metav1.TypeMeta{APIVersion: constants.GatewayV1, Kind: constants.KindGRPCRoute}
	GRPCRouteGroupVersion = metav1.GroupVersionKind(grpcRouteTypeMeta.GetObjectKind().GroupVersionKind())
)

// getServiceName returns the service name as "workloadInitials-workloadName-suffix"
func getServiceName(workloadInitials string, workloadName string, suffix string) string {
	return fmt.Sprintf("%s-%s-%s", workloadInitials, workloadName, suffix)
}

type ComposeUrlOptions struct {
	SubdomainSupported bool
}

func ComposeExposedUrlInt(ctx context.Context, suffix string, jobName string) (*apis.URL, error) {
	clusterDomain := config.GetClusterDomain()

	projectName, err := project.GetProjectFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("project name not found in context")
	}

	options := ComposeUrlOptions{
		SubdomainSupported: config.SubdomainSupportEnabled(),
	}

	return ComposeExposedUrl(suffix, jobName, projectName, clusterDomain, options)

}

// ComposeExposedUrl composes URL for ExposedUrl whose URL is not provided by the researcher.
func ComposeExposedUrl(suffix string, jobName string, projectName string, clusterDomain string, options ComposeUrlOptions) (*apis.URL, error) {
	parsedUrlPattern, err := apis.ParseURL(clusterDomain)
	if err != nil || parsedUrlPattern == nil {
		return nil, fmt.Errorf("invalid cluster domain '%s': %v", clusterDomain, err)
	}

	if options.SubdomainSupported {
		subdomain := fmt.Sprintf("%s-%s", projectName, jobName)
		if suffix != "" {
			subdomain = fmt.Sprintf("%s-%s", subdomain, suffix)
		}
		subdomain = shortenDNSLabel(subdomain)
		parsedUrlPattern.Host = fmt.Sprintf("%s.%s", subdomain, parsedUrlPattern.Host)
	} else {
		path := fmt.Sprintf("/%s/%s", projectName, jobName)
		if suffix != "" {
			path = fmt.Sprintf("%s/%s", path, suffix)
		}
		if parsedUrlPattern.Path != "" {
			path = fmt.Sprintf("%s/%s", path, parsedUrlPattern.Path)
		}
		parsedUrlPattern.Path = path
	}
	if parsedUrlPattern.Path == "" {
		parsedUrlPattern.Path = "/"
	}
	return parsedUrlPattern, nil
}

func shortenDNSLabel(label string) string {
	if len(label) <= maxDNSLabelLength {
		return label
	}

	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(label)))[:dnsLabelHashLength]
	return fmt.Sprintf("%s-%s", label[:maxDNSLabelLength-dnsLabelHashLength-1], digest)
}

// ComposeJupyterOrTensorboardJobUrl composes exposed URL for Jupyter/Tensorboard URL based on
func ComposeJupyterOrTensorboardJobUrl(ctx context.Context, jobName string) (*apis.URL, error) {
	return ComposeExposedUrlInt(ctx, "", jobName)
}

// ComposeWorkspaceUrl composes URL for accessing workspace directly. Sample URL:
// http://domain.com/workspaces?columnFilter=[{"term":"work-100","name":"name"}]&clusterId=5ee7956e-9e76-4244-9f7c-7fb1cedd38a0
func ComposeWorkspaceUrl(_ context.Context, workloadKind string, jobName string, clusterId string) (*apis.URL, error) {
	clusterDomain := config.GetClusterDomain()
	parsedUrlPattern, err := apis.ParseURL(clusterDomain)
	if err != nil || parsedUrlPattern == nil {
		return nil, fmt.Errorf("invalid cluster domain '%s': %v", clusterDomain, err)
	}
	if workloadKind == constants.KindInteractiveWorkload {
		parsedUrlPattern.Path = "workspaces"
	} else {
		parsedUrlPattern.Path = "trainings"
	}
	parsedUrlPattern.RawQuery = fmt.Sprintf("columnFilter=[{\"term\":\"%s\",\"name\":\"name\"}]&clusterId=%s",
		jobName, clusterId)
	return parsedUrlPattern, nil
}

func networkingV1TypeMeta(obj runtime.Object) metav1.TypeMeta {
	gvk := networkingv1.SchemeGroupVersion.WithKind(reflect.TypeOf(obj).Elem().Name())
	return metav1.TypeMeta{
		Kind:       gvk.Kind,
		APIVersion: gvk.GroupVersion().String(),
	}
}

func AddJobUrlToAnnotations(resource *metav1.ObjectMeta, jobUrlToAdd string) {
	if resource.Annotations == nil {
		resource.Annotations = map[string]string{}
	}

	if config.IsOpenshift() && config.SubdomainSupportEnabled() {
		// OpenShift host-based routing mechanism already supports wildcard certificate authorization.
		// We just need to trim the 'runai' str from our subdomain routes.
		jobUrlToAdd = strings.Replace(jobUrlToAdd, config.RunaiTrimStr, "", 1)
	}

	jobUrlToAdd = strings.TrimSuffix(jobUrlToAdd, "/")
	if jobUrls, ok := resource.Annotations[constants.AnnotationJobUrl]; ok {
		for _, jobUrl := range strings.Split(resource.Annotations[constants.AnnotationJobUrl], ", ") {
			if jobUrl == jobUrlToAdd {
				return // prevent duplicate URLs
			}
		}
		resource.Annotations[constants.AnnotationJobUrl] = jobUrls + ", " + jobUrlToAdd
	} else {
		resource.Annotations[constants.AnnotationJobUrl] = jobUrlToAdd
	}
}

func UrlToPort(url *apis.URL) (int32, error) {
	if url == nil {
		return DefaultIngressHttpPort, nil
	}
	port := url.URL().Port()
	if port == "" {
		return DefaultIngressHttpPort, nil
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return DefaultIngressHttpPort, err
	}
	return int32(portNum), nil
}
