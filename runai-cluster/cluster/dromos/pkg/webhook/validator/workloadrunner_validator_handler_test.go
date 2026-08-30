package validator

import (
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestStatusErrorFor(t *testing.T) {
	base := field.NewPath("spec", "template")

	t.Run("only field errors -> Invalid (422) with causes", func(t *testing.T) {
		errs := field.ErrorList{
			field.Required(base.Child("metadata", "name"), "name is required"),
			field.Forbidden(base, "immutable"),
		}
		se := statusErrorFor("runner", errs)

		if se.ErrStatus.Reason != metav1.StatusReasonInvalid {
			t.Fatalf("Reason = %q, want %q", se.ErrStatus.Reason, metav1.StatusReasonInvalid)
		}
		if se.ErrStatus.Code != 422 {
			t.Fatalf("Code = %d, want 422", se.ErrStatus.Code)
		}
		if !apierrors.IsInvalid(se) {
			t.Fatal("IsInvalid() = false, want true")
		}
		if se.ErrStatus.Details == nil || len(se.ErrStatus.Details.Causes) != 2 {
			t.Fatalf("Details.Causes = %+v, want 2 causes", se.ErrStatus.Details)
		}
	})

	t.Run("any internal cause -> Internal error (500)", func(t *testing.T) {
		errs := field.ErrorList{
			field.Required(base.Child("metadata", "name"), "name is required"),
			field.InternalError(base.Child("sourceSecret"), assertErr("boom")),
		}
		se := statusErrorFor("runner", errs)

		if se.ErrStatus.Reason != metav1.StatusReasonInternalError {
			t.Fatalf("Reason = %q, want %q", se.ErrStatus.Reason, metav1.StatusReasonInternalError)
		}
		if se.ErrStatus.Code != 500 {
			t.Fatalf("Code = %d, want 500", se.ErrStatus.Code)
		}
		if !apierrors.IsInternalError(se) {
			t.Fatal("IsInternalError() = false, want true")
		}
	})
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
