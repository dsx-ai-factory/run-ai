package services

import (
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
)

// Error field constants for UI parsing
const (
	ErrorFieldGVK            = "gvk"
	ErrorFieldProject        = "project"
	ErrorFieldCategory       = "category"
	ErrorFieldPriority       = "priority"
	ErrorFieldPreemptibility = "preemptibility"
	ErrorFieldName           = "name"
	ErrorFieldManifest       = "manifest"
	ErrorFieldCluster        = "cluster"
	ErrorFieldConfiguration  = "configuration"
)

// NewSubmissionErrorWithField creates a SubmissionErrorV2 with a specific field identifier
func NewSubmissionErrorWithField(code int, message string, field string) *api.SubmissionErrorV2 {
	return &api.SubmissionErrorV2{
		Code:    int32(code),
		Message: message,
		ComplianceIssues: []api.ComplianceIssuesV2ComplianceIssuesInner{
			{
				Field:   field,
				Details: message,
			},
		},
	}
}

// NewGVKError creates a SubmissionErrorV2 for GVK issues
func NewGVKError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldGVK)
}

// NewProjectError creates a SubmissionErrorV2 for project issues (400 Bad Request)
func NewProjectError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldProject)
}

// NewProjectErrorWithCode creates a SubmissionErrorV2 for project issues with a specific status code
func NewProjectErrorWithCode(code int, message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(code, message, ErrorFieldProject)
}

// NewCategoryError creates a SubmissionErrorV2 for category issues
func NewCategoryError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldCategory)
}

// NewPriorityError creates a SubmissionErrorV2 for priority issues
func NewPriorityError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldPriority)
}

// NewPreemptibilityError creates a SubmissionErrorV2 for preemptibility issues
func NewPreemptibilityError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldPreemptibility)
}

// NewNameError creates a SubmissionErrorV2 for name issues
func NewNameError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldName)
}

// NewManifestError creates a SubmissionErrorV2 for manifest issues
func NewManifestError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldManifest)
}

// NewClusterError creates a SubmissionErrorV2 for cluster issues
func NewClusterError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldCluster)
}

// NewConfigurationError creates a SubmissionErrorV2 for configuration issues
func NewConfigurationError(message string) *api.SubmissionErrorV2 {
	return NewSubmissionErrorWithField(http.StatusBadRequest, message, ErrorFieldConfiguration)
}

// ValidationErrors accumulates multiple validation errors during validation
type ValidationErrors struct {
	errors []api.SubmissionErrorV2
}

// NewValidationErrors creates a new ValidationErrors accumulator
func NewValidationErrors() *ValidationErrors {
	return &ValidationErrors{
		errors: make([]api.SubmissionErrorV2, 0),
	}
}

// Add adds a validation error for a specific field
func (v *ValidationErrors) Add(err *api.SubmissionErrorV2) {
	if err != nil {
		v.errors = append(v.errors, *err)
	}
}

// HasErrors returns true if there are any validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.errors) > 0
}

// ToSubmissionErrors returns the accumulated errors as a slice
func (v *ValidationErrors) ToSubmissionErrors() []api.SubmissionErrorV2 {
	return v.errors
}

// SingleError wraps a single error into an array for consistent API response format
func SingleError(err *api.SubmissionErrorV2) []api.SubmissionErrorV2 {
	if err == nil {
		return nil
	}
	return []api.SubmissionErrorV2{*err}
}

// ToSubmissionErrorV2 converts an api.Error to SubmissionErrorV2 (for generic errors)
func ToSubmissionErrorV2(err *api.Error) *api.SubmissionErrorV2 {
	return &api.SubmissionErrorV2{
		Code:    err.Code,
		Message: err.Message,
		Details: err.Details,
	}
}

// SubmissionErrorToApiError converts a SubmissionErrorV2 back to api.Error
// Use this for functions that still need to return api.Error
func SubmissionErrorToApiError(err *api.SubmissionErrorV2) *api.Error {
	return &api.Error{
		Code:    err.Code,
		Message: err.Message,
		Details: err.Details,
	}
}
