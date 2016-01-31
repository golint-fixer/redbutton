package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

// RouteWrapper - Little helper to reduce some boilerplate when defining routes and their handlers
type RouteWrapper struct {
	Router *mux.Router
}

// NewRouteWrapper creates new RouteWrapper
func NewRouteWrapper(router *mux.Router) *RouteWrapper {
	return &RouteWrapper{Router: router}
}

func (wrapper *RouteWrapper) Route(method string, path string, handler func(c *HTTPHandlerContext)) {
	wrapper.Router.Path(path).Methods(method).HandlerFunc(wrapHandlerToConventional(handler))
}

func (wrapper *RouteWrapper) Get(path string, handler func(c *HTTPHandlerContext)) {
	wrapper.Route("GET", path, handler)
}

func (wrapper *RouteWrapper) Post(path string, handler func(c *HTTPHandlerContext)) {
	wrapper.Route("POST", path, handler)
}

func (wrapper *RouteWrapper) Put(path string, handler func(c *HTTPHandlerContext)) {
	wrapper.Route("PUT", path, handler)
}

func (wrapper *RouteWrapper) Delete(path string, handler func(c *HTTPHandlerContext)) {
	wrapper.Route("DELETE", path, handler)
}

func wrapHandlerToConventional(handler func(c *HTTPHandlerContext)) func(resp http.ResponseWriter, req *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		context := NewHandler(req)
		defer println("[", req.Method, "] "+req.RequestURI, context.status)

		handler(context)

		if context.result == nil {
			context.Error(http.StatusInternalServerError, "response was not created")
		}

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
