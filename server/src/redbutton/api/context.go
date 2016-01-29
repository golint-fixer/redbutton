package api
import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
)

// provides methods for parsing JSON request, and sending responses
type HttpHandlerContext struct {
	req            *http.Request
	routeVariables map[string]string
	status         int
	result         interface{}
}

func NewHandler(req *http.Request) *HttpHandlerContext {
	return &HttpHandlerContext{req: req, status: http.StatusOK}
}

func (this *HttpHandlerContext) Status(status int) *HttpHandlerContext {
	this.status = status
	return this
}

func (this *HttpHandlerContext) Result(result interface{}) *HttpHandlerContext {
	this.result = result
	return this
}

func (this *HttpHandlerContext) ParseRequest(r interface{}) bool {
	decoder := json.NewDecoder(this.req.Body)
	err := decoder.Decode(r)
	if err != nil {
		this.Error(http.StatusBadRequest, "could not parse request: " + err.Error())
		return false
	}
	return true
}

func (this *HttpHandlerContext) PathParam(name string) string {
	if this.routeVariables == nil {
		this.routeVariables = mux.Vars(this.req)
		if this.routeVariables == nil {
			this.routeVariables = map[string]string{}
		}
	}

	return this.routeVariables[name]
}

func (this *HttpHandlerContext) Error(status int, message string) {
	println("error: ", status, message)
	this.Status(status)
	this.Result(map[string]string{"message": message})
}
