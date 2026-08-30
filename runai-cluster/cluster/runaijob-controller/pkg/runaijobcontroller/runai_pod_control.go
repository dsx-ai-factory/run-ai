package runaijobcontroller

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/runaijob-controller/pkg/common"

	runaijob "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller"
)

type PodControlInterface interface {
	// CreateStatefulPod create a Pod in a StatefulSet. Any PVCs necessary for the Pod are created prior to creating
	// the Pod. If the returned error is nil the Pod and its PVCs have been created.
	CreatePod(job *runaijob.RunaiJob, podName string, podIndex int64, podAttempt int64) error
	// UpdateStatefulPod Updates a Pod in a StatefulSet. If the Pod already has the correct identity and stable
	// storage this method is a no-op. If the Pod must be mutated to conform to the Set, it is mutated and updated.
	// pod is an in-out parameter, and any updates made to the pod are reflected as mutations to this parameter. If
	// the create is successful, the returned error is nil.
	// UpdateStatefulPod(set *apps.StatefulSet, pod *v1.Pod) error
	// DeleteStatefulPod deletes a Pod in a StatefulSet. The pods PVCs are not deleted. If the delete is successful,
	// the returned error is nil.
	DeletePod(namespace string, podID string, object runtime.Object) error
}

type RealPodControl struct {
	KubeClient clientset.Interface
	Recorder   record.EventRecorder
}

func NewPodControl(
	client clientset.Interface,
	recorder record.EventRecorder,
) PodControlInterface {
	return &RealPodControl{client, recorder}
}

func (r RealPodControl) DeletePod(namespace string, podID string, object runtime.Object) error {
	accessor, err := meta.Accessor(object)
	if err != nil {
		return fmt.Errorf("object does not have ObjectMeta, %v", err)
	}
	klog.V(2).InfoS("Deleting pod", "controller", accessor.GetName(), "pod", klog.KRef(namespace, podID))
	if err := r.KubeClient.CoreV1().Pods(namespace).Delete(context.TODO(), podID, metav1.DeleteOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			klog.V(4).Infof("pod %v/%v has already been deleted.", namespace, podID)
			return err
		}
		r.Recorder.Eventf(object, v1.EventTypeWarning, controller.FailedDeletePodReason, "Error deleting: %v", err)
		return fmt.Errorf("unable to delete pods: %v", err)
	}
	r.Recorder.Eventf(object, v1.EventTypeNormal, controller.SuccessfulDeletePodReason, "Deleted pod: %v", podID)

	return nil
}

func (r RealPodControl) CreatePod(job *runaijob.RunaiJob, podName string, podIndex int64, podAttempt int64) error {
	controllerRef := metav1.NewControllerRef(job, controllerKind)
	pod, err := GetPodFromTemplate(&job.Spec.Template, job, controllerRef)
	if err != nil {
		return err
	}
	if len(labels.Set(pod.Labels)) == 0 {
		return fmt.Errorf("unable to create pods, no labels")
	}

	pod.Labels[labelOfPodIndex] = fmt.Sprintf("%d-%d", podIndex, podAttempt)
	pod.Name = podName
	for containerIndex := range pod.Spec.Containers {
		pod.Spec.Containers[containerIndex].Env = append(pod.Spec.Containers[containerIndex].Env, v1.EnvVar{
			Name:  "POD_INDEX",
			Value: fmt.Sprintf("%d", podIndex),
		})
	}
	newPod, err := r.KubeClient.CoreV1().Pods(job.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	// sink already exists errors
	if apierrors.IsAlreadyExists(err) {
		klog.V(4).Infof("Attempted to create pod for job: <%v/%v> with th name: <%v> but a pod with that name already exists", job.Namespace, job.Name, pod.Name)
		return err
	}
	if err != nil {
		// only send an event if the namespace isn't terminating
		if !apierrors.HasStatusCause(err, v1.NamespaceTerminatingCause) {
			r.Recorder.Eventf(job, v1.EventTypeWarning, controller.FailedCreatePodReason, "Error creating: %v", err)
		}
		return err
	}
	accessor, err := meta.Accessor(job)
	if err != nil {
		klog.Errorf("parentObject does not have ObjectMeta, %v", err)
		return nil
	}
	klog.V(4).Infof("Controller %v created pod %v", accessor.GetName(), newPod.Name)
	r.Recorder.Eventf(job, v1.EventTypeNormal, controller.SuccessfulCreatePodReason, "Created pod: %v", newPod.Name)
	return nil
}

func GetPodFromTemplate(template *v1.PodTemplateSpec, job *runaijob.RunaiJob, controllerRef *metav1.OwnerReference) (*v1.Pod, error) {
	desiredLabels := common.GetPodsLabelSet(template)
	desiredFinalizers := common.GetPodsFinalizers(template)
	desiredAnnotations := common.GetPodsAnnotationSet(template)
	accessor, err := meta.Accessor(job)
	if err != nil {
		return nil, fmt.Errorf("parentObject does not have ObjectMeta, %v", err)
	}
	prefix := common.GetPodsPrefix(accessor.GetName())

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:       desiredLabels,
			Annotations:  desiredAnnotations,
			GenerateName: prefix,
			Finalizers:   desiredFinalizers,
		},
	}
	if controllerRef != nil {
		pod.OwnerReferences = append(pod.OwnerReferences, *controllerRef)
	}
	pod.Spec = *template.Spec.DeepCopy()
	return pod, nil
}
