package arm

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"go.viam.com/rdk/logging"
	"go.viam.com/test"
)

func TestStartProxyDisabled(t *testing.T) {
	ctx := context.Background()
	x := &xArm{
		conf:   &Config{Host: "127.0.0.1"},
		logger: logging.NewTestLogger(t),
	}
	err := x.startProxy(ctx)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, x.proxyServer, test.ShouldBeNil)
}

func TestStartProxyEnabled(t *testing.T) {
	ctx := context.Background()

	// Pick a free port for the proxy to listen on.
	proxyPort := getFreePort(t)

	x := &xArm{
		conf: &Config{
			Host:            "127.0.0.1",
			StudioProxy:     true,
			StudioProxyPort: proxyPort,
		},
		logger: logging.NewTestLogger(t),
	}

	err := x.startProxy(ctx)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, x.proxyServer, test.ShouldNotBeNil)
	defer x.stopProxy()

	// Verify the proxy is listening by making a request.
	// It will proxy to Host:18333 which has no backend, so expect 502.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/", proxyPort), nil)
	test.That(t, err, test.ShouldBeNil)
	resp, err := http.DefaultClient.Do(req)
	test.That(t, err, test.ShouldBeNil)
	resp.Body.Close()
	test.That(t, resp.StatusCode, test.ShouldEqual, http.StatusBadGateway)
}

func TestStartProxyPortConflict(t *testing.T) {
	ctx := context.Background()

	// Occupy a port on all interfaces (matching startProxy's ":port" bind) and keep it open.
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", ":0")
	test.That(t, err, test.ShouldBeNil)
	port := ln.Addr().(*net.TCPAddr).Port
	defer ln.Close()

	x := &xArm{
		conf: &Config{
			Host:            "127.0.0.1",
			StudioProxy:     true,
			StudioProxyPort: port,
		},
		logger: logging.NewTestLogger(t),
	}

	err = x.startProxy(ctx)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, err.Error(), test.ShouldContainSubstring, "failed to listen")
}

func TestStopProxyIdempotent(t *testing.T) {
	x := &xArm{
		conf:   &Config{Host: "127.0.0.1"},
		logger: logging.NewTestLogger(t),
	}
	x.stopProxy()
	x.stopProxy()
}

func TestProxyForwardsRequests(t *testing.T) {
	ctx := context.Background()

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "proxied")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "path=%s", r.URL.Path)
	}))
	defer backend.Close()

	backendHost, backendPort, err := net.SplitHostPort(backend.Listener.Addr().String())
	test.That(t, err, test.ShouldBeNil)

	var lc net.ListenConfig
	proxyLn, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	test.That(t, err, test.ShouldBeNil)
	proxyPort := proxyLn.Addr().(*net.TCPAddr).Port

	x := &xArm{
		conf: &Config{
			Host:            backendHost,
			StudioProxy:     true,
			StudioProxyPort: proxyPort,
		},
		logger: logging.NewTestLogger(t),
	}

	x.proxyServer = buildTestProxy(t, backendHost, backendPort, proxyLn)
	defer x.stopProxy()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/some/path", proxyPort), nil)
	test.That(t, err, test.ShouldBeNil)
	resp, err := http.DefaultClient.Do(req)
	test.That(t, err, test.ShouldBeNil)
	defer resp.Body.Close()

	test.That(t, resp.StatusCode, test.ShouldEqual, http.StatusOK)
	test.That(t, resp.Header.Get("X-Test"), test.ShouldEqual, "proxied")

	body, err := io.ReadAll(resp.Body)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, string(body), test.ShouldEqual, "path=/some/path")
}

func buildTestProxy(t *testing.T, host, port string, ln net.Listener) *http.Server {
	t.Helper()
	proxy := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
		Handler: &httputil.ReverseProxy{
			Rewrite: func(r *httputil.ProxyRequest) {
				r.SetURL(&url.URL{Scheme: "http", Host: net.JoinHostPort(host, port)})
				r.Out.Host = net.JoinHostPort(host, port)
			},
		},
	}
	go func() {
		if err := proxy.Serve(ln); err != nil && err != http.ErrServerClosed {
			t.Errorf("test proxy error: %v", err)
		}
	}()
	return proxy
}

func getFreePort(t *testing.T) int {
	t.Helper()
	ctx := context.Background()
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
	test.That(t, err, test.ShouldBeNil)
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}
