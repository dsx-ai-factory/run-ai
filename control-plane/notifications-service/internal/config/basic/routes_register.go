package basic

import "github.com/go-chi/chi/v5"

type RoutesRegister struct{}

func NewRoutesRegister() *RoutesRegister {
	return &RoutesRegister{}
}

func (*RoutesRegister) GetName() string {
	return "basic_routes_register"
}

func (*RoutesRegister) Register(_ *chi.Mux) {
}
