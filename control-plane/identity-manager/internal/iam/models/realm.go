/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"strconv"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils/pointer"
)

var (
	EnabledEventTypes = []string{"LOGIN", "CLIENT_LOGIN", "UPDATE_CREDENTIAL"}
	EventsListeners   = []string{"jboss-logging", "last-login-event", "password-audit-event"}
	EventsExpiration  = int64(90 * 24 * 60 * 60)
	DefaultTheme      = "runai-theme"
)

const TenantIdRealmAttribute = "runai.tenant.id"

type RealmDefaults struct {
	Theme         *string
	SmtpHost      string
	SmtpPort      int
	SmtpFromEmail string
}

type UserProfileAttributePermissions struct {
	View []string `json:"view"`
	Edit []string `json:"edit"`
}

type UserProfileAttributeMetadata struct {
	Name        *string                         `json:"name"`
	DisplayName *string                         `json:"displayName"`
	Multivalued *bool                           `json:"multivalued"`
	Required    *interface{}                    `json:"required"`
	Annotations *interface{}                    `json:"annotations"`
	Validations *interface{}                    `json:"validations"`
	Permissions UserProfileAttributePermissions `json:"permissions"`
}

type UserProfileMetadata struct {
	Attributes []UserProfileAttributeMetadata `json:"attributes"`
	Groups     []interface{}                  `json:"groups"`
}

type Realm struct {
	ID                   *string
	Name                 *string
	Enabled              *bool
	RememberMe           *bool
	ResetPasswordAllowed *bool
	EnabledEventTypes    *[]string
	EventsListeners      *[]string
	EventsEnabled        *bool
	EventsExpiration     *int64
	LoginTheme           *string
	EmailTheme           *string
	SMTPServer           *map[string]string
	TenantId             *string
}

func NewRunaiRealm(name string, defaults *RealmDefaults) *Realm {
	r := &Realm{
		Name: &name, LoginTheme: defaults.Theme, EmailTheme: defaults.Theme,
	}
	r.SetRealmDefaults(defaults)
	return r
}

func (r *Realm) SetRealmDefaults(defaults *RealmDefaults) {
	r.Enabled = pointer.ToBool(true)
	r.RememberMe = pointer.ToBool(true)
	r.ResetPasswordAllowed = pointer.ToBool(true)
	r.EnabledEventTypes = &EnabledEventTypes
	r.EventsListeners = &EventsListeners
	r.EventsEnabled = pointer.ToBool(true)
	r.EventsExpiration = &EventsExpiration

	if defaults.Theme != nil && *defaults.Theme != "" {
		r.LoginTheme = defaults.Theme
		r.EmailTheme = defaults.Theme
	} else {
		if r.LoginTheme == nil || *r.LoginTheme == "" {
			r.LoginTheme = pointer.ToString(DefaultTheme)
		}
		if r.EmailTheme == nil || *r.EmailTheme == "" {
			r.EmailTheme = pointer.ToString(DefaultTheme)
		}
	}

	smtpSettings := map[string]string{
		"host":     defaults.SmtpHost,
		"port":     strconv.Itoa(defaults.SmtpPort),
		"starttls": "false",
		"auth":     "false",
		"ssl":      "false",
		"from":     defaults.SmtpFromEmail,
	}
	r.SMTPServer = &smtpSettings
}
