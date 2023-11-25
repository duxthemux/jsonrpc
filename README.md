# jsonrpc
Simple Json RPC Impl

The idea is to be simple and keep it simple.

The jsonRpc server allows registering handlers and convert them into http routes, while taking care of marshalling
and unmarshalling.

handlers are **POINTERS** to data structures. The **PUBLIC** methods in these data structures that have the following
signature:

```go
func(h *Handler) SomeMethod(ctx context.Context, in *SomeDataIn, out *SomeDataOut) error{}
```
Will be converted into routes. 

Please note: 

1. ctx is the http.Request context received by the http endpoint. 
2. in and out are data structures created by you, and will me marshalled/unmarshalled automatically. If anything fails 
just return and error and get it propagated to the caller. 
3. If in your RPC call you need access to request and response objects, they can be retrieved by using:
```go
jsonrpc.HttpRequest(ctx context.Context) *http.Request

jsonrpc.HttpResponseWriter(ctx context.Context) http.ResponseWriter
```

Please refer to [example](./example) for further details.

The package has also a client wrapper - in the test file you can see it in action, but basically does this:

```go
    cli := jsonrpc.NewClient(srv.URL + "/${HANDLER}/${METHOD}")
	in := &TestIn{
		A:  3,
		B:  4,
		Op: "NonValidOp",
	}
	out := &TestOut{}

	err = cli.Call("namedtesthandler", "testmethod", in, out)
```

When the client is created you point it to some server, and from there you can call multiple endpoints.