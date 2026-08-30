package handler

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/providers"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/application-controller/pkg/common"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientCache "k8s.io/client-go/tools/cache"
	ctrlCache "sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
)

func StartApplicationHandler(ctx context.Context, client ctrlClient.Client, cache ctrlCache.Cache) {
	addInformers(ctx, client, cache)
	if !cache.WaitForCacheSync(ctx) {
		log.Fatal().Msgf("Failed waiting for cache to sync")
	}
}

func addInformers(ctx context.Context, client ctrlClient.Client, cache ctrlCache.Cache) {
	// Get all application source factories
	factories := getApplicationFactories()

	// Set up an informer for each source resource type
	for _, factory := range factories {
		addInformerForFactory(ctx, client, cache, factory)
	}
}

func addInformerForFactory(ctx context.Context, client ctrlClient.Client, cache ctrlCache.Cache, factory common.ApplicationFactory) {
	// Get the source resource type to watch
	sourceResourceType := factory.GetSourceResource()

	informer, err := cache.GetInformer(ctx, sourceResourceType)
	if err != nil {
		log.Warn().Msgf("Failed to get informer for source %s, error: %s", factory.GetApplicationSource(), err.Error())
		return
	}

	_, err = informer.AddEventHandler(clientCache.FilteringResourceEventHandler{
		FilterFunc: func(object interface{}) bool {
			sourceResource, ok := object.(ctrlClient.Object)
			if !ok || sourceResource == nil {
				return false
			}
			return factory.FilterSourceResource(ctx, client, sourceResource)
		},
		Handler: clientCache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				sourceResource := obj.(ctrlClient.Object)
				genericControllerErrorHandler(handleAdd(ctx, client, factory, sourceResource), "Add", factory.GetApplicationSource(), sourceResource.GetName())
			},
			UpdateFunc: func(oldObj any, newObj any) {
				sourceResource := newObj.(ctrlClient.Object)
				genericControllerErrorHandler(handleUpdate(ctx, client, factory, sourceResource), "Update", factory.GetApplicationSource(), sourceResource.GetName())
			},
		},
	})
	if err != nil {
		log.Warn().Msgf("Failed to add event handler for source %s, error: %s", factory.GetApplicationSource(), err.Error())
	}
}

// helmSourceProvider is the singleton instance used by the handler
// Exposed for testing purposes to allow injecting mock helpers
var helmSourceProvider = providers.NewHelmSourceProvider()

// GetHelmSourceProvider returns the Helm source provider instance (useful for testing)
func GetHelmSourceProvider() *providers.HelmSourceProvider {
	return helmSourceProvider
}

func getApplicationFactories() []common.ApplicationFactory {
	return []common.ApplicationFactory{
		helmSourceProvider,
	}
}

func handleAdd(ctx context.Context, client ctrlClient.Client, factory common.ApplicationFactory, sourceResource ctrlClient.Object) error {
	return handleUpsert(ctx, client, factory, sourceResource)
}

func handleUpdate(ctx context.Context, client ctrlClient.Client, factory common.ApplicationFactory, sourceResource ctrlClient.Object) error {
	return handleUpsert(ctx, client, factory, sourceResource)
}

