package nano

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Client of HTTP POST rpc
type Client struct {
	URL string // Base URL without / (slash) at the end
}

// Invoke remote method with context. Param should be JSON-encodable and result should be JSON-decodable
func (c *Client) InvokeContext(ctx context.Context, method string, param interface{}, result interface{}) error {
	data, err := json.Marshal(param)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.URL+"/"+method, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(data)))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		text, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("%v: code %v: status %v: %v", method, res.StatusCode, res.Status, string(text))
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

// Invoke with timeout context
func (c *Client) InvokeTimeout(timeout time.Duration, method string, param interface{}, result interface{}) error {
	ctx, cls := context.WithTimeout(context.Background(), timeout)
	defer cls()
	return c.InvokeContext(ctx, method, param, result)
}

// Invoke remote method with default context
func (c *Client) Invoke(method string, param interface{}, result interface{}) error {
	return c.InvokeTimeout(30*time.Second, method, param, result)
}
