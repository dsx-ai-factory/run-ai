/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalurl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	ColonDelimiter = ":"
)

func NewExternalUrlList(portList []string) ([]openapi.ExposedUrl, error) {
	var portObjList = make([]openapi.ExposedUrl, 0)
	for _, port := range portList {
		portObj, err := NewExternalUrl(port)
		if err != nil {
			return nil, err
		}
		portObjList = append(portObjList, *portObj)
	}
	return portObjList, nil
}

func NewExternalUrl(eurl string) (*openapi.ExposedUrl, error) {
	euParser := NewParser(eurl)
	err := euParser.Parse()
	if err != nil {
		return nil, err
	}

	eu, err := createExternalUrl(euParser)
	if err != nil {
		return nil, err
	}

	return setEUrlObj(eu, euParser)
}

func createExternalUrl(gp *parser.AttributeMapper) (*openapi.ExposedUrl, error) {
	var container int64
	err := parser.CheckMandatoryFields(gp, Container)
	if err != nil {
		return nil, err
	}

	container, err = strconv.ParseInt(*gp.Get(Container), 10, 32)
	if err != nil {
		return nil, err
	}
	return &openapi.ExposedUrl{
		Container: *openapi.NewNullableInt32(openapi.PtrInt32(int32(container))),
	}, nil
}

func setEUrlObj(eUrlObj *openapi.ExposedUrl, eu *parser.AttributeMapper) (*openapi.ExposedUrl, error) {
	if eu.Get(Url) != nil {
		eUrlObj.SetUrl(*eu.Get(Url))
	}
	if eu.Get(AuthUsers) != nil {
		users, err := getAuthList(*eu.Get(AuthUsers))
		if err != nil {
			return nil, err
		}
		eUrlObj.SetAuthorizedUsers(users)
	}

	if eu.Get(AuthGroups) != nil {
		groups, err := getAuthList(*eu.Get(AuthGroups))
		if err != nil {
			return nil, err
		}
		eUrlObj.SetAuthorizedGroups(groups)
	}

	return eUrlObj, nil
}

func getAuthList(authStr string) ([]string, error) {
	var split []string
	authSplit := strings.Split(authStr, ColonDelimiter)
	for _, auth := range authSplit {
		auth = strings.TrimSpace(auth)
		if auth == specset.EmptyString {
			return nil, fmt.Errorf(consts.ErrorAttributeEmptyValue, authStr)
		}
		split = append(split, auth)
	}
	return split, nil
}
