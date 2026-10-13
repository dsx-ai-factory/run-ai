/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// staleWebhookError mimics what the API server returns while the 2.24 copy of
// runai-project-validation still points at runai-project-controller:9443, a port
// the new operator no longer serves: an InternalError wrapping the webhook call
// failure. init-ca re-points the webhook a little later.
func staleWebhookError() error {
	return apierrors.NewInternalError(errors.New(
		`failed calling webhook "runai-project-validation": failed to call webhook: ` +
			`Post "https://runai-project-controller.runai.svc:9443/validate?timeout=10s": ` +
			`no service port 9443 found for service "runai-project-controller"`))
}

// shortenWebhookRetry makes the retry loop fast enough for unit tests and restores
// the production values afterwards.
func shortenWebhookRetry(t *testing.T, timeout time.Duration) {
	t.Helper()
	prevInterval, prevTimeout := webhookRetryInterval, webhookRetryTimeout
	webhookRetryInterval = time.Millisecond
	webhookRetryTimeout = timeout
	t.Cleanup(func() {
		webhookRetryInterval, webhookRetryTimeout = prevInterval, prevTimeout
	})
}

// failingDeletes returns interceptor funcs whose Delete fails with err for the first
// failures calls that match isTarget and delegates otherwise. attempts counts the
// matching Delete calls.
func failingDeletes(isTarget func(client.Object) bool, failures int, err error, attempts *int) interceptor.Funcs {
	return interceptor.Funcs{
		Delete: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			if !isTarget(obj) {
				return cl.Delete(ctx, obj, opts...)
			}
			*attempts++
			if *attempts <= failures {
				return err
			}
			return cl.Delete(ctx, obj, opts...)
		},
	}
}

func isLegacyDepartment(obj client.Object) bool {
	_, ok := obj.(*schedv2alpha2.Department)
	return ok
}

func isLegacyProject(obj client.Object) bool {
	_, ok := obj.(*runaiv2.Project)
	return ok
}

func TestDepartmentCleanupRetriesDeleteWhileWebhookIsUnreachable(t *testing.T) {
	shortenWebhookRetry(t, time.Second)
	attempts := 0
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartment("default"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
	).WithInterceptorFuncs(failingDeletes(isLegacyDepartment, 2, staleWebhookError(), &attempts)).Build()

	require.NoError(t, hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background())))

	require.Equal(t, 3, attempts, "two webhook failures then success")
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, &schedv2alpha2.Department{})))
	require.False(t, oldDepartmentCRDExists(t, k8sClient))
}

func TestProjectCleanupRetriesDeleteWhileWebhookIsUnreachable(t *testing.T) {
	shortenWebhookRetry(t, time.Second)
	attempts := 0
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiProjectScheme(t)).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
		departmentCRD(newProjectCRDName),
		legacyProject("proj-a"),
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}},
	).WithInterceptorFuncs(failingDeletes(isLegacyProject, 2, staleWebhookError(), &attempts)).Build()

	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))

	require.Equal(t, 3, attempts, "two webhook failures then success")
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &runaiv2.Project{})))
	require.False(t, oldProjectCRDExists(t, k8sClient))
}

func TestDepartmentCleanupDoesNotRetryNonWebhookDeleteErrors(t *testing.T) {
	shortenWebhookRetry(t, time.Second)
	attempts := 0
	hook := cleanupHook(true)
	forbidden := apierrors.NewForbidden(schedv2alpha2.GroupVersion.WithResource("departments").GroupResource(), "default", errors.New("denied"))
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartment("default"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
	).WithInterceptorFuncs(failingDeletes(isLegacyDepartment, 100, forbidden, &attempts)).Build()

	err := hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background()))

	require.ErrorContains(t, err, `failed to delete legacy department "default"`)
	require.Equal(t, 1, attempts, "a non-webhook error is returned on the first attempt")
	// The legacy department and CRD are left in place.
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, &schedv2alpha2.Department{}))
	require.True(t, oldDepartmentCRDExists(t, k8sClient))
}

