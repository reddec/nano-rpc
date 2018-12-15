# Nano HTTP RPC lib

[![](https://godoc.org/github.com/reddec/nano-rpc?status.svg)](http://godoc.org/github.com/reddec/nano-rpc)


`import "github.com/reddec/nano-rpc"`

No dependency, super-tiny library that lets you expose any suitable function to HTTP POST interface.

Function should have one input argument and returns tuple of value and error.

Ex:

**good**

* `func hello(name string) (string, error)`
* `func bye(data MyStruct) (bool, error)`
* `func sample(a *int) (*MyStruct, error)`

**bad**

* `func hello(name, lastname string) (string, error)` - two arguments
* `func sample() (string, error)` - no arguments
* `func close(data string) error` - only one return
* `func shutdown(delay int) (int, int)` - last argument is not an error 

Methods are expose as POST methods using method name as last part of path.

Ex:

Method `greetings` will be exposed as `POST /greetings`


## Server usage

```go
package main

import (
	"github.com/reddec/nano-rpc"
	"net/http"
)

func main() {
	server := &nano.Server{}
	server.AddFunc("greetings", func(name string) (string, error) {
    		return "hello " + name, nil
    })
	http.ListenAndServe(":8080", server)
}
```

Instead of `AddFunc` you can use `Add` that will scan all methods in object and expose suitable ones.


## Client usage

```go
package main

import (
	"fmt"
	"github.com/reddec/nano-rpc"
)

func main() {
	client := &nano.Client{
		URL: "http://127.0.0.1:8080",
	}
	var ans string
	err := client.Invoke("greetings", "Foo", &ans)
	if err!=nil{
		panic(err)
	}
	fmt.Println(ans)
}
```