package api
import (
"github.com/gorilla/mux"
)

type RouteWrapper struct {
	Router *mux.Router
}

func NewRouteWrapper(router *mux.Router) *RouteWrapper {
	return &RouteWrapper{Router:router}
}

func (this *RouteWrapper) Route(method string, path string, handler func(c *HttpHandlerContext)){
	this.Router.Path(path).Methods(method).HandlerFunc(Wrap(handler))
}

func (this *RouteWrapper) Get(path string, handler func(c *HttpHandlerContext)){
	this.Route("GET",path,handler)
}

func (this *RouteWrapper) Post(path string, handler func(c *HttpHandlerContext)){
	this.Route("POST",path,handler)
}

func (this *RouteWrapper) Put(path string, handler func(c *HttpHandlerContext)){
	this.Route("PUT",path,handler)
}

func (this *RouteWrapper) Delete(path string, handler func(c *HttpHandlerContext)){
	this.Route("DELETE",path,handler)
}
