/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"bytes"
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	"github.com/ghodss/yaml"
	openshiftV1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SharedResources holds resources which are common to all workloads
type SharedResources struct {
	Pvcs     []corev1.PersistentVolumeClaim
	Services []corev1.Service
	Secrets  []corev1.Secret

	//
	//    following are needed for jupyter jobs. route is used in openshift, ingress in all
	//    the other environments
	//
	Ingress []networkingv1.Ingress
	Route   []openshiftV1.Route
}

var fixedDate = metav1.NewTime(time.Date(2022, 2, 22, 17, 0, 0, 0, time.UTC))
var yamlSeparator = []byte("---\n")

func (self *OldSharedBuilder) ToYaml(_ context.Context, result *bytes.Buffer) error {

	for _, pvc := range self.SharedResources.Pvcs {
		pvc.Kind = constants.KindPvc
		pvc.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(pvc)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, service := range self.SharedResources.Services {
		service.Kind = constants.KindService
		service.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(service)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, ingress := range self.SharedResources.Ingress {
		ingress.Kind = constants.KindIngress
		ingress.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(ingress)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, route := range self.SharedResources.Route {
		route.Kind = constants.KindRoute
		route.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(route)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	for _, secret := range self.SharedResources.Secrets {
		secret.Kind = constants.KindSecret
		secret.CreationTimestamp = fixedDate
		marshalled, err := yaml.Marshal(secret)
		if err != nil {
			return err
		}
		result.Write(yamlSeparator)
		result.Write(marshalled)
	}

	return nil
}
