package openapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func AddSwaggerEndpoints(router chi.Router) {
	router.Group(func(router chi.Router) {
		router.Get("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/swagger/swagger.yaml")
		})
		router.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/swagger/swagger.json")
		})
		// Serve the Swagger UI page.
		router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "/swagger/index.html")
		})
	})
}
