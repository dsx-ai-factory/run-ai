/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	sdkglobal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	commonutils "github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	corev1 "k8s.io/api/core/v1"
)

const (
	UsernameLoaderInitContainerName = "runai-auth-user-loader"
	// UsernameLoaderRunAsUser is the nobody UID used so the loader works on
	// clusters that enforce runAsNonRoot without providing a runAsUser (e.g. Kyverno).
	// ubi-minimal defaults to root, so an explicit UID is required.
	UsernameLoaderRunAsUser = int64(65534)
)

// UsernameLoaderInitContainerResources returns configured resource requirements,
// preserving the legacy defaults only when configuration is unset.
func UsernameLoaderInitContainerResources() (corev1.ResourceRequirements, error) {
	return workloadconfig.UsernameLoaderResources(sdkglobal.DefaultUsernameLoaderInitContainerResources())
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
