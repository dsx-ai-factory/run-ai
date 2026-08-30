package authz

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/auth"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/authz/authorizers"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	authv1 "k8s.io/api/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RunAIGroup                            = "run.ai"
	TrainingWorkloadsResource             = "trainingworkloads"
	InferenceWorkloadsResource            = "inferenceworkloads"
	DistributedInferenceWorkloadsResource = "distributedinferenceworkloads"
	InteractiveWorkloadsResource          = "interactiveworkloads"
	DistributedWorkloadsResource          = "distributedworkloads"
	GetVerb                               = "get"
	CreateVerb                            = "create"
)

type ResourceAccessControl interface {
	Authorize(ctx context.Context, request *AccessRequest) error
}

type Resource struct {
	Namespace string
	Name      string
	Group     string
}

type AccessRequest struct {
	Resource *Resource
	Subject  string
	Groups   []string
	Action   string
}

type K8sAuthorizer struct {
	resourceAccessControl ResourceAccessControl
}

func NewK8sAuthorizer(resourceAccessControl ResourceAccessControl) *K8sAuthorizer {
	log.Debug().Msg("Creating new K8sAuthorizer")

	return &K8sAuthorizer{
		resourceAccessControl: resourceAccessControl,
	}
}

func (ka *K8sAuthorizer) Authorize(ctx context.Context, request *authorizers.Request) error {
	log.Info().
		Str("action", request.Action.String()).
		Str("kind", request.Kind.String()).
		Str("namespace", request.Namespace).
		Str("workload", request.Workload).
		Msg("Authorizing request")

	switch request.Action {
	case api.PortForward, api.Logs, api.Status:
		return ka.authorizeReadWorkload(ctx, request)
	case api.Exec, api.Attach:
		return ka.authorizeExecWorkload(ctx, request)
	default:
		log.Error().Str("action", request.Action.String()).Msg("Invalid workload action")
		return fmt.Errorf("invalid workload action: %v", request.Action)
	}
}

func (ka *K8sAuthorizer) authorizeReadWorkload(ctx context.Context, request *authorizers.Request) error {
	log.Info().
		Str("kind", request.Kind.String()).
		Str("namespace", request.Namespace).
		Str("workload", request.Workload).
		Msg("Authorizing read workload")

	rbacRequest, err := createResourceFromRequest(request)
	if err != nil {
		log.Error().Err(err).Msg("Error creating resource from request")
		return err
	}

	return ka.authorizeRequest(ctx, request.AuthInfo, &rbacRequest, GetVerb)
}

func (ka *K8sAuthorizer) authorizeExecWorkload(ctx context.Context, request *authorizers.Request) error {
	log.Info().
		Str("kind", request.Kind.String()).
		Str("namespace", request.Namespace).
		Str("workload", request.Workload).
		Msg("Authorizing exec workload")

	rbacRequest, err := createResourceFromRequest(request)
	if err != nil {
		log.Error().Err(err).Msg("Error creating resource from request")
		return err
	}

	return ka.authorizeRequest(ctx, request.AuthInfo, &rbacRequest, CreateVerb)
}

func createResourceFromRequest(request *authorizers.Request) (Resource, error) {
	log.Info().Str("kind", request.Kind.String()).Msg("Creating resource from request")

	var resource string

	switch request.Kind {
	case api.Training:
		resource = TrainingWorkloadsResource
	case api.Workspace:
		resource = InteractiveWorkloadsResource
	case api.Distributed:
		resource = DistributedWorkloadsResource
	case api.Inference:
		resource = InferenceWorkloadsResource
	case api.DistributedInference:
		resource = DistributedInferenceWorkloadsResource
	default:
		log.Error().Str("kind", request.Kind.String()).Msg("Invalid workload kind")
		return Resource{}, fmt.Errorf("invalid workload kind: %v", request.Kind)
	}

	return Resource{
		Namespace: request.Namespace,
		Group:     RunAIGroup,
		Name:      resource,
	}, nil
}

