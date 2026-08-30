package handlers

import (
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
)

type ProjectResourceHandler interface {
	HandleResource(project kaiv1alpha1.Project) ([]kaiv1alpha1.ProjectCondition, error)
}
