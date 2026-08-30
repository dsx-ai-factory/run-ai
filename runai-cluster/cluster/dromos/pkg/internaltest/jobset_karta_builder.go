package internaltest

import (
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobSetKartaBuilder struct {
	karta *kartav1alpha1.Karta
}

func NewJobSetKartaBuilder(name string) *JobSetKartaBuilder {
	return &JobSetKartaBuilder{
		karta: &kartav1alpha1.Karta{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: kartav1alpha1.KartaSpec{
				StructureDefinition: kartav1alpha1.StructureDefinition{
					RootComponent: kartav1alpha1.ComponentDefinition{
						StatusDefinition: &kartav1alpha1.StatusDefinition{
							StatusMappings: kartav1alpha1.StatusMappings{},
						},
					},
				},
			},
		},
	}
}

func (b *JobSetKartaBuilder) WithDefaults() *JobSetKartaBuilder {
	defaultKarta := test.NewValidJobSetKarta()
	defaultKarta.Name = b.karta.Name
	b.karta = defaultKarta
	return b
}

func (b *JobSetKartaBuilder) WithStatusMatcher(phase kartav1alpha1.ResourceStatus, conditions []kartav1alpha1.ExpectedCondition) *JobSetKartaBuilder {
	matcher := kartav1alpha1.StatusMatcher{
		ByConditions: conditions,
	}

	if b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition == nil {
		b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition = &kartav1alpha1.StatusDefinition{}
	}

	switch phase {
	case kartav1alpha1.InitializingStatus:
		b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Initializing = []kartav1alpha1.StatusMatcher{matcher}
	case kartav1alpha1.RunningStatus:
		b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Running = []kartav1alpha1.StatusMatcher{matcher}
	case kartav1alpha1.CompletedStatus:
		b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Completed = []kartav1alpha1.StatusMatcher{matcher}
	case kartav1alpha1.FailedStatus:
		b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Failed = []kartav1alpha1.StatusMatcher{matcher}
	case kartav1alpha1.SuspendedStatus:
		b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Suspended = []kartav1alpha1.StatusMatcher{matcher}
	case kartav1alpha1.SuspendingStatus:
		b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Suspending = []kartav1alpha1.StatusMatcher{matcher}
	case kartav1alpha1.ResumingStatus:
		b.karta.Spec.StructureDefinition.RootComponent.StatusDefinition.StatusMappings.Resuming = []kartav1alpha1.StatusMatcher{matcher}
	}

	return b
}

func (b *JobSetKartaBuilder) Build() *kartav1alpha1.Karta {
	return b.karta
}
