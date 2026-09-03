/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package operands

import (
	"context"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceFunction func(context.Context, client.Reader, *runv1.RunaiConfig) (client.Object, error)
type ResourceStatusFunction func(context.Context, client.Reader, *runv1.RunaiConfig) (Status, error)

// Operand is an interface each operand has to implement in order to
// smoothly integrate with the controller.
// An Operand is defined as a group of Kubernetes resources that
// together form a logical unit, such as a micro-service, database, etc
// TODO review
type Operand interface {
	Name() string
	Resources() []Resource
	Status(ctx context.Context) (Status, error)
	Enabled() bool
}

type OperandOld interface {
	Name() string
	ResourcesForRunaiConfig(ctx context.Context, hybridClient kc.HybridClient,
		runaiConfig *runv1.RunaiConfig) ([]ResourceOld, error)
	Status(ctx context.Context, hybridClient kc.HybridClient,
		runaiConfig *runv1.RunaiConfig) (Status, error)
}

// ResourceOld wraps a reource with it's operand, mainly used for logging
type ResourceOld struct {
	Object client.Object

	OperandName string
	Enabled     bool
}

type Resource interface {
	OperandName() string
	Object(ctx context.Context) (client.Object, error)
	Status(ctx context.Context) error
	Enabled() bool
}

// ResourceForOperand structs a Resource from a k8s resource and it's operand
func ResourceForOperand(resource client.Object, operandName string) ResourceOld {
	return ResourceOld{
		Object:      resource,
		OperandName: operandName,
		Enabled:     true,
	}
}

type Status = runv1.OperandStatus

func ReadyStatus() Status {
	return Status{
		Ready: true,
	}
}

func NotReadyStatus(reasons ...string) Status {
	return Status{
		Reasons: reasons,
	}
}
