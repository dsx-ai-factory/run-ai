/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

type NotifyBuilderer interface {
	WithTitle(title string) NotifyBuilderer
	WithDescription(description string) NotifyBuilderer
	WithTenant(tenant string) NotifyBuilderer
	WithUser(user string) NotifyBuilderer
	WithProperty(label string, val interface{}) NotifyBuilderer
	Notify()
}

type NotifyBuilder struct {
	Notification *Notification
}

func NewNotifyBuilder(category string, typ string) *NotifyBuilder {
	return &NotifyBuilder{
		Notification: &Notification{
			Category:             category,
			Type:                 typ,
			AdditionalProperties: map[string]interface{}{},
		},
	}
}

func (n *NotifyBuilder) WithTitle(title string) NotifyBuilderer {
	n.Notification.Title = title
	return n
}

func (n *NotifyBuilder) WithDescription(description string) NotifyBuilderer {
	n.Notification.Description = description
	return n
}

func (n *NotifyBuilder) WithTenant(tenant string) NotifyBuilderer {
	n.Notification.Tenant = tenant
	return n
}

func (n *NotifyBuilder) WithUser(user string) NotifyBuilderer {
	n.Notification.User = user
	return n
}

func (n *NotifyBuilder) WithProperty(property string, val interface{}) NotifyBuilderer {
	n.Notification.AdditionalProperties[property] = val
	return n
}

func (n *NotifyBuilder) Notify() {
	if globalClient == nil {
		return
	}

	globalClient.Notify(n.Notification)
}
