package legacycleanup

import (
	"context"
	"fmt"

	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	authorizationclient "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

type requiredPermission struct {
	Verb      string
	Group     string
	Resource  string
	Namespace string
	Name      string
}

type permissionReviewer interface {
	Review(context.Context, requiredPermission) (allowed bool, reason string, err error)
}

type selfSubjectAccessReviewer struct {
	client authorizationclient.AuthorizationV1Interface
}

func newSelfSubjectAccessReviewer(client authorizationclient.AuthorizationV1Interface) permissionReviewer {
	return &selfSubjectAccessReviewer{client: client}
}

func (s *selfSubjectAccessReviewer) Review(ctx context.Context, permission requiredPermission) (bool, string, error) {
	review, err := s.client.SelfSubjectAccessReviews().Create(ctx, &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb: permission.Verb, Group: permission.Group, Resource: permission.Resource,
				Namespace: permission.Namespace, Name: permission.Name,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return false, "", err
	}
	return review.Status.Allowed, review.Status.Reason, nil
}

func checkPermissions(ctx context.Context, reviewer permissionReviewer, permissions []requiredPermission) error {
	for _, permission := range permissions {
		allowed, reason, err := reviewer.Review(ctx, permission)
		if err != nil {
			return fmt.Errorf("SelfSubjectAccessReview for %s: %w", describePermission(permission), err)
		}
		if !allowed {
			return fmt.Errorf("permission denied for %s: %s", describePermission(permission), reason)
		}
	}
	return nil
}

func describePermission(permission requiredPermission) string {
	group := permission.Group
	if group == "" {
		group = "core"
	}
	scope := "cluster scope"
	if permission.Namespace != "" {
		scope = fmt.Sprintf("namespace %q", permission.Namespace)
	}
	name := ""
	if permission.Name != "" {
		name = fmt.Sprintf(" name %q", permission.Name)
	}
	return fmt.Sprintf("%s %s/%s at %s%s", permission.Verb, group, permission.Resource, scope, name)
}
