package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/olivere/env"
	"github.com/olivere/httputil"
	"golang.org/x/sync/errgroup"
)

var (
	// httpClient to use for HTTP requests to other services
	httpClient *http.Client = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}
)

func main() {
	if err := runMain(); err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	var (
		port       = flag.String("port", env.String("8080", "PORT"), "HTTP port number")
		backendURL = flag.String("backend-url", env.String("http://backend.local", "BACKEND_URL"), "HTTP Backend address")
	)
	flag.Parse()

	lis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", *port))
	if err != nil {
		return err
	}
	defer lis.Close()

	r := mux.NewRouter()
	r.Path("/").Handler(rootHandler(*backendURL))

	httpSrv := &http.Server{
		Handler:      r,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 300,
	}
	idleConnsClosed := make(chan struct{})

	errCh := make(chan error, 1)

	go func() {
		errCh <- httpSrv.Serve(lis)
	}()

	go func() {
		defer close(idleConnsClosed)
		for {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			switch sig := <-c; sig {
			case syscall.SIGINT, syscall.SIGTERM:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				errCh <- httpSrv.Shutdown(ctx)
				return
			}
		}
	}()

	switch err := <-errCh; err {
	case http.ErrServerClosed:
		err = nil
	case nil:
	default:
		log.Printf("error %v", err)
	}

	return nil
}

func rootHandler(backendURL string) http.Handler {
	hostname, _ := os.Hostname()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Service  string    `json:"service"`
			Time     time.Time `json:"time"`
			Hostname string    `json:"hostname"`
			Backend  struct {
				Status int    `json:"status,omitempty"`
				Took   string `json:"took,omitempty"`
				Error  string `json:"error,omitempty"`
			} `json:"backend,omitempty"`
		}{
			Service:  "frontend",
			Time:     time.Now(),
			Hostname: hostname,
		}

		if httputil.QueryBool(r, "check", false) {
			g, ctx := errgroup.WithContext(r.Context())

			g.Go(func() error {
				start := time.Now()
				req, _ := http.NewRequestWithContext(ctx, "GET", backendURL, http.NoBody)
				resp, err := httpClient.Do(req)
				if err != nil {
					data.Backend.Error = err.Error()
					data.Backend.Status = http.StatusServiceUnavailable
				} else {
					data.Backend.Status = resp.StatusCode
				}
				data.Backend.Took = time.Since(start).String()
				return nil
			})

			if err := g.Wait(); err != nil {
				log.Printf("unable to check service: %v", err)
			}
		}

		httputil.WriteJSON(w, data)
	})
}
