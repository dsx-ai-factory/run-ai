/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

const (
	UrlAlreadyExistsMsg = "admission webhook \"validate.nginx.ingress.kubernetes.io\" denied the request: host <HOST> and path <PATH> is already defined in ingress %s"
)

// FakeClientWithIngressUrlErrors tracks ingress creation and returns error if other ingress with same url already exist
type FakeClientWithIngressUrlErrors struct {
	client.Client
	existingIngressUrls sync.Map // [<url string>]<ingress name>
}

func NewFakeClientWithIngressUrlErrors(scheme *runtime.Scheme, statusObjects []client.Object, objects ...client.Object) *FakeClientWithIngressUrlErrors {
	return &FakeClientWithIngressUrlErrors{Client: InitFakeClientWithObjectWithStatus(scheme, statusObjects, objects...)}
}

func (c *FakeClientWithIngressUrlErrors) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if ingress, ok := obj.(*networkingv1.Ingress); ok {
		urlIngress := getUrlFromIngress(ingress)
		if existingIngress, exists := c.existingIngressUrls.Load(urlIngress); exists {
			return &k8serrors.StatusError{ErrStatus: metav1.Status{Message: fmt.Sprintf(UrlAlreadyExistsMsg, existingIngress), Code: http.StatusBadRequest}}
		}
		c.existingIngressUrls.Store(urlIngress, fmt.Sprintf("%s/%s", ingress.Namespace, ingress.Name))
	}

	return c.Client.Create(ctx, obj, opts...)
}

func getUrlFromIngress(ingress *networkingv1.Ingress) string {
	if len(ingress.Spec.Rules) == 0 || len(ingress.Spec.Rules[0].HTTP.Paths) == 0 {
		return ""
	}

	urlStr, _ := url.JoinPath(ingress.Spec.Rules[0].Host, ingress.Spec.Rules[0].HTTP.Paths[0].Path)
	return urlStr
}
