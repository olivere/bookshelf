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
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/olivere/env"
	"github.com/olivere/httputil"
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
		port = flag.String("port", env.String("8080", "PORT"), "HTTP port number")
	)
	flag.Parse()

	lis, err := net.Listen("tcp", net.JoinHostPort("0.0.0.0", *port))
	if err != nil {
		return err
	}
	defer lis.Close()

	r := mux.NewRouter()
	r.Path("/").Handler(rootHandler())

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

func rootHandler() http.Handler {
	hostname, _ := os.Hostname()

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
		}{
			Service:  "backend",
			Time:     time.Now(),
			Hostname: hostname,
		}
		data.Runtime.Version = runtime.Version()
		info, ok := debug.ReadBuildInfo()
		if ok {
			data.Build.Settings = make(map[string]interface{})
			for _, setting := range info.Settings {
				data.Build.Settings[setting.Key] = setting.Value
			}
		}

		httputil.WriteJSON(w, data)
	})
}
