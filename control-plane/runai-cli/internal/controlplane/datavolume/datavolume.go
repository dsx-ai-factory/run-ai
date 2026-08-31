/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolume

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datavolume/client"
	dvlist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datavolume/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var listClient dvlist.Lister

func SetClient(lc dvlist.Lister) {
	listClient = lc
}

func GetNameToIdMap(txn *transaction.Txn, domain, token, projectId string, names []string) (map[string]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	dvClient, err := getClient(domain, token)
	if err != nil {
		return nil, err
	}

	list, err := dvClient.List(txn.GetCtx(), &projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to list datavolumes: %w", err)
	}

	nameToIdMap := make(map[string]string)
	for _, dv := range list.Datavolumes {
		for _, requestedName := range names {
			if dv.GetName() == requestedName {
				nameToIdMap[requestedName] = dv.GetId()
				break
			}
		}
	}

	for _, name := range names {
		if _, found := nameToIdMap[name]; !found {
			return nil, fmt.Errorf("datavolume '%s' not found", name)
		}
	}

	return nameToIdMap, nil
}

func getClient(domain, token string) (dvlist.Lister, error) {
	if listClient != nil {
		return listClient, nil
	}
	return client.NewClient(domain, token)
}
