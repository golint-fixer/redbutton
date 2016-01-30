package api
import (
	"github.com/gorilla/mux"
	"net/http"
	"encoding/json"
)

// Little helper to reduce some boilerplate when defining routes and their handlers
type RouteWrapper struct {
	Router *mux.Router
}

func NewRouteWrapper(router *mux.Router) *RouteWrapper {
	return &RouteWrapper{Router:router}
}

func (this *RouteWrapper) Route(method string, path string, handler func(c *HttpHandlerContext)) {
	this.Router.Path(path).Methods(method).HandlerFunc(wrapHandlerToConventional(handler))
}

func (this *RouteWrapper) Get(path string, handler func(c *HttpHandlerContext)) {
	this.Route("GET", path, handler)
}

func (this *RouteWrapper) Post(path string, handler func(c *HttpHandlerContext)) {
	this.Route("POST", path, handler)
}

func (this *RouteWrapper) Put(path string, handler func(c *HttpHandlerContext)) {
	this.Route("PUT", path, handler)
}

func (this *RouteWrapper) Delete(path string, handler func(c *HttpHandlerContext)) {
	this.Route("DELETE", path, handler)
}


func wrapHandlerToConventional(handler func(c *HttpHandlerContext)) func(resp http.ResponseWriter, req *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		context := NewHandler(req)
		handler(context)

		defer println("[", req.Method, "] " + req.RequestURI, context.status)

		result, err := json.MarshalIndent(context.result, "", "  ")
		if err == nil {
			resp.Header().Set("Content-Type", "application/json; charset=utf-8")
			resp.WriteHeader(context.status)
			resp.Write(result)
			return
		}

		// write error handler
		println("failed to respond properly:", err.Error())
		http.Error(resp, err.Error(), http.StatusInternalServerError)
	}
}