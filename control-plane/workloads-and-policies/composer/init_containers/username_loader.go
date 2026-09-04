/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package init_containers

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	corev1 "k8s.io/api/core/v1"
)

const (
	usernameLoaderContainerName = "runai-auth-user-loader"
	usernameLoaderServingNick   = "serving-port"
	// usernameLoaderRunAsUser is the nobody UID used so the loader works on
	// clusters that enforce runAsNonRoot without providing a runAsUser (e.g. Kyverno).
	// ubi-minimal defaults to root, so an explicit UID is required.
	usernameLoaderRunAsUser = int64(65534)
)

func BuildUsernameLoaderInitContainer(info *composer.ExposedUrlInfo) (*corev1.Container, *workloadsApi.Error) {
	if info == nil || info.Config == nil {
		return nil, nil
	}

	externalAuthEnabled := info.Config.ExternalAuthUrlEnabled()
	nicks := make([]string, 0, len(info.Items)+1)
	for nick, item := range info.Items {
		if item == nil {
			continue
		}
		if !externalAuthEnabled && !item.BuildWhenExternalAuthURLDisabled {
			continue
		}
		nicks = append(nicks, nick)
	}
	if externalAuthEnabled && info.ServingPortAuthorizedUsers != nil {
		nicks = append(nicks, usernameLoaderServingNick)
	}
	sort.Strings(nicks)

	command := ""
	for _, nick := range nicks {
		authorizedUsers := usernameLoaderAuthorizedUsers(info, nick)
		if authorizedUsers == nil || *authorizedUsers == "" {
			continue
		}

		if command != "" {
			command += " && "
		}
		command += fmt.Sprintf("echo -e \"%s\" | base64 -d > %s/%s",
			cleanAndBase64Encode(*authorizedUsers), oauth2AuthVolumePath, oauth2SafeFileNameFromNick(nick))
	}

	if command == "" {
		return nil, nil
	}

	image := info.Config.OAuth2ProxyUserLoaderImage()
	if image == "" {
		return nil, composer.ErrClusterValueMissing("runaiOauth2ProxyUsernameLoaderImage")
	}

	resources, err := info.Config.UsernameLoaderResources()
	if err != nil {
		return nil, composer.ErrClusterValueInvalid("usernameLoaderResources", err)
	}

	return &corev1.Container{
		Name:            usernameLoaderContainerName,
		Image:           image,
		ImagePullPolicy: corev1.PullIfNotPresent,
		Command:         []string{"sh", "-c", command},
		VolumeMounts: []corev1.VolumeMount{
			{Name: oauth2AuthVolumeName, MountPath: oauth2AuthVolumePath},
		},
		Resources:       resources,
		SecurityContext: usernameLoaderSecurityContext(info.Config.IsOpenshift()),
	}, nil
}

func usernameLoaderAuthorizedUsers(info *composer.ExposedUrlInfo, nick string) *string {
	if nick == usernameLoaderServingNick && info.ServingPortAuthorizedUsers != nil {
		return info.ServingPortAuthorizedUsers
	}
	if item := info.Items[nick]; item != nil {
		return item.AuthorizedUsers
	}
	return nil
}

// usernameLoaderSecurityContext returns a non-root security context for the
// auth username loader init container. On vanilla Kubernetes, an explicit
// runAsUser is required because the default ubi-minimal image runs as root.
// On OpenShift, the SCC assigns a UID from the namespace's allowed range.
func usernameLoaderSecurityContext(isOpenShift bool) *corev1.SecurityContext {
	securityContext := &corev1.SecurityContext{
		AllowPrivilegeEscalation: convutils.Ref(false),
		RunAsNonRoot:             convutils.Ref(true),
	}
	if !isOpenShift {
		securityContext.RunAsUser = convutils.Ref(usernameLoaderRunAsUser)
	}
	return securityContext
}

func cleanAndBase64Encode(emailsSeparatedByComma string) string {
	lines := strings.Split(emailsSeparatedByComma, ",")
	filteredLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filteredLines = append(filteredLines, line)
		}
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(filteredLines, "\n")))
}
