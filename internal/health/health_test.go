package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cypher682/platform-cli/internal/config"
)

func TestCheckStatuses(t *testing.T) {
	cases := map[int]Status{200: Healthy, 404: Degraded, 503: Down}
	for code, want := range cases {
		code := code
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(code)
		}))
		r := Check(context.Background(), config.Service{Name: "x", Health: srv.URL}, time.Second)
		srv.Close()
		if r.Status != want {
			t.Errorf("code %d: want %s got %s", code, want, r.Status)
		}
	}
}

func TestCheckTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer srv.Close()
	r := Check(context.Background(), config.Service{Health: srv.URL}, 50*time.Millisecond)
	if r.Status != Down || r.Err == nil {
		t.Fatalf("expected Down with error, got %+v", r)
	}
}
