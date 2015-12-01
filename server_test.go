package hpc_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/rpc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tj/go-hpc"
)

type AddInput struct {
	A int
	B int
}

type AddOutput struct {
	Value int `json:"value"`
}

type Math struct {
}

func (m *Math) Add(r *http.Request, in *AddInput, out *AddOutput) error {
	out.Value = in.A + in.B
	return nil
}

type GetStatsOutput struct {
	Requests int `json:"requests"`
}

func (m *Math) GetStats(r *http.Request, in *struct{}, out *GetStatsOutput) error {
	out.Requests = 5
	return nil
}

func (m *Math) Error(r *http.Request, in *struct{}, out *struct{}) error {
	return hpc.NewError(400, "Boom")
}

func (m *Math) InternalError(r *http.Request, in *struct{}, out *struct{}) error {
	return errors.New("boom")
}

func TestCodec_request(t *testing.T) {
	r := rpc.NewServer()
	r.RegisterCodec(hpc.NewCodec(), "application/json")
	r.RegisterService(&Math{}, "")

	s := httptest.NewServer(r)
	defer s.Close()

	body := strings.NewReader(`{ "a": 5, "b": 10 }`)
	res, err := http.Post(s.URL+"/math/add", "application/json", body)
	assert.NoError(t, err, "error posting")

	b, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "error reading")
	defer res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "{\"value\":15}\n", string(b))
}

func TestCodec_underscores(t *testing.T) {
	r := rpc.NewServer()
	r.RegisterCodec(hpc.NewCodec(), "application/json")
	r.RegisterService(&Math{}, "")

	s := httptest.NewServer(r)
	defer s.Close()

	body := strings.NewReader(`{ "a": 5, "b": 10 }`)
	res, err := http.Post(s.URL+"/math/get_stats", "application/json", body)
	assert.NoError(t, err, "error posting")

	b, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "error reading")
	defer res.Body.Close()

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "{\"requests\":5}\n", string(b))
}

func TestCodec_errors(t *testing.T) {
	r := rpc.NewServer()
	r.RegisterCodec(hpc.NewCodec(), "application/json")
	r.RegisterService(&Math{}, "")

	s := httptest.NewServer(r)
	defer s.Close()

	body := strings.NewReader(`{ "a": 5, "b": 10 }`)
	res, err := http.Post(s.URL+"/math/error", "application/json", body)
	assert.NoError(t, err, "error posting")

	b, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "error reading")
	defer res.Body.Close()

	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, "{\"error\":\"Boom\"}\n", string(b))
}

func TestCodec_internalErrors(t *testing.T) {
	r := rpc.NewServer()
	r.RegisterCodec(hpc.NewCodec(), "application/json")
	r.RegisterService(&Math{}, "")

	s := httptest.NewServer(r)
	defer s.Close()

	body := strings.NewReader(`{ "a": 5, "b": 10 }`)
	res, err := http.Post(s.URL+"/math/internal_error", "application/json", body)
	assert.NoError(t, err, "error posting")
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "error reading")

	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, "{\"error\":\"Bad Request\"}\n", string(b))
}
