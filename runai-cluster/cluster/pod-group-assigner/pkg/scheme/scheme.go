// Package scheme builds the runtime scheme for pod-group-assigner.
//
// Add a group here only when the service actually Gets, Lists or Watches one of
// its kinds — an unregistered kind fails at runtime with
// "no kind is registered for the type ... in scheme".
package scheme

import (
	kaitopologyv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// AddToScheme registers every kind pod-group-assigner reads:
//   - core/v1                   Namespace, Pod
//   - kai.resources/v1alpha1    NodePool, Project
//   - kai.scheduler/v1alpha1    Topology
//   - scheduling/v2             Queue
//   - scheduling/v2alpha2       PodGroup
func AddToScheme(s *runtime.Scheme) error {
	for _, add := range []func(*runtime.Scheme) error{
		clientgoscheme.AddToScheme,
		v1alpha1.AddToScheme,
		kaitopologyv1alpha1.AddToScheme,
		kaiv2.AddToScheme,
		kaiv2alpha2.AddToScheme,
	} {
		if err := add(s); err != nil {
			return err
		}
	}
	return nil
}

// Scheme returns a new scheme with all of the service's kinds registered. It
// panics on failure, which can only happen if a registration function is broken.
func Scheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(AddToScheme(s))
	return s
}
