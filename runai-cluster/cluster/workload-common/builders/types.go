package builders

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource struct {
	Kind      metav1.GroupVersionKind
	Name      string
	Namespace string
}

func (r Resource) String() string {
	return fmt.Sprintf("%s/%s", r.Kind.String(), r.Name)
}
