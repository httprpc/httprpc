package httprpc

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

type (
	// Context httprpc context
	Context struct {
		client *http.Client
		req    *http.Request
		*http.Response
		err        error
		readOnce   *sync.Once
		doOnce     *sync.Once
		bytes      []byte
		middleware []MiddlewareFunc
	}
	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc
	// HandlerFunc defines a function to handle HTTP requests.
	HandlerFunc func(*Context) error
)

// NewContext returns a Context instance.
func NewContext() (c *Context) {
	c = new(Context)
	c.readOnce = new(sync.Once)
	c.doOnce = new(sync.Once)
	c.middleware = make([]MiddlewareFunc, 0)
	return
}

// WithClient replace default http client
func (c *Context) WithClient(client *http.Client) {
	c.client = client
}

// CleanMiddleware remove before used middlewares
func (c *Context) CleanMiddleware() *Context {
	c.middleware = make([]MiddlewareFunc, 0)
	return c
}

// Use adds middleware to the chain
func (c *Context) Use(middleware ...MiddlewareFunc) *Context {
	c.middleware = append(c.middleware, middleware...)
	return c
}

// Do exec chain
func (c *Context) Do() (_ *Context, err error) {
	if c.err != nil {
		err = c.err
		return
	}

	if c.req == nil {
		err = errors.New("nil request")
		return c, err
	}

	if c.client == nil {
		c.client = http.DefaultClient
	}

	cb := applyMiddleware(func(c *Context) (doErr error) {
		c.Response, doErr = c.client.Do(c.req)
		return
	}, c.middleware...)

	c.doOnce.Do(func() {
		err = cb(c)
	})

	return c, err
}

// Request return Context's HTTP request
func (c *Context) Request() *http.Request {
	return c.req
}

//String Returns http.Response.Body as string
func (c *Context) String() (str string, err error) {
	bs, err := c.Bytes()
	return string(bs), err
}

// Bytes Returns http.Response.Body as bytes
func (c *Context) Bytes() (bs []byte, err error) {
	c.readOnce.Do(func() {
		_, err = c.Do()
		if err != nil {
			return
		}
		c.bytes, err = ioutil.ReadAll(c.Body)
		if err != nil {
			return
		}
		defer c.Body.Close()
	})
	return c.bytes, err
}

// IntoJSON json.Unmarshal HTTP response body and filling the data of the incoming structure
func (c *Context) IntoJSON(v interface{}) (err error) {
	bs, err := c.Bytes()
	if err != nil {
		return
	}
	err = json.Unmarshal(bs, v)
	return err
}

// IntoXML xml.Unmarshal HTTP response body and filling the data of the incoming structure
func (c *Context) IntoXML(v interface{}) (err error) {
	bs, err := c.Bytes()
	if err != nil {
		return
	}
	err = xml.Unmarshal(bs, v)
	return err
}

// Get make a *Context with GET method
func Get(ctx context.Context, url string) (c *Context) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c = NewContext()
		c.err = err
		return
	}
	return Request(ctx, req)
}

//Post make a *Context with POST method
func Post(ctx context.Context, url, contentType string, body io.Reader) (c *Context) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c = NewContext()
		c.err = err
		return
	}
	req.Header.Set("Content-Type", contentType)
	return Request(ctx, req)
}

// JSON make a *Context
func JSON(ctx context.Context, url string, i interface{}) (c *Context) {
	bs, err := json.Marshal(i)
	if err != nil {
		c = NewContext()
		c.err = err
		return
	}
	return Post(ctx, url, "application/json", bytes.NewReader(bs))
}

// XML make a *Context
func XML(ctx context.Context, url string, i interface{}) (c *Context) {
	bs, err := xml.Marshal(i)
	if err != nil {
		c = NewContext()
		c.err = err
		return
	}
	return Post(ctx, url, "application/xml", bytes.NewReader(bs))
}

// Request make a *Context Use HTTP request
func Request(ctx context.Context, req *http.Request) (c *Context) {
	c = NewContext()
	c.req = req.WithContext(ctx)
	return c
}

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
