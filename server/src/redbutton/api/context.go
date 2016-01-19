package api
import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
)

type HttpHandlerContext struct {
	req            *http.Request
	routeVariables map[string]string
	status         int
	result         interface{}
}

func NewHandler(req *http.Request) *HttpHandlerContext {
	return &HttpHandlerContext{req: req, status: http.StatusOK}
}

func Wrap(handler func(c *HttpHandlerContext)) func(resp http.ResponseWriter, req *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		context := NewHandler(req)
		handler(context)

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


func (this *HttpHandlerContext) Status(status int) *HttpHandlerContext{
	this.status = status
	return this
}

func (this *HttpHandlerContext) Result(result interface{}) *HttpHandlerContext{
	this.result = result
	return this
}

func (this *HttpHandlerContext) ParseRequest(r interface{}) bool {
	decoder := json.NewDecoder(this.req.Body)
	err := decoder.Decode(r)
	if err != nil {
		this.Error(http.StatusBadRequest, "could not parse request: "+err.Error())
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
	this.Status(status)
	this.Result(map[string]string{"message": message})
}
