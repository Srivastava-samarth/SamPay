package routes

import "github.com/Srivastava-samarth/sampay/controllers"

type Router struct {
	Controller *controllers.Controller
}

func NewRouter(
	controller *controllers.Controller,
) *Router {
	return &Router{
		Controller: controller,
	}
}
