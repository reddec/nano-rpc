package nano

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
)

func TestServer_ServeHTTP(t *testing.T) {
	list, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err, "make listener")
	}
	defer list.Close()

	s := &Server{}

	hServer := httptest.NewServer(s)
	defer hServer.Close()

	err = s.AddFunc("helloP", func(name *string) (string, error) {
		return "hello " + *name, nil
	})
	if err != nil {
		t.Error(err)
		return
	}
	err = s.AddFunc("hello", func(name string) (string, error) {
		return "hello " + name, nil
	})
	if err != nil {
		t.Error(err)
		return
	}

	cl := hServer.Client()
	res, err := cl.Post(hServer.URL+"/hello", "application/json", bytes.NewBufferString(`"reddec"`))
	if err != nil {
		t.Error(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		text, _ := ioutil.ReadAll(res.Body)
		t.Error("code: ", res.StatusCode, res.Status, string(text))
		return
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}
	if string(data) != `"hello reddec"` {
		t.Error("answer:", string(data))
		return
	}

	// pointer
	res, err = cl.Post(hServer.URL+"/helloP", "application/json", bytes.NewBufferString(`"reddec"`))
	if err != nil {
		t.Error(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		text, _ := ioutil.ReadAll(res.Body)
		t.Error("code: ", res.StatusCode, res.Status, string(text))
		return
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}
	if string(data) != `"hello reddec"` {
		t.Error("answer:", string(data))
		return
	}
}

type st struct{}

func (s *st) Hello(string) (string, error)   { return "", nil }
func (s *st) HelloP(*string) (string, error) { return "", nil }
func (s *st) spam(string) (string, error)    { return "", nil }

func TestServer_Add(t *testing.T) {
	s := &Server{}
	err := s.Add(&st{})
	if err != nil {
		t.Fatal(err)
	}

	names := s.Names()
	if len(names) != 2 {
		t.Fatal("len(names) = ", len(names))
	}
	sort.Strings(names)
	var rep = []string{"Hello", "HelloP"}
	for i, name := range names {
		if rep[i] != name {
			t.Error(i, "=>", name, " instead of ", rep[i])
		}
	}
}
