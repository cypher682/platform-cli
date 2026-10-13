package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cypher682/platform-cli/internal/config"
)

type Status string

const (
	Healthy  Status = "healthy"
	Degraded Status = "degraded"
	Down     Status = "down"
	Unknown  Status = "unknown"
	Skipped  Status = "skipped"
)

type Result struct {
	Service config.Service
	Status  Status
	Code    int
	Latency time.Duration
	Err     error
}

// Check probes one service.
func Check(ctx context.Context, svc config.Service, timeout time.Duration) Result {
	res := Result{Service: svc, Status: Down}

	hc := svc.HealthCheck
	if hc == nil && svc.Health != "" {
		hc = &config.HealthCheck{Type: "http", URL: svc.Health}
	}
	if hc == nil || hc.Type == "" || strings.ToLower(hc.Type) == "none" {
		res.Status = Skipped
		return res
	}

	t := strings.ToLower(hc.Type)
	switch t {
	case "http", "https":
		return checkHTTP(ctx, svc, hc, timeout, res)
	case "tcp":
		return checkTCP(ctx, svc, hc, timeout, res)
	default:
		res.Status = Unknown
		res.Err = fmt.Errorf("unsupported health check type %q", hc.Type)
		return res
	}
}

func checkHTTP(ctx context.Context, svc config.Service, hc *config.HealthCheck, timeout time.Duration, res Result) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	url := hc.URL
	if url == "" {
		url = svc.Health
	}
	if url == "" {
		res.Status = Unknown
		res.Err = fmt.Errorf("http health check missing URL")
		return res
	}
	method := hc.Method
	if method == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		res.Err = err
		res.Status = Down
		return res
	}
	for k, v := range hc.Headers {
		req.Header.Add(k, v)
	}
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	res.Latency = time.Since(start)
	if err != nil {
		res.Err = err
		res.Status = Down
		return res
	}
	defer resp.Body.Close()
	res.Code = resp.StatusCode
	exp := hc.ExpectStatus
	if exp == 0 {
		exp = 200
	}
	// simple: healthy if matches or 2xx when default
	if resp.StatusCode == exp || (exp == 200 && resp.StatusCode >= 200 && resp.StatusCode < 300) {
		res.Status = Healthy
		return res
	}
	if resp.StatusCode >= 500 {
		res.Status = Down
	} else {
		res.Status = Degraded
	}
	return res
}

func checkTCP(ctx context.Context, svc config.Service, hc *config.HealthCheck, timeout time.Duration, res Result) Result {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	addr := hc.Addr
	if addr == "" && hc.Port > 0 {
		addr = fmt.Sprintf("localhost:%d", hc.Port)
	}
	if addr == "" && svc.Port > 0 {
		addr = fmt.Sprintf("localhost:%d", svc.Port)
	}
	if addr == "" {
		res.Status = Unknown
		res.Err = fmt.Errorf("tcp health check missing address/port")
		return res
	}
	d := net.Dialer{Timeout: timeout}
	start := time.Now()
	c, err := d.DialContext(ctx, "tcp", addr)
	res.Latency = time.Since(start)
	if err != nil {
		res.Err = err
		res.Status = Down
		return res
	}
	_ = c.Close()
	res.Status = Healthy
	return res
}

// CheckAll fans out
func CheckAll(ctx context.Context, svcs []config.Service, timeout time.Duration) []Result {
	results := make([]Result, len(svcs))
	var wg sync.WaitGroup
	for i, s := range svcs {
		wg.Add(1)
		go func(i int, s config.Service) {
			defer wg.Done()
			results[i] = Check(ctx, s, timeout)
		}(i, s)
	}
	wg.Wait()
	return results
}

// WaitHealthy polls
func WaitHealthy(ctx context.Context, svcs []config.Service, perCheck time.Duration) ([]Result, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		rs := CheckAll(ctx, svcs, perCheck)
		all := true
		for _, r := range rs {
			if r.Status != Healthy && r.Status != Skipped {
				all = false
			}
		}
		if all {
			return rs, nil
		}
		select {
		case <-ctx.Done():
			return rs, ctx.Err()
		case <-ticker.C:
		}
	}
}
