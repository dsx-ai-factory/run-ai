package operands

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceOld wraps a resource with its operand, mainly used for logging
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

// ResourceForOperand structs a Resource from a k8s resource and its operand
func ResourceForOperand(resource client.Object, operandName string) ResourceOld {
	return ResourceOld{
		Object:      resource,
		OperandName: operandName,
		Enabled:     true,
	}
}

// Status reports whether an operand's resources are ready, and why not if they
// are not.
type Status struct {
	// Ready specifies if the operand is ready or not
	Ready bool

	// Reasons specifies the reasons why the operand is not ready
	Reasons []string
}

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
