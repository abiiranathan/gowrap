// Simple wrapper around http.Server with graceful shutdown and easy tls setup.
package server

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type server struct {
	*http.Server
	tlsEnabled bool
}

// functional option to configure the server
type option func(*server)

// Set server address to addr
func WithAddr(addr string) option {
	return func(s *server) {
		s.Addr = addr
	}
}

// set server handler to h
func WithHandler(h http.Handler) option {
	return func(s *server) {
		s.Handler = h
	}
}

// set server ReadTimeout to timeout
func WithReadTimeout(timeout time.Duration) option {
	return func(s *server) {
		s.ReadTimeout = timeout
	}
}

// set server WriteTimeout to timeout
func WithWriteTimeout(timeout time.Duration) option {
	return func(s *server) {
		s.WriteTimeout = timeout
	}
}

// Start the server with tls enabled by certificates
func WithTLS(certFile, keyFile string) option {
	certs, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("could not load TLS certificate: %v\n", err)
	}

	return func(s *server) {
		config := &tls.Config{
			MinVersion:       tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			Certificates:     []tls.Certificate{certs},
			// Set NextProtos to make sure the http2 protocol is used.
			// This is necessary for the websocket transport to work.
			NextProtos: []string{"http/1.1", "h2"},
		}

		s.TLSConfig = config
		s.tlsEnabled = true
	}
}

// Create a new http server instance.
// Run the server by calling the Run method on it.
func NewServer(addr string, options ...option) *server {
	srv := &server{
		Server: &http.Server{
			Addr:           addr,
			MaxHeaderBytes: 1 << 20,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   60 * time.Second,
		},
	}

	for _, opt := range options {
		opt(srv)
	}

	return srv
}

// Setup tls on the server before running it
func (s *server) WithTLS(certfile, keyfile string) {
	opt := WithTLS(certfile, keyfile)
	opt(s)
}

func (s *server) Run() {
	timeout := 10 * time.Second // 10 seconds to shutdown gracefully
	done := make(chan error, 1) // channel to signal shutdown

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		done <- s.Shutdown(ctx)
	}()

	var err error

	if s.tlsEnabled {
		log.Printf("Listening on https %s", s.Addr)
		// TLS certs already populated in TLSConfig
		err = s.ListenAndServeTLS("", "")
	} else {
		log.Printf("Listening on http %s", s.Addr)
		err = s.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server stopped with error: %v\n", err)
	}

	log.Println("Server stopped gracefully")
}
