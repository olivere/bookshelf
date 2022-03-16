package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
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

	type service struct {
		Status   int         `json:"status,omitempty"`
		Took     string      `json:"took,omitempty"`
		Response interface{} `json:"response,omitempty"`
		Error    string      `json:"error,omitempty"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Service  string    `json:"service"`
			Time     time.Time `json:"time"`
			Hostname string    `json:"hostname"`
			Runtime  struct {
				Version string `json:"version"`
			} `json:"runtime,omitempty"`
			Build struct {
				Settings map[string]interface{} `json:"settings,omitempty"`
			} `json:"build,omitempty"`
			Services map[string]service `json:"services,omitempty"`
		}{
			Service:  "frontend",
			Time:     time.Now(),
			Hostname: hostname,
			Services: make(map[string]service),
		}
		data.Runtime.Version = runtime.Version()
		info, ok := debug.ReadBuildInfo()
		if ok {
			data.Build.Settings = make(map[string]interface{})
			for _, setting := range info.Settings {
				data.Build.Settings[setting.Key] = setting.Value
			}
		}

		// Check the services
		if svcs := httputil.QueryStringArray(r, "services", nil); len(svcs) > 0 {
			g, ctx := errgroup.WithContext(r.Context())

			for _, svc := range svcs {
				// TODO Make this independent of the backendURL
				var url string
				if svc == "backend" {
					url = backendURL
				}
				if url != "" {
					g.Go(func() error {
						start := time.Now()
						req, _ := http.NewRequestWithContext(ctx, "GET", backendURL, http.NoBody)
						resp, err := httpClient.Do(req)
						if err != nil {
							data.Services[svc] = service{
								Error:  err.Error(),
								Status: http.StatusServiceUnavailable,
								Took:   time.Since(start).String(),
							}
						} else {
							defer resp.Body.Close()
							var body map[string]interface{}
							_ = json.NewDecoder(resp.Body).Decode(&body)
							data.Services[svc] = service{
								Status:   resp.StatusCode,
								Took:     time.Since(start).String(),
								Response: body,
							}
						}
						return nil
					})
				}
			}

			if len(svcs) > 0 {
				if err := g.Wait(); err != nil {
					log.Printf("unable to check service: %v", err)
				}
			}
		}

		httputil.WriteJSON(w, data)
	})
}
