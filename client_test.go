package nano

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_InvokeTimeout(t *testing.T) {
	s := &Server{}
	err := s.AddFunc("hello", func(name string) (string, error) {
		return "hello " + name, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	hServer := httptest.NewServer(s)
	defer hServer.Close()

	cln := &Client{URL: hServer.URL}
	var ans string
	err = cln.InvokeTimeout(3*time.Second, "hello", "reddec", &ans)
	if err != nil {
		t.Error("invoke:", err)
		return
	}
	if ans != "hello reddec" {
		t.Error(ans, " instead of ", "hello reddec")
		return
	}
}
