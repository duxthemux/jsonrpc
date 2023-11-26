package jsonrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type Namer interface {
	Name() string
}

type Server struct {
	endpoints map[string]any
	router    *http.ServeMux
	routes    []string
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.router.ServeHTTP(writer, request)
}

// ----------------------------------------------------------------------------

type reqKey struct {
}

type resKey struct {
}

var (
	ctxReqKey = reqKey{}
	ctxResKey = resKey{}
)

func HttpRequest(ctx context.Context) *http.Request {
	return ctx.Value(ctxReqKey).(*http.Request)
}
func HttpResponseWriter(ctx context.Context) http.ResponseWriter {
	return ctx.Value(ctxResKey).(http.ResponseWriter)
}

// ----------------------------------------------------------------------------

func handleCall(handler any, method reflect.Method) http.HandlerFunc {
	return func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		ctx = context.WithValue(ctx, ctxReqKey, req)
		ctx = context.WithValue(ctx, ctxResKey, writer)

		inParamType := method.Type.In(2)

		outParamType := method.Type.In(3)

		inParam := reflect.New(inParamType.Elem())

		inParamIface := inParam.Interface()

		if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodPatch {
			if err := json.NewDecoder(req.Body).Decode(inParamIface); err != nil {
				http.Error(writer, err.Error(), http.StatusBadRequest)
				return
			}
		}

		inParam = reflect.ValueOf(inParamIface)

		outParam := reflect.New(outParamType.Elem())

		ctxValue := reflect.ValueOf(ctx)

		outValues := method.Func.Call([]reflect.Value{reflect.ValueOf(handler), ctxValue, inParam, outParam})

		if len(outValues) > 0 {
			if !outValues[0].IsNil() {
				outValueZero := outValues[0].Interface()
				switch x := outValueZero.(type) {
				case error:
					http.Error(writer, x.Error(), http.StatusInternalServerError)
					return
				default:
					http.Error(writer, fmt.Sprintf("%v", x), http.StatusInternalServerError)
					return
				}
			}
		}
		writer.Header().Add("content-type", "application/json")
		if err := json.NewEncoder(writer).Encode(outParam.Interface()); err != nil {
			http.Error(writer, "error encoding response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) registerAHandler(h any) {
	typeOfH := reflect.TypeOf(h)

	for i := 0; i < typeOfH.NumMethod(); i++ {
		method := typeOfH.Method(i)

		if !method.IsExported() {
			continue
		}

		if method.Type.NumIn() != 4 {
			continue
		}

		if method.Type.NumOut() != 1 {
			continue
		}

		name := strings.ToLower(typeOfH.Elem().Name())
		if typeOfH, ok := h.(Namer); ok {
			name = strings.ToLower(typeOfH.Name())
		}

		route := fmt.Sprintf("/%s/%s", name, strings.ToLower(method.Name))
		s.routes = append(s.routes, route)

		s.router.HandleFunc(route, handleCall(h, method))
	}

}

func (s *Server) Register(handlers ...any) error {
	for _, handler := range handlers {
		rname := reflect.TypeOf(handler).Elem().Name()
		_, alreadyExists := s.endpoints[rname]
		if alreadyExists {
			return fmt.Errorf("handler called %s already registered", rname)
		}
		s.registerAHandler(handler)
	}

	return nil
}

func (s *Server) RegisterAs(handler any, hName string) error {
	_, alreadyExists := s.endpoints[hName]
	if alreadyExists {
		return fmt.Errorf("handler called %s already registered", hName)
	}
	s.endpoints[hName] = handler

	return nil
}

func (s *Server) Routes() []string {
	return s.routes
}

func New() *Server {
	ret := &Server{
		endpoints: map[string]any{},
		router:    http.NewServeMux()}
	return ret
}

type Client struct {
	BaseUrl    string
	Headers    http.Header
	Proto      string
	HttpClient *http.Client
}

func NewClient(baseUrl string) *Client {
	ret := &Client{
		BaseUrl:    baseUrl,
		Headers:    http.Header{},
		HttpClient: &http.Client{},
	}
	ret.Proto = "http"
	ret.Headers.Add("content-type", "application/json")

	return ret
}

func (c *Client) Call(handler string, method string, in any, out any) error {

	finalUrl := strings.ReplaceAll(strings.ReplaceAll(c.BaseUrl, "${HANDLER}", handler), "${METHOD}", method)

	callUrl, err := url.Parse(finalUrl)

	pi, po := io.Pipe()

	go func() {
		json.NewEncoder(po).Encode(in)
		po.Close()
	}()

	req := http.Request{
		Method: http.MethodPost,
		URL:    callUrl,
		Proto:  c.Proto,
		Header: c.Headers,
		Body:   pi,
	}
	res, err := c.HttpClient.Do(&req)
	if err != nil {
		return err
	}
	if res.StatusCode > 399 {
		return fmt.Errorf("http error %v: %s", res.StatusCode, res.Status)
	}

	if err = json.NewDecoder(res.Body).Decode(out); err != nil {
		return err
	}

	return nil
}