func TestDepartmentCleanupGivesUpWhenWebhookStaysUnreachable(t *testing.T) {
	shortenWebhookRetry(t, 20*time.Millisecond)
	attempts := 0
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartment("default"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
	).WithInterceptorFuncs(failingDeletes(isLegacyDepartment, 1_000_000, staleWebhookError(), &attempts)).Build()

	err := hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background()))

	require.ErrorContains(t, err, `failed to delete legacy department "default"`)
	require.ErrorContains(t, err, "failed calling webhook")
	require.Greater(t, attempts, 1, "the delete is retried until the timeout")
	require.True(t, oldDepartmentCRDExists(t, k8sClient), "the legacy CRD is kept when a department could not be deleted")
}

// blockingDeletes returns interceptor funcs whose Delete, for objects matching
// isTarget, blocks until the request context is done and returns its error, the way
// a real API server call that never answers behaves.
func blockingDeletes(isTarget func(client.Object) bool) interceptor.Funcs {
	return interceptor.Funcs{
		Delete: func(ctx context.Context, cl client.WithWatch, obj client.Object, opts ...client.DeleteOption) error {
			if !isTarget(obj) {
				return cl.Delete(ctx, obj, opts...)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}
}

func TestDepartmentCleanupCancelsABlockedDeleteAtTheRetryTimeout(t *testing.T) {
	shortenWebhookRetry(t, 50*time.Millisecond)
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartment("default"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
	).WithInterceptorFuncs(blockingDeletes(isLegacyDepartment)).Build()

	done := make(chan error, 1)
	go func() {
		done <- hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background()))
	}()

	select {
	case err := <-done:
		require.ErrorContains(t, err, `failed to delete legacy department "default"`)
		require.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(2 * time.Second):
		t.Fatal("cleanup did not return after the retry timeout: the delete is not bounded by it")
	}
	require.True(t, oldDepartmentCRDExists(t, k8sClient), "the legacy CRD is kept when a department could not be deleted")
}

func unavailableAgentWebhookError() error {
	return apierrors.NewInternalError(errors.New(
		`failed calling webhook "runai-project-validation": failed to call webhook: ` +
			`Post "https://runai-agent.runai.svc:9443/validate-project?timeout=10s": ` +
			`dial tcp 10.0.0.1:9443: connect: connection refused`))
}

func failingFinalizerPatches(failures int, patchErr error, attempts *int) interceptor.Funcs {
	return interceptor.Funcs{
		Patch: func(ctx context.Context, cl client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
			*attempts++
			if *attempts <= failures {
				return patchErr
			}
			return cl.Patch(ctx, obj, patch, opts...)
		},
	}
}

func TestDepartmentCleanupRetriesFinalizerPatchWhileAgentWebhookIsUnreachable(t *testing.T) {
	shortenWebhookRetry(t, time.Second)
	attempts := 0
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartmentWithFinalizer("default"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
	).WithInterceptorFuncs(failingFinalizerPatches(2, unavailableAgentWebhookError(), &attempts)).Build()

	require.NoError(t, hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background())))

	require.Equal(t, 3, attempts)
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, &schedv2alpha2.Department{})))
	require.False(t, oldDepartmentCRDExists(t, k8sClient))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, &kaiv1alpha1.Department{}))
}

func TestProjectCleanupRetriesFinalizerPatchWhileAgentWebhookIsUnreachable(t *testing.T) {
	shortenWebhookRetry(t, time.Second)
	attempts := 0
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiProjectScheme(t)).WithObjects(
		departmentCRD(hooks.OldProjectCRDName),
		departmentCRD(newProjectCRDName),
		legacyProjectWithFinalizer("proj-a"),
		&kaiv1alpha1.Project{ObjectMeta: metav1.ObjectMeta{Name: "proj-a"}},
	).WithInterceptorFuncs(failingFinalizerPatches(2, unavailableAgentWebhookError(), &attempts)).Build()

	require.NoError(t, hook.cleanupLegacyProjects(context.Background(), k8sClient))

	require.Equal(t, 3, attempts)
	require.True(t, apierrors.IsNotFound(
		k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &runaiv2.Project{})))
	require.False(t, oldProjectCRDExists(t, k8sClient))
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "proj-a"}, &kaiv1alpha1.Project{}))
}

