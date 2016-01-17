package redbutton
import (
	"github.com/martini-contrib/render"
	"net/http"
	"strconv"
	"log"
)


type ApiError struct {
	Message string	`json:"message"`
	httpCode int
}

func (this ApiError) Error() string {
	return this.Message+"("+strconv.Itoa(this.httpCode)+")"
}

func (this ApiError) httpError(message string, code int) ApiError {
	this.Message = message
	this.httpCode = code
	return this
}

func (this ApiError) badRequest(message string) ApiError {
	return this.httpError(message, http.StatusBadRequest)
}

func (this ApiError) notFound(message string) ApiError {
	return this.httpError(message, http.StatusNotFound)
}

// analyses error contents and returns an error
func respondWithError(err error, r render.Render) {
	apiError, ok := err.(ApiError)
	if !ok {
		apiError = ApiError{Message: err.Error(), httpCode: http.StatusInternalServerError}
	}
	log.Println("responding with error: ", apiError.httpCode, apiError.Message)
	r.JSON(apiError.httpCode,apiError)

}