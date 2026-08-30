package project

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

const (
	projectKey                = "projectName"
	failedToTranslateErrorMsg = "Failed to translate project to namespace"
)

type Translator interface {
	Translate(ctx context.Context, projectName string) (string, error)
}

func NewNamespaceMiddleware(translator Translator) *NamespaceMiddleware {
	log.Debug().Msg("Creating new NamespaceMiddleware")

	return &NamespaceMiddleware{
		translator: translator,
	}
}

type NamespaceMiddleware struct {
	translator Translator
}

func (n *NamespaceMiddleware) Middleware(next http.Handler) http.Handler {
	log.Debug().Msg("Creating middleware")
	return n.middleware(next)
}

func (n *NamespaceMiddleware) middleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		log.Debug().Msg("Processing request")

		projectName := chi.URLParam(r, projectKey)
		log.Debug().Str("projectName", projectName).Msg("Retrieved project name from URL")

		if projectName == "" {
			log.Error().Msg("Project name is empty")
			http.Error(w, "project name is required", http.StatusBadRequest)

			return
		}

		projectCtx := WithProject(ctx, projectName)
		log.Debug().Str("projectName", projectName).Msg("Added project to context")

		namespace, err := n.translator.Translate(ctx, projectName)
		if err != nil {
			log.Error().Err(err).Str("projectName", projectName).Msg(failedToTranslateErrorMsg)
			http.Error(w, failedToTranslateErrorMsg, http.StatusInternalServerError)

			return
		}

		log.Debug().Str("projectName", projectName).Str("namespace", namespace).Msg("Translated project to namespace")

		namespaceCtx := WithNamespace(projectCtx, namespace)
		log.Debug().Str("namespace", namespace).Msg("Added namespace to context")

		log.Debug().Msg("Calling next handler")
		next.ServeHTTP(w, r.WithContext(namespaceCtx))
	}
}