func (ka *K8sAuthorizer) authorizeRequest(
	ctx context.Context, authInfo *auth.Info, rbacRequest *Resource, action string,
) error {
	log.Info().
		Str("resource", rbacRequest.Name).
		Str("action", action).
		Msg("Authorizing request")

	subjects := map[string]struct{}{}
	for _, possibleSubject := range []string{
		authInfo.Subject,
		authInfo.OriginalSubject,
		authInfo.Username,
	} {
		subjects[possibleSubject] = struct{}{}
	}

	groups := authInfo.Groups

	var wg sync.WaitGroup

	wg.Add(len(subjects))
	errChan := make(chan error, len(subjects))

	for subj := range subjects {
		go ka.authorizeSubjectUsingRBAC(ctx, subj, groups, rbacRequest, errChan, &wg, action)
	}

	wg.Wait()
	close(errChan)

	var subjectAccessReviewsErrors []error

	for err := range errChan {
		if err != nil {
			subjectAccessReviewsErrors = append(subjectAccessReviewsErrors, err)
		}
	}

	if len(subjectAccessReviewsErrors) >= len(subjects) {
		log.Warn().
			Str("action", action).
			Str("resource", rbacRequest.Name).
			Str("namespace", rbacRequest.Namespace).
			Msg("User is not authorized")

		return NewUnauthorizedError(fmt.Sprintf(
			"user is not authorized to perform the [%s] action on resource [%s] in namespace [%s]: %v",
			action, rbacRequest.Name, rbacRequest.Namespace, subjectAccessReviewsErrors))
	}

	log.Info().
		Str("action", action).
		Str("resource", rbacRequest.Name).
		Str("namespace", rbacRequest.Namespace).
		Msg("User is authorized")

	return nil
}

func (ka *K8sAuthorizer) authorizeSubjectUsingRBAC(
	ctx context.Context, subject string, groups []string,
	rbacRequest *Resource, errChan chan error, wg *sync.WaitGroup, action string,
) {
	defer wg.Done()

	log.Info().
		Str("subject", subject).
		Str("resource", rbacRequest.Name).
		Str("action", action).
		Msg("Authorizing subject using RBAC")

	err := ka.resourceAccessControl.Authorize(ctx, &AccessRequest{
		Resource: rbacRequest,
		Subject:  subject,
		Groups:   groups,
		Action:   action,
	})

	errChan <- err
}

type RBACAuthorizer struct {
	k8sClient client.Client
}

type RBACAuthorizeResponse struct {
	AccessAllowed bool
}

// NewRBACAuthorizer creates a new RBAC authorizer.
// if cacheExpirationPeriod is 0, a non-cached authorizer is returned
func NewRBACAuthorizer(k8sClient client.Client) *RBACAuthorizer {
	log.Info().Msg("Creating new RBACAuthorizer")

	return &RBACAuthorizer{
		k8sClient: k8sClient,
	}
}

func (ra *RBACAuthorizer) Authorize(ctx context.Context, request *AccessRequest) error {
	log.Info().
		Str("subject", request.Subject).
		Str("resource", request.Resource.Name).
		Str("action", request.Action).
		Msg("Authorizing request using RBAC")

	cacheKey := cacheKeyForRequest(request)

	var accessAllowed bool

	cachedResponse, _ := cache.Get[RBACAuthorizeResponse](ctx, cacheKey)
	if cachedResponse != nil {
		log.Debug().Str("subject", request.Subject).Msg("using cached response")

		accessAllowed = cachedResponse.AccessAllowed
	} else {
		log.Debug().Str("subject", request.Subject).Msg("using fresh response and caching it")

		response, err := queryKubernetesRBACAuthorizer(ctx, ra.k8sClient, request)
		if err != nil {
			return fmt.Errorf("failed querying kubernetes rbac authorizer: %v", err)
		}

		accessAllowed = response

		err = cache.Set[RBACAuthorizeResponse](ctx, cacheKey, RBACAuthorizeResponse{AccessAllowed: accessAllowed})
		if err != nil {
			log.Debug().Str("subject", request.Subject).Msg("failed to set cache for RBACAuthorizeResponse")
		}
	}

	if !accessAllowed {
		log.Warn().Str("subject", request.Subject).Msg("Access denied for subject")
		return NewUnauthorizedError(request.Subject)
	}

	log.Info().Str("subject", request.Subject).Msg("Access allowed for subject")

	return nil
}

func cacheKeyForRequest(request *AccessRequest) string {
	return fmt.Sprintf("%s_%s_%s_%s_%s_%v",
		request.Resource.Namespace, request.Action, request.Resource.Group,
		request.Resource.Name, request.Subject, request.Groups)
}

func subjectAccessReviewForRequest(request *AccessRequest) *authv1.SubjectAccessReview {
	return &authv1.SubjectAccessReview{
		Spec: authv1.SubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Namespace: request.Resource.Namespace,
				Verb:      request.Action,
				Group:     request.Resource.Group,
				Resource:  request.Resource.Name,
			},
			User:   request.Subject,
			Groups: request.Groups,
		},
	}
}

func queryKubernetesRBACAuthorizer(ctx context.Context, k8s client.Client, request *AccessRequest) (bool, error) {
	sar := subjectAccessReviewForRequest(request)

	err := k8s.Create(ctx, sar, &client.CreateOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Error creating SubjectAccessReview")
		return false, err
	}

	return sar.Status.Allowed, nil
}
