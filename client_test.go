package hpc_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/rpc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tj/go-hpc"
)

func TestClient_Call(t *testing.T) {
	r := rpc.NewServer()
	r.RegisterCodec(hpc.NewCodec(), "application/json")
	r.RegisterService(&Math{}, "")

	s := httptest.NewServer(r)
	defer s.Close()

	c := hpc.NewClient(hpc.NewConfig(s.URL))

	var out AddOutput
	err := c.Call("math", "add", AddInput{5, 10}, &out)
	assert.NoError(t, err, "error calling")

	assert.Equal(t, 15, out.Value)
}
