package testutil

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

// WithSSASupport wraps a fake client.WithWatch so that Server-Side Apply
// (ApplyPatchType) patches are handled via Get+Create/Update, which the
// controller-runtime fake client does not support natively.
//
// This is a workaround, not the ideal solution. The fake client in
// controller-runtime v0.21.0 (our current version) does not support
// ApplyPatchType natively. Starting from controller-runtime v0.22.0,
// the fake client has native SSA support — Apply patches work out of
// the box, including field-level ownership tracking via managedFields.
// Once we upgrade, this helper can be removed.
func WithSSASupport(c client.WithWatch) client.WithWatch {
	return interceptor.NewClient(c, interceptor.Funcs{
		Patch: func(ctx context.Context, cl client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			if patch.Type() != types.ApplyPatchType {
				return cl.Patch(ctx, obj, patch, opts...)
			}
			existing := obj.DeepCopyObject().(client.Object)
			key := client.ObjectKeyFromObject(obj)
			if err := cl.Get(ctx, key, existing); err != nil {
				return cl.Create(ctx, obj)
			}
			obj.SetResourceVersion(existing.GetResourceVersion())
			return cl.Update(ctx, obj)
		},
	})
}
