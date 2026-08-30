package workload

import (
	"context"
	"fmt"
	"net/http"

	projectApi "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/project"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	TypeSupportedFrameworksMsg = "workload type %s supported frameworks are %v"
	FrameworkMismatchMsg       = "could not find workload with the specified framework %s"
)

type Locator interface {
	Locate(ctx context.Context, workloadType api.WorkloadType, workloadName, namespace string) (client.Object, error)
}

type ExistMiddleware struct {
	locator Locator
}

func (e *ExistMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(e.middleware(next))
}

func (e *ExistMiddleware) middleware(next http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Debug().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("Starting middleware processing")

		namespace := projectApi.NamespaceFromContext(ctx)
		if namespace == "" {
			log.Error().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Namespace is required but not in context")
			http.Error(w, "namespace is required, but not in context", http.StatusInternalServerError)

			return
		}

		log.Debug().Str("namespace", namespace).Msg("Namespace retrieved from context")

		workloadType, err := api.WorkloadTypeFromContext(ctx)
		if err != nil || workloadType == "" {
			log.Error().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Workload type is required but not in context")
			http.Error(w, "workload type is required, but not in context", http.StatusInternalServerError)

			return
		}

		log.Debug().Str("workloadType", workloadType.String()).Msg("Workload type retrieved from context")

		workloadFramework, err := api.WorkloadFrameworkFromContext(ctx)
		if err != nil || workloadFramework == "" {
			log.Error().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Workload framework is required but not in context")
			http.Error(w, "workload framework is required, but not in context", http.StatusInternalServerError)

			return
		}

		log.Debug().Str("workloadFramework", workloadFramework.String()).Msg("Workload framework retrieved from context")

		err = validateWorkloadTypeAndFrameworkMatch(workloadType, workloadFramework)
		if err != nil {
			log.Error().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		workloadName := api.WorkloadNameFromContext(ctx)
		if workloadName == "" {
			log.Error().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Workload name is required but not in context")
			http.Error(w, "workload name is required, but not in context", http.StatusInternalServerError)

			return
		}

		log.Debug().Str("workloadName", workloadName).Msg("Workload name retrieved from context")

		project := projectApi.FromContext(ctx)
		if project == "" {
			log.Error().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Project is required but not in context")
			http.Error(w, "project is required, but not in context", http.StatusInternalServerError)

			return
		}

		log.Debug().Str("project", project).Msg("Project retrieved from context")

		updatedRequest := locateLogic(ctx, e.locator, w, r, workloadType, workloadName, namespace, project, workloadFramework)
		if updatedRequest == nil {
			log.Debug().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("locateLogic returned nil, stopping middleware processing")

			return
		}

		log.Debug().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("Proceeding to next handler")
		next.ServeHTTP(w, updatedRequest)
	}
}

