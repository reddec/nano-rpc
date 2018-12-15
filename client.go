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
	URL        string        // Base URL without / (slash) at the end
	RetryNum   int           // Network retry numbers of attempts (means errors of encode/decode errors and non-200 code are not counted). Negative value means unlimited
	RetryDelay time.Duration // Delay between attempts
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

	num := 0
	for {
		success, err := c.invoke(req, result)
		if success {
			return err
		}
		if c.RetryNum >= 0 && num >= c.RetryNum {
			return err
		}
		num++
		select {
		case <-time.After(c.RetryDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
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

func (c *Client) invoke(req *http.Request, result interface{}) (success bool, err error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		text, _ := ioutil.ReadAll(res.Body)
		return true, fmt.Errorf("code %v: status %v: %v", res.StatusCode, res.Status, string(text))
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(data, result)
}
