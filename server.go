package nano

import (
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	lock    sync.RWMutex
	methods map[string]*method
}

// Names of all registered methods
func (s *Server) Names() []string {
	var ans = make([]string, 0, len(s.methods))
	s.lock.RLock()
	defer s.lock.RUnlock()
	for name := range s.methods {
		ans = append(ans, name)
	}
	return ans
}

// Handles HTTP requests
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("only POST allowed"))
		return
	}
	methodName := path.Base(r.URL.Path)
	s.lock.RLock()
	method := s.methods[methodName]
	s.lock.RUnlock()
	if method == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("method " + methodName + " not found"))
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("interrupted"))
		return
	}
	res, err := method.call(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(res)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

// Add single function to handler. Functions should accept one pointer parameters and returns tuple of (value, error)
func (s *Server) AddFunc(name string, fn interface{}) error {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return errors.New("is not a function")
	}
	if !isMethodApplicable(v.Type()) {
		return errors.New("not suitable function")
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.methods == nil {
		s.methods = make(map[string]*method)
	}
	var arg = v.Type().In(0)
	var argIsRef bool
	if argIsRef = v.Type().In(0).Kind() == reflect.Ptr; argIsRef {
		arg = arg.Elem()
	}
	s.methods[name] = &method{
		Method: v,
		Arg:    arg,
		ArgRef: argIsRef,
	}
	return nil
}

// Add all exported and suitable functions found in the object to the server
func (s *Server) Add(iface interface{}) error {
	v := reflect.ValueOf(iface)
	if v.Kind() != reflect.Ptr {
		return errors.New("is not a pointer")
	}
	n := v.NumMethod()
	for i := 0; i < n; i++ {
		m := v.Method(i)
		name := v.Type().Method(i).Name
		if !ast.IsExported(name) {
			continue
		}
		_ = s.AddFunc(name, m.Interface())
	}
	return nil
}

type method struct {
	Arg    reflect.Type
	Method reflect.Value
	ArgRef bool
}

func (m *method) call(data []byte) ([]byte, error) {
	inV := reflect.New(m.Arg)
	var in = inV.Interface()
	err := json.Unmarshal(data, in)
	if err != nil {
		return nil, err
	}
	if !m.ArgRef {
		inV = inV.Elem()
	}
	res := m.Method.Call([]reflect.Value{inV})
	ans := res[0]
	if !res[1].IsNil() {
		err = res[1].Interface().(error)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(ans.Interface())
}

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// check method. Method name should be exportable, has 1 in argument as a pointer and returns values with error
func isMethodApplicable(m reflect.Type) bool {
	//if ! ast.IsExported(m.Name) {
	//	return false
	//}

	if m.NumIn() != 1 {
		return false
	}
	if m.NumOut() != 2 {
		return false
	}
	if !m.Out(1).Implements(errorInterface) {
		return false
	}
	return true
}

func ListenAndServe(ctx context.Context, binding string, handler http.Handler) error {
	done := make(chan struct{})
	srv := &http.Server{
		Addr:    binding,
		Handler: handler,
	}

	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
		case <-done:
		}
		child, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = srv.Shutdown(child)
	}()

	return srv.ListenAndServe()
}

var defaultServer = &Server{}

func AddFunc(name string, fn interface{}) error {
	return defaultServer.AddFunc(name, fn)
}

func MustAddFunc(name string, fn interface{}) {
	err := AddFunc(name, fn)
	if err != nil {
		panic(err)
	}
}

func AddObject(st interface{}) error {
	return defaultServer.Add(st)
}

func MustAddObject(st interface{}) {
	err := AddObject(st)
	if err != nil {
		panic(err)
	}
}

func Run(ctx context.Context, binding string) error {
	return ListenAndServe(ctx, binding, defaultServer)
}