func locateLogic(ctx context.Context, locator Locator, w http.ResponseWriter, r *http.Request,
	workloadType api.WorkloadType, workloadName string, namespace, project string, workloadFramework api.WorkloadFramework,
) *http.Request {
	log.Debug().
		Str("workloadType", workloadType.String()).
		Str("workloadName", workloadName).
		Str("workloadFramework", workloadFramework.String()).
		Str("namespace", namespace).
		Str("project", project).
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Msg("Starting locate logic")

	workloadResource, err := locator.Locate(ctx, workloadType, workloadName, namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Error().
				Str("workloadName", workloadName).
				Str("namespace", namespace).
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Workload not found")
			http.Error(w, "workload not found", http.StatusNotFound)

			return nil
		} else {
			log.Error().
				Err(err).
				Str("workloadName", workloadName).
				Str("namespace", namespace).
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Msg("Error locating workload")
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return nil
		}
	}

	log.Debug().
		Str("workloadResourceType", fmt.Sprintf("%T", workloadResource)).
		Msg("Workload resource located successfully")

	workload, err := FromK8sResource(workloadResource, workloadType, project)
	if err != nil {
		log.Error().
			Err(err).
			Str("workloadResourceType", fmt.Sprintf("%T", workloadResource)).
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("Error converting K8s resource to workload")

		http.Error(w, err.Error(), http.StatusBadRequest)

		return nil
	}

	log.Debug().
		Interface("workload", workload).
		Msg("K8s resource converted to workload successfully")

	if err = validateLocatedFramework(workload, workloadFramework); err != nil {
		log.Error().
			Err(err).
			Str("workloadFramework", workloadFramework.String()).
			Str("jobType", workload.JobType.String()).
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("Located workload's framework does not match the expected")
		http.Error(w, err.Error(), http.StatusNotFound)

		return nil
	}

	log.Debug().
		Str("workloadFramework", workloadFramework.String()).
		Str("jobType", workload.JobType.String()).
		Msg("Located workload's framework matches the expected")

	// add workload to context
	ctx = SetIntoContext(ctx, workload)
	updateRequest := r.WithContext(ctx)
	log.Debug().
		Str("method", updateRequest.Method).
		Str("url", updateRequest.URL.String()).
		Msg("Workload added to context, returning updated request")

	return updateRequest
}

func validateWorkloadTypeAndFrameworkMatch(
	workloadType api.WorkloadType, workloadFramework api.WorkloadFramework,
) error {
	switch workloadType {
	case api.Workspace:
		if workloadFramework != api.RunaiFramework {
			return fmt.Errorf(TypeSupportedFrameworksMsg, api.Workspace, []api.WorkloadFramework{api.RunaiFramework})
		}
	case api.Training:
		if workloadFramework == api.ExternalFramework ||
			workloadFramework == api.KnativeFramework ||
			workloadFramework == api.LwsFramework {
			return fmt.Errorf(TypeSupportedFrameworksMsg, api.Distributed, []api.WorkloadFramework{
				api.RunaiFramework, api.MPIFramework,
				api.PyTorchFramework, api.TFFramework, api.XGBoostFramework, api.JAXFrameworkName, api.AnyDistributedFramework,
			})
		}
	case api.Distributed:
		if workloadFramework == api.RunaiFramework || workloadFramework == api.ExternalFramework ||
			workloadFramework == api.KnativeFramework || workloadFramework == api.LwsFramework {
			return fmt.Errorf(TypeSupportedFrameworksMsg, api.Distributed, []api.WorkloadFramework{
				api.MPIFramework,
				api.PyTorchFramework, api.TFFramework, api.XGBoostFramework, api.JAXFrameworkName, api.AnyDistributedFramework,
			})
		}
	case api.Inference:
		if workloadFramework != api.KnativeFramework {
			return fmt.Errorf(TypeSupportedFrameworksMsg, api.Inference, []api.WorkloadFramework{api.KnativeFramework})
		}
	case api.DistributedInference:
		return nil
	case api.External:
		if workloadFramework != api.ExternalFramework {
			return fmt.Errorf(TypeSupportedFrameworksMsg,
				api.ExternalFrameworkName, []api.WorkloadFramework{api.ExternalFramework})
		}
	}

	return nil
}

func validateLocatedFramework(located *Workload, workloadFramework api.WorkloadFramework) error {
	// 'any' doesn't have a job type, need to handle separately
	if workloadFramework == api.AnyDistributedFramework {
		if located.JobType == api.RunaiJob {
			return fmt.Errorf(FrameworkMismatchMsg, workloadFramework.String())
		} else {
			return nil
		}
	}

	if located.JobType == api.DistributedInferenceJob {
		return nil
	}

	frameworkJobType, err := workloadFramework.JobType()
	if err != nil {
		return err
	}

	if located.JobType != frameworkJobType {
		return fmt.Errorf(FrameworkMismatchMsg, workloadFramework.String())
	}

	return nil
}

func NewExistMiddleware(locator Locator) *ExistMiddleware {
	return &ExistMiddleware{
		locator: locator,
	}
}
