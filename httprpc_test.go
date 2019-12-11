package httprpc

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestContext_String(t *testing.T) {
	innerHTML := "Hello, client"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, innerHTML)
	}))
	defer ts.Close()
	ctx := context.TODO()
	body, err := Get(ctx, ts.URL).String()
	if err != nil {
		t.Error(err)
	}
	if body != innerHTML {
		t.Error("err resp body")
	}
}

func TestResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"hello":"world"}`)
	}))
	defer ts.Close()
	ctx := context.TODO()
	if _, err := Get(ctx, ts.URL).Do(); err != nil {
		t.Error(err)
	}
}

func TestContext_IntoJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"hello":"world"}`)
	}))
	defer ts.Close()

	ctx := context.TODO()
	result := make(map[string]string)
	err := JSON(ctx, ts.URL, nil).IntoJSON(&result)
	if err != nil {
		t.Error(err)
	}
	if result["hello"] != "world" {
		t.Error("err resp body")
	}
}

func TestContext_IntoXML(t *testing.T) {
	type Hello struct {
		Hello string
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		bs, _ := xml.Marshal(Hello{Hello: "world"})
		w.Write(bs)
		return
	}))
	defer ts.Close()

	ctx := context.TODO()
	result := new(Hello)
	err := XML(ctx, ts.URL, nil).IntoXML(&result)
	if err != nil {
		t.Error(err)
	}
	if result.Hello != "world" {
		t.Error("err resp body")
	}
}

func TestTimeoutContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"hello":"world"}`)
	}))
	defer ts.Close()

	var err error
	var start = time.Now()
	var c = make(chan int)
	go func() {
		ctx := context.TODO()
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		_, err = Get(ctx, ts.URL).Do()
		c <- 1
	}()
	<-c
	if !os.IsTimeout(err) {
		t.Errorf("fail err,got: %v", err)
	}
	if time.Now().Sub(start) > 200*time.Millisecond {
		t.Error("request timeout")
	}
}

func TestGet(t *testing.T) {
	ctx := context.TODO()
	_, err := Get(ctx, "h://hello").Do()
	if err == nil {
		t.Error("test illegal request,got nil")
	}
}

func TestContext_Do(t *testing.T) {
	var k string = "k"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(k, r.Header.Get(k)+"b")
		w.WriteHeader(http.StatusOK)
		return
	}))
	defer ts.Close()

	m := func(next HandlerFunc) HandlerFunc {
		return func(c *Context) (err error) {
			req := c.Request()
			req.Header.Set(k, "a")
			err = next(c)
			if err != nil {
				return
			}
			c.Response.Header.Set(k, c.Response.Header.Get(k)+"c")
			return nil
		}
	}

	hctx, err := Get(context.TODO(), ts.URL).Use(m).Do()
	if err != nil {
		t.Error(err)
	}
	if hctx.Header.Get(k) != "abc" {
		t.Errorf("want: %s,got: %s", "abc", hctx.Header.Get(k))
	}
}
