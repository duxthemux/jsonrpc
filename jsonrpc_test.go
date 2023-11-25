package jsonrpc_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/duxthemux/jsonrpc"
)

type TestIn struct {
	A  int
	B  int
	Op string
}
type TestOut struct {
	Res int
}
type TestHandler struct {
}

func (t *TestHandler) Name() string {
	return "namedTestHandler"
}

func (t *TestHandler) TestMethod(_ context.Context, in *TestIn, out *TestOut) error {
	switch in.Op {
	case "+":
		out.Res = in.A + in.B
		return nil
	default:
		return fmt.Errorf("op unknown")
	}

	return nil
}

func TestServer_Call(t *testing.T) {
	server := jsonrpc.New()

	err := server.Register(&TestHandler{})
	assert.NoError(t, err)

	srv := httptest.NewServer(server)
	defer srv.Close()

	cli := jsonrpc.NewClient(srv.URL + "/${HANDLER}/${METHOD}")
	in := &TestIn{
		A:  3,
		B:  4,
		Op: "+",
	}
	out := &TestOut{}

	err = cli.Call("namedtesthandler", "testmethod", in, out)
	assert.NoError(t, err)
	assert.Equal(t, 7, out.Res)
}

func TestServer_CallWError(t *testing.T) {

	server := jsonrpc.New()

	err := server.Register(&TestHandler{})
	assert.NoError(t, err)

	srv := httptest.NewServer(server)
	defer srv.Close()

	cli := jsonrpc.NewClient(srv.URL + "/${HANDLER}/${METHOD}")
	in := &TestIn{
		A:  3,
		B:  4,
		Op: "NonValidOp",
	}
	out := &TestOut{}

	err = cli.Call("namedtesthandler", "testmethod", in, out)
	assert.Error(t, err)
	assert.Equal(t, 0, out.Res)
}
