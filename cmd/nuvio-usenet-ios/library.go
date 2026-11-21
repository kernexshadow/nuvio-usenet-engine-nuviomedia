// Command nuvio-usenet-ios builds the engine as a static library for iOS, where an
// app cannot start a child process. The app links it and runs the same loopback
// HTTP server as the Android sidecar inside its own process: start takes the
// bootstrap record the sidecar reads from stdin and returns the readiness record
// the sidecar writes to stdout.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/NuvioMedia/nuvio-usenet-engine/internal/engine"
)

type bootstrap struct {
	Token        string   `json:"token"`
	Certificates []string `json:"certificates"`
	MemoryMiB    int64    `json:"memoryMiB"`
	NZBCacheDir  string   `json:"nzbCacheDir"`
}

type instance struct {
	cancel context.CancelFunc
	done   chan struct{}
}

var (
	mu      sync.Mutex
	current *instance
)

// start stops the running server, if any, and starts a new one on a free loopback
// port. Restarting is how the app recovers after iOS reclaimed the listening
// socket of a suspended app.
func start(record []byte) ([]byte, error) {
	if len(record) > 4<<20 {
		return nil, errors.New("bootstrap record too large")
	}
	var init bootstrap
	if err := json.Unmarshal(record, &init); err != nil {
		return nil, err
	}
	if len(init.Token) < 32 {
		return nil, errors.New("bootstrap token too short")
	}
	mu.Lock()
	defer mu.Unlock()
	stopLocked()
	limit := init.MemoryMiB
	if limit < 64 {
		limit = 128
	}
	if limit > 512 {
		limit = 512
	}
	debug.SetMemoryLimit(limit << 20)
	roots, _ := x509.SystemCertPool()
	if roots == nil {
		roots = x509.NewCertPool()
	}
	for _, pem := range init.Certificates {
		roots.AppendCertsFromPEM([]byte(pem))
	}
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}, ResponseHeaderTimeout: 30 * time.Second, IdleConnTimeout: 60 * time.Second}
	client := &http.Client{Transport: transport, Timeout: 2 * time.Minute}
	app := engine.NewServer(ctx, init.Token, roots, client)
	app.SetNZBCacheDir(init.NZBCacheDir)
	srv := &http.Server{Handler: app, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10, BaseContext: func(net.Listener) context.Context { return ctx }}
	done := make(chan struct{})
	go func() { <-ctx.Done(); srv.Close() }()
	go func() {
		defer close(done)
		srv.Serve(ln)
		cancel()
		app.Close()
	}()
	current = &instance{cancel: cancel, done: done}
	return json.Marshal(map[string]any{"protocol": 1, "port": ln.Addr().(*net.TCPAddr).Port})
}

// stop closes the server and its sessions and returns once they are closed.
func stop() {
	mu.Lock()
	defer mu.Unlock()
	stopLocked()
}

func stopLocked() {
	if current == nil {
		return
	}
	current.cancel()
	<-current.done
	current = nil
}
