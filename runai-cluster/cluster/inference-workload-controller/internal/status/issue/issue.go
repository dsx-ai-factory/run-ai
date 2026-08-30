package issue

import (
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

type InferenceResourceGrouper struct {
	issue.DefaultResourceGrouper
}

func (rg *InferenceResourceGrouper) GetObjectResource(resource client.Object) (metav1.GroupVersionKind, error) {
	defaultResourceGrouper := issue.DefaultResourceGrouper{}

	if resource == nil {
		return defaultResourceGrouper.GetObjectResource(resource)
	}

	var objectResource metav1.GroupVersionKind
	switch resource.(type) {
	case *servingv1.Service:
		objectResource = metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeServiceKind,
		}
	case *servingv1.Configuration:
		objectResource = metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeConfigurationKind,
		}
	case *servingv1.Revision:
		objectResource = metav1.GroupVersionKind{
			Group:   constants.KnativeServiceGroup,
			Version: constants.KnativeServiceVersion,
			Kind:    constants.KnativeRevisionKind,
		}
	case *lwsv1.LeaderWorkerSet:
		objectResource = metav1.GroupVersionKind{
			Group:   constants.LwsGroup,
			Version: constants.LwsVersion,
			Kind:    constants.LwsKind,
		}
	default:
		return defaultResourceGrouper.GetObjectResource(resource)
	}

	return objectResource, nil
}

func (rg *InferenceResourceGrouper) GetIssueGroupByGvk(gvk metav1.GroupVersionKind) issue.Group {
	if gvk.Group == constants.KnativeServiceGroup &&
		(gvk.Kind == constants.KnativeServiceKind ||
			gvk.Kind == constants.KnativeConfigurationKind ||
			gvk.Kind == constants.KnativeRevisionKind) {
		return issue.Compute
	}

	if gvk.Group == constants.LwsGroup && gvk.Kind == constants.LwsKind {
		return issue.Compute
	}

	return rg.DefaultResourceGrouper.GetIssueGroupByGvk(gvk)
}

func NewIssue(object client.Object, err error) *issue.Issue {
	return issue.New(object, err, false, "", &InferenceResourceGrouper{})
}
