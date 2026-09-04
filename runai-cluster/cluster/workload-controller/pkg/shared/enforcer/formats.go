/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enforcer

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"

	docker "github.com/novln/docker-parser"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/third_party/validation"
	"k8s.io/apimachinery/pkg/api/resource"
	k8svalidation "k8s.io/apimachinery/pkg/util/validation"
)

// validate a value matches a format
func enforceFormat(format string, value interface{}) (bool, error) {

	strValue := fmt.Sprintf("%v", value)

	errs := make([]string, 0)
	var err error

	switch format {
	case workload_fields.FormatEnvVar:
		errs = k8svalidation.IsEnvVarName(strValue)
	case workload_fields.FormatShellCommand:
		err = validation.Validate(value, validation.By(validateShellCommand))
	case workload_fields.FormatPassword:
		// this is a frontend format, no need to validate
	case workload_fields.FormatPort:
		// this is frontend format, in backend it is premised to be integer and runai:rules enforce the range
	case workload_fields.FormatKubeName:
		errs = k8svalidation.IsDNS1035Label(strValue)
	case workload_fields.FormatMemory:
		if strValue != "" {
			_, err = resource.ParseQuantity(strValue)
		}
	case workload_fields.FormatGpuMemory:
		err = validation.Validate(value, validation.By(validateGpuMemory))
	case workload_fields.FormatRequestUrl:
		err = validation.Validate(value, validation.RequestURL)
	case workload_fields.FormatImageName:
		_, err = docker.Parse(strValue)
	case workload_fields.FormatDuration:
		_, err = time.ParseDuration(strValue)
	case workload_fields.FormatGroups:
		err = validation.Validate(value, validation.By(validateGroups))
	case workload_fields.FormatTopologyKey:
		err = validation.Validate(value, validation.By(validateTopologyKey))
	}

	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) != 0 {
		return false, errors.New(strings.Join(errs, "; "))
	}

	return true, nil
}

func validateTopologyKey(value interface{}) error {

	const topologyKeyFmt = "([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]((\\/)(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]))?"
	var topologyKeyRegexp = regexp.MustCompile("^" + topologyKeyFmt + "$")

	valueStr := fmt.Sprintf("%v", value)
	if !topologyKeyRegexp.MatchString(valueStr) {
		return fmt.Errorf("%s is not a valid topology key: must consist of alphanumeric characters, '-', '_', '.' or '/' allowed only once, and must start and end with an alphanumeric character", value)
	}

	return nil
}

func validateGroups(value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	valueStr = strings.ReplaceAll(valueStr, ",", ";") // we support both , and ;
	results, err := utils.DecodeInt64List(valueStr, common.GroupsSeparator)
	if err != nil {
		return err
	}
	for _, result := range results {
		if result < 0 {
			return fmt.Errorf("'%d' is not a valid gid: should be no less than zero", result)
		}
	}
	return nil
}

func validateShellCommand(value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	_, err := workloadbuilders.ParseCommandLine(valueStr)
	if err != nil {
		return fmt.Errorf("'%s' cannot be used to form a valid command: %v", valueStr, err)
	}
	return nil
}

func validateGpuMemory(value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	_, err := resource.ParseQuantity(valueStr)
	return err
}
