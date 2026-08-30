package client

import (
	"context"
	"fmt"
	workloadcommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// The FakeClientWithEventUIDIndex is a decorator for a fake client, adding to it the option
// to list events by
type FakeClientWithEventUIDIndex struct {
	client.Client

	// field to track the number of times list events was invoked
	EventListCalls int
}

// List supports any k8s object that the fake client handles, but adds special handling for event list
// the reason we need all this mess is that the native fake client does not support this functionality.
// In the future, this function can support additional types, not only events.
func (f *FakeClientWithEventUIDIndex) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)

	uidField := workloadcommon.InvolvedObjectUidField

	// Fallback for non-events or missing selector
	_, isEventList := list.(*corev1.EventList)
	if !isEventList || listOpts.FieldSelector == nil {
		return f.Client.List(ctx, list, opts...)
	}

	// Extract UID value from MatchingFields selector
	selectorStr := listOpts.FieldSelector.String()
	if !strings.Contains(selectorStr, uidField) {
		// Not filtering by UID → delegate
		return f.Client.List(ctx, list, opts...)
	}

	// Increment test counter
	f.EventListCalls++

	// Parse "involvedObject.uid=<value>"
	parts := strings.SplitN(selectorStr, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("unexpected selector format: %q", selectorStr)
	}
	filterUID := parts[1]

	// List all events
	all := &corev1.EventList{}
	if err := f.Client.List(ctx, all); err != nil {
		return err
	}

	// Filter manually
	var objs []runtime.Object
	for i := range all.Items {
		ev := &all.Items[i]
		if string(ev.InvolvedObject.UID) == filterUID &&
			(listOpts.Namespace == "" || ev.Namespace == listOpts.Namespace) {
			objs = append(objs, ev)
		}
	}

	return apimeta.SetList(list, objs)
}
