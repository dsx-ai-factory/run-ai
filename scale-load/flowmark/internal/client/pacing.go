/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"net/http"

	"golang.org/x/time/rate"
)

type Limits struct {
	KubeRequestsPerSecond         int
	ControlPlaneRequestsPerSecond int
}

var DefaultLimits = Limits{KubeRequestsPerSecond: 50, ControlPlaneRequestsPerSecond: 25}

const burstSeconds = 4

func burstOf(requestsPerSecond int) int {
	return requestsPerSecond * burstSeconds
}

type pacedTransport struct {
	limiter *rate.Limiter
	next    http.RoundTripper
}

func paced(client *http.Client, requestsPerSecond int) *http.Client {
	client.Transport = &pacedTransport{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burstOf(requestsPerSecond)),
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
