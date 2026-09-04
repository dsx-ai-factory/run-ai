//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"os"
	"strings"
)

const (
	NotificationHost = "notifications-service"
	// NotificationHost = "localhost"
	NotificationPort = "5000"
	ConfigPort       = "8085"
	SmtpHost         = "smtp"
	// SmtpHost    = "localhost"
	SmtpApiPort = 8025
	SmtpPort    = 1025
	ListenerId  = "http-runai"
	TenantId    = "tenant1"
)

var (
	notificationHost = NotificationHost
	smtpHost         = SmtpHost
)

func FindFirstInArray[V interface{}](arr []V, predicate func(V) bool) *V {
	for _, element := range arr {
		if predicate(element) {
			return &element
		}
	}
	return nil
}

func FindAllInArray[V interface{}](arr []V, predicate func(V) bool) []*V {
	elements := []*V{}
	for _, element := range arr {
		if predicate(element) {
			elements = append(elements, &element)
		}
	}
	return elements
}

func GetEnv(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Trim(value, " ")
	}
	return defaultVal
}
