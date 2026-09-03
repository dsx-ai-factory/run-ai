/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package itemized

import (
	"fmt"
	workloadApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"regexp"
	"strings"
)

// ItemizedNumbering is responsible for composing nicknames for itemized fields
// It is important that the nickname will be consistent b/w policy and workloads,
// and in policy - between cluster and control plane policy, that's why we're
// doing it in one single point of truth.
// in general, if an itemized instance has a given name, this name will be used
// otherwise, we assign it a nickname with prefix + increasing number
// for example
//
//	hostPath:
//	   - name: hp1   -->  nickname host-path-hp1
//	     path: /path/hp1
//	     mountPath: /mount/path/hp1
//	   - path: /path/noname1  --> nickname host-path-1 in workload, or host-path-pol-1 in policy
//	     mountPath: /mount/path/noname
//	   - path: /path/noname2  --> nickname host-path-2 in workload, or host-path-pol-2 in policy
//	     mountPath: /mount/path/noname
//	   - name: hp2   --> nickname host-path-hp2
//	     path: /path/hp2
//	     mountPath: /mount/path/hp2
type ItemizedNumbering map[string]int

const (
	PolicyNamingIndicator   = "pol"
	TemplateNamingIndicator = "templt"
	dummyNamingIndicator    = "dummy"
)

// Compile the regex to match anything that is not a letter or digit
var reNormalize = regexp.MustCompile(`[^a-zA-Z0-9]`)

func (in *ItemizedNumbering) Compose(prefix string, key *string, replicaType *workloadApi.ReplicaType, forPolicy bool) string {
	instanceIndex, prefix := in.getIndexOf(prefix, replicaType)

	if key != nil {
		// normalize the key to have letters and digits only
		normalizedKey := reNormalize.ReplaceAllString(*key, "-")
		return fmt.Sprintf("%s-%s", prefix, normalizedKey)
	} else if forPolicy {
		return fmt.Sprintf("%s-%s-%d", prefix, PolicyNamingIndicator, instanceIndex)
	} else if instanceIndex != 0 {
		return fmt.Sprintf("%s-%d", prefix, instanceIndex)
	} else {
		return prefix
	}
}

func (in *ItemizedNumbering) DefaultPolicyNick(namingIndicator string, prefix string, replicaType *workloadApi.ReplicaType) string {
	instanceIndex, _ := in.getIndexOf(prefix, replicaType)
	return fmt.Sprintf("%s-%d", namingIndicator, instanceIndex)
}

// DummyNick composes nickname for workload instances that do not have key. we do not care much
// about those nicknames, so prefix then with dummy-
func (in *ItemizedNumbering) DummyNick(prefix string) string {
	instanceIndex, _ := in.getIndexOf(prefix, nil)
	return fmt.Sprintf("%s-%d", dummyNamingIndicator, instanceIndex)
}

func (in *ItemizedNumbering) getIndexOf(prefix string, replicaType *workloadApi.ReplicaType) (int, string) {
	if replicaType != nil {
		prefix += fmt.Sprintf("-%s", strings.ToLower(string(*replicaType)))
	}

	indexOfInstance, prefixAlreadyIn := (*in)[prefix]
	if prefixAlreadyIn {
		indexOfInstance += 1
	}

	(*in)[prefix] = indexOfInstance

	return indexOfInstance, prefix
}
