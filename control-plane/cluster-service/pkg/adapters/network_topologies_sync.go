/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package adapters

import (
	"slices"

	"github.com/lib/pq"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
)

type NetworkTopologiesSyncAdapter []api.NetworkTopologySyncFields

// NewNetworkTopologiesSyncAdapter creates a new NetworkTopologyLevelsAdapter from a slice of strings
func NewNetworkTopologiesSyncAdapter(networkTopologies []api.NetworkTopologySyncFields) *NetworkTopologiesSyncAdapter {
	return new(NetworkTopologiesSyncAdapter(networkTopologies))
}

func (nta *NetworkTopologiesSyncAdapter) BySources(sources ...string) *NetworkTopologiesSyncAdapter {
	filtered := make([]api.NetworkTopologySyncFields, 0, len(*nta))
	for _, field := range *nta {
		if slices.Contains(sources, field.Source) {
			filtered = append(filtered, field)
		}
	}
	return NewNetworkTopologiesSyncAdapter(filtered)
}

func (nta *NetworkTopologiesSyncAdapter) ToNames() []string {
	names := make([]string, 0, len(*nta))
	for _, field := range *nta {
		names = append(names, field.Name)
	}
	return names
}

func (nta *NetworkTopologiesSyncAdapter) ToFields() []api.NetworkTopologySyncFields {
	return *nta
}

type NetworkTopologySyncGroveAdapter api.NetworkTopologySyncGrove

func NewNetworkTopologySyncGroveAdapter(g *api.NetworkTopologySyncGrove) *NetworkTopologySyncGroveAdapter {
	if g == nil {
		return nil
	}
	return new(NetworkTopologySyncGroveAdapter(*g))
}

func (a *NetworkTopologySyncGroveAdapter) ToDBModel() *models.Grove {
	if a == nil {
		return nil
	}
	grove := &models.Grove{
		Name:    a.Name,
		Domains: pq.StringArray(a.Domains),
	}
	if len(a.Conditions) > 0 {
		conditions := make(models.GroveTopologyConditions, len(a.Conditions))
		for i, c := range a.Conditions {
			conditions[i] = models.GroveTopologyCondition{
				Type:               c.Type,
				Status:             c.Status,
				Reason:             c.Reason,
				Message:            c.Message,
				ObservedGeneration: c.ObservedGeneration,
			}
		}
		grove.Conditions = conditions
	}
	return grove
}

type GroveDBModelAdapter models.Grove

func NewGroveDBModelAdapter(g *models.Grove) *GroveDBModelAdapter {
	if g == nil {
		return nil
	}
	return new(GroveDBModelAdapter(*g))
}

func (a *GroveDBModelAdapter) ToAPIModel() *api.NetworkTopologyGrove {
	if a == nil {
		return nil
	}
	grove := &api.NetworkTopologyGrove{
		Name:    a.Name,
		Domains: []string(a.Domains),
	}
	if len(a.Conditions) > 0 {
		conditions := make([]api.NetworkTopologyGroveCondition, len(a.Conditions))
		for i, c := range a.Conditions {
			conditions[i] = api.NetworkTopologyGroveCondition{
				Type:               c.Type,
				Status:             c.Status,
				Reason:             c.Reason,
				Message:            c.Message,
				ObservedGeneration: c.ObservedGeneration,
			}
		}
		grove.Conditions = conditions
	}
	return grove
}
