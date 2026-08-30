package common

import (
	"context"
	"fmt"
	"time"

	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"

	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OldSharedBuilder is used for building and creating resources for any kind of builder
// TODO: this struct should be refactor and removed
type OldSharedBuilder struct {
	//  the input of the builder
	ObjectNaming                   *metav1.ObjectMeta
	SharedFields                   *field_groups.WorkloadFields
	DistSharedFields               *field_groups.WorkloadFields // for the master
	SharedStatus                   *common.WorkloadStatus
	InternalSharedResourcesBuilder *SharedResourcesBuilder
	//  shared resources that the builder builds and creates
	SharedResources *SharedResources
	Workload        *job.JobWorkload
	Options         factory.BuilderOptions
	K8sClient       client.Client
}

// new builder

// ReportEvent allows the builder to report events to the status of the Workload
func (self *OldSharedBuilder) ReportEvent(event common.Event) {
	self.SharedStatus.LogEvent(event)
}

func (self *OldSharedBuilder) ReportCreatedResource(object client.Object) {
	if object == nil {
		return
	}
	var resourceTypeAndName common.ResourceTypeAndName

	resourceTypeAndName.Namespace = object.GetNamespace()
	resourceTypeAndName.Name = object.GetName()

	result, _, err := self.K8sClient.Scheme().ObjectKinds(object)
	if err != nil || len(result) == 0 {
		log.Error(err, "failed to determine APIVersion", "resource", resourceTypeAndName)
	} else {
		resourceTypeAndName.Kind = result[0].Kind
		if result[0].Group != "" {
			resourceTypeAndName.APIVersion = fmt.Sprintf("%s/%s", result[0].Group, result[0].Version)
		} else {
			resourceTypeAndName.APIVersion = result[0].Version
		}
	}

	self.SharedStatus.LogResource(resourceTypeAndName)
}

// Build shared resources pvc and services
func (self *OldSharedBuilder) Build(ctx context.Context, _ *metav1.ObjectMeta, _ string, isJobSupportAuth bool) error {
	var err error
	pvcList, err := self.InternalSharedResourcesBuilder.BuildPVC(ctx, self.Workload, ResourceBuilderOptions{})
	if err != nil {
		return err
	}
	for _, pvc := range pvcList {
		pvc := pvc
		if pvc == nil {
			continue
		}
		self.SharedResources.Pvcs = append(self.SharedResources.Pvcs, *pvc.(*corev1.PersistentVolumeClaim))
	}

	secretList, err := self.InternalSharedResourcesBuilder.BuildCredentials(ctx, self.Workload, ResourceBuilderOptions{})
	if err != nil {
		return err
	}
	for _, secret := range secretList {
		secret := secret
		if secret == nil {
			continue
		}
		self.SharedResources.Secrets = append(self.SharedResources.Secrets, *secret.(*corev1.Secret))
	}

	services, err := self.InternalSharedResourcesBuilder.BuildServices(ctx, self.Workload, ResourceBuilderOptions{})
	if err != nil {
		return err
	}
	for _, service := range services {
		service := service
		if service == nil {
			continue
		}
		self.SharedResources.Services = append(self.SharedResources.Services, *service)
	}

	objects, err := self.InternalSharedResourcesBuilder.BuildExternalUrls(ctx, self.Workload, &ExternalBuilderOptions{
		DisableAuth: ptr.To(!isJobSupportAuth),
	})
	if err != nil {
		return err
	}
	for i := range objects {
		if objects[i] == nil {
			continue
		}
		switch castObject := objects[i].(type) {
		case *corev1.Service:
			self.SharedResources.Services = append(self.SharedResources.Services, *castObject)
		case *corev1.PersistentVolumeClaim:
			self.SharedResources.Pvcs = append(self.SharedResources.Pvcs, *castObject)
		case *v1.Ingress:
			self.SharedResources.Ingress = append(self.SharedResources.Ingress, *castObject)
		case *openshiftV1.Route:
			self.SharedResources.Route = append(self.SharedResources.Route, *castObject)
		case *corev1.Secret:
			self.SharedResources.Secrets = append(self.SharedResources.Secrets, *castObject)
		}
	}

	return nil
}

// BuildObjectMeta initialize the object meta and put some basic stuff in it which is relevant for any resource
func (self *OldSharedBuilder) BuildObjectMeta(_ context.Context, om *metav1.ObjectMeta, workloadType *metav1.TypeMeta) {

	if om.Annotations == nil {
		om.Annotations = map[string]string{}
	}

	if om.Labels == nil {
		om.Labels = map[string]string{}
	}

	workloadbuilders.AddUserDefinedAnnotationsAndLabels(om, &self.SharedFields.ContainerFields)

	om.Namespace = self.ObjectNaming.Namespace
	om.CreationTimestamp = metav1.Time{Time: time.Now()}

	om.Labels[constants.LabelWorkloadName] = self.ObjectNaming.Name
	om.Labels[constants.LabelWorkloadKind] = workloadType.Kind

	om.Annotations[constants.AnnotationWorkloadApiVersion] = workloadType.APIVersion
	if self.Workload != nil {
		workloadbuilders.CopyTopologyAnnotations(om, self.Workload.Annotations)
	}
}

// HandleDryRun return creation options and log prefix based on "dry run" option
func (self *OldSharedBuilder) HandleDryRun() (client.CreateOptions, client.UpdateOptions, string) {
	var logPrefix string

	createOptions := client.CreateOptions{}
	updateOptions := client.UpdateOptions{}

	if self.Options.DryRun {
		createOptions.DryRun = []string{"All"}
		updateOptions.DryRun = []string{"All"}
		logPrefix = "dry-run completed"
	} else {
		logPrefix = "created"
	}

	return createOptions, updateOptions, logPrefix
}

func (self *OldSharedBuilder) IsWorkloadActive(stateFields field_groups.StateFields) bool {
	// if not set - assume active
	return !stateFields.Active.IsValueInitialized() || stateFields.Active.Deref()
}
