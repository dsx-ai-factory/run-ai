/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"net/http"

	"golang.org/x/time/rate"
)

const (
	controlPlaneQPS   = 25
	controlPlaneBurst = 50
)

type pacedTransport struct {
	limiter *rate.Limiter
	next    http.RoundTripper
}

func paced(client *http.Client) *http.Client {
	client.Transport = &pacedTransport{
		limiter: rate.NewLimiter(rate.Limit(controlPlaneQPS), controlPlaneBurst),
		next:    client.Transport,
	}
	return client
}

func (t *pacedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if err := t.limiter.Wait(request.Context()); err != nil {
		return nil, err
	}
	return t.next.RoundTrip(request)
}