func TestClearLegacyFinalizersDoesNotRetryOtherPatchErrors(t *testing.T) {
	resource := schedv2alpha2.GroupVersion.WithResource("departments").GroupResource()
	for _, tc := range []struct {
		name     string
		patchErr error
		wantErr  bool
	}{
		{"validation rejection", apierrors.NewForbidden(resource, "default", errors.New("admission webhook denied the request")), true},
		{"unrelated internal error", apierrors.NewInternalError(errors.New("storage unavailable")), true},
		{"already deleted", apierrors.NewNotFound(resource, "default"), false},
		{"CRD already removed", &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: resource.Group, Kind: "Department"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			shortenWebhookRetry(t, time.Second)
			attempts := 0
			dep := legacyDepartmentWithFinalizer("default")
			k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(dep).
				WithInterceptorFuncs(failingFinalizerPatches(100, tc.patchErr, &attempts)).Build()

			err := clearLegacyFinalizers(context.Background(), k8sClient, dep, "department", log.FromContext(context.Background()))

			if tc.wantErr {
				require.ErrorIs(t, err, tc.patchErr)
				require.ErrorContains(t, err, `failed to clear finalizers on legacy department "default"`)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, 1, attempts)
		})
	}
}

func TestDepartmentCleanupKeepsFinalizersAndCRDWhenPatchRetryTimesOut(t *testing.T) {
	shortenWebhookRetry(t, 20*time.Millisecond)
	attempts := 0
	webhookErr := unavailableAgentWebhookError()
	hook := cleanupHook(true)
	k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(
		departmentCRD(hooks.OldDepartmentCRDName),
		departmentCRD(newDepartmentCRDName),
		legacyDepartmentWithFinalizer("default"),
		&kaiv1alpha1.Department{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
	).WithInterceptorFuncs(failingFinalizerPatches(1_000_000, webhookErr, &attempts)).Build()

	err := hook.cleanupLegacyDepartments(context.Background(), k8sClient, log.FromContext(context.Background()))

	require.ErrorContains(t, err, "admission webhook still unreachable after")
	require.ErrorIs(t, err, webhookErr)
	require.Greater(t, attempts, 1)
	var remaining schedv2alpha2.Department
	require.NoError(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, &remaining))
	require.Equal(t, []string{"run.ai.finalizers.department"}, remaining.Finalizers)
	require.Nil(t, remaining.DeletionTimestamp)
	require.True(t, oldDepartmentCRDExists(t, k8sClient))
}

func TestClearLegacyFinalizersBoundsBlockedPatch(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cancel  bool
		wantErr error
	}{
		{"retry timeout", false, context.DeadlineExceeded},
		{"caller cancellation", true, context.Canceled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			shortenWebhookRetry(t, 50*time.Millisecond)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			dep := legacyDepartmentWithFinalizer("default")
			k8sClient := fake.NewClientBuilder().WithScheme(kaiDepartmentScheme(t)).WithObjects(dep).
				WithInterceptorFuncs(interceptor.Funcs{
					Patch: func(ctx context.Context, _ client.WithWatch, _ client.Object, _ client.Patch, _ ...client.PatchOption) error {
						if tc.cancel {
							cancel()
						}
						<-ctx.Done()
						return ctx.Err()
					},
				}).Build()

			done := make(chan error, 1)
			go func() {
				done <- clearLegacyFinalizers(ctx, k8sClient, dep, "department", log.FromContext(ctx))
			}()
			select {
			case err := <-done:
				require.ErrorIs(t, err, tc.wantErr)
				require.ErrorContains(t, err, `failed to clear finalizers on legacy department "default"`)
			case <-time.After(2 * time.Second):
				t.Fatal("finalizer patch did not stop when its context ended")
			}
		})
	}
}