func handleUpsert(ctx context.Context, client ctrlClient.Client, factory common.ApplicationFactory, sourceResource ctrlClient.Object) error {
	log.Debug().Msgf("Handling upsert of %s source resource %s", factory.GetApplicationSource(), sourceResource.GetName())

	// Check if this CRD should create an application
	shouldCreate, applicationName, err := factory.ShouldCreateApplication(ctx, client, sourceResource)
	if err != nil || !shouldCreate || applicationName == "" {
		// Not an error - just means we shouldn't create an application for this resource
		return nil
	}

	// Try to get existing Application
	application := &v1alpha1.Application{}
	err = client.Get(ctx, ctrlClient.ObjectKey{
		Namespace: sourceResource.GetNamespace(),
		Name:      applicationName,
	}, application)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info().Msgf("Application for %s source %s in namespace %s does not exist - creating a new one.",
				factory.GetApplicationSource(), sourceResource.GetName(), sourceResource.GetNamespace())
			application, err = createApplicationFromFactory(ctx, client, factory, sourceResource)
			if err != nil {
				log.Error().Msgf("Failed to create application with name %s in namespace %s error: %v",
					applicationName, sourceResource.GetNamespace(), err)
				return err
			}
			// Fetch the latest version to get the correct resource version before updating status
			err = client.Get(ctx, ctrlClient.ObjectKey{
				Namespace: sourceResource.GetNamespace(),
				Name:      applicationName,
			}, application)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Backfill managed fields missing on Applications created by older cluster versions.
	backfillApplicationFields(ctx, client, factory, sourceResource, application)

	// Cast factory to provider (they're the same object implementing both interfaces)
	provider, ok := factory.(common.ApplicationSourceProvider)
	if !ok {
		return fmt.Errorf("factory for source %s does not implement ApplicationSourceProvider", factory.GetApplicationSource())
	}

	applicationPhase := provider.GetApplicationPhase(sourceResource)
	return common.UpdateApplicationStatus(ctx, client, application, &applicationPhase, provider)
}

func createApplicationFromFactory(ctx context.Context, client ctrlClient.Client, factory common.ApplicationFactory, sourceResource ctrlClient.Object) (*v1alpha1.Application, error) {
	// Get application metadata from the factory
	metadata, err := factory.CreateApplicationMetadata(ctx, client, sourceResource)
	if err != nil {
		return nil, fmt.Errorf("failed to create application metadata: %w", err)
	}

	application := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:            metadata.Name,
			Namespace:       metadata.Namespace,
			Labels:          metadata.Labels,
			OwnerReferences: metadata.OwnerReferences,
		},
		Spec: v1alpha1.ApplicationSpec{
			Type:   metadata.Type,
			Source: metadata.Source,
		},
		Status: v1alpha1.ApplicationStatus{},
	}

	err = client.Create(ctx, application)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, err
	}
	log.Info().Msgf("Application for %s source %s in namespace %s created successfully.",
		factory.GetApplicationSource(), sourceResource.GetName(), sourceResource.GetNamespace())
	return application, nil
}

// backfillApplicationFields backfills fields that newer cluster versions set on
// Application creation but that are missing on Applications created by older versions
// (backward-compatibility for upgraded clusters). Currently only labels are backfilled;
// spec fields can be added the same way in the future.
func backfillApplicationFields(ctx context.Context, client ctrlClient.Client, factory common.ApplicationFactory, sourceResource ctrlClient.Object, application *v1alpha1.Application) {
	departmentName, ok := application.Labels[clusterconstants.DepartmentNameLabel]
	if ok && departmentName != "" {
		return
	}

	metadata, err := factory.CreateApplicationMetadata(ctx, client, sourceResource)
	if err != nil {
		log.Warn().Msgf("Skipping label backfill for application %s: failed to build metadata: %v", application.Name, err)
		return
	}

	patch := ctrlClient.MergeFrom(application.DeepCopy())

	if application.Labels == nil {
		application.Labels = map[string]string{}
	}
	application.Labels[clusterconstants.DepartmentNameLabel] = metadata.Labels[clusterconstants.DepartmentNameLabel]

	if err = client.Patch(ctx, application, patch); err != nil {
		log.Warn().Msgf("Failed to backfill labels on application %s: %v", application.Name, err)
		return
	}

	log.Info().Msgf("Backfilled missing managed labels on application %s in namespace %s", application.Name, application.Namespace)
}

func genericControllerErrorHandler(err error, operation, resourceType, resourceName string) {
	if err != nil {
		zap.S().Errorf("Had an error while trying to %s %s %s: %v ", operation, resourceType, resourceName, err)
	}
}
