package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

// HTTPHandlerContext provides methods for parsing JSON request, and sending responses
type HTTPHandlerContext struct {
	Req            *http.Request
	routeVariables map[string]string
	status         int
	result         interface{}
}

func NewHandler(req *http.Request) *HTTPHandlerContext {
	return &HTTPHandlerContext{Req: req, status: http.StatusOK}
}

func (c *HTTPHandlerContext) Status(status int) *HTTPHandlerContext {
	c.status = status
	return c
}

func (c *HTTPHandlerContext) Result(result interface{}) *HTTPHandlerContext {
	c.result = result
	return c
}

func (c *HTTPHandlerContext) ParseRequest(r interface{}) bool {
	decoder := json.NewDecoder(c.Req.Body)
	err := decoder.Decode(r)
	if err != nil {
		c.Error(http.StatusBadRequest, "could not parse request: "+err.Error())
		return false
	}
	return true
}

func (c *HTTPHandlerContext) PathParam(name string) string {
	if c.routeVariables == nil {
		c.routeVariables = mux.Vars(c.Req)
		if c.routeVariables == nil {
			c.routeVariables = map[string]string{}
		}
	}

	return c.routeVariables[name]
}

func (c *HTTPHandlerContext) Error(status int, message string) {
	c.Status(status)
	c.Result(map[string]string{"message": message})
}
