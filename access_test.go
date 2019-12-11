package httprpc

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessLog(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"hello":"world"}`)
	}))
	defer ts.Close()
	ctx := context.TODO()
	if _, err := Get(ctx, ts.URL).Use(AccessLog).Do(); err != nil {
		t.Error(err)
	}
}
