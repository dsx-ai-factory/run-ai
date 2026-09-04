/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	UsernameLoaderInitContainerName = "runai-auth-user-loader"
	// UsernameLoaderRunAsUser is the nobody UID used so the loader works on
	// clusters that enforce runAsNonRoot without providing a runAsUser (e.g. Kyverno).
	// ubi-minimal defaults to root, so an explicit UID is required.
	UsernameLoaderRunAsUser = int64(65534)
)

// DefaultUsernameLoaderInitContainerResources returns the legacy resource requirements.
func DefaultUsernameLoaderInitContainerResources() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("50m"),
			corev1.ResourceMemory: resource.MustParse("50Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("100Mi"),
		},
	}
}

// UsernameLoaderInitContainerResources returns configured resource requirements,
// preserving the legacy defaults only when configuration is unset.
func UsernameLoaderInitContainerResources() (corev1.ResourceRequirements, error) {
	return workloadconfig.UsernameLoaderResources(DefaultUsernameLoaderInitContainerResources())
}

// UsernameLoaderSecurityContext returns a non-root security context for the
// auth username loader init container. On vanilla Kubernetes, an explicit
// runAsUser is required because the default ubi-minimal image runs as root.
// On OpenShift, the SCC assigns a UID from the namespace's allowed range.
func UsernameLoaderSecurityContext() *corev1.SecurityContext {
	securityContext := &corev1.SecurityContext{
		AllowPrivilegeEscalation: commonutils.Ref(false),
		RunAsNonRoot:             commonutils.Ref(true),
	}
	if !workloadconfig.IsOpenshift() {
		securityContext.RunAsUser = commonutils.Ref(UsernameLoaderRunAsUser)
	}
	return securityContext
}

// BuildUsernameLoaderInitContainer creates a simple init container that loads authorized users to the auth volume
// This is a simplified version for use cases with a single set of authorized users
//
// Deprecated: use BuildUsernameLoaderInitContainerWithError. This compatibility
// wrapper panics when injected-container resource configuration is invalid.
func BuildUsernameLoaderInitContainer(authorizedUsers string, nick string) *corev1.Container {
	container, err := BuildUsernameLoaderInitContainerWithError(authorizedUsers, nick)
	if err != nil {
		panic(err)
	}
	return container
}

// BuildUsernameLoaderInitContainerWithError creates a username-loader init
// container and returns resource configuration errors to the caller.
func BuildUsernameLoaderInitContainerWithError(authorizedUsers string, nick string) (*corev1.Container, error) {
	if authorizedUsers == "" {
		return nil, nil
	}

	encodedUsers := CleanAndBase64Encode(authorizedUsers)
	filename := GetSafeFileNameFromNick(nick)
	command := fmt.Sprintf("echo -e \"%s\" | base64 -d > %s/%s", encodedUsers, constants.ExternalUrlAuthVolumePath, filename)
	resources, err := UsernameLoaderInitContainerResources()
	if err != nil {
		return nil, fmt.Errorf("get username-loader resources: %w", err)
	}

	return &corev1.Container{
		Name:            UsernameLoaderInitContainerName,
		Image:           workloadconfig.GetOauth2ProxyUsernameLoaderImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"sh", "-c", command},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      constants.ExternalUrlAuthVolumeName,
				MountPath: constants.ExternalUrlAuthVolumePath,
			},
		},
		Resources:       resources,
		SecurityContext: UsernameLoaderSecurityContext(),
	}, nil
}

func CleanAndBase64Encode(emailsSeparatedByComa string) string {
	// Replace commas with newlines and remove empty strings
	lines := strings.Split(emailsSeparatedByComa, ",")
	var filteredLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filteredLines = append(filteredLines, line)
		}
	}
	result := strings.Join(filteredLines, "\n")

	// Get base64 encoding of the resulting string
	base64Encoded := base64.StdEncoding.EncodeToString([]byte(result))

	return base64Encoded
}

func GetSafeFileNameFromNick(nick string) string {
	// from security perspective, we want to make sure that the nick is not used as a file name
	// so we encode it to base64
	return base64.StdEncoding.EncodeToString([]byte(nick))
}

func GetSortedNicknames(exposedUrls map[string]field_types.ExposedUrlItemAndRules) []string {
	// Sort Exposed URLs by Nickname (Key)
	sortedNicknames := make([]string, 0, len(exposedUrls))
	for nick := range exposedUrls {
		sortedNicknames = append(sortedNicknames, nick)
	}
	sort.Strings(sortedNicknames)
	return sortedNicknames
}

// GetUsernameLoaderInitContainer creates an init container that loads the emails to the auth volume,
// to prevent potential security issues with special characters in the email (e.g. of putting a user-input directly in the command line!)
// it decodes the emails that were encoded in base64 format
//
// Deprecated: use GetUsernameLoaderInitContainerWithError. This compatibility
// wrapper panics when injected-container resource configuration is invalid.
func GetUsernameLoaderInitContainer(exposedUrls map[string]field_types.ExposedUrlItemAndRules, servingPortAuthorizedUsers *string) *corev1.Container {
	container, err := GetUsernameLoaderInitContainerWithError(exposedUrls, servingPortAuthorizedUsers)
	if err != nil {
		panic(err)
	}
	return container
}

// GetUsernameLoaderInitContainerWithError composes a username-loader init
// container and returns resource configuration errors to the caller.
func GetUsernameLoaderInitContainerWithError(exposedUrls map[string]field_types.ExposedUrlItemAndRules, servingPortAuthorizedUsers *string) (*corev1.Container, error) {
	if !workloadconfig.ExternalAuthUrlEnabled() {
		return nil, nil
	}

	sortedNicknames := GetSortedNicknames(exposedUrls)

	if servingPortAuthorizedUsers != nil {
		sortedNicknames = append(sortedNicknames, "serving-port")
		sort.Strings(sortedNicknames)
	}

	command := ""
	for _, nick := range sortedNicknames {
		var authorizedUsers string
		if nick == "serving-port" {
			authorizedUsers = *servingPortAuthorizedUsers
		} else {
			entry := exposedUrls[nick]
			if entry.Value != nil && entry.Value.AuthorizedUsers != nil {
				authorizedUsers = *exposedUrls[nick].Value.AuthorizedUsers
			}
		}

		if authorizedUsers != "" {
			if command != "" {
				command += " && "
			}
			command = command + fmt.Sprintf("echo -e \"%s\" | base64 -d > /etc/runai/auth/users/%s",
				CleanAndBase64Encode(authorizedUsers), GetSafeFileNameFromNick(nick))
		}
	}

	if command == "" {
		return nil, nil
	}

	resources, err := UsernameLoaderInitContainerResources()
	if err != nil {
		return nil, fmt.Errorf("get username-loader resources: %w", err)
	}

	return &corev1.Container{
		Name:            UsernameLoaderInitContainerName,
		Image:           workloadconfig.GetOauth2ProxyUsernameLoaderImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"sh", "-c", command},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      constants.ExternalUrlAuthVolumeName,
				MountPath: constants.ExternalUrlAuthVolumePath,
			},
		},
		Resources:       resources,
		SecurityContext: UsernameLoaderSecurityContext(),
	}, nil
}
