package router

import (
	"net/http"

	"github.com/run-ai/runai/runai-installer/provisioner-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/services"
)

func Create() http.Handler {
	// Create the provision service
	provisionService := services.NewProvisionerService()

	// Wrap it with the adapter to implement the generated API interface
	serviceAdapter := services.NewProvisionServiceAdapter(provisionService)

	// Create the API controller with the adapter
	provisionController := api.NewProvisionAPIController(serviceAdapter)

	// Create and return the chi router with all routes
	router := api.NewRouter(provisionController)

	// Add health check endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	return router
}
