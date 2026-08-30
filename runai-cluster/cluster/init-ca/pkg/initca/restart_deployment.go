package initca

import (
	"context"

	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// from kubectl:
// https://github.com/kubernetes/kubectl/blob/c1df07341a15fb36264a018e130ee397332bdfc0/pkg/polymorphichelpers/objectrestarter.go#L51
func rolloutRestartDeployment(ctx context.Context, client kubernetes.Interface, deploymentName, namespace string) error {
	deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			return nil
		}
		return err
	}

	selector := deployment.Spec.Selector
	err = client.CoreV1().Pods(namespace).DeleteCollection(ctx, metav1.DeleteOptions{}, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(selector.MatchLabels),
	})
	if err != nil {
		return err
	}

	return nil
}
