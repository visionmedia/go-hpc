// Package hpc implements a Gorilla RPC Codec which implements HTTP RPC via "/<service>/<method>" and JSON bodies.
package hpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/rpc/v2"
	"github.com/zhgo/nameconv"
)

// Errors.
var (
	ErrServiceMissing = errors.New("hpc: service name missing")
	ErrMethodMissing  = errors.New("hpc: method name missing")
)

// StatusError represents an error with HTTP status code, it is
// native to the API and may be exposed in the response.
type StatusError interface {
	StatusCode() int
	error
}

// Status error.
type statusError struct {
	Message string `json:"error"`
	Status  int    `json:"-"`
}

// Error implements error.
func (e *statusError) Error() string {
	return e.Message
}

// StatusCode implements StatusError.
func (e *statusError) StatusCode() int {
	return e.Status
}

// NewError with status code and message. Use this for public errors.
func NewError(status int, msg string) StatusError {
	return &statusError{
		Message: msg,
		Status:  status,
	}
}

// Codec implements Gorilla's rpc.Codec.
type Codec struct{}

// NewCodec returns a Gorilla RPC codec.
func NewCodec() *Codec {
	return &Codec{}
}

// NewRequest implements rpc.Codec.
func (c *Codec) NewRequest(r *http.Request) rpc.CodecRequest {
	return &codecRequest{r}
}

// Codec request.
type codecRequest struct {
	r *http.Request
}

// Method parses the service name and method from the request url.
func (c *codecRequest) Method() (string, error) {
	p := strings.Split(c.r.URL.Path, "/")

	if len(p) < 2 {
		return "", ErrServiceMissing
	}

	if len(p) < 3 {
		return "", ErrMethodMissing
	}

	service := nameconv.UnderscoreToCamelcase(p[1], true)
	method := nameconv.UnderscoreToCamelcase(p[2], true)
	s := fmt.Sprintf("%s.%s", service, method)
	return s, nil
}

// ReadRequest reads the JSON request body.
func (c *codecRequest) ReadRequest(args interface{}) error {
	return json.NewDecoder(c.r.Body).Decode(args)
}

// WriteResponse writes the JSON response body.
func (c *codecRequest) WriteResponse(w http.ResponseWriter, reply interface{}) {
	json.NewEncoder(w).Encode(reply)
}

// WriteError handles writing request errors.
func (c *codecRequest) WriteError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")

	if err, ok := err.(StatusError); ok {
		w.WriteHeader(err.StatusCode())
		json.NewEncoder(w).Encode(err)
		return
	}

	w.WriteHeader(status)

	json.NewEncoder(w).Encode(&statusError{
		Message: http.StatusText(status),
	})
}
