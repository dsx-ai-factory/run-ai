/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package grafana

type Instance struct {
	Id     int    `json:"id"`
	UserId int    `json:"hmInstancePromId"`
	Slug   string `json:"slug"`
	Region string `json:"regionSlug"`
	Url    string `json:"hmInstancePromUrl"`
}

type Policy struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Token struct {
	Id    string `json:"id"`
	Token string `json:"token"`
}

type CreateInstanceBody struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Url  string `json:"url"`
}

type AccessPolicyRealm struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type CreateAccessPolicyBody struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"displayName"`
	Scopes      []string            `json:"scopes"`
	Realms      []AccessPolicyRealm `json:"realms"`
}

type CreateTokenBody struct {
	AccessPolicyId string `json:"accessPolicyId"`
	Name           string `json:"name"`
	DisplayName    string `json:"displayName"`
}
