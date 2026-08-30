package patch

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AddAnnotationToObjectMeta(objectMeta *metav1.ObjectMeta, key string, value string) {
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = make(map[string]string)
	}
	objectMeta.Annotations[key] = value
}
